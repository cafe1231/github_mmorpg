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
	// Gestion des défis
	CreateChallenge(challengerID uuid.UUID, req *models.CreateChallengeRequest) (*models.PvPChallenge, error)
	GetChallenge(id uuid.UUID) (*models.PvPChallenge, error)
	GetChallenges(req *models.GetChallengesRequest) ([]*models.PvPChallenge, error)
	RespondToChallenge(challengeID uuid.UUID, req *models.RespondToChallengeRequest) (*models.ChallengeResponse, error)
	CancelChallenge(challengeID, playerID uuid.UUID) error

	// Classements et statistiques
	GetRankings(req *models.GetRankingsRequest) (*models.RankingsResponse, error)
	GetPlayerStatistics(playerID uuid.UUID, season string) (*models.PvPStatistics, error)
	UpdatePlayerStatistics(playerID uuid.UUID, result *models.MatchResult) error
	GetCurrentSeasonInfo() (*models.SeasonInfo, error)

	// File d'attente et matchmaking
	JoinQueue(req *models.JoinQueueRequest) (*models.QueueResponse, error)
	LeaveQueue(playerID uuid.UUID) error
	GetQueueStatus(playerID uuid.UUID) (*models.QueueStatus, error)
	ProcessMatchmaking() error

	// Gestion des matches
	StartMatch(player1ID, player2ID uuid.UUID, matchType models.ChallengeType) (*models.PvPMatch, error)
	EndMatch(matchID uuid.UUID, result *models.MatchResult) error

	// Nettoyage et maintenance
	CleanupExpiredChallenges() error
	CleanupOldQueue() error
	StartCleanupRoutine()
}

// PvPService implémente l'interface PvPServiceInterface
type PvPService struct {
	pvpRepo     repository.PvPRepositoryInterface
	combatRepo  repository.CombatRepositoryInterface
	config      *config.Config
	queueTicker *time.Ticker
}

// NewPvPService crée un nouveau service PvP
func NewPvPService(
	pvpRepo repository.PvPRepositoryInterface,
	combatRepo repository.CombatRepositoryInterface,
	config *config.Config,
) PvPServiceInterface {
	service := &PvPService{
		pvpRepo:    pvpRepo,
		combatRepo: combatRepo,
		config:     config,
	}

	// Démarrer le matchmaking automatique
	service.StartMatchmakingRoutine()

	return service
}

// CreateChallenge crée un nouveau défi PvP
func (s *PvPService) CreateChallenge(challengerID uuid.UUID, req *models.CreateChallengeRequest) (*models.PvPChallenge, error) {
	// Validation de base
	if challengerID == req.ChallengedID {
		return nil, fmt.Errorf("cannot challenge yourself")
	}

	// Vérifier que le joueur n'a pas déjà un défi en attente avec cette personne
	existingChallenges, err := s.pvpRepo.GetChallengesByPlayer(challengerID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing challenges: %w", err)
	}

	for _, challenge := range existingChallenges {
		if challenge.ChallengedID == req.ChallengedID && challenge.Status == models.ChallengeStatusPending {
			return nil, fmt.Errorf("challenge already pending with this player")
		}
	}

	// Créer le défi
	challenge := &models.PvPChallenge{
		ID:            uuid.New(),
		ChallengerID:  challengerID,
		ChallengedID:  req.ChallengedID,
		ChallengeType: req.ChallengeType,
		Message:       &req.Message,
		Stakes: func() models.PvPStakes {
			if req.Stakes != nil {
				return *req.Stakes
			} else {
				return models.PvPStakes{}
			}
		}(),
		Status:    models.ChallengeStatusPending,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Duration(config.DefaultJWTExpiration) * time.Hour), // 24h d'expiration
	}

	if err := s.pvpRepo.CreateChallenge(challenge); err != nil {
		return nil, fmt.Errorf("failed to create challenge: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"challenge_id":  challenge.ID,
		"challenger_id": challengerID,
		"challenged_id": req.ChallengedID,
		"type":          req.ChallengeType,
	}).Info("PvP challenge created")

	return challenge, nil
}

// GetChallenge récupère un défi par son ID
func (s *PvPService) GetChallenge(id uuid.UUID) (*models.PvPChallenge, error) {
	return s.pvpRepo.GetChallengeByID(id)
}

// GetChallenges récupère les défis d'un joueur
func (s *PvPService) GetChallenges(req *models.GetChallengesRequest) ([]*models.PvPChallenge, error) {
	switch req.Type {
	case "sent":
		return s.pvpRepo.GetChallengesByPlayer(req.PlayerID)
	case "received":
		return s.pvpRepo.GetPendingChallenges(req.PlayerID)
	default:
		// Récupérer tous les défis (envoyés et reçus)
		sent, err := s.pvpRepo.GetChallengesByPlayer(req.PlayerID)
		if err != nil {
			return nil, err
		}

		received, err := s.pvpRepo.GetPendingChallenges(req.PlayerID)
		if err != nil {
			return nil, err
		}

		// Combiner et dédupliquer
		sent = append(sent, received...)
		return s.deduplicateChallenges(sent), nil
	}
}

// RespondToChallenge répond à un défi PvP
func (s *PvPService) RespondToChallenge(challengeID uuid.UUID, req *models.RespondToChallengeRequest) (*models.ChallengeResponse, error) {
	challenge, err := s.pvpRepo.GetChallengeByID(challengeID)
	if err != nil {
		return nil, fmt.Errorf("challenge not found: %w", err)
	}

	// Vérifier que le joueur peut répondre à ce défi
	if challenge.ChallengedID != req.PlayerID {
		return nil, fmt.Errorf("only the challenged player can respond")
	}

	if challenge.Status != models.ChallengeStatusPending {
		return nil, fmt.Errorf("challenge is no longer pending")
	}

	if time.Now().After(challenge.ExpiresAt) {
		return nil, fmt.Errorf("challenge has expired")
	}

	response := &models.ChallengeResponse{
		Success:   true,
		Challenge: challenge,
		Message:   req.Message,
	}

	if req.Accept {
		// Accepter le défi - créer un combat
		challenge.Status = models.ChallengeStatusAccepted
		now := time.Now()
		challenge.RespondedAt = &now

		// Créer un combat PvP
		combat, err := s.createPvPCombat(challenge)
		if err != nil {
			return nil, fmt.Errorf("failed to create PvP combat: %w", err)
		}

		challenge.CombatID = &combat.ID
		response.Match = nil // ou response.Match = ... si tu veux retourner le match

		logrus.WithFields(logrus.Fields{
			"challenge_id": challengeID,
			"combat_id":    combat.ID,
			"challenger":   challenge.ChallengerID,
			"challenged":   challenge.ChallengedID,
		}).Info("PvP challenge accepted and combat created")
	} else {
		// Refuser le défi
		challenge.Status = models.ChallengeStatusDeclined
		now := time.Now()
		challenge.RespondedAt = &now

		logrus.WithFields(logrus.Fields{
			"challenge_id": challengeID,
			"challenger":   challenge.ChallengerID,
			"challenged":   challenge.ChallengedID,
		}).Info("PvP challenge declined")
	}

	if err := s.pvpRepo.UpdateChallenge(challenge); err != nil {
		return nil, fmt.Errorf("failed to update challenge: %w", err)
	}

	return response, nil
}

// CancelChallenge annule un défi PvP
func (s *PvPService) CancelChallenge(challengeID, playerID uuid.UUID) error {
	challenge, err := s.pvpRepo.GetChallengeByID(challengeID)
	if err != nil {
		return fmt.Errorf("challenge not found: %w", err)
	}

	// Vérifier que le joueur peut annuler ce défi
	if challenge.ChallengerID != playerID {
		return fmt.Errorf("only the challenger can cancel the challenge")
	}

	if challenge.Status != models.ChallengeStatusPending {
		return fmt.Errorf("can only cancel pending challenges")
	}

	challenge.Status = models.ChallengeStatusCancelled

	if err := s.pvpRepo.UpdateChallenge(challenge); err != nil {
		return fmt.Errorf("failed to update challenge: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"challenge_id": challengeID,
		"challenger":   challenge.ChallengerID,
	}).Info("PvP challenge canceled")

	return nil
}

// GetRankings récupère les classements PvP
func (s *PvPService) GetRankings(req *models.GetRankingsRequest) (*models.RankingsResponse, error) {
	rankings, err := s.pvpRepo.GetTopPlayers(req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get rankings: %w", err)
	}

	// TODO: Enrichir avec les noms des joueurs depuis le service player
	for _, ranking := range rankings {
		ranking.PlayerName = fmt.Sprintf("Player-%s", ranking.PlayerID.String()[:8])
	}

	response := &models.RankingsResponse{
		Season:   req.Season,
		Rankings: rankings,
		Total:    len(rankings),
	}

	return response, nil
}

// GetPlayerStatistics récupère les statistiques PvP d'un joueur
func (s *PvPService) GetPlayerStatistics(playerID uuid.UUID, season string) (*models.PvPStatistics, error) {
	stats, err := s.pvpRepo.GetPvPStatistics(playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get PvP statistics: %w", err)
	}

	return stats, nil
}

// UpdatePlayerStatistics met à jour les statistiques PvP d'un joueur
func (s *PvPService) UpdatePlayerStatistics(playerID uuid.UUID, result *models.MatchResult) error {
	stats, err := s.pvpRepo.GetPvPStatistics(playerID)
	if err != nil {
		// Créer de nouvelles statistiques si elles n'existent pas
		stats = &models.PvPStatistics{
			PlayerID:      playerID,
			CurrentRating: config.DefaultPvPRating, // Rating de départ
		}
	}

	// Mettre à jour les statistiques en fonction du résultat
	switch result.ResultType {
	case models.ResultTypeVictory:
		stats.BattlesWon++
		if playerID == result.WinnerID {
			stats.CurrentRating = result.WinnerRating
		} else {
			stats.CurrentRating = result.LoserRating
		}
	case models.ResultTypeDefeat:
		stats.BattlesLost++
		if playerID == result.LoserID {
			stats.CurrentRating = result.LoserRating
		} else {
			stats.CurrentRating = result.WinnerRating
		}
	case models.ResultTypeDraw:
		stats.Draws++
	}
	stats.TotalMatches = stats.BattlesWon + stats.BattlesLost + stats.Draws
	if stats.TotalMatches > 0 {
		stats.WinRate = float64(stats.BattlesWon) / float64(stats.TotalMatches) * config.DefaultPercentageMultiplier
	}
	// Streaks (simplifié)
	if result.ResultType == models.ResultTypeVictory {
		stats.CurrentStreak++
		if stats.CurrentStreak > stats.BestStreak {
			stats.BestStreak = stats.CurrentStreak
		}
	} else if result.ResultType == models.ResultTypeDefeat {
		stats.CurrentStreak = 0
	}
	now := time.Now()
	stats.LastMatchAt = &now

	// Sauvegarder
	if err := s.pvpRepo.UpdatePvPStatistics(stats); err != nil {
		return fmt.Errorf("failed to update PvP statistics: %w", err)
	}

	return nil
}

// GetCurrentSeasonInfo récupère les informations de la saison actuelle
func (s *PvPService) GetCurrentSeasonInfo() (*models.SeasonInfo, error) {
	// TODO: Implémenter la gestion des saisons
	return &models.SeasonInfo{
		Name:      "Season 1",
		StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
		IsActive:  true,
	}, nil
}

// JoinQueue rejoint la file d'attente PvP
func (s *PvPService) JoinQueue(req *models.JoinQueueRequest) (*models.QueueResponse, error) {
	// Vérifier si le joueur est déjà en file d'attente
	existing, err := s.pvpRepo.GetQueueEntry(req.PlayerID)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("already in queue")
	}

	// Créer l'entrée de file d'attente
	entry := &models.PvPQueueEntry{
		PlayerID:    req.PlayerID,
		QueueType:   req.QueueType,
		JoinedAt:    time.Now(),
		Preferences: req.Preferences,
	}

	if err := s.pvpRepo.AddToQueue(entry); err != nil {
		return nil, fmt.Errorf("failed to join queue: %w", err)
	}

	// Estimer le temps d'attente
	queueCount, _ := s.getQueueCount(req.QueueType)
	estimatedWait := s.estimateWaitTime(queueCount)

	response := &models.QueueResponse{
		Success:       true,
		QueueEntry:    entry,
		Position:      queueCount + 1,
		EstimatedWait: estimatedWait,
		Message:       "Player joined PvP queue",
	}

	logrus.WithFields(logrus.Fields{
		"player_id":      req.PlayerID,
		"queue_type":     req.QueueType,
		"estimated_wait": estimatedWait,
	}).Info("Player joined PvP queue")

	return response, nil
}

// LeaveQueue quitte la file d'attente PvP
func (s *PvPService) LeaveQueue(playerID uuid.UUID) error {
	if err := s.pvpRepo.RemoveFromQueue(playerID); err != nil {
		return fmt.Errorf("failed to leave queue: %w", err)
	}

	logrus.WithField("player_id", playerID).Info("Player left PvP queue")
	return nil
}

// GetQueueStatus récupère le statut de la file d'attente
func (s *PvPService) GetQueueStatus(playerID uuid.UUID) (*models.QueueStatus, error) {
	entry, err := s.pvpRepo.GetQueueEntry(playerID)
	if err != nil {
		return &models.QueueStatus{
			InQueue: false,
		}, nil
	}

	if entry == nil {
		return &models.QueueStatus{
			InQueue: false,
		}, nil
	}

	// Calculer la position et le temps d'attente
	queueCount, _ := s.getQueueCount(entry.QueueType)
	waitTime := time.Since(entry.JoinedAt)
	estimatedRemaining := s.estimateWaitTime(queueCount) - waitTime

	if estimatedRemaining < 0 {
		estimatedRemaining = 0
	}

	return &models.QueueStatus{
		InQueue:       true,
		QueueEntry:    entry,
		Position:      queueCount, // Approximation
		EstimatedWait: estimatedRemaining,
		QueueSize:     queueCount,
	}, nil
}

// ProcessMatchmaking traite le matchmaking automatique
func (s *PvPService) ProcessMatchmaking() error {
	// Récupérer toutes les files d'attente
	queueTypes := []models.ChallengeType{"ranked", "casual"}

	for _, queueType := range queueTypes {
		entries, err := s.pvpRepo.GetQueueByType(queueType)
		if err != nil {
			logrus.WithError(err).Error("Failed to get queue entries")
			continue
		}

		// Essayer de créer des matches
		s.createMatches(entries)
	}

	return nil
}

// StartMatch démarre un match PvP
func (s *PvPService) StartMatch(player1ID, player2ID uuid.UUID, matchType models.ChallengeType) (*models.PvPMatch, error) {
	// Créer le combat
	combatReq := &models.CreateCombatRequest{
		CombatType:      models.CombatTypePvP,
		MaxParticipants: config.DefaultMaxParticipantsPvP,
		TurnTimeLimit:   config.DefaultTurnTimeLimitPvP,
		MaxDuration:     config.DefaultMaxDurationPvP,
	}

	combat := &models.CombatInstance{
		ID:              uuid.New(),
		CombatType:      combatReq.CombatType,
		Status:          "pending",
		MaxParticipants: combatReq.MaxParticipants,
		TurnTimeLimit:   combatReq.TurnTimeLimit,
		MaxDuration:     combatReq.MaxDuration,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	err := s.combatRepo.Create(combat)
	if err != nil {
		return nil, fmt.Errorf("failed to create combat: %w", err)
	}

	// Créer le match PvP
	match := &models.PvPMatch{
		ID:        uuid.New(),
		CombatID:  combat.ID,
		Players:   []*models.PlayerSummary{{ID: player1ID, Name: "Player 1"}, {ID: player2ID, Name: "Player 2"}},
		MatchType: matchType,
	}

	logrus.WithFields(logrus.Fields{
		"match_id":  match.ID,
		"combat_id": combat.ID,
		"player1":   player1ID,
		"player2":   player2ID,
		"type":      matchType,
	}).Info("PvP match started")

	return match, nil
}

// EndMatch termine un match PvP
func (s *PvPService) EndMatch(matchID uuid.UUID, result *models.MatchResult) error {
	// Mettre à jour les statistiques des deux joueurs
	if err := s.UpdatePlayerStatistics(result.WinnerID, &models.MatchResult{
		ResultType:   result.ResultType,
		WinnerID:     result.WinnerID,
		LoserID:      result.LoserID,
		WinnerRating: result.WinnerRating,
		LoserRating:  result.LoserRating,
	}); err != nil {
		logrus.WithError(err).Error("Failed to update player 1 statistics")
	}

	if err := s.UpdatePlayerStatistics(result.LoserID, &models.MatchResult{
		ResultType:   result.ResultType,
		WinnerID:     result.LoserID,
		LoserID:      result.WinnerID,
		WinnerRating: result.LoserRating,
		LoserRating:  result.WinnerRating,
	}); err != nil {
		logrus.WithError(err).Error("Failed to update player 2 statistics")
	}

	logrus.WithFields(logrus.Fields{
		"match_id":       matchID,
		"player1_result": result.ResultType,
		"player2_result": result.ResultType,
		"duration":       result.Duration,
	}).Info("PvP match ended")

	return nil
}

// CleanupExpiredChallenges nettoie les défis expirés
func (s *PvPService) CleanupExpiredChallenges() error {
	return s.pvpRepo.CleanupExpiredChallenges()
}

// CleanupOldQueue nettoie l'ancienne file d'attente
func (s *PvPService) CleanupOldQueue() error {
	return s.pvpRepo.CleanupOldQueue()
}

// StartCleanupRoutine démarre les routines de nettoyage
func (s *PvPService) StartCleanupRoutine() {
	// Nettoyage toutes les 30 minutes
	ticker := time.NewTicker(time.Duration(config.DefaultCombatTurnTimeout) * time.Minute)
	go func() {
		defer ticker.Stop()
		for range ticker.C {
			if err := s.CleanupExpiredChallenges(); err != nil {
				logrus.WithError(err).Error("Failed to cleanup expired challenges")
			}
			if err := s.CleanupOldQueue(); err != nil {
				logrus.WithError(err).Error("Failed to cleanup old queue")
			}
		}
	}()
}

// StartMatchmakingRoutine démarre le matchmaking automatique
func (s *PvPService) StartMatchmakingRoutine() {
	// Matchmaking toutes les 10 secondes
	s.queueTicker = time.NewTicker(time.Duration(config.DefaultQueueTicker2) * time.Second)
	go func() {
		defer s.queueTicker.Stop()
		for range s.queueTicker.C {
			if err := s.ProcessMatchmaking(); err != nil {
				logrus.WithError(err).Error("Failed to process matchmaking")
			}
		}
	}()
}

// Méthodes utilitaires privées

func (s *PvPService) createPvPCombat(_ *models.PvPChallenge) (*models.CombatInstance, error) {
	combat := &models.CombatInstance{
		ID:              uuid.New(),
		CombatType:      models.CombatTypePvP,
		Status:          "pending",
		MaxParticipants: config.DefaultMaxParticipantsPvP,
		CurrentTurn:     0,
		TurnTimeLimit:   config.DefaultTurnTimeLimitPvP,
		MaxDuration:     config.DefaultMaxDurationPvP,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.combatRepo.Create(combat); err != nil {
		return nil, err
	}

	return combat, nil
}

func (s *PvPService) deduplicateChallenges(challenges []*models.PvPChallenge) []*models.PvPChallenge {
	seen := make(map[uuid.UUID]bool)
	result := []*models.PvPChallenge{}

	for _, challenge := range challenges {
		if !seen[challenge.ID] {
			seen[challenge.ID] = true
			result = append(result, challenge)
		}
	}

	return result
}

func (s *PvPService) getQueueCount(queueType models.ChallengeType) (int, error) {
	entries, err := s.pvpRepo.GetQueueByType(queueType)
	if err != nil {
		return 0, err
	}
	return len(entries), nil
}

func (s *PvPService) estimateWaitTime(queueSize int) time.Duration {
	// Estimation simple basée sur la taille de la file d'attente
	baseWait := time.Duration(config.DefaultCombatTurnTimeout) * time.Second
	return baseWait + time.Duration(queueSize)*10*time.Second
}

func (s *PvPService) createMatches(entries []*models.PvPQueueEntry) {
	// Algorithme de matchmaking simple
	for i := 0; i < len(entries)-1; i += 2 {
		player1 := entries[i]
		player2 := entries[i+1]

		// Vérifier si les joueurs sont compatibles
		if s.arePlayersCompatible(player1, player2) {
			// Créer le match
			match, err := s.StartMatch(player1.PlayerID, player2.PlayerID, player1.QueueType)
			if err != nil {
				logrus.WithError(err).Error("Failed to start match")
				continue
			}

			// Retirer les joueurs de la file d'attente
			if err := s.pvpRepo.RemoveFromQueue(player1.PlayerID); err != nil {
				logrus.WithError(err).Error("Failed to remove player1 from queue")
			}
			if err := s.pvpRepo.RemoveFromQueue(player2.PlayerID); err != nil {
				logrus.WithError(err).Error("Failed to remove player2 from queue")
			}

			logrus.WithFields(logrus.Fields{
				"match_id": match.ID,
				"player1":  player1.PlayerID,
				"player2":  player2.PlayerID,
			}).Info("Match created from queue")
		}
	}
}

func (s *PvPService) arePlayersCompatible(p1, p2 *models.PvPQueueEntry) bool {
	// Vérifier la différence de rating
	ratingDiff := abs(p1.Rating - p2.Rating)
	maxRatingDiff := 200 // Maximum 200 points de différence

	// Augmenter la tolérance avec le temps d'attente
	waitTime1 := time.Since(p1.JoinedAt)
	waitTime2 := time.Since(p2.JoinedAt)
	avgWaitTime := (waitTime1 + waitTime2) / config.DefaultMaxParticipantsPvP

	// Augmenter la tolérance de 50 points par minute d'attente
	maxRatingDiff += int(avgWaitTime.Minutes()) * config.DefaultRatingMultiplier

	return ratingDiff <= maxRatingDiff
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
