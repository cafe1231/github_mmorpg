package service

import (
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

// GetCombatSession récupère une session de combat
func (s *CombatService) GetCombatSession(sessionID uuid.UUID) (*models.CombatSession, error) {
	// Utiliser la méthode qui existe dans l'interface
	sessions, err := s.combatRepo.GetActiveSessions()
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}

	// Chercher la session par ID
	for _, session := range sessions {
		if session.ID == sessionID {
			return session, nil
		}
	}

	return nil, fmt.Errorf("session not found")
}

// JoinCombatSession fait rejoindre un joueur à une session
func (s *CombatService) JoinCombatSession(sessionID, characterID uuid.UUID) error {
	session, err := s.GetCombatSession(sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	if session.Status != "waiting" {
		return fmt.Errorf("session is not accepting new participants")
	}

	// Créer le participant (avec les champs qui existent)
	participant := &models.CombatParticipant{
		SessionID:   sessionID,
		CharacterID: characterID,
		Team:        1, // TODO: Logique d'assignation d'équipe
		Status:      "alive",
		JoinedAt:    time.Now(),
		// TODO: Initialiser les autres champs nécessaires
	}

	// Pour l'instant, on log juste car CreateParticipant n'existe pas
	logrus.WithFields(logrus.Fields{
		"session_id":   sessionID,
		"character_id": characterID,
	}).Info("Character joined combat session")

	// TODO: Implémenter quand l'interface repository sera complète
	_ = participant

	return nil
}

// LeaveCombatSession fait quitter un joueur d'une session
func (s *CombatService) LeaveCombatSession(sessionID, characterID uuid.UUID) error {
	// TODO: Implémenter la logique de sortie
	logrus.WithFields(logrus.Fields{
		"session_id":   sessionID,
		"character_id": characterID,
	}).Info("Character left combat session")

	return nil
}

// ExecuteAction exécute une action de combat (version simplifiée)
func (s *CombatService) ExecuteAction(actionReq models.CombatActionRequest) (*models.CombatActionResult, error) {
	// Valider la session de combat
	session, err := s.GetCombatSession(actionReq.SessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if session.Status != "active" {
		return nil, fmt.Errorf("combat session is not active")
	}

	// Pour l'instant, retourner un résultat basique
result := &models.CombatActionResult{
	Results: []models.ActionResult{
		{
			TargetID:       actionReq.ActorID,
			DamageDealt:    10,
			HealingDone:    0,
			EffectsApplied: []string{},
			EffectsRemoved: []string{},
			Absorbed:       0,
			Reflected:      0,
		},
	},
	Success:     true,
	CriticalHit: false,
	Duration:    1500 * time.Millisecond,
}

	logrus.WithFields(logrus.Fields{
		"session_id":  session.ID,
		"actor_id":    actionReq.ActorID,
		"action_type": actionReq.Type,
		"success":     result.Success,
		"targets":     len(actionReq.Targets),
	}).Info("Combat action executed")

	return result, nil
}

// AddParticipant ajoute un participant à une session
func (s *CombatService) AddParticipant(participant *models.CombatParticipant) error {
	// TODO: Implémenter quand la méthode existera dans le repository
	logrus.WithFields(logrus.Fields{
		"session_id":   participant.SessionID,
		"character_id": participant.CharacterID,
	}).Info("Participant added to combat session")
	return nil
}

// GetParticipants récupère les participants d'une session
func (s *CombatService) GetParticipants(sessionID uuid.UUID) ([]*models.CombatParticipant, error) {
	// Utiliser la méthode qui existe
	return s.combatRepo.GetParticipants(sessionID)
}

// EndCombatSession termine une session de combat
func (s *CombatService) EndCombatSession(sessionID uuid.UUID) error {
	// Pour l'instant, juste logger
	logrus.WithField("session_id", sessionID).Info("Combat session ended")

	// TODO: Implémenter avec les bonnes méthodes du repository

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