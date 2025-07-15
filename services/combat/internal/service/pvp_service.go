package service

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"combat/internal/config"
	"combat/internal/models"
	"combat/internal/repository"
	"combat/internal/external"
)

// PvPServiceInterface définit les méthodes du service PvP
type PvPServiceInterface interface {
	CreateChallenge(challengerID, targetID uuid.UUID, matchType string) (*models.PvPChallenge, error)
	AcceptChallenge(challengeID, accepterID uuid.UUID) (*models.CombatSession, error)
	DeclineChallenge(challengeID, declinerID uuid.UUID) error
	CancelChallenge(challengeID, cancellerID uuid.UUID) error
	GetPendingChallenges(playerID uuid.UUID) ([]*models.PvPChallenge, error)
	
	FindMatch(playerID uuid.UUID, matchType string) (*models.CombatSession, error)
	LeaveMatchmaking(playerID uuid.UUID) error
	GetMatchmakingStatus(playerID uuid.UUID) (*models.MatchmakingStatus, error)
	
	ProcessMatchResult(sessionID uuid.UUID, results *models.MatchResult) error
	GetRankings(season string, limit int, offset int) ([]*models.PvPRanking, error)
	GetPlayerRanking(playerID uuid.UUID, season string) (*models.PvPRanking, error)
	UpdatePlayerStats(playerID uuid.UUID, matchResult *models.MatchResult) error
	
	StartNewSeason() error
	GetSeasonInfo(season string) (*models.SeasonInfo, error)
	CalculateRewards(playerID uuid.UUID, season string) (*models.SeasonRewards, error)
}

// PvPService implémente l'interface PvPServiceInterface
type PvPService struct {
	config        *config.Config
	pvpRepo       repository.PvPRepositoryInterface
	combatService CombatServiceInterface
	playerClient  external.PlayerClientInterface
	worldClient   external.WorldClientInterface
	matchmaker    *Matchmaker
}

// NewPvPService crée une nouvelle instance du service PvP
func NewPvPService(
	cfg *config.Config,
	pvpRepo repository.PvPRepositoryInterface,
	combatService CombatServiceInterface,
	playerClient external.PlayerClientInterface,
	worldClient external.WorldClientInterface,
) PvPServiceInterface {
	matchmaker := NewMatchmaker(cfg)
	
	return &PvPService{
		config:        cfg,
		pvpRepo:       pvpRepo,
		combatService: combatService,
		playerClient:  playerClient,
		worldClient:   worldClient,
		matchmaker:    matchmaker,
	}
}

// CreateChallenge crée un défi PvP direct entre deux joueurs
func (s *PvPService) CreateChallenge(challengerID, targetID uuid.UUID, matchType string) (*models.PvPChallenge, error) {
	// Vérifier que les joueurs existent et sont connectés
	challenger, err := s.playerClient.GetPlayerInfo(challengerID)
	if err != nil {
		return nil, fmt.Errorf("challenger not found: %w", err)
	}
	
	target, err := s.playerClient.GetPlayerInfo(targetID)
	if err != nil {
		return nil, fmt.Errorf("target not found: %w", err)
	}
	
	// Vérifier que le joueur ne se défie pas lui-même
	if challengerID == targetID {
		return nil, fmt.Errorf("cannot challenge yourself")
	}
	
	// Vérifier qu'il n'y a pas déjà un défi en cours
	existingChallenges, err := s.pvpRepo.GetActiveChallenges(challengerID, targetID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing challenges: %w", err)
	}
	
	if len(existingChallenges) > 0 {
		return nil, fmt.Errorf("challenge already exists between these players")
	}
	
	// Vérifier les niveaux pour certains types de match
	if err := s.validateLevelRequirement(challenger, target, matchType); err != nil {
		return nil, fmt.Errorf("level requirement not met: %w", err)
	}
	
	// Créer le défi
	challenge := &models.PvPChallenge{
		ID:           uuid.New(),
		ChallengerID: challengerID,
		TargetID:     targetID,
		MatchType:    matchType,
		Status:       "pending",
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(5 * time.Minute), // Expire après 5 minutes
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

// AcceptChallenge accepte un défi et crée une session de combat
func (s *PvPService) AcceptChallenge(challengeID, accepterID uuid.UUID) (*models.CombatSession, error) {
	// Récupérer le défi
	challenge, err := s.pvpRepo.GetChallenge(challengeID)
	if err != nil {
		return nil, fmt.Errorf("challenge not found: %w", err)
	}
	
	// Vérifier que c'est bien la cible qui accepte
	if challenge.TargetID != accepterID {
		return nil, fmt.Errorf("only the target can accept this challenge")
	}
	
	// Vérifier que le défi est toujours valide
	if challenge.Status != "pending" {
		return nil, fmt.Errorf("challenge is no longer pending")
	}
	
	if time.Now().After(challenge.ExpiresAt) {
		return nil, fmt.Errorf("challenge has expired")
	}
	
	// Marquer le défi comme accepté
	challenge.Status = "accepted"
	challenge.AcceptedAt = time.Now()
	
	if err := s.pvpRepo.UpdateChallenge(challenge); err != nil {
		return nil, fmt.Errorf("failed to update challenge: %w", err)
	}
	
	// Créer la session de combat PvP
	session, err := s.createPvPSession(challenge)
	if err != nil {
		return nil, fmt.Errorf("failed to create PvP session: %w", err)
	}
	
	logrus.WithFields(logrus.Fields{
		"challenge_id": challengeID,
		"session_id":   session.ID,
		"challenger":   challenge.ChallengerID,
		"target":       challenge.TargetID,
	}).Info("PvP challenge accepted")
	
	return session, nil
}

// DeclineChallenge refuse un défi
func (s *PvPService) DeclineChallenge(challengeID, declinerID uuid.UUID) error {
	challenge, err := s.pvpRepo.GetChallenge(challengeID)
	if err != nil {
		return fmt.Errorf("challenge not found: %w", err)
	}
	
	if challenge.TargetID != declinerID {
		return fmt.Errorf("only the target can decline this challenge")
	}
	
	if challenge.Status != "pending" {
		return fmt.Errorf("challenge is no longer pending")
	}
	
	challenge.Status = "declined"
	challenge.DeclinedAt = time.Now()
	
	if err := s.pvpRepo.UpdateChallenge(challenge); err != nil {
		return fmt.Errorf("failed to update challenge: %w", err)
	}
	
	logrus.WithFields(logrus.Fields{
		"challenge_id": challengeID,
		"decliner_id":  declinerID,
	}).Info("PvP challenge declined")
	
	return nil
}

// CancelChallenge annule un défi
func (s *PvPService) CancelChallenge(challengeID, cancellerID uuid.UUID) error {
	challenge, err := s.pvpRepo.GetChallenge(challengeID)
	if err != nil {
		return fmt.Errorf("challenge not found: %w", err)
	}
	
	if challenge.ChallengerID != cancellerID {
		return fmt.Errorf("only the challenger can cancel this challenge")
	}
	
	if challenge.Status != "pending" {
		return fmt.Errorf("challenge is no longer pending")
	}
	
	challenge.Status = "cancelled"
	challenge.CancelledAt = time.Now()
	
	if err := s.pvpRepo.UpdateChallenge(challenge); err != nil {
		return fmt.Errorf("failed to update challenge: %w", err)
	}
	
	return nil
}

// GetPendingChallenges récupère les défis en attente pour un joueur
func (s *PvPService) GetPendingChallenges(playerID uuid.UUID) ([]*models.PvPChallenge, error) {
	challenges, err := s.pvpRepo.GetPlayerChallenges(playerID, "pending")
	if err != nil {
		return nil, fmt.Errorf("failed to get pending challenges: %w", err)
	}
	
	// Filtrer les défis expirés
	now := time.Now()
	validChallenges := make([]*models.PvPChallenge, 0)
	
	for _, challenge := range challenges {
		if now.Before(challenge.ExpiresAt) {
			validChallenges = append(validChallenges, challenge)
		} else {
			// Marquer comme expiré
			challenge.Status = "expired"
			s.pvpRepo.UpdateChallenge(challenge)
		}
	}
	
	return validChallenges, nil
}

// FindMatch trouve un match via le système de matchmaking
func (s *PvPService) FindMatch(playerID uuid.UUID, matchType string) (*models.CombatSession, error) {
	// Récupérer les informations du joueur
	player, err := s.playerClient.GetPlayerInfo(playerID)
	if err != nil {
		return nil, fmt.Errorf("player not found: %w", err)
	}
	
	// Récupérer le ranking pour le matchmaking
	ranking, err := s.GetPlayerRanking(playerID, s.getCurrentSeason())
	if err != nil {
		// Créer un ranking par défaut si le joueur n'en a pas
		ranking = &models.PvPRanking{
			PlayerID: playerID,
			Season:   s.getCurrentSeason(),
			Rating:   1000, // Rating de départ
			Tier:     "Bronze",
			Division: 5,
		}
	}
	
	// Ajouter le joueur à la file d'attente
	matchRequest := &models.MatchmakingRequest{
		PlayerID:     playerID,
		MatchType:    matchType,
		Rating:       ranking.Rating,
		Level:        player.Level,
		QueuedAt:     time.Now(),
		Preferences:  map[string]interface{}{},
	}
	
	// Chercher un match
	match, err := s.matchmaker.FindMatch(matchRequest)
	if err != nil {
		return nil, fmt.Errorf("matchmaking failed: %w", err)
	}
	
	if match == nil {
		// Aucun match trouvé, ajouter à la file
		s.matchmaker.AddToQueue(matchRequest)
		return nil, nil
	}
	
	// Match trouvé, créer la session
	session, err := s.createMatchmadeSession(match)
	if err != nil {
		return nil, fmt.Errorf("failed to create matched session: %w", err)
	}
	
	return session, nil
}

// LeaveMatchmaking retire un joueur de la file d'attente
func (s *PvPService) LeaveMatchmaking(playerID uuid.UUID) error {
	s.matchmaker.RemoveFromQueue(playerID)
	
	logrus.WithField("player_id", playerID).Info("Player left matchmaking queue")
	return nil
}

// GetMatchmakingStatus récupère le statut du matchmaking pour un joueur
func (s *PvPService) GetMatchmakingStatus(playerID uuid.UUID) (*models.MatchmakingStatus, error) {
	return s.matchmaker.GetPlayerStatus(playerID), nil
}

// ProcessMatchResult traite le résultat d'un match PvP
func (s *PvPService) ProcessMatchResult(sessionID uuid.UUID, results *models.MatchResult) error {
	// Créer l'enregistrement du match
	pvpMatch := &models.PvPMatch{
		ID:        uuid.New(),
		SessionID: sessionID,
		MatchType: results.MatchType,
		Season:    s.getCurrentSeason(),
		StartedAt: results.StartedAt,
		EndedAt:   results.EndedAt,
		Duration:  results.Duration,
		WinnerID:  results.WinnerID,
		LoserID:   results.LoserID,
		Result:    results.Result,
		CreatedAt: time.Now(),
	}
	
	if err := s.pvpRepo.CreateMatch(pvpMatch); err != nil {
		return fmt.Errorf("failed to create PvP match record: %w", err)
	}
	
	// Mettre à jour les classements
	if err := s.updateRatings(results); err != nil {
		logrus.WithError(err).Error("Failed to update player ratings")
	}
	
	// Mettre à jour les statistiques
	if err := s.UpdatePlayerStats(results.WinnerID, results); err != nil {
		logrus.WithError(err).Error("Failed to update winner stats")
	}
	
	if err := s.UpdatePlayerStats(results.LoserID, results); err != nil {
		logrus.WithError(err).Error("Failed to update loser stats")
	}
	
	logrus.WithFields(logrus.Fields{
		"session_id": sessionID,
		"winner_id":  results.WinnerID,
		"loser_id":   results.LoserID,
		"duration":   results.Duration,
	}).Info("PvP match result processed")
	
	return nil
}

// GetRankings récupère les classements PvP
func (s *PvPService) GetRankings(season string, limit int, offset int) ([]*models.PvPRanking, error) {
	if season == "" {
		season = s.getCurrentSeason()
	}
	
	rankings, err := s.pvpRepo.GetRankings(season, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get rankings: %w", err)
	}
	
	return rankings, nil
}

// GetPlayerRanking récupère le classement d'un joueur
func (s *PvPService) GetPlayerRanking(playerID uuid.UUID, season string) (*models.PvPRanking, error) {
	if season == "" {
		season = s.getCurrentSeason()
	}
	
	ranking, err := s.pvpRepo.GetPlayerRanking(playerID, season)
	if err != nil {
		return nil, fmt.Errorf("failed to get player ranking: %w", err)
	}
	
	return ranking, nil
}

// UpdatePlayerStats met à jour les statistiques d'un joueur
func (s *PvPService) UpdatePlayerStats(playerID uuid.UUID, matchResult *models.MatchResult) error {
	stats, err := s.pvpRepo.GetPlayerStats(playerID, s.getCurrentSeason())
	if err != nil {
		// Créer de nouvelles stats si elles n'existent pas
		stats = &models.PvPStats{
			PlayerID: playerID,
			Season:   s.getCurrentSeason(),
		}
	}
	
	// Mettre à jour les statistiques
	stats.TotalMatches++
	
	if playerID == matchResult.WinnerID {
		stats.Wins++
		stats.WinStreak++
		if stats.WinStreak > stats.BestWinStreak {
			stats.BestWinStreak = stats.WinStreak
		}
		stats.LossStreak = 0
	} else {
		stats.Losses++
		stats.LossStreak++
		stats.WinStreak = 0
	}
	
	// Calculer le taux de victoire
	if stats.TotalMatches > 0 {
		stats.WinRate = float64(stats.Wins) / float64(stats.TotalMatches)
	}
	
	stats.UpdatedAt = time.Now()
	
	return s.pvpRepo.UpdatePlayerStats(stats)
}

// StartNewSeason démarre une nouvelle saison
func (s *PvPService) StartNewSeason() error {
	// Archiver la saison actuelle
	currentSeason := s.getCurrentSeason()
	
	// Calculer les récompenses pour tous les joueurs
	if err := s.distributeSeasonRewards(currentSeason); err != nil {
		logrus.WithError(err).Error("Failed to distribute season rewards")
	}
	
	// Créer la nouvelle saison
	newSeason := s.generateNewSeasonName()
	
	// Reset partiel des ratings (soft reset)
	if err := s.performSeasonReset(newSeason); err != nil {
		return fmt.Errorf("failed to perform season reset: %w", err)
	}
	
	logrus.WithFields(logrus.Fields{
		"old_season": currentSeason,
		"new_season": newSeason,
	}).Info("New PvP season started")
	
	return nil
}

// GetSeasonInfo récupère les informations d'une saison
func (s *PvPService) GetSeasonInfo(season string) (*models.SeasonInfo, error) {
	if season == "" {
		season = s.getCurrentSeason()
	}
	
	info, err := s.pvpRepo.GetSeasonInfo(season)
	if err != nil {
		return nil, fmt.Errorf("failed to get season info: %w", err)
	}
	
	return info, nil
}

// CalculateRewards calcule les récompenses de fin de saison
func (s *PvPService) CalculateRewards(playerID uuid.UUID, season string) (*models.SeasonRewards, error) {
	ranking, err := s.GetPlayerRanking(playerID, season)
	if err != nil {
		return nil, fmt.Errorf("failed to get player ranking: %w", err)
	}
	
	rewards := &models.SeasonRewards{
		PlayerID: playerID,
		Season:   season,
		Tier:     ranking.Tier,
		Division: ranking.Division,
		Rating:   ranking.Rating,
	}
	
	// Calculer les récompenses selon le tier
	rewards.Gold = s.calculateGoldReward(ranking.Tier, ranking.Division)
	rewards.Items = s.calculateItemRewards(ranking.Tier, ranking.Division)
	rewards.Title = s.calculateTitleReward(ranking.Tier, ranking.Division)
	
	return rewards, nil
}

// Méthodes privées

// validateLevelRequirement vérifie les exigences de niveau
func (s *PvPService) validateLevelRequirement(challenger, target *external.PlayerInfo, matchType string) error {
	levelRequirements := map[string]int{
		"duel":   10,
		"arena":  15,
		"ranked": 20,
	}
	
	requiredLevel, exists := levelRequirements[matchType]
	if !exists {
		return nil // Pas d'exigence pour ce type
	}
	
	if challenger.Level < requiredLevel {
		return fmt.Errorf("challenger level %d is below required level %d", challenger.Level, requiredLevel)
	}
	
	if target.Level < requiredLevel {
		return fmt.Errorf("target level %d is below required level %d", target.Level, requiredLevel)
	}
	
	return nil
}

// createPvPSession crée une session de combat PvP
func (s *PvPService) createPvPSession(challenge *models.PvPChallenge) (*models.CombatSession, error) {
	// Déterminer la zone PvP appropriée
	zoneID := s.selectPvPZone(challenge.MatchType)
	
	// Créer la requête de combat
	combatReq := models.StartCombatRequest{
		Type:            "pvp",
		ZoneID:          zoneID,
		MaxParticipants: 2,
		IsPrivate:       true,
		Rules: map[string]interface{}{
			"match_type":    challenge.MatchType,
			"challenge_id":  challenge.ID,
			"time_limit":    600, // 10 minutes
		},
	}
	
	// Créer la session via le service de combat
	session, err := s.combatService.CreateCombatSession(challenge.ChallengerID, combatReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create combat session: %w", err)
	}
	
	// Ajouter les participants
	challengerParticipant := &models.CombatParticipant{
		SessionID:   session.ID,
		CharacterID: challenge.ChallengerID,
		Team:        1,
		Status:      "alive",
		JoinedAt:    time.Now(),
	}
	
	targetParticipant := &models.CombatParticipant{
		SessionID:   session.ID,
		CharacterID: challenge.TargetID,
		Team:        2,
		Status:      "alive",
		JoinedAt:    time.Now(),
	}
	
	// TODO: Récupérer les stats des joueurs et les initialiser
	// Pour l'instant, utilisation de stats par défaut
	
	if err := s.combatService.AddParticipant(challengerParticipant); err != nil {
		return nil, fmt.Errorf("failed to add challenger: %w", err)
	}
	
	if err := s.combatService.AddParticipant(targetParticipant); err != nil {
		return nil, fmt.Errorf("failed to add target: %w", err)
	}
	
	return session, nil
}

// createMatchmadeSession crée une session pour un match trouvé par matchmaking
func (s *PvPService) createMatchmadeSession(match *models.Match) (*models.CombatSession, error) {
	zoneID := s.selectPvPZone(match.MatchType)
	
	combatReq := models.StartCombatRequest{
		Type:            "pvp",
		ZoneID:          zoneID,
		MaxParticipants: len(match.Players),
		IsPrivate:       false,
		Rules: map[string]interface{}{
			"match_type":  match.MatchType,
			"matchmade":   true,
			"time_limit":  900, // 15 minutes
		},
	}
	
	session, err := s.combatService.CreateCombatSession(match.Players[0], combatReq)
	if err != nil {
		return nil, err
	}
	
	// Ajouter tous les joueurs
	for i, playerID := range match.Players {
		participant := &models.CombatParticipant{
			SessionID:   session.ID,
			CharacterID: playerID,
			Team:        (i % 2) + 1, // Alterner les équipes
			Status:      "alive",
			JoinedAt:    time.Now(),
		}
		
		if err := s.combatService.AddParticipant(participant); err != nil {
			logrus.WithError(err).Error("Failed to add matchmade participant")
		}
	}
	
	return session, nil
}

// updateRatings met à jour les ratings ELO des joueurs
func (s *PvPService) updateRatings(results *models.MatchResult) error {
	season := s.getCurrentSeason()
	
	// Récupérer les rankings actuels
	winnerRanking, err := s.GetPlayerRanking(results.WinnerID, season)
	if err != nil {
		winnerRanking = s.createDefaultRanking(results.WinnerID, season)
	}
	
	loserRanking, err := s.GetPlayerRanking(results.LoserID, season)
	if err != nil {
		loserRanking = s.createDefaultRanking(results.LoserID, season)
	}
	
	// Calculer les nouveaux ratings avec l'algorithme ELO
	newWinnerRating, newLoserRating := s.calculateELOChanges(
		winnerRanking.Rating,
		loserRanking.Rating,
		true, // winner won
	)
	
	// Mettre à jour les rankings
	winnerRanking.Rating = newWinnerRating
	winnerRanking.Wins++
	winnerRanking.TotalMatches++
	winnerRanking.UpdateTierAndDivision()
	winnerRanking.UpdatedAt = time.Now()
	
	loserRanking.Rating = newLoserRating
	loserRanking.Losses++
	loserRanking.TotalMatches++
	loserRanking.UpdateTierAndDivision()
	loserRanking.UpdatedAt = time.Now()
	
	// Sauvegarder
	if err := s.pvpRepo.UpdateRanking(winnerRanking); err != nil {
		return fmt.Errorf("failed to update winner ranking: %w", err)
	}
	
	if err := s.pvpRepo.UpdateRanking(loserRanking); err != nil {
		return fmt.Errorf("failed to update loser ranking: %w", err)
	}
	
	logrus.WithFields(logrus.Fields{
		"winner_id":         results.WinnerID,
		"loser_id":          results.LoserID,
		"winner_old_rating": winnerRanking.Rating - (newWinnerRating - winnerRanking.Rating),
		"winner_new_rating": newWinnerRating,
		"loser_old_rating":  loserRanking.Rating - (newLoserRating - loserRanking.Rating),
		"loser_new_rating":  newLoserRating,
	}).Info("Player ratings updated")
	
	return nil
}

// calculateELOChanges calcule les changements de rating ELO
func (s *PvPService) calculateELOChanges(winnerRating, loserRating int, winnerWon bool) (int, int) {
	// Facteur K (influence du match sur le rating)
	kFactor := 32.0
	
	// Calcul des probabilités de victoire
	expectedWinner := 1.0 / (1.0 + math.Pow(10, float64(loserRating-winnerRating)/400.0))
	expectedLoser := 1.0 / (1.0 + math.Pow(10, float64(winnerRating-loserRating)/400.0))
	
	// Calcul des nouveaux ratings
	var actualWinner, actualLoser float64
	if winnerWon {
		actualWinner = 1.0
		actualLoser = 0.0
	} else {
		actualWinner = 0.0
		actualLoser = 1.0
	}
	
	newWinnerRating := float64(winnerRating) + kFactor*(actualWinner-expectedWinner)
	newLoserRating := float64(loserRating) + kFactor*(actualLoser-expectedLoser)
	
	// S'assurer que les ratings ne descendent jamais en dessous de 0
	if newWinnerRating < 0 {
		newWinnerRating = 0
	}
	if newLoserRating < 0 {
		newLoserRating = 0
	}
	
	return int(newWinnerRating), int(newLoserRating)
}

// createDefaultRanking crée un ranking par défaut pour un nouveau joueur
func (s *PvPService) createDefaultRanking(playerID uuid.UUID, season string) *models.PvPRanking {
	ranking := &models.PvPRanking{
		ID:       uuid.New(),
		PlayerID: playerID,
		Season:   season,
		Rating:   1000, // Rating de départ
		Tier:     "Bronze",
		Division: 5,
		Wins:     0,
		Losses:   0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	// Sauvegarder le nouveau ranking
	if err := s.pvpRepo.CreateRanking(ranking); err != nil {
		logrus.WithError(err).Error("Failed to create default ranking")
	}
	
	return ranking
}

// selectPvPZone sélectionne une zone PvP appropriée
func (s *PvPService) selectPvPZone(matchType string) string {
	zoneMap := map[string]string{
		"duel":   "pvp_arena_small",
		"arena":  "pvp_arena_medium",
		"ranked": "pvp_arena_ranked",
	}
	
	if zone, exists := zoneMap[matchType]; exists {
		return zone
	}
	
	return "pvp_arena_default"
}

// getCurrentSeason retourne la saison actuelle
func (s *PvPService) getCurrentSeason() string {
	// Format: "Season-YYYY-Q" (ex: "Season-2024-1")
	now := time.Now()
	quarter := (now.Month()-1)/3 + 1
	return fmt.Sprintf("Season-%d-%d", now.Year(), quarter)
}

// generateNewSeasonName génère le nom de la nouvelle saison
func (s *PvPService) generateNewSeasonName() string {
	now := time.Now()
	quarter := (now.Month()-1)/3 + 1
	
	// Si on est dans le dernier trimestre, passer à l'année suivante
	if quarter == 4 {
		return fmt.Sprintf("Season-%d-1", now.Year()+1)
	}
	
	return fmt.Sprintf("Season-%d-%d", now.Year(), quarter+1)
}

// distributeSeasonRewards distribue les récompenses de fin de saison
func (s *PvPService) distributeSeasonRewards(season string) error {
	// Récupérer tous les joueurs qui ont participé à la saison
	players, err := s.pvpRepo.GetSeasonParticipants(season)
	if err != nil {
		return fmt.Errorf("failed to get season participants: %w", err)
	}
	
	for _, playerID := range players {
		rewards, err := s.CalculateRewards(playerID, season)
		if err != nil {
			logrus.WithError(err).WithField("player_id", playerID).Error("Failed to calculate rewards")
			continue
		}
		
		// Distribuer les récompenses (via le service Player ou Inventory)
		if err := s.distributeRewardsToPlayer(playerID, rewards); err != nil {
			logrus.WithError(err).WithField("player_id", playerID).Error("Failed to distribute rewards")
		}
	}
	
	return nil
}

// performSeasonReset effectue le reset de saison
func (s *PvPService) performSeasonReset(newSeason string) error {
	// Soft reset: réduire tous les ratings de 10%
	return s.pvpRepo.PerformSeasonReset(newSeason, 0.9) // 90% du rating précédent
}

// calculateGoldReward calcule la récompense en or
func (s *PvPService) calculateGoldReward(tier string, division int) int {
	baseRewards := map[string]int{
		"Bronze":   100,
		"Silver":   250,
		"Gold":     500,
		"Platinum": 1000,
		"Diamond":  2000,
		"Master":   3000,
		"Grandmaster": 5000,
	}
	
	baseGold, exists := baseRewards[tier]
	if !exists {
		baseGold = 50
	}
	
	// Bonus selon la division (division 1 = meilleure)
	divisionBonus := (6 - division) * 50
	
	return baseGold + divisionBonus
}

// calculateItemRewards calcule les récompenses d'objets
func (s *PvPService) calculateItemRewards(tier string, division int) []string {
	items := make([]string, 0)
	
	// Ajouter des objets selon le tier
	switch tier {
	case "Bronze":
		items = append(items, "bronze_badge")
	case "Silver":
		items = append(items, "silver_badge", "pvp_potion_small")
	case "Gold":
		items = append(items, "gold_badge", "pvp_potion_medium", "pvp_gear_token")
	case "Platinum":
		items = append(items, "platinum_badge", "pvp_potion_large", "pvp_gear_token", "elite_pvp_mount")
	case "Diamond":
		items = append(items, "diamond_badge", "legendary_pvp_weapon", "pvp_gear_set", "diamond_mount")
	case "Master":
		items = append(items, "master_badge", "artifact_pvp_weapon", "master_gear_set", "master_mount", "master_title")
	case "Grandmaster":
		items = append(items, "grandmaster_badge", "mythic_pvp_weapon", "grandmaster_gear_set", "legendary_mount", "grandmaster_title", "season_trophy")
	}
	
	// Bonus division 1
	if division == 1 {
		items = append(items, "division_1_bonus_chest")
	}
	
	return items
}

// calculateTitleReward calcule le titre de récompense
func (s *PvPService) calculateTitleReward(tier string, division int) string {
	if tier == "Grandmaster" && division == 1 {
		return "Arena Grandmaster"
	} else if tier == "Master" {
		return "Arena Master"
	} else if tier == "Diamond" && division <= 2 {
		return "Diamond Gladiator"
	} else if tier == "Platinum" && division == 1 {
		return "Platinum Champion"
	}
	
	return "" // Pas de titre
}

// distributeRewardsToPlayer distribue les récompenses à un joueur
func (s *PvPService) distributeRewardsToPlayer(playerID uuid.UUID, rewards *models.SeasonRewards) error {
	// TODO: Intégration avec le service Player/Inventory pour donner les récompenses
	// Pour l'instant, on log simplement
	
	logrus.WithFields(logrus.Fields{
		"player_id": playerID,
		"season":    rewards.Season,
		"tier":      rewards.Tier,
		"division":  rewards.Division,
		"gold":      rewards.Gold,
		"items":     rewards.Items,
		"title":     rewards.Title,
	}).Info("Season rewards distributed")
	
	return nil
}

// Matchmaker gère le système de matchmaking
type Matchmaker struct {
	config      *config.Config
	queue       map[string][]*models.MatchmakingRequest // queue par type de match
	playerQueue map[uuid.UUID]*models.MatchmakingRequest // track des joueurs en queue
}

// NewMatchmaker crée un nouveau matchmaker
func NewMatchmaker(cfg *config.Config) *Matchmaker {
	return &Matchmaker{
		config:      cfg,
		queue:       make(map[string][]*models.MatchmakingRequest),
		playerQueue: make(map[uuid.UUID]*models.MatchmakingRequest),
	}
}

// FindMatch cherche un match pour un joueur
func (m *Matchmaker) FindMatch(request *models.MatchmakingRequest) (*models.Match, error) {
	matchType := request.MatchType
	
	// Chercher dans la queue existante
	queue, exists := m.queue[matchType]
	if !exists || len(queue) == 0 {
		return nil, nil // Pas de match possible
	}
	
	// Trouver le meilleur adversaire
	bestMatch := m.findBestOpponent(request, queue)
	if bestMatch == nil {
		return nil, nil
	}
	
	// Créer le match
	match := &models.Match{
		ID:        uuid.New(),
		MatchType: matchType,
		Players:   []uuid.UUID{request.PlayerID, bestMatch.PlayerID},
		CreatedAt: time.Now(),
		Status:    "matched",
	}
	
	// Retirer les joueurs de la queue
	m.removeFromQueue(bestMatch.PlayerID, matchType)
	
	logrus.WithFields(logrus.Fields{
		"match_id":   match.ID,
		"player_1":   request.PlayerID,
		"player_2":   bestMatch.PlayerID,
		"match_type": matchType,
		"rating_diff": abs(request.Rating - bestMatch.Rating),
	}).Info("Match found")
	
	return match, nil
}

// AddToQueue ajoute un joueur à la queue
func (m *Matchmaker) AddToQueue(request *models.MatchmakingRequest) {
	matchType := request.MatchType
	
	// Retirer le joueur s'il est déjà en queue
	if existing, exists := m.playerQueue[request.PlayerID]; exists {
		m.removeFromQueue(request.PlayerID, existing.MatchType)
	}
	
	// Ajouter à la nouvelle queue
	if _, exists := m.queue[matchType]; !exists {
		m.queue[matchType] = make([]*models.MatchmakingRequest, 0)
	}
	
	m.queue[matchType] = append(m.queue[matchType], request)
	m.playerQueue[request.PlayerID] = request
	
	logrus.WithFields(logrus.Fields{
		"player_id":   request.PlayerID,
		"match_type":  matchType,
		"rating":      request.Rating,
		"queue_size":  len(m.queue[matchType]),
	}).Info("Player added to matchmaking queue")
}

// RemoveFromQueue retire un joueur de la queue
func (m *Matchmaker) RemoveFromQueue(playerID uuid.UUID) {
	if request, exists := m.playerQueue[playerID]; exists {
		m.removeFromQueue(playerID, request.MatchType)
	}
}

// GetPlayerStatus récupère le statut d'un joueur dans le matchmaking
func (m *Matchmaker) GetPlayerStatus(playerID uuid.UUID) *models.MatchmakingStatus {
	if request, exists := m.playerQueue[playerID]; exists {
		return &models.MatchmakingStatus{
			InQueue:      true,
			MatchType:    request.MatchType,
			QueueTime:    time.Since(request.QueuedAt),
			EstimatedWait: m.estimateWaitTime(request),
			QueuePosition: m.getQueuePosition(playerID, request.MatchType),
		}
	}
	
	return &models.MatchmakingStatus{
		InQueue: false,
	}
}

// findBestOpponent trouve le meilleur adversaire dans la queue
func (m *Matchmaker) findBestOpponent(request *models.MatchmakingRequest, queue []*models.MatchmakingRequest) *models.MatchmakingRequest {
	var bestMatch *models.MatchmakingRequest
	bestRatingDiff := math.MaxInt32
	
	// Calculer la tolérance de rating basée sur le temps d'attente
	waitTime := time.Since(request.QueuedAt)
	ratingTolerance := m.calculateRatingTolerance(waitTime)
	
	for _, candidate := range queue {
		// Ne pas se matcher avec soi-même
		if candidate.PlayerID == request.PlayerID {
			continue
		}
		
		// Vérifier la différence de rating
		ratingDiff := abs(request.Rating - candidate.Rating)
		if ratingDiff > ratingTolerance {
			continue
		}
		
		// Vérifier la différence de niveau
		levelDiff := abs(request.Level - candidate.Level)
		if levelDiff > m.getMaxLevelDifference() {
			continue
		}
		
		// Garder le meilleur match (plus proche en rating)
		if ratingDiff < bestRatingDiff {
			bestRatingDiff = ratingDiff
			bestMatch = candidate
		}
	}
	
	return bestMatch
}

// removeFromQueue retire un joueur d'une queue spécifique
func (m *Matchmaker) removeFromQueue(playerID uuid.UUID, matchType string) {
	queue, exists := m.queue[matchType]
	if !exists {
		return
	}
	
	// Trouver et retirer le joueur
	for i, request := range queue {
		if request.PlayerID == playerID {
			// Retirer de la slice
			m.queue[matchType] = append(queue[:i], queue[i+1:]...)
			break
		}
	}
	
	// Retirer du track des joueurs
	delete(m.playerQueue, playerID)
}

// calculateRatingTolerance calcule la tolérance de rating selon le temps d'attente
func (m *Matchmaker) calculateRatingTolerance(waitTime time.Duration) int {
	// Commencer avec une tolérance de 50 points
	baseTolerance := 50
	
	// Augmenter de 25 points toutes les 30 secondes
	minutes := int(waitTime.Seconds() / 30)
	tolerance := baseTolerance + (minutes * 25)
	
	// Maximum de 300 points de différence
	if tolerance > 300 {
		tolerance = 300
	}
	
	return tolerance
}

// getMaxLevelDifference retourne la différence de niveau maximale autorisée
func (m *Matchmaker) getMaxLevelDifference() int {
	return 5 // Maximum 5 niveaux de différence
}

// estimateWaitTime estime le temps d'attente
func (m *Matchmaker) estimateWaitTime(request *models.MatchmakingRequest) time.Duration {
	queue := m.queue[request.MatchType]
	if len(queue) <= 1 {
		return 2 * time.Minute // Estimation par défaut
	}
	
	// Estimer basé sur la taille de la queue et l'activité
	estimatedMinutes := len(queue) / 2 // 2 joueurs par match
	if estimatedMinutes < 1 {
		estimatedMinutes = 1
	}
	
	return time.Duration(estimatedMinutes) * time.Minute
}

// getQueuePosition retourne la position dans la queue
func (m *Matchmaker) getQueuePosition(playerID uuid.UUID, matchType string) int {
	queue := m.queue[matchType]
	
	for i, request := range queue {
		if request.PlayerID == playerID {
			return i + 1 // Position basée sur 1
		}
	}
	
	return 0
}

// abs retourne la valeur absolue d'un entier
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Méthodes d'extension pour les modèles

// UpdateTierAndDivision met à jour le tier et la division selon le rating
func (r *models.PvPRanking) UpdateTierAndDivision() {
	rating := r.Rating
	
	switch {
	case rating >= 2400:
		r.Tier = "Grandmaster"
		r.Division = 1
	case rating >= 2200:
		r.Tier = "Master"
		r.Division = min(5, max(1, (2400-rating)/40+1))
	case rating >= 1800:
		r.Tier = "Diamond"
		r.Division = min(5, max(1, (2200-rating)/80+1))
	case rating >= 1400:
		r.Tier = "Platinum"
		r.Division = min(5, max(1, (1800-rating)/80+1))
	case rating >= 1000:
		r.Tier = "Gold"
		r.Division = min(5, max(1, (1400-rating)/80+1))
	case rating >= 600:
		r.Tier = "Silver"
		r.Division = min(5, max(1, (1000-rating)/80+1))
	default:
		r.Tier = "Bronze"
		r.Division = min(5, max(1, (600-rating)/80+1))
	}
}

// min retourne le minimum de deux entiers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max retourne le maximum de deux entiers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}