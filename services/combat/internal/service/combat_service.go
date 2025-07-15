package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"combat/internal/config"
	"combat/internal/models"
	"combat/internal/repository"
	"combat/internal/external"
)

// CombatService gère la logique métier du combat (version simplifiée)
type CombatService struct {
	config        *config.Config
	combatRepo    repository.CombatRepositoryInterface
	spellRepo     repository.SpellRepositoryInterface
	effectRepo    repository.EffectRepositoryInterface
	combatLogRepo repository.CombatLogRepositoryInterface
	playerClient  external.PlayerClientInterface
	worldClient   external.WorldClientInterface
	damageCalc    DamageCalculatorInterface
	effectService EffectServiceInterface
}

// NewCombatService crée une nouvelle instance du service de combat
func NewCombatService(
	config *config.Config,
	combatRepo repository.CombatRepositoryInterface,
	spellRepo repository.SpellRepositoryInterface,
	effectRepo repository.EffectRepositoryInterface,
	combatLogRepo repository.CombatLogRepositoryInterface,
	playerClient external.PlayerClientInterface,
	worldClient external.WorldClientInterface,
	damageCalc DamageCalculatorInterface,
	effectService EffectServiceInterface,
) CombatServiceInterface {
	return &CombatService{
		config:        config,
		combatRepo:    combatRepo,
		spellRepo:     spellRepo,
		effectRepo:    effectRepo,
		combatLogRepo: combatLogRepo,
		playerClient:  playerClient,
		worldClient:   worldClient,
		damageCalc:    damageCalc,
		effectService: effectService,
	}
}

// CreateCombatSession crée une nouvelle session de combat
func (s *CombatService) CreateCombatSession(creatorID uuid.UUID, req models.StartCombatRequest) (*models.CombatSession, error) {
	// Valider la requête
	if err := s.validateStartCombatRequest(req); err != nil {
		return nil, fmt.Errorf("invalid combat request: %w", err)
	}

	// Vérifier que le créateur existe et peut participer
	// TODO: Vérifier avec le service Player quand il sera disponible

	// Créer la session
	session := &models.CombatSession{
		ID:              uuid.New(),
		Type:            req.Type,
		Status:          "waiting",
		ZoneID:          req.ZoneID,
		CreatedBy:       creatorID,
		MaxParticipants: req.MaxParticipants,
		IsPrivate:       req.IsPrivate,
		LevelRange:      req.LevelRange,
		Rules:           req.Rules,
		LastActionAt:    time.Now(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Enregistrer en base
	if err := s.combatRepo.CreateSession(session); err != nil {
		return nil, fmt.Errorf("failed to create combat session: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"session_id": session.ID,
		"type":       session.Type,
		"zone_id":    session.ZoneID,
		"creator_id": creatorID,
	}).Info("Combat session created")

	return session, nil
}

// GetCombatSession récupère une session de combat par ID
func (s *CombatService) GetCombatSession(sessionID uuid.UUID) (*models.CombatSession, error) {
	session, err := s.combatRepo.GetSessionByID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get combat session: %w", err)
	}

	// Charger les participants
	participants, err := s.combatRepo.GetSessionParticipants(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participants: %w", err)
	}

	session.Participants = make([]models.CombatParticipant, len(participants))
	for i, p := range participants {
		session.Participants[i] = *p
	}

	return session, nil
}

// JoinCombatSession permet à un joueur de rejoindre une session
func (s *CombatService) JoinCombatSession(sessionID, characterID uuid.UUID) error {
	// Récupérer la session
	session, err := s.combatRepo.GetSessionByID(sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	// Vérifier que la session accepte de nouveaux joueurs
	if session.Status != "waiting" {
		return fmt.Errorf("session is not accepting new players")
	}

	// Vérifier le nombre de participants
	participants, err := s.combatRepo.GetSessionParticipants(sessionID)
	if err != nil {
		return fmt.Errorf("failed to check participants: %w", err)
	}

	if len(participants) >= session.MaxParticipants {
		return fmt.Errorf("session is full")
	}

	// TODO: Récupérer les stats du personnage depuis le service Player
	// Pour l'instant, utiliser des stats par défaut
	participant := &models.CombatParticipant{
		ID:            uuid.New(),
		SessionID:     sessionID,
		CharacterID:   characterID,
		PlayerID:      characterID, // TODO: récupérer le vrai player ID
		Team:          s.assignTeam(participants),
		Position:      models.Position{X: 0, Y: 0, Z: 0}, // Position de départ
		Status:        "alive",
		CurrentHealth: 100,
		MaxHealth:     100,
		CurrentMana:   50,
		MaxMana:       50,
		Damage:        20,
		Defense:       10,
		CritChance:    0.05,
		AttackSpeed:   1.0,
		JoinedAt:      time.Now(),
	}

	if err := s.combatRepo.AddParticipant(participant); err != nil {
		return fmt.Errorf("failed to add participant: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"session_id":   sessionID,
		"character_id": characterID,
		"team":         participant.Team,
	}).Info("Character joined combat session")

	return nil
}

// LeaveCombatSession permet à un joueur de quitter une session
func (s *CombatService) LeaveCombatSession(sessionID, characterID uuid.UUID) error {
	if err := s.combatRepo.RemoveParticipant(sessionID, characterID); err != nil {
		return fmt.Errorf("failed to remove participant: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"session_id":   sessionID,
		"character_id": characterID,
	}).Info("Character left combat session")

	return nil
}

// ExecuteAction exécute une action de combat
func (s *CombatService) ExecuteAction(actionReq models.PerformActionRequest) (*models.CombatActionResult, error) {
	// Récupérer la session
	session, err := s.combatRepo.GetSessionByID(actionReq.SessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if session.Status != "active" {
		return nil, fmt.Errorf("session is not active")
	}

	// Récupérer le participant qui fait l'action
	participant, err := s.combatRepo.GetParticipantByCharacter(actionReq.SessionID, actionReq.CharacterID)
	if err != nil {
		return nil, fmt.Errorf("participant not found: %w", err)
	}

	if participant.Status != "alive" {
		return nil, fmt.Errorf("dead participants cannot act")
	}

	// Créer l'action de combat
	action := &models.CombatAction{
		ID:         uuid.New(),
		SessionID:  actionReq.SessionID,
		ActorID:    actionReq.CharacterID, // Utiliser CharacterID depuis PerformActionRequest
		Type:       actionReq.Type,
		ActionData: actionReq.ActionData, // ActionData est déjà json.RawMessage
		Targets:    actionReq.Targets,
		ExecutedAt: time.Now(),
	}

	// Exécuter selon le type d'action
	switch actionReq.Type {
	case "attack":
		return s.executeAttack(action, participant)
	case "spell":
		return s.executeSpell(action, participant)
	case "move":
		return s.executeMovement(action, participant)
	case "item":
		return s.executeItemUse(action, participant)
	case "defend":
		return s.executeDefend(action, participant)
	default:
		return nil, fmt.Errorf("unknown action type: %s", actionReq.Type)
	}
}

// executeAttack exécute une attaque
func (s *CombatService) executeAttack(action *models.CombatAction, attacker *models.CombatParticipant) (*models.CombatActionResult, error) {
	// Pour l'instant, attaque basique
	results := make([]models.ActionResult, len(action.Targets))
	
	for i, targetID := range action.Targets {
		target, err := s.combatRepo.GetParticipantByCharacter(action.SessionID, targetID)
		if err != nil {
			logrus.WithError(err).Error("Failed to get target participant")
			continue
		}

		// Calcul des dégâts basique
		damage := attacker.Damage
		
		// Appliquer la défense
		finalDamage := damage - target.Defense
		if finalDamage < 1 {
			finalDamage = 1
		}

		// Appliquer les dégâts
		target.CurrentHealth -= finalDamage
		if target.CurrentHealth < 0 {
			target.CurrentHealth = 0
			target.Status = "dead"
		}

		// Mettre à jour le participant
		if err := s.combatRepo.UpdateParticipant(target); err != nil {
			logrus.WithError(err).Error("Failed to update target participant")
		}

		results[i] = models.ActionResult{
			TargetID:    targetID,
			DamageDealt: finalDamage,
		}
	}

	action.Results = results
	action.Success = true

	// Enregistrer l'action
	if err := s.combatRepo.RecordAction(action); err != nil {
		logrus.WithError(err).Error("Failed to record attack action")
	}

	return &models.CombatActionResult{
		Results:     results,
		Success:     true,
		CriticalHit: false,
		Duration:    1000 * time.Millisecond,
	}, nil
}

// executeSpell exécute un sort
func (s *CombatService) executeSpell(action *models.CombatAction, caster *models.CombatParticipant) (*models.CombatActionResult, error) {
	// TODO: Implémenter la logique des sorts avec le SpellRepository
	
	results := make([]models.ActionResult, len(action.Targets))
	
	for i, targetID := range action.Targets {
		results[i] = models.ActionResult{
			TargetID:    targetID,
			DamageDealt: 15, // Dégâts magiques par défaut
		}
	}

	action.Results = results
	action.Success = true

	if err := s.combatRepo.RecordAction(action); err != nil {
		logrus.WithError(err).Error("Failed to record spell action")
	}

	return &models.CombatActionResult{
		Results:     results,
		Success:     true,
		CriticalHit: false,
		Duration:    2000 * time.Millisecond,
	}, nil
}

// executeMovement exécute un mouvement
func (s *CombatService) executeMovement(action *models.CombatAction, participant *models.CombatParticipant) (*models.CombatActionResult, error) {
	// Extraire la nouvelle position des ActionData
	var actionData map[string]interface{}
	if err := json.Unmarshal(action.ActionData, &actionData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal action data: %w", err)
	}

	positionData, ok := actionData["position"]
	if !ok {
		return nil, fmt.Errorf("missing position data")
	}

	positionMap, ok := positionData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid position format")
	}

	newPosition := models.Position{
		X: positionMap["x"].(float64),
		Y: positionMap["y"].(float64),
		Z: positionMap["z"].(float64),
	}

	// Mettre à jour la position
	participant.Position = newPosition
	now := time.Now()
	participant.LastActionAt = &now

	if err := s.combatRepo.UpdateParticipant(participant); err != nil {
		return nil, fmt.Errorf("failed to update participant position: %w", err)
	}

	action.Success = true

	if err := s.combatRepo.RecordAction(action); err != nil {
		logrus.WithError(err).Error("Failed to record movement action")
	}

	return &models.CombatActionResult{
		Results:  []models.ActionResult{},
		Success:  true,
		Duration: 500 * time.Millisecond,
	}, nil
}

// executeItemUse exécute l'utilisation d'un objet
func (s *CombatService) executeItemUse(action *models.CombatAction, participant *models.CombatParticipant) (*models.CombatActionResult, error) {
	// TODO: Implémenter avec le service d'inventaire
	
	action.Success = true

	if err := s.combatRepo.RecordAction(action); err != nil {
		logrus.WithError(err).Error("Failed to record item use action")
	}

	return &models.CombatActionResult{
		Results:  []models.ActionResult{},
		Success:  true,
		Duration: 1000 * time.Millisecond,
	}, nil
}

// executeDefend exécute une action de défense
func (s *CombatService) executeDefend(action *models.CombatAction, participant *models.CombatParticipant) (*models.CombatActionResult, error) {
	// TODO: Implémenter la logique de défense (augmenter temporairement la défense)
	
	action.Success = true

	if err := s.combatRepo.RecordAction(action); err != nil {
		logrus.WithError(err).Error("Failed to record defend action")
	}

	return &models.CombatActionResult{
		Results:  []models.ActionResult{},
		Success:  true,
		Duration: 500 * time.Millisecond,
	}, nil
}

// AddParticipant ajoute un participant à une session
func (s *CombatService) AddParticipant(participant *models.CombatParticipant) error {
	if err := s.combatRepo.AddParticipant(participant); err != nil {
		return fmt.Errorf("failed to add participant: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"session_id":   participant.SessionID,
		"character_id": participant.CharacterID,
	}).Info("Participant added to combat session")
	
	return nil
}

// GetParticipants récupère les participants d'une session
func (s *CombatService) GetParticipants(sessionID uuid.UUID) ([]*models.CombatParticipant, error) {
	// Utiliser la méthode correcte qui existe dans le repository
	return s.combatRepo.GetSessionParticipants(sessionID)
}

// EndCombatSession termine une session de combat
func (s *CombatService) EndCombatSession(sessionID uuid.UUID) error {
	if err := s.combatRepo.EndSession(sessionID); err != nil {
		return fmt.Errorf("failed to end combat session: %w", err)
	}

	logrus.WithField("session_id", sessionID).Info("Combat session ended")

	return nil
}

// Méthodes utilitaires

// validateStartCombatRequest valide une requête de création de combat
func (s *CombatService) validateStartCombatRequest(req models.StartCombatRequest) error {
	if req.Type == "" {
		return fmt.Errorf("combat type is required")
	}

	if req.ZoneID == "" {
		return fmt.Errorf("zone ID is required")
	}

	if req.MaxParticipants < 1 {
		return fmt.Errorf("max participants must be at least 1")
	}

	return nil
}

// assignTeam assigne une équipe à un nouveau participant
func (s *CombatService) assignTeam(existingParticipants []*models.CombatParticipant) int {
	if len(existingParticipants) == 0 {
		return 1 // Première équipe
	}

	// Compter les participants par équipe
	teamCounts := make(map[int]int)
	for _, p := range existingParticipants {
		teamCounts[p.Team]++
	}

	// Assigner à l'équipe avec le moins de participants
	minCount := int(^uint(0) >> 1) // Max int
	bestTeam := 1

	for team, count := range teamCounts {
		if count < minCount {
			minCount = count
			bestTeam = team
		}
	}

	// Si toutes les équipes sont pleines, créer une nouvelle
	if minCount > 0 {
		maxTeam := 0
		for team := range teamCounts {
			if team > maxTeam {
				maxTeam = team
			}
		}
		return maxTeam + 1
	}

	return bestTeam
}