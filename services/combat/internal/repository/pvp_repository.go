package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"combat/internal/database"
	"combat/internal/models"
)

// PvPRepositoryInterface définit les méthodes du repository PvP
type PvPRepositoryInterface interface {
	// Gestion des défis
	CreateChallenge(challenge *models.PvPChallenge) error
	GetChallengeByID(id uuid.UUID) (*models.PvPChallenge, error)
	UpdateChallenge(challenge *models.PvPChallenge) error
	DeleteChallenge(id uuid.UUID) error
	
	// Recherche de défis
	GetChallengesByPlayer(playerID uuid.UUID) ([]*models.PvPChallenge, error)
	GetPendingChallenges(playerID uuid.UUID) ([]*models.PvPChallenge, error)
	GetChallengesByStatus(status models.ChallengeStatus) ([]*models.PvPChallenge, error)
	GetExpiredChallenges() ([]*models.PvPChallenge, error)
	
	// Statistiques PvP
	GetPvPStatistics(playerID uuid.UUID) (*models.PvPStatistics, error)
	UpdatePvPStatistics(stats *models.PvPStatistics) error
	CreatePvPStatistics(stats *models.PvPStatistics) error
	
	// Classements
	GetTopPlayers(limit int) ([]*models.PvPRanking, error)
	GetPlayerRank(playerID uuid.UUID) (int, error)
	GetPlayersInRatingRange(minRating, maxRating int) ([]*models.PvPRanking, error)
	
	// Matchmaking et file d'attente
	AddToQueue(entry *models.PvPQueueEntry) error
	RemoveFromQueue(playerID uuid.UUID) error
	GetQueueEntry(playerID uuid.UUID) (*models.PvPQueueEntry, error)
	GetQueueByType(queueType models.ChallengeType) ([]*models.PvPQueueEntry, error)
	FindMatchmakingCandidates(entry *models.PvPQueueEntry) ([]*models.PvPQueueEntry, error)
	
	// Nettoyage et maintenance
	CleanupExpiredChallenges() error
	CleanupOldQueue() error
}

// PvPRepository implémente l'interface PvPRepositoryInterface
type PvPRepository struct {
	db *database.DB
}

// NewPvPRepository crée une nouvelle instance du repository PvP
func NewPvPRepository(db *database.DB) PvPRepositoryInterface {
	return &PvPRepository{db: db}
}

// CreateChallenge crée un nouveau défi PvP
func (r *PvPRepository) CreateChallenge(challenge *models.PvPChallenge) error {
	// Sérialiser les stakes en JSON
	stakesJSON, err := json.Marshal(challenge.Stakes)
	if err != nil {
		return fmt.Errorf("failed to marshal stakes: %w", err)
	}

	query := `
		INSERT INTO pvp_challenges (
			id, challenger_id, challenged_id, combat_id, challenge_type, message, stakes,
			status, winner_id, loser_id, result_type,
			created_at, responded_at, expires_at, completed_at
		) VALUES (
			:id, :challenger_id, :challenged_id, :combat_id, :challenge_type, :message, :stakes,
			:status, :winner_id, :loser_id, :result_type,
			:created_at, :responded_at, :expires_at, :completed_at
		)`

	data := map[string]interface{}{
		"id":             challenge.ID,
		"challenger_id":  challenge.ChallengerID,
		"challenged_id":  challenge.ChallengedID,
		"combat_id":      challenge.CombatID,
		"challenge_type": challenge.ChallengeType,
		"message":        challenge.Message,
		"stakes":         stakesJSON,
		"status":         challenge.Status,
		"winner_id":      challenge.WinnerID,
		"loser_id":       challenge.LoserID,
		"result_type":    challenge.ResultType,
		"created_at":     challenge.CreatedAt,
		"responded_at":   challenge.RespondedAt,
		"expires_at":     challenge.ExpiresAt,
		"completed_at":   challenge.CompletedAt,
	}

	_, err = r.db.NamedExec(query, data)
	if err != nil {
		return fmt.Errorf("failed to create challenge: %w", err)
	}

	return nil
}

// GetChallengeByID récupère un défi par son ID
func (r *PvPRepository) GetChallengeByID(id uuid.UUID) (*models.PvPChallenge, error) {
	var challenge models.PvPChallenge
	var stakesJSON []byte

	query := `
		SELECT id, challenger_id, challenged_id, combat_id, challenge_type, message, stakes,
		       status, winner_id, loser_id, result_type,
		       created_at, responded_at, expires_at, completed_at
		FROM pvp_challenges 
		WHERE id = $1`

	row := r.db.QueryRow(query, id)
	err := row.Scan(
		&challenge.ID, &challenge.ChallengerID, &challenge.ChallengedID, &challenge.CombatID,
		&challenge.ChallengeType, &challenge.Message, &stakesJSON, &challenge.Status,
		&challenge.WinnerID, &challenge.LoserID, &challenge.ResultType,
		&challenge.CreatedAt, &challenge.RespondedAt, &challenge.ExpiresAt, &challenge.CompletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("challenge not found")
		}
		return nil, fmt.Errorf("failed to get challenge: %w", err)
	}

	// Désérialiser les stakes
	if err := json.Unmarshal(stakesJSON, &challenge.Stakes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal stakes: %w", err)
	}

	return &challenge, nil
}

// UpdateChallenge met à jour un défi
func (r *PvPRepository) UpdateChallenge(challenge *models.PvPChallenge) error {
	// Sérialiser les stakes
	stakesJSON, err := json.Marshal(challenge.Stakes)
	if err != nil {
		return fmt.Errorf("failed to marshal stakes: %w", err)
	}

	query := `
		UPDATE pvp_challenges SET
			combat_id = :combat_id,
			status = :status,
			winner_id = :winner_id,
			loser_id = :loser_id,
			result_type = :result_type,
			responded_at = :responded_at,
			completed_at = :completed_at,
			stakes = :stakes
		WHERE id = :id`

	data := map[string]interface{}{
		"id":           challenge.ID,
		"combat_id":    challenge.CombatID,
		"status":       challenge.Status,
		"winner_id":    challenge.WinnerID,
		"loser_id":     challenge.LoserID,
		"result_type":  challenge.ResultType,
		"responded_at": challenge.RespondedAt,
		"completed_at": challenge.CompletedAt,
		"stakes":       stakesJSON,
	}

	result, err := r.db.NamedExec(query, data)
	if err != nil {
		return fmt.Errorf("failed to update challenge: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("challenge not found")
	}

	return nil
}

// GetChallengesByPlayer récupère les défis d'un joueur
func (r *PvPRepository) GetChallengesByPlayer(playerID uuid.UUID) ([]*models.PvPChallenge, error) {
	var challenges []*models.PvPChallenge

	query := `
		SELECT id, challenger_id, challenged_id, combat_id, challenge_type, message, stakes,
		       status, winner_id, loser_id, result_type,
		       created_at, responded_at, expires_at, completed_at
		FROM pvp_challenges 
		WHERE challenger_id = $1 OR challenged_id = $1 
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get challenges by player: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var challenge models.PvPChallenge
		var stakesJSON []byte

		err := rows.Scan(
			&challenge.ID, &challenge.ChallengerID, &challenge.ChallengedID, &challenge.CombatID,
			&challenge.ChallengeType, &challenge.Message, &stakesJSON, &challenge.Status,
			&challenge.WinnerID, &challenge.LoserID, &challenge.ResultType,
			&challenge.CreatedAt, &challenge.RespondedAt, &challenge.ExpiresAt, &challenge.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan challenge: %w", err)
		}

		// Désérialiser les stakes
		if err := json.Unmarshal(stakesJSON, &challenge.Stakes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal stakes: %w", err)
		}

		challenges = append(challenges, &challenge)
	}

	return challenges, nil
}

// GetPendingChallenges récupère les défis en attente pour un joueur
func (r *PvPRepository) GetPendingChallenges(playerID uuid.UUID) ([]*models.PvPChallenge, error) {
	var challenges []*models.PvPChallenge

	query := `
		SELECT id, challenger_id, challenged_id, combat_id, challenge_type, message, stakes,
		       status, winner_id, loser_id, result_type,
		       created_at, responded_at, expires_at, completed_at
		FROM pvp_challenges 
		WHERE challenged_id = $1 AND status = 'pending' AND expires_at > CURRENT_TIMESTAMP
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending challenges: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var challenge models.PvPChallenge
		var stakesJSON []byte

		err := rows.Scan(
			&challenge.ID, &challenge.ChallengerID, &challenge.ChallengedID, &challenge.CombatID,
			&challenge.ChallengeType, &challenge.Message, &stakesJSON, &challenge.Status,
			&challenge.WinnerID, &challenge.LoserID, &challenge.ResultType,
			&challenge.CreatedAt, &challenge.RespondedAt, &challenge.ExpiresAt, &challenge.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan challenge: %w", err)
		}

		// Désérialiser les stakes
		if err := json.Unmarshal(stakesJSON, &challenge.Stakes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal stakes: %w", err)
		}

		challenges = append(challenges, &challenge)
	}

	return challenges, nil
}

// GetChallengesByStatus récupère les défis par statut
func (r *PvPRepository) GetChallengesByStatus(status models.ChallengeStatus) ([]*models.PvPChallenge, error) {
	var challenges []*models.PvPChallenge

	query := `
		SELECT id, challenger_id, challenged_id, combat_id, challenge_type, message, stakes,
		       status, winner_id, loser_id, result_type,
		       created_at, responded_at, expires_at, completed_at
		FROM pvp_challenges 
		WHERE status = $1 
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get challenges by status: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var challenge models.PvPChallenge
		var stakesJSON []byte

		err := rows.Scan(
			&challenge.ID, &challenge.ChallengerID, &challenge.ChallengedID, &challenge.CombatID,
			&challenge.ChallengeType, &challenge.Message, &stakesJSON, &challenge.Status,
			&challenge.WinnerID, &challenge.LoserID, &challenge.ResultType,
			&challenge.CreatedAt, &challenge.RespondedAt, &challenge.ExpiresAt, &challenge.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan challenge: %w", err)
		}

		// Désérialiser les stakes
		if err := json.Unmarshal(stakesJSON, &challenge.Stakes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal stakes: %w", err)
		}

		challenges = append(challenges, &challenge)
	}

	return challenges, nil
}

// GetExpiredChallenges récupère les défis expirés
func (r *PvPRepository) GetExpiredChallenges() ([]*models.PvPChallenge, error) {
	var challenges []*models.PvPChallenge

	query := `
		SELECT id, challenger_id, challenged_id, combat_id, challenge_type, message, stakes,
		       status, winner_id, loser_id, result_type,
		       created_at, responded_at, expires_at, completed_at
		FROM pvp_challenges 
		WHERE status = 'pending' AND expires_at < CURRENT_TIMESTAMP`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get expired challenges: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var challenge models.PvPChallenge
		var stakesJSON []byte

		err := rows.Scan(
			&challenge.ID, &challenge.ChallengerID, &challenge.ChallengedID, &challenge.CombatID,
			&challenge.ChallengeType, &challenge.Message, &stakesJSON, &challenge.Status,
			&challenge.WinnerID, &challenge.LoserID, &challenge.ResultType,
			&challenge.CreatedAt, &challenge.RespondedAt, &challenge.ExpiresAt, &challenge.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan challenge: %w", err)
		}

		// Désérialiser les stakes
		if err := json.Unmarshal(stakesJSON, &challenge.Stakes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal stakes: %w", err)
		}

		challenges = append(challenges, &challenge)
	}

	return challenges, nil
}

// GetPvPStatistics récupère les statistiques PvP d'un joueur
func (r *PvPRepository) GetPvPStatistics(playerID uuid.UUID) (*models.PvPStatistics, error) {
	// Pour l'instant, on récupère depuis combat_statistics
	// Dans une vraie implémentation, on aurait une table séparée pour les stats PvP
	query := `
		SELECT character_id, user_id, pvp_battles_won, pvp_battles_lost, pvp_draws, pvp_rating,
		       total_damage_dealt, total_damage_taken, total_healing_done,
		       updated_at
		FROM combat_statistics 
		WHERE character_id = $1`

	var stats models.PvPStatistics
	err := r.db.QueryRow(query, playerID).Scan(
		&stats.PlayerID, &stats.UserID, &stats.BattlesWon, &stats.BattlesLost, &stats.Draws, &stats.CurrentRating,
		&stats.TotalDamageDealt, &stats.TotalDamageTaken, &stats.TotalHealingDone,
		&stats.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Créer des statistiques par défaut
			return &models.PvPStatistics{
				PlayerID:      playerID,
				CurrentRating: 1000, // Rating par défaut
				HighestRating: 1000,
				UpdatedAt:     time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("failed to get pvp statistics: %w", err)
	}

	// Calculer les statistiques dérivées
	totalMatches := stats.BattlesWon + stats.BattlesLost + stats.Draws
	if totalMatches > 0 {
		stats.WinRate = float64(stats.BattlesWon) / float64(totalMatches) * 100
	}

	stats.TotalMatches = totalMatches
	stats.HighestRating = stats.CurrentRating // TODO: Stocker le plus haut rating séparément
	stats.RankName = models.GetRankFromRating(stats.CurrentRating)

	return &stats, nil
}

// UpdatePvPStatistics met à jour les statistiques PvP
func (r *PvPRepository) UpdatePvPStatistics(stats *models.PvPStatistics) error {
	stats.UpdatedAt = time.Now()

	query := `
		UPDATE combat_statistics SET
			pvp_battles_won = $2,
			pvp_battles_lost = $3,
			pvp_draws = $4,
			pvp_rating = $5,
			updated_at = $6
		WHERE character_id = $1`

	result, err := r.db.Exec(query, stats.PlayerID, stats.BattlesWon, stats.BattlesLost, 
		stats.Draws, stats.CurrentRating, stats.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update pvp statistics: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("statistics not found")
	}

	return nil
}

// CreatePvPStatistics crée de nouvelles statistiques PvP
func (r *PvPRepository) CreatePvPStatistics(stats *models.PvPStatistics) error {
	query := `
		INSERT INTO combat_statistics (
			id, character_id, user_id, pvp_battles_won, pvp_battles_lost, pvp_draws, pvp_rating,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		) ON CONFLICT (character_id) DO UPDATE SET
			pvp_battles_won = EXCLUDED.pvp_battles_won,
			pvp_battles_lost = EXCLUDED.pvp_battles_lost,
			pvp_draws = EXCLUDED.pvp_draws,
			pvp_rating = EXCLUDED.pvp_rating,
			updated_at = EXCLUDED.updated_at`

	_, err := r.db.Exec(query, uuid.New(), stats.PlayerID, stats.UserID, 
		stats.BattlesWon, stats.BattlesLost, stats.Draws, stats.CurrentRating,
		time.Now(), time.Now())
	if err != nil {
		return fmt.Errorf("failed to create pvp statistics: %w", err)
	}

	return nil
}

// GetTopPlayers récupère le classement des meilleurs joueurs
func (r *PvPRepository) GetTopPlayers(limit int) ([]*models.PvPRanking, error) {
	var rankings []*models.PvPRanking

	query := `
		SELECT character_id, pvp_rating, pvp_battles_won, pvp_battles_lost, pvp_draws,
		       ROW_NUMBER() OVER (ORDER BY pvp_rating DESC) as rank
		FROM combat_statistics 
		WHERE pvp_rating > 0 
		ORDER BY pvp_rating DESC 
		LIMIT $1`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top players: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var ranking models.PvPRanking

		err := rows.Scan(
			&ranking.PlayerID, &ranking.Rating, &ranking.Wins, &ranking.Losses, &ranking.Draws,
			&ranking.Rank,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ranking: %w", err)
		}

		// Calculer le taux de victoire
		totalMatches := ranking.Wins + ranking.Losses + ranking.Draws
		if totalMatches > 0 {
			ranking.WinRate = float64(ranking.Wins) / float64(totalMatches) * 100
		}

		// TODO: Récupérer le nom du joueur depuis le service player
		ranking.PlayerName = "Player" // Placeholder

		rankings = append(rankings, &ranking)
	}

	return rankings, nil
}

// GetPlayerRank récupère le rang d'un joueur spécifique
func (r *PvPRepository) GetPlayerRank(playerID uuid.UUID) (int, error) {
	query := `
		WITH ranked_players AS (
			SELECT character_id, 
			       ROW_NUMBER() OVER (ORDER BY pvp_rating DESC) as rank
			FROM combat_statistics 
			WHERE pvp_rating > 0
		)
		SELECT rank FROM ranked_players WHERE character_id = $1`

	var rank int
	err := r.db.Get(&rank, query, playerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("player not ranked")
		}
		return 0, fmt.Errorf("failed to get player rank: %w", err)
	}

	return rank, nil
}

// GetPlayersInRatingRange récupère les joueurs dans une plage de rating
func (r *PvPRepository) GetPlayersInRatingRange(minRating, maxRating int) ([]*models.PvPRanking, error) {
	var rankings []*models.PvPRanking

	query := `
		SELECT character_id, pvp_rating, pvp_battles_won, pvp_battles_lost, pvp_draws
		FROM combat_statistics 
		WHERE pvp_rating BETWEEN $1 AND $2 
		ORDER BY pvp_rating DESC`

	rows, err := r.db.Query(query, minRating, maxRating)
	if err != nil {
		return nil, fmt.Errorf("failed to get players in rating range: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var ranking models.PvPRanking

		err := rows.Scan(
			&ranking.PlayerID, &ranking.Rating, &ranking.Wins, &ranking.Losses, &ranking.Draws,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ranking: %w", err)
		}

		// Calculer le taux de victoire
		totalMatches := ranking.Wins + ranking.Losses + ranking.Draws
		if totalMatches > 0 {
			ranking.WinRate = float64(ranking.Wins) / float64(totalMatches) * 100
		}

		rankings = append(rankings, &ranking)
	}

	return rankings, nil
}

// Note: Les méthodes de file d'attente suivantes nécessiteraient une table séparée
// Pour l'instant, on utilise une implémentation en mémoire ou on peut créer une table temporaire

// AddToQueue ajoute un joueur à la file d'attente
func (r *PvPRepository) AddToQueue(entry *models.PvPQueueEntry) error {
	// TODO: Implémenter avec une vraie table de file d'attente
	// Pour l'instant, retourner une implémentation stub
	return fmt.Errorf("queue functionality not implemented yet")
}

// RemoveFromQueue retire un joueur de la file d'attente
func (r *PvPRepository) RemoveFromQueue(playerID uuid.UUID) error {
	// TODO: Implémenter
	return fmt.Errorf("queue functionality not implemented yet")
}

// GetQueueEntry récupère l'entrée de file d'attente d'un joueur
func (r *PvPRepository) GetQueueEntry(playerID uuid.UUID) (*models.PvPQueueEntry, error) {
	// TODO: Implémenter
	return nil, fmt.Errorf("queue functionality not implemented yet")
}

// GetQueueByType récupère les entrées de file d'attente par type
func (r *PvPRepository) GetQueueByType(queueType models.ChallengeType) ([]*models.PvPQueueEntry, error) {
	// TODO: Implémenter
	return nil, fmt.Errorf("queue functionality not implemented yet")
}

// FindMatchmakingCandidates trouve des candidats pour le matchmaking
func (r *PvPRepository) FindMatchmakingCandidates(entry *models.PvPQueueEntry) ([]*models.PvPQueueEntry, error) {
	// TODO: Implémenter un algorithme de matchmaking basé sur le rating
	return nil, fmt.Errorf("matchmaking not implemented yet")
}

// CleanupExpiredChallenges nettoie les défis expirés
func (r *PvPRepository) CleanupExpiredChallenges() error {
	query := `
		UPDATE pvp_challenges 
		SET status = 'expired' 
		WHERE status = 'pending' AND expires_at < CURRENT_TIMESTAMP`

	result, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired challenges: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected > 0 {
		fmt.Printf("Marked %d challenges as expired\n", rowsAffected)
	}

	return nil
}

// CleanupOldQueue nettoie les anciennes entrées de file d'attente
func (r *PvPRepository) CleanupOldQueue() error {
	// TODO: Implémenter
	return nil
}

// GetChallengeStatistics récupère des statistiques sur les défis
func (r *PvPRepository) GetChallengeStatistics() (*ChallengeStatistics, error) {
	query := `
		SELECT 
			COUNT(*) as total_challenges,
			COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_challenges,
			COUNT(CASE WHEN status = 'accepted' THEN 1 END) as accepted_challenges,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_challenges,
			COUNT(CASE WHEN status = 'declined' THEN 1 END) as declined_challenges,
			COUNT(CASE WHEN status = 'expired' THEN 1 END) as expired_challenges,
			AVG(EXTRACT(EPOCH FROM (responded_at - created_at))/60) as avg_response_time_minutes
		FROM pvp_challenges 
		WHERE created_at > CURRENT_TIMESTAMP - INTERVAL '7 days'`

	var stats ChallengeStatistics
	err := r.db.Get(&stats, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get challenge statistics: %w", err)
	}

	return &stats, nil
}

// ChallengeStatistics représente les statistiques des défis
type ChallengeStatistics struct {
	TotalChallenges     int     `json:"total_challenges" db:"total_challenges"`
	PendingChallenges   int     `json:"pending_challenges" db:"pending_challenges"`
	AcceptedChallenges  int     `json:"accepted_challenges" db:"accepted_challenges"`
	CompletedChallenges int     `json:"completed_challenges" db:"completed_challenges"`
	DeclinedChallenges  int     `json:"declined_challenges" db:"declined_challenges"`
	ExpiredChallenges   int     `json:"expired_challenges" db:"expired_challenges"`
	AvgResponseTime     float64 `json:"avg_response_time_minutes" db:"avg_response_time_minutes"`
}

// DeleteChallenge supprime un défi
func (r *PvPRepository) DeleteChallenge(id uuid.UUID) error {
	query := `DELETE FROM pvp_challenges WHERE id = $1`
	
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete challenge: %w", err)
	}

	rowsAffecte