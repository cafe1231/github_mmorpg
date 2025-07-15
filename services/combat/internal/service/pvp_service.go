// internal/service/pvp_service.go
package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"combat/internal/config"
	"combat/internal/models"
	"combat/internal/repository"
)

// PvPServiceInterface définit les méthodes du service PvP
type PvPServiceInterface interface {
	CreateChallenge(challengerID, targetID uuid.UUID, matchType string) (*models.PvPChallenge, error)
	AcceptChallenge(challengeID, playerID uuid.UUID) (*models.CombatSession, error)
	DeclineChallenge(challengeID, playerID uuid.UUID) error
	CancelChallenge(challengeID, playerID uuid.UUID) error
	GetPlayerChallenges(playerID uuid.UUID) ([]*models.PvPChallenge, error)
	StartMatchmaking(playerID uuid.UUID, preferences models.MatchmakingRequest) error
	StopMatchmaking(playerID uuid.UUID) error
}

// PvPService gère la logique métier du PvP
type PvPService struct {
	config        *config.Config
	pvpRepo       repository.PvPRepositoryInterface
	combatService CombatServiceInterface
}

// NewPvPService crée une nouvelle instance du service PvP
func NewPvPService(
	cfg *config.Config,
	pvpRepo repository.PvPRepositoryInterface,
	combatService CombatServiceInterface,
) PvPServiceInterface {
	return &PvPService{
		config:        cfg,
		pvpRepo:       pvpRepo,
		combatService: combatService,
	}
}

// CreateChallenge crée un défi PvP
func (s *PvPService) CreateChallenge(challengerID, targetID uuid.UUID, matchType string) (*models.PvPChallenge, error) {
	// TODO: Vérifier que les joueurs existent et peuvent se défier
	
	challenge := &models.PvPChallenge{
		ID:           uuid.New(),
		ChallengerID: challengerID,
		ChallengedID: targetID,      // Utiliser le bon nom de champ du modèle existant
		Type:         matchType,     // Utiliser Type au lieu de MatchType
		Status:       "pending",
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(5 * time.Minute), // Expire dans 5 minutes
	}

	if err := s.pvpRepo.CreateChallenge(challenge); err != nil {
		return nil, fmt.Errorf("failed to create challenge: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"challenge_id":  challenge.ID,
		"challenger_id": challengerID,
		"target_id":     targetID,
		"match_type":    matchType,
	}).Info("PvP challenge created")

	return challenge, nil
}

// AcceptChallenge accepte un défi PvP
func (s *PvPService) AcceptChallenge(challengeID, playerID uuid.UUID) (*models.CombatSession, error) {
	challenge, err := s.pvpRepo.GetChallenge(challengeID)
	if err != nil {
		return nil, fmt.Errorf("challenge not found: %w", err)
	}

	if challenge.Status != "pending" {
		return nil, fmt.Errorf("challenge is not pending")
	}

	if challenge.ChallengedID != playerID { // Utiliser ChallengedID
		return nil, fmt.Errorf("not authorized to accept this challenge")
	}

	// Marquer le défi comme accepté
	challenge.Status = "accepted"
	now := time.Now()
	challenge.RespondedAt = &now // Utiliser RespondedAt du modèle existant

	if err := s.pvpRepo.UpdateChallenge(challenge); err != nil {
		return nil, fmt.Errorf("failed to update challenge: %w", err)
	}

	// Créer la session de combat
	session, err := s.createChallengeSession(challenge)
	if err != nil {
		return nil, fmt.Errorf("failed to create combat session: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"challenge_id": challengeID,
		"session_id":   session.ID,
		"player_id":    playerID,
	}).Info("PvP challenge accepted")

	return session, nil
}

// DeclineChallenge refuse un défi PvP
func (s *PvPService) DeclineChallenge(challengeID, playerID uuid.UUID) error {
	challenge, err := s.pvpRepo.GetChallenge(challengeID)
	if err != nil {
		return fmt.Errorf("challenge not found: %w", err)
	}

	if challenge.ChallengedID != playerID { // Utiliser ChallengedID
		return fmt.Errorf("not authorized to decline this challenge")
	}

	challenge.Status = "declined"
	now := time.Now()
	challenge.RespondedAt = &now

	if err := s.pvpRepo.UpdateChallenge(challenge); err != nil {
		return fmt.Errorf("failed to update challenge: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"challenge_id": challengeID,
		"player_id":    playerID,
	}).Info("PvP challenge declined")

	return nil
}

// CancelChallenge annule un défi PvP
func (s *PvPService) CancelChallenge(challengeID, playerID uuid.UUID) error {
	challenge, err := s.pvpRepo.GetChallenge(challengeID)
	if err != nil {
		return fmt.Errorf("challenge not found: %w", err)
	}

	if challenge.ChallengerID != playerID {
		return fmt.Errorf("not authorized to cancel this challenge")
	}

	challenge.Status = "cancelled"
	now := time.Now()
	challenge.RespondedAt = &now

	if err := s.pvpRepo.UpdateChallenge(challenge); err != nil {
		return fmt.Errorf("failed to update challenge: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"challenge_id": challengeID,
		"player_id":    playerID,
	}).Info("PvP challenge cancelled")

	return nil
}

// GetPlayerChallenges récupère les défis d'un joueur
func (s *PvPService) GetPlayerChallenges(playerID uuid.UUID) ([]*models.PvPChallenge, error) {
	// Pour l'instant, retourner une liste vide
	// TODO: Implémenter quand les méthodes du repository seront disponibles
	return []*models.PvPChallenge{}, nil
}

// StartMatchmaking démarre le matchmaking pour un joueur
func (s *PvPService) StartMatchmaking(playerID uuid.UUID, preferences models.MatchmakingRequest) error {
	// TODO: Implémenter la logique de matchmaking
	logrus.WithFields(logrus.Fields{
		"player_id":  playerID,
		"match_type": preferences.MatchType,
		"rating":     preferences.Rating,
	}).Info("Matchmaking started")

	return nil
}

// StopMatchmaking arrête le matchmaking pour un joueur
func (s *PvPService) StopMatchmaking(playerID uuid.UUID) error {
	// TODO: Implémenter l'arrêt du matchmaking
	logrus.WithField("player_id", playerID).Info("Matchmaking stopped")

	return nil
}

// Méthodes privées

// createChallengeSession crée une session de combat pour un défi
func (s *PvPService) createChallengeSession(challenge *models.PvPChallenge) (*models.CombatSession, error) {
	// Déterminer la zone PvP appropriée
	zoneID := s.selectPvPZone(challenge.Type)
	
	// Créer la requête de combat
	combatReq := models.StartCombatRequest{
		Type:            "pvp",
		ZoneID:          zoneID,
		MaxParticipants: 2,
		IsPrivate:       true,
		LevelRange:      models.LevelRange{Min: 1, Max: 100},
		Rules: models.CombatRules{
			FriendlyFire:   false,
			AllowItems:     true,
			AllowFleeing:   false,
			TurnBased:      false,
			TimeLimit:      10 * time.Minute,
			RespawnAllowed: false,
			ExperienceGain: 1.0,
		},
	}
	
	// Créer la session via le service de combat
	session, err := s.combatService.CreateCombatSession(challenge.ChallengerID, combatReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create combat session: %w", err)
	}
	
	// TODO: Ajouter les participants
	logrus.WithFields(logrus.Fields{
		"session_id":    session.ID,
		"challenger_id": challenge.ChallengerID,
		"challenged_id": challenge.ChallengedID,
	}).Info("PvP session created")

	return session, nil
}

// selectPvPZone sélectionne une zone PvP appropriée
func (s *PvPService) selectPvPZone(matchType string) string {
	zoneMap := map[string]string{
		"duel":       "arena_1v1",
		"team_match": "arena_team",
		"ranked":     "ranked_arena",
		"casual":     "casual_arena",
	}
	
	if zone, exists := zoneMap[matchType]; exists {
		return zone
	}
	
	return "default_arena" // Zone par défaut
}