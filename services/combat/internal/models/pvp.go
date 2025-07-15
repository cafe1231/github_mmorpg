// internal/models/pvp.go
package models

import (
	"time"

	"github.com/google/uuid"
)

// PvPChallenge représente un défi PvP entre deux joueurs
type PvPChallenge struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	ChallengerID   uuid.UUID  `json:"challenger_id" db:"challenger_id"`
	TargetPlayerID uuid.UUID  `json:"target_player_id" db:"target_player_id"` // Corrigé le nom
	Type           string     `json:"type" db:"type"` // duel, ranked, casual, tournament
	Status         string     `json:"status" db:"status"` // pending, accepted, declined, cancelled, expired
	Message        string     `json:"message" db:"message"` // Message optionnel du challenger
	
	// Timing
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt    time.Time  `json:"expires_at" db:"expires_at"`
	AcceptedAt   *time.Time `json:"accepted_at" db:"accepted_at"`
	DeclinedAt   *time.Time `json:"declined_at" db:"declined_at"`
	CancelledAt  *time.Time `json:"cancelled_at" db:"cancelled_at"`
	
	// Métadonnées
	SessionID    *uuid.UUID `json:"session_id" db:"session_id"` // ID de la session créée si accepté
	
	// Relations (chargées séparément)
	Challenger   *Character `json:"challenger,omitempty" db:"-"`
	TargetPlayer *Character `json:"target_player,omitempty" db:"-"`
}

// MatchmakingPreferences représente les préférences de matchmaking d'un joueur
type MatchmakingPreferences struct {
	PlayerID         uuid.UUID `json:"player_id"`
	MatchTypes       []string  `json:"match_types"` // duel, ranked, casual
	PreferredRating  int       `json:"preferred_rating"`
	RatingRange      int       `json:"rating_range"` // ±range autour du rating préféré
	MaxWaitTime      int       `json:"max_wait_time"` // en secondes
	AllowCrossFaction bool     `json:"allow_cross_faction"`
}

// MatchmakingQueue représente un joueur dans la queue de matchmaking
type MatchmakingQueue struct {
	ID           uuid.UUID              `json:"id" db:"id"`
	PlayerID     uuid.UUID              `json:"player_id" db:"player_id"`
	MatchType    string                 `json:"match_type" db:"match_type"`
	Rating       int                    `json:"rating" db:"rating"`
	Preferences  MatchmakingPreferences `json:"preferences" db:"preferences"`
	QueuedAt     time.Time              `json:"queued_at" db:"queued_at"`
	EstimatedWait time.Duration         `json:"estimated_wait" db:"estimated_wait"`
	Status       string                 `json:"status" db:"status"` // waiting, matched, cancelled
}

// PvPMatch représente un match trouvé par le système de matchmaking
type PvPMatch struct {
	ID          uuid.UUID   `json:"id" db:"id"`
	MatchType   string      `json:"match_type" db:"match_type"`
	Players     []uuid.UUID `json:"players" db:"players"`
	SessionID   *uuid.UUID  `json:"session_id" db:"session_id"`
	Status      string      `json:"status" db:"status"` // created, active, completed, cancelled
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	StartedAt   *time.Time  `json:"started_at" db:"started_at"`
	CompletedAt *time.Time  `json:"completed_at" db:"completed_at"`
}

// PvPRating représente le rating PvP d'un joueur
type PvPRating struct {
	ID           uuid.UUID `json:"id" db:"id"`
	PlayerID     uuid.UUID `json:"player_id" db:"player_id"`
	MatchType    string    `json:"match_type" db:"match_type"`
	Rating       int       `json:"rating" db:"rating"`
	Wins         int       `json:"wins" db:"wins"`
	Losses       int       `json:"losses" db:"losses"`
	Draws        int       `json:"draws" db:"draws"`
	Streak       int       `json:"streak" db:"streak"` // Streak actuelle (positif = wins, négatif = losses)
	BestRating   int       `json:"best_rating" db:"best_rating"`
	LastMatchAt  *time.Time `json:"last_match_at" db:"last_match_at"`
	Season       string    `json:"season" db:"season"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// PvPStats représente les statistiques PvP détaillées d'un joueur
type PvPStats struct {
	PlayerID          uuid.UUID `json:"player_id"`
	TotalMatches      int       `json:"total_matches"`
	TotalWins         int       `json:"total_wins"`
	TotalLosses       int       `json:"total_losses"`
	TotalDraws        int       `json:"total_draws"`
	WinRate           float64   `json:"win_rate"`
	AverageMatchTime  float64   `json:"average_match_time"` // en minutes
	TotalDamageDealt  int       `json:"total_damage_dealt"`
	TotalHealingDone  int       `json:"total_healing_done"`
	TotalKills        int       `json:"total_kills"`
	TotalDeaths       int       `json:"total_deaths"`
	KillDeathRatio    float64   `json:"kill_death_ratio"`
	LongestWinStreak  int       `json:"longest_win_streak"`
	CurrentStreak     int       `json:"current_streak"`
	FavoriteMatchType string    `json:"favorite_match_type"`
}

// PvPLeaderboard représente une entrée du classement PvP
type PvPLeaderboardEntry struct {
	Rank         int       `json:"rank"`
	PlayerID     uuid.UUID `json:"player_id"`
	PlayerName   string    `json:"player_name"`
	Rating       int       `json:"rating"`
	Wins         int       `json:"wins"`
	Losses       int       `json:"losses"`
	WinRate      float64   `json:"win_rate"`
	Streak       int       `json:"streak"`
	LastMatchAt  *time.Time `json:"last_match_at"`
}

// Méthodes utilitaires

// IsExpired vérifie si un défi a expiré
func (c *PvPChallenge) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// CanBeAccepted vérifie si un défi peut être accepté
func (c *PvPChallenge) CanBeAccepted() bool {
	return c.Status == "pending" && !c.IsExpired()
}

// CanBeCancelled vérifie si un défi peut être annulé
func (c *PvPChallenge) CanBeCancelled() bool {
	return c.Status == "pending"
}

// GetWinRate calcule le pourcentage de victoires
func (r *PvPRating) GetWinRate() float64 {
	total := r.Wins + r.Losses + r.Draws
	if total == 0 {
		return 0.0
	}
	return float64(r.Wins) / float64(total) * 100.0
}

// GetTotalMatches retourne le nombre total de matchs
func (r *PvPRating) GetTotalMatches() int {
	return r.Wins + r.Losses + r.Draws
}

// IsOnWinStreak vérifie si le joueur est sur une série de victoires
func (r *PvPRating) IsOnWinStreak() bool {
	return r.Streak > 0
}

// IsOnLossStreak vérifie si le joueur est sur une série de défaites
func (r *PvPRating) IsOnLossStreak() bool {
	return r.Streak < 0
}

// GetEstimatedRatingChange estime le changement de rating pour un match
func (r *PvPRating) GetEstimatedRatingChange(opponentRating int, isWin bool) int {
	// Formule simple basée sur la différence de rating
	ratingDiff := opponentRating - r.Rating
	baseChange := 25
	
	if isWin {
		if ratingDiff > 0 {
			// Victoire contre un adversaire plus fort
			return baseChange + (ratingDiff / 10)
		} else {
			// Victoire contre un adversaire plus faible
			return baseChange - (-ratingDiff / 15)
		}
	} else {
		if ratingDiff > 0 {
			// Défaite contre un adversaire plus fort
			return -(baseChange - (ratingDiff / 15))
		} else {
			// Défaite contre un adversaire plus faible
			return -(baseChange + (-ratingDiff / 10))
		}
	}
}

// Constants pour les types de matchs PvP
const (
	MatchTypeDuel       = "duel"
	MatchTypeRanked     = "ranked"
	MatchTypeCasual     = "casual"
	MatchTypeTournament = "tournament"
	MatchTypeArena      = "arena"
	MatchTypeBattleground = "battleground"
)

// Constants pour les statuts de défi
const (
	ChallengeStatusPending   = "pending"
	ChallengeStatusAccepted  = "accepted"
	ChallengeStatusDeclined  = "declined"
	ChallengeStatusCancelled = "cancelled"
	ChallengeStatusExpired   = "expired"
	ChallengeStatusCompleted = "completed"
)

// Constants pour les statuts de matchmaking
const (
	MatchmakingStatusWaiting   = "waiting"
	MatchmakingStatusMatched   = "matched"
	MatchmakingStatusCancelled = "cancelled"
)

// Constants pour les statuts de match
const (
	MatchStatusCreated   = "created"
	MatchStatusActive    = "active"
	MatchStatusCompleted = "completed"
	MatchStatusCancelled = "cancelled"
)