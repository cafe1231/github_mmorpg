package models

import (
	"time"

	"github.com/google/uuid"
)

// MatchmakingRequest représente une demande de matchmaking
type MatchmakingRequest struct {
	PlayerID     uuid.UUID              `json:"player_id"`
	MatchType    string                 `json:"match_type"`
	Rating       int                    `json:"rating"`
	Level        int                    `json:"level"`
	QueuedAt     time.Time              `json:"queued_at"`
	Preferences  map[string]interface{} `json:"preferences"`
}

// MatchmakingStatus représente le statut d'un joueur dans le matchmaking
type MatchmakingStatus struct {
	InQueue       bool          `json:"in_queue"`
	MatchType     string        `json:"match_type,omitempty"`
	QueueTime     time.Duration `json:"queue_time,omitempty"`
	EstimatedWait time.Duration `json:"estimated_wait,omitempty"`
	QueuePosition int           `json:"queue_position,omitempty"`
}

// Match représente un match trouvé par le matchmaking
type Match struct {
	ID        uuid.UUID     `json:"id"`
	MatchType string        `json:"match_type"`
	Players   []uuid.UUID   `json:"players"`
	CreatedAt time.Time     `json:"created_at"`
	Status    string        `json:"status"`
}

// MatchResult représente le résultat d'un match PvP
type MatchResult struct {
	SessionID uuid.UUID     `json:"session_id"`
	MatchType string        `json:"match_type"`
	WinnerID  uuid.UUID     `json:"winner_id"`
	LoserID   uuid.UUID     `json:"loser_id"`
	StartedAt time.Time     `json:"started_at"`
	EndedAt   time.Time     `json:"ended_at"`
	Duration  time.Duration `json:"duration"`
	Result    string        `json:"result"` // victory, defeat, draw
}

// NOTE: PvPChallenge est déjà défini dans combat_log.go
// NOTE: PvPRanking est déjà défini dans combat_log.go  
// NOTE: PvPMatch est déjà défini dans combat_log.go

// PvPStats représente les statistiques PvP d'un joueur
type PvPStats struct {
	PlayerID       uuid.UUID `json:"player_id" db:"player_id"`
	Season         string    `json:"season" db:"season"`
	TotalMatches   int       `json:"total_matches" db:"total_matches"`
	Wins           int       `json:"wins" db:"wins"`
	Losses         int       `json:"losses" db:"losses"`
	WinRate        float64   `json:"win_rate" db:"win_rate"`
	WinStreak      int       `json:"win_streak" db:"win_streak"`
	LossStreak     int       `json:"loss_streak" db:"loss_streak"`
	BestWinStreak  int       `json:"best_win_streak" db:"best_win_streak"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// SeasonInfo représente les informations d'une saison
type SeasonInfo struct {
	Season       string    `json:"season" db:"season"`
	StartDate    time.Time `json:"start_date" db:"start_date"`
	EndDate      time.Time `json:"end_date" db:"end_date"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	PlayerCount  int       `json:"player_count" db:"player_count"`
	MatchCount   int       `json:"match_count" db:"match_count"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// SeasonRewards représente les récompenses de fin de saison
type SeasonRewards struct {
	PlayerID uuid.UUID `json:"player_id"`
	Season   string    `json:"season"`
	Tier     string    `json:"tier"`
	Division int       `json:"division"`
	Rating   int       `json:"rating"`
	Gold     int       `json:"gold"`
	Items    []string  `json:"items"`
	Title    string    `json:"title,omitempty"`
}

// NOTE: StartCombatRequest est déjà défini dans combat.go