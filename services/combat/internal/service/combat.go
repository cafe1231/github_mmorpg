// internal/service/combat.go
package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"combat/internal/config"
	"combat/internal/external"
	"combat/internal/models"
	"combat/internal/repository"
)

// CombatServiceInterface définit les méthodes du service combat
type CombatServiceInterface interface {
	// Gestion des sessions
	CreateCombatSession(creatorID uuid.UUID, req models.StartCombatRequest) (*models.CombatSession, error)
	GetCombatSession(sessionID uuid.UUID) (*models.CombatStatusResponse, error)
	JoinCombat(sessionID uuid.UUID, req models.JoinCombatRequest) error
	LeaveCombat(sessionID uuid.UUID, characterID uuid.UUID) error
	EndCombat(sessionID uuid.UUID, winnerTeam int) error
	
	// Actions de combat
	PerformAction(req models.PerformActionRequest) (*models.CombatAction, error)
	ValidateAction(sessionID uuid.UUID, characterID uuid.UUID, actionType string, targets []uuid.UUID) error
	
	// Gestion des participants
	GetParticipants(sessionID uuid.UUID) ([]*models.CombatParticipant, error)
	UpdateParticipantStats(participant *models.CombatParticipant) error
	HandleParticipantDeath(sessionID uuid.UUID, characterID uuid.UUID) error
	
	// Utilitaires
	GetActiveCombats() ([]*models.CombatSession, error)
	GetCombatsByZone(zoneID string) ([]*models.CombatSession, error)
	GetCombatStatistics() (*models.CombatStatistics, error)
	CleanupInactiveSessions() error
}

// CombatService implémente l'interface CombatServiceInterface
type CombatService struct {
	config          *config.Config
	combatRepo      repository.CombatRepositoryInterface
	spellRepo       repository.SpellRepositoryInterface
	effectRepo      repository.EffectRepositoryInterface
	combatLogRepo   repository.CombatLogRepositoryInterface
	playerClient    external.PlayerClientInterface
	worldClient     external.WorldClientInterface
	damageCalc      DamageCalculatorInterface
	effectService   EffectServiceInterface
}

// NewCombatService crée une nouvelle instance du service combat
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

	// Vérifier que le créateur peut créer un combat dans cette zone
	canCreate, err := s.worldClient.CanCreateCombat(creatorID, req.ZoneID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate zone access: %w", err)
	}
	if !canCreate {
		return nil, fmt.Errorf("cannot create combat in this zone")
	}

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

	if err := s.combatRepo.CreateSession(session); err != nil {
		return nil, fmt.Errorf("failed to create combat session: %w", err)
	}

	// Log de création
	logEntry := &models.CombatLog{
		ID:        uuid.New(),
		SessionID: session.ID,
		EventType: "session_created",
		Message:   fmt.Sprintf("Combat session created by character %s", creatorID),
		Timestamp: time.Now(),
	}
	s.combatLogRepo.CreateLog(logEntry)

	logrus.WithFields(logrus.Fields{
		"session_id": session.ID,
		"type":       session.Type,
		"zone_id":    session.ZoneID,
		"creator_id": creatorID,
	}).Info("Combat session created")

	return session, nil
}

// GetCombatSession récupère le statut complet d'une session
func (s *CombatService) GetCombatSession(sessionID uuid.UUID) (*models.CombatStatusResponse, error) {
	session, err := s.combatRepo.GetSessionByID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	participants, err := s.combatRepo.GetSessionParticipants(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participants: %w", err)
	}

	response := &models.CombatStatusResponse{
		Session:      *session,
		Participants: convertParticipantSlice(participants),
		RecentActions: convertActionSlice([]*models.CombatAction{}),
	}

	return response, nil
}

// JoinCombat permet à un personnage de rejoindre un combat
func (s *CombatService) JoinCombat(sessionID uuid.UUID, req models.JoinCombatRequest) error {
	// Récupérer la session
	session, err := s.combatRepo.GetSessionByID(sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	if session.Status != "waiting" {
		return fmt.Errorf("combat has already started")
	}

	// Récupérer les informations du personnage
	character, err := s.playerClient.GetCharacter(req.CharacterID)
	if err != nil {
		return fmt.Errorf("failed to get character: %w", err)
	}

	// Valider le join
	if err := s.validateJoinCombat(session, character, req); err != nil {
		return err
	}

	// Créer le participant
	participant := &models.CombatParticipant{
		ID:          uuid.New(),
		SessionID:   sessionID,
		CharacterID: req.CharacterID,
		Team:        req.Team,
		Position:    models.Position{X: 0, Y: 0, Z: 0}, // Position par défaut
		Status:      "alive",
		JoinedAt:    time.Now(),
	}

	if err := s.combatRepo.AddParticipant(participant); err != nil {
		return fmt.Errorf("failed to add participant: %w", err)
	}

	// Log du join
	logEntry := &models.CombatLog{
		ID:        uuid.New(),
		SessionID: sessionID,
		ActorID:   &req.CharacterID,
		EventType: "participant_joined",
		Message:   fmt.Sprintf("%s joined the combat", character.Name),
		Color:     "#00FF00",
		Timestamp: time.Now(),
	}
	s.combatLogRepo.CreateLog(logEntry)

	// Vérifier si le combat peut commencer
	if err := s.checkStartConditions(session); err == nil {
		s.startCombatSession(sessionID)
	}

	return nil
}

// LeaveCombat permet à un personnage de quitter un combat
func (s *CombatService) LeaveCombat(sessionID uuid.UUID, characterID uuid.UUID) error {
	participant, err := s.combatRepo.GetParticipantByCharacter(sessionID, characterID)
	if err != nil {
		return fmt.Errorf("participant not found: %w", err)
	}

	// Marquer comme ayant fui
	participant.Status = "fled"
	leftAt := time.Now()
	// Note: LeftAt field doesn't exist in the model, using a workaround
	participant.LastActionAt = &leftAt

	if err := s.combatRepo.UpdateParticipant(participant); err != nil {
		return fmt.Errorf("failed to update participant: %w", err)
	}

	// Log de départ
	logEntry := &models.CombatLog{
		ID:        uuid.New(),
		SessionID: sessionID,
		ActorID:   &characterID,
		EventType: "participant_left",
		Message:   "left the combat",
		Color:     "#FF8C00",
		Timestamp: time.Now(),
	}
	s.combatLogRepo.CreateLog(logEntry)

	// Vérifier si le combat doit se terminer
	session, err := s.combatRepo.GetSessionByID(sessionID)
	if err == nil {
		participants, err := s.combatRepo.GetSessionParticipants(sessionID)
		if err == nil && s.shouldEndCombat(session, participants) {
			winnerTeam := s.determineWinnerTeam(participants)
			s.endCombatSession(sessionID, winnerTeam)
		}
	}

	return nil
}

// EndCombat termine un combat avec une équipe gagnante
func (s *CombatService) EndCombat(sessionID uuid.UUID, winnerTeam int) error {
	return s.endCombatSession(sessionID, winnerTeam)
}

// PerformAction exécute une action de combat
func (s *CombatService) PerformAction(req models.PerformActionRequest) (*models.CombatAction, error) {
	// Valider l'action
	if err := s.ValidateAction(req.SessionID, req.CharacterID, req.Type, req.Targets); err != nil {
		return nil, fmt.Errorf("invalid action: %w", err)
	}

	// Récupérer le participant
	participant, err := s.combatRepo.GetParticipantByCharacter(req.SessionID, req.CharacterID)
	if err != nil {
		return nil, fmt.Errorf("participant not found: %w", err)
	}

	if participant.Status != "alive" {
		return nil, fmt.Errorf("character is not alive")
	}

	// Créer l'action
	action := &models.CombatAction{
		ID:         uuid.New(),
		SessionID:  req.SessionID,
		ActorID:    req.CharacterID,
		Type:       req.Type,
		ActionData: req.ActionData,
		Targets:    req.Targets,
		Success:    true,
		ExecutedAt: time.Now(),
	}

	// Traiter l'action selon son type
	switch req.Type {
	case "attack":
		err = s.processAttackAction(action, participant)
	case "spell":
		err = s.processSpellAction(action, participant)
	case "move":
		err = s.processMoveAction(action, participant)
	case "item":
		err = s.processItemAction(action, participant)
	case "defend":
		err = s.processDefendAction(action, participant)
	default:
		return nil, fmt.Errorf("unknown action type: %s", req.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to process action: %w", err)
	}

	// Sauvegarder l'action
	if err := s.combatRepo.RecordAction(action); err != nil {
		logrus.WithError(err).Error("Failed to save combat action")
	}

	// Mettre à jour le timestamp de la session
	s.updateSessionTimestamp(req.SessionID)

	return action, nil
}

// ValidateAction valide qu'une action peut être exécutée
func (s *CombatService) ValidateAction(sessionID uuid.UUID, characterID uuid.UUID, actionType string, targets []uuid.UUID) error {
	// Vérifier que la session existe et est active
	session, err := s.combatRepo.GetSessionByID(sessionID)
	if err != nil {
		return fmt.Errorf("session not found")
	}

	if session.Status != "active" {
		return fmt.Errorf("combat is not active")
	}

	// Vérifier que le personnage est dans le combat
	participant, err := s.combatRepo.GetParticipantByCharacter(sessionID, characterID)
	if err != nil {
		return fmt.Errorf("character not in combat")
	}

	if participant.Status != "alive" {
		return fmt.Errorf("character cannot act")
	}

	// Validation spécifique par type d'action
	switch actionType {
	case "attack":
		if len(targets) != 1 {
			return fmt.Errorf("attack requires exactly one target")
		}

	case "spell":
		if len(targets) == 0 {
			// Vérifier si le sort peut être lancé sans cible
			// Cette validation sera affinée avec les données du sort
		}

	case "move":
		if len(targets) > 0 {
			return fmt.Errorf("move action cannot have targets")
		}

	case "item":
		// La validation des objets sera faite dans processItemAction
		break

	case "defend":
		// Défense peut être sans cible (défense globale)
		break

	default:
		return fmt.Errorf("unknown action type: %s", actionType)
	}

	// Valider les cibles
	if len(targets) > 0 {
		if err := s.validateTargets(sessionID, characterID, targets); err != nil {
			return fmt.Errorf("invalid targets: %w", err)
		}
	}

	return nil
}

// GetParticipants récupère les participants d'un combat
func (s *CombatService) GetParticipants(sessionID uuid.UUID) ([]*models.CombatParticipant, error) {
	return s.combatRepo.GetSessionParticipants(sessionID)
}

// UpdateParticipantStats met à jour les statistiques d'un participant
func (s *CombatService) UpdateParticipantStats(participant *models.CombatParticipant) error {
	return s.combatRepo.UpdateParticipant(participant)
}

// HandleParticipantDeath gère la mort d'un participant
func (s *CombatService) HandleParticipantDeath(sessionID uuid.UUID, characterID uuid.UUID) error {
	participant, err := s.combatRepo.GetParticipantByCharacter(sessionID, characterID)
	if err != nil {
		return err
	}

	participant.Status = "dead"
	if err := s.combatRepo.UpdateParticipant(participant); err != nil {
		return err
	}

	// Log de mort
	logEntry := &models.CombatLog{
		ID:        uuid.New(),
		SessionID: sessionID,
		TargetID:  &characterID,
		EventType: "participant_death",
		Message:   "has been defeated",
		Color:     "#FF0000",
		Timestamp: time.Now(),
	}
	s.combatLogRepo.CreateLog(logEntry)

	// Vérifier si le combat doit se terminer
	session, err := s.combatRepo.GetSessionByID(sessionID)
	if err == nil {
		participants, err := s.combatRepo.GetSessionParticipants(sessionID)
		if err == nil && s.shouldEndCombat(session, participants) {
			winnerTeam := s.determineWinnerTeam(participants)
			s.endCombatSession(sessionID, winnerTeam)
		}
	}

	return nil
}

// GetActiveCombats récupère tous les combats actifs
func (s *CombatService) GetActiveCombats() ([]*models.CombatSession, error) {
	return s.combatRepo.GetActiveSessions()
}

// GetCombatsByZone récupère les combats d'une zone
func (s *CombatService) GetCombatsByZone(zoneID string) ([]*models.CombatSession, error) {
	return s.combatRepo.GetSessionsByZone(zoneID)
}

// GetCombatStatistics récupère les statistiques de combat
func (s *CombatService) GetCombatStatistics() (*models.CombatStatistics, error) {
	return s.combatRepo.GetCombatStatistics()
}

// CleanupInactiveSessions nettoie les sessions inactives
func (s *CombatService) CleanupInactiveSessions() error {
	count, err := s.combatRepo.CleanupInactiveSessions()
	if err != nil {
		return err
	}

	if count > 0 {
		logrus.WithField("count", count).Info("Cleaned up inactive combat sessions")
	}

	return nil
}

// Fonctions de conversion pour éviter les erreurs de types
func convertParticipantSlice(participants []*models.CombatParticipant) []models.CombatParticipant {
	if participants == nil {
		return []models.CombatParticipant{}
	}
	
	result := make([]models.CombatParticipant, len(participants))
	for i, p := range participants {
		if p != nil {
			result[i] = *p
		}
	}
	return result
}

func convertActionSlice(actions []*models.CombatAction) []models.CombatAction {
	if actions == nil {
		return []models.CombatAction{}
	}
	
	result := make([]models.CombatAction, len(actions))
	for i, a := range actions {
		if a != nil {
			result[i] = *a
		}
	}
	return result
}

// updateSessionTimestamp met à jour le timestamp d'une session (méthode helper privée)
func (s *CombatService) updateSessionTimestamp(sessionID uuid.UUID) {
	// Mettre à jour la session avec le timestamp actuel
	session, err := s.combatRepo.GetSessionByID(sessionID)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get session for timestamp update")
		return
	}
	
	session.LastActionAt = time.Now()
	if err := s.combatRepo.UpdateSession(session); err != nil {
		logrus.WithError(err).Warn("Failed to update session timestamp")
	}
}

// UpdateSessionTimestamp met à jour le timestamp d'une session (méthode publique)
func (s *CombatService) UpdateSessionTimestamp(sessionID uuid.UUID) {
	s.updateSessionTimestamp(sessionID)
}

// Méthodes auxiliaires privées

// validateStartCombatRequest valide une requête de création de combat
func (s *CombatService) validateStartCombatRequest(req models.StartCombatRequest) error {
	if req.Type == "" {
		return fmt.Errorf("combat type is required")
	}

	if req.ZoneID == "" {
		return fmt.Errorf("zone ID is required")
	}

	if req.MaxParticipants < 2 || req.MaxParticipants > s.config.Combat.MaxParticipants {
		return fmt.Errorf("invalid max participants count")
	}

	if req.LevelRange.Min > req.LevelRange.Max {
		return fmt.Errorf("invalid level range")
	}

	return nil
}

// validateJoinCombat valide qu'un personnage peut rejoindre un combat
func (s *CombatService) validateJoinCombat(session *models.CombatSession, character *models.Character, req models.JoinCombatRequest) error {
	// Vérifier le niveau
	if session.LevelRange.Min > 0 && character.Level < session.LevelRange.Min {
		return fmt.Errorf("character level too low")
	}

	if session.LevelRange.Max > 0 && character.Level > session.LevelRange.Max {
		return fmt.Errorf("character level too high")
	}

	// Vérifier l'équipe
	if req.Team < 0 || req.Team > 2 {
		return fmt.Errorf("invalid team number")
	}

	// Vérifier que le personnage n'est pas déjà dans un autre combat
	activeCombats, err := s.combatRepo.GetActiveSessions()
	if err == nil {
		for _, combat := range activeCombats {
			if combat.ID != session.ID {
				participant, _ := s.combatRepo.GetParticipantByCharacter(combat.ID, req.CharacterID)
				if participant != nil {
					return fmt.Errorf("character is already in another combat")
				}
			}
		}
	}

	return nil
}

// validateTargets valide les cibles d'une action
func (s *CombatService) validateTargets(sessionID uuid.UUID, actorID uuid.UUID, targets []uuid.UUID) error {
	for _, targetID := range targets {
		target, err := s.combatRepo.GetParticipantByCharacter(sessionID, targetID)
		if err != nil {
			return fmt.Errorf("target %s not found", targetID)
		}

		if target.Status != "alive" {
			return fmt.Errorf("target %s is not alive", targetID)
		}
	}

	return nil
}

// checkStartConditions vérifie si le combat peut commencer
func (s *CombatService) checkStartConditions(session *models.CombatSession) error {
	participants, err := s.combatRepo.GetSessionParticipants(session.ID)
	if err != nil {
		return err
	}

	if len(participants) < 2 {
		return fmt.Errorf("not enough participants")
	}

	// Vérifier qu'il y a au moins 2 équipes différentes
	teams := make(map[int]bool)
	for _, p := range participants {
		teams[p.Team] = true
	}

	if len(teams) < 2 {
		return fmt.Errorf("need participants from at least 2 teams")
	}

	return nil
}

// startCombatSession démarre une session de combat
func (s *CombatService) startCombatSession(sessionID uuid.UUID) error {
	now := time.Now()
	session := &models.CombatSession{
		ID:        sessionID,
		Status:    "active",
		StartedAt: &now,
		UpdatedAt: now,
	}

	if err := s.combatRepo.UpdateSession(session); err != nil {
		return err
	}

	// Log de début
	logEntry := &models.CombatLog{
		ID:        uuid.New(),
		SessionID: sessionID,
		EventType: "combat_started",
		Message:   "Combat has begun!",
		Color:     "#FFD700",
		Timestamp: now,
	}
	s.combatLogRepo.CreateLog(logEntry)

	logrus.WithField("session_id", sessionID).Info("Combat started")

	return nil
}

// shouldEndCombat vérifie si le combat doit se terminer
func (s *CombatService) shouldEndCombat(session *models.CombatSession, participants []*models.CombatParticipant) bool {
	if session.Status != "active" {
		return false
	}

	// Compter les participants vivants par équipe
	aliveByTeam := make(map[int]int)
	for _, p := range participants {
		if p.Status == "alive" {
			aliveByTeam[p.Team]++
		}
	}

	// Le combat se termine s'il ne reste qu'une équipe ou aucune
	teamsAlive := 0
	for _, count := range aliveByTeam {
		if count > 0 {
			teamsAlive++
		}
	}

	return teamsAlive <= 1
}

// endCombatSession termine une session de combat
func (s *CombatService) endCombatSession(sessionID uuid.UUID, winnerTeam int) error {
	now := time.Now()
	session := &models.CombatSession{
		ID:        sessionID,
		Status:    "ended",
		EndedAt:   &now,
		UpdatedAt: now,
	}

	if err := s.combatRepo.UpdateSession(session); err != nil {
		return err
	}

	// Log de fin
	message := "Combat ended"
	if winnerTeam > 0 {
		message = fmt.Sprintf("Combat ended - Team %d wins!", winnerTeam)
	}

	logEntry := &models.CombatLog{
		ID:        uuid.New(),
		SessionID: sessionID,
		EventType: "combat_ended",
		Message:   message,
		Color:     "#FFD700",
		Timestamp: now,
	}
	s.combatLogRepo.CreateLog(logEntry)

	logrus.WithFields(logrus.Fields{
		"session_id":   sessionID,
		"winner_team": winnerTeam,
	}).Info("Combat ended")

	return nil
}

// determineWinnerTeam détermine l'équipe gagnante
func (s *CombatService) determineWinnerTeam(participants []*models.CombatParticipant) int {
	teamScores := make(map[int]int)

	for _, p := range participants {
		if p.Status == "alive" {
			teamScores[p.Team]++
		}
	}

	// Retourner l'équipe avec le plus de survivants
	winnerTeam := 0
	maxScore := 0

	for team, score := range teamScores {
		if score > maxScore {
			maxScore = score
			winnerTeam = team
		}
	}

	return winnerTeam
}

// Actions de combat - implémentations basiques

// processAttackAction traite une action d'attaque
func (s *CombatService) processAttackAction(action *models.CombatAction, participant *models.CombatParticipant) error {
	// Implémentation basique - à développer
	logEntry := &models.CombatLog{
		ID:        uuid.New(),
		SessionID: action.SessionID,
		ActorID:   &action.ActorID,
		EventType: "attack",
		Message:   "performed an attack",
		Color:     "#FF4500",
		Timestamp: time.Now(),
	}
	return s.combatLogRepo.CreateLog(logEntry)
}

// processSpellAction traite une action de sort
func (s *CombatService) processSpellAction(action *models.CombatAction, participant *models.CombatParticipant) error {
	// Implémentation basique - à développer
	logEntry := &models.CombatLog{
		ID:        uuid.New(),
		SessionID: action.SessionID,
		ActorID:   &action.ActorID,
		EventType: "spell",
		Message:   "cast a spell",
		Color:     "#4169E1",
		Timestamp: time.Now(),
	}
	return s.combatLogRepo.CreateLog(logEntry)
}

// processMoveAction traite une action de mouvement
func (s *CombatService) processMoveAction(action *models.CombatAction, participant *models.CombatParticipant) error {
	// Implémentation basique - à développer
	logEntry := &models.CombatLog{
		ID:        uuid.New(),
		SessionID: action.SessionID,
		ActorID:   &action.ActorID,
		EventType: "move",
		Message:   "moved to a new position",
		Color:     "#32CD32",
		Timestamp: time.Now(),
	}
	return s.combatLogRepo.CreateLog(logEntry)
}

// processItemAction traite une action d'objet
func (s *CombatService) processItemAction(action *models.CombatAction, participant *models.CombatParticipant) error {
	// Implémentation basique - à développer
	logEntry := &models.CombatLog{
		ID:        uuid.New(),
		SessionID: action.SessionID,
		ActorID:   &action.ActorID,
		EventType: "item",
		Message:   "used an item",
		Color:     "#FFD700",
		Timestamp: time.Now(),
	}
	return s.combatLogRepo.CreateLog(logEntry)
}

// processDefendAction traite une action de défense
func (s *CombatService) processDefendAction(action *models.CombatAction, participant *models.CombatParticipant) error {
	// Implémentation basique - à développer
	logEntry := &models.CombatLog{
		ID:        uuid.New(),
		SessionID: action.SessionID,
		ActorID:   &action.ActorID,
		EventType: "defend",
		Message:   "took a defensive stance",
		Color:     "#808080",
		Timestamp: time.Now(),
	}
	return s.combatLogRepo.CreateLog(logEntry)
}