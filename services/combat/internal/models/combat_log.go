// internal/models/combat_log.go
package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// CombatLog représente une entrée dans les logs de combat
type CombatLog struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	SessionID   uuid.UUID       `json:"session_id" db:"session_id"`
	ActionID    *uuid.UUID      `json:"action_id" db:"action_id"` // référence à l'action qui a généré ce log
	
	// Participants
	ActorID     *uuid.UUID      `json:"actor_id" db:"actor_id"` // qui a fait l'action
	TargetID    *uuid.UUID      `json:"target_id" db:"target_id"` // qui a reçu l'action
	
	// Type d'événement
	EventType   string          `json:"event_type" db:"event_type"` // damage, heal, death, spell_cast, effect_applied, etc.
	
	// Messages
	Message     string          `json:"message" db:"message"` // message formaté pour l'affichage
	RawData     json.RawMessage `json:"raw_data" db:"raw_data"` // données brutes de l'événement
	
	// Valeurs numériques
	Value       int             `json:"value" db:"value"` // dégâts, soins, etc.
	OldValue    int             `json:"old_value" db:"old_value"` // valeur avant l'action
	NewValue    int             `json:"new_value" db:"new_value"` // valeur après l'action
	
	// Contexte
	IsCritical  bool            `json:"is_critical" db:"is_critical"`
	IsResisted  bool            `json:"is_resisted" db:"is_resisted"`
	IsAbsorbed  bool            `json:"is_absorbed" db:"is_absorbed"`
	
	// Métadonnées
	Color       string          `json:"color" db:"color"` // couleur pour l'affichage
	Icon        string          `json:"icon" db:"icon"`
	Priority    int             `json:"priority" db:"priority"` // pour le tri d'affichage
	
	Timestamp   time.Time       `json:"timestamp" db:"timestamp"`
}

// CombatSummary représente un résumé de combat
type CombatSummary struct {
	SessionID       uuid.UUID               `json:"session_id"`
	Type            string                  `json:"type"`
	Status          string                  `json:"status"`
	Duration        time.Duration           `json:"duration"`
	
	// Participants
	ParticipantCount int                    `json:"participant_count"`
	Winners         []uuid.UUID             `json:"winners"`
	Losers          []uuid.UUID             `json:"losers"`
	
	// Statistiques générales
	TotalDamage     int64                   `json:"total_damage"`
	TotalHealing    int64                   `json:"total_healing"`
	ActionsCount    int                     `json:"actions_count"`
	SpellsCast      int                     `json:"spells_cast"`
	
	// Statistiques par participant
	ParticipantStats []ParticipantSummary   `json:"participant_stats"`
	
	StartedAt       time.Time               `json:"started_at"`
	EndedAt         time.Time               `json:"ended_at"`
}

// ParticipantSummary statistiques d'un participant
type ParticipantSummary struct {
	CharacterID     uuid.UUID `json:"character_id"`
	CharacterName   string    `json:"character_name"`
	Team            int       `json:"team"`
	
	// Statistiques de combat
	DamageDealt     int64     `json:"damage_dealt"`
	DamageTaken     int64     `json:"damage_taken"`
	HealingDone     int64     `json:"healing_done"`
	HealingReceived int64     `json:"healing_received"`
	
	// Actions
	ActionsPerformed int      `json:"actions_performed"`
	SpellsCast       int      `json:"spells_cast"`
	CriticalHits     int      `json:"critical_hits"`
	
	// Résultat
	Survived        bool      `json:"survived"`
	KillCount       int       `json:"kill_count"`
	DeathCount      int       `json:"death_count"`
	
	// Temps
	TimeInCombat    time.Duration `json:"time_in_combat"`
}

// LogEvent représente un événement spécifique pour les logs
type LogEvent struct {
	Type        string      `json:"type"`
	ActorName   string      `json:"actor_name,omitempty"`
	TargetName  string      `json:"target_name,omitempty"`
	SpellName   string      `json:"spell_name,omitempty"`
	ItemName    string      `json:"item_name,omitempty"`
	Value       int         `json:"value,omitempty"`
	IsCritical  bool        `json:"is_critical,omitempty"`
	Timestamp   time.Time   `json:"timestamp"`
}

// internal/models/pvp.go

// PvPMatch représente un match PvP
type PvPMatch struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	SessionID       uuid.UUID       `json:"session_id" db:"session_id"`
	Type            string          `json:"type" db:"type"` // duel, team_vs_team, battle_royale
	
	// Configuration
	MatchRules      PvPRules        `json:"match_rules" db:"match_rules"`
	RatingType      string          `json:"rating_type" db:"rating_type"` // casual, ranked, tournament
	
	// Participants
	Participants    []PvPParticipant `json:"participants,omitempty" db:"-"`
	
	// Résultats
	WinnerTeam      int             `json:"winner_team" db:"winner_team"`
	Results         []PvPResult     `json:"results" db:"results"`
	
	// Timing
	StartedAt       time.Time       `json:"started_at" db:"started_at"`
	EndedAt         *time.Time      `json:"ended_at" db:"ended_at"`
	Duration        time.Duration   `json:"duration" db:"duration"`
	
	// État
	Status          string          `json:"status" db:"status"` // active, completed, cancelled
	
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

// PvPParticipant représente un participant PvP
type PvPParticipant struct {
	ID              uuid.UUID   `json:"id" db:"id"`
	MatchID         uuid.UUID   `json:"match_id" db:"match_id"`
	CharacterID     uuid.UUID   `json:"character_id" db:"character_id"`
	PlayerID        uuid.UUID   `json:"player_id" db:"player_id"`
	
	// Équipe et position
	Team            int         `json:"team" db:"team"`
	Position        int         `json:"position" db:"position"` // position dans l'équipe
	
	// Rating avant le match
	RatingBefore    int         `json:"rating_before" db:"rating_before"`
	RatingAfter     int         `json:"rating_after" db:"rating_after"`
	RatingChange    int         `json:"rating_change" db:"rating_change"`
	
	// Résultat
	IsWinner        bool        `json:"is_winner" db:"is_winner"`
	Placement       int         `json:"placement" db:"placement"` // classement final
	
	JoinedAt        time.Time   `json:"joined_at" db:"joined_at"`
}

// PvPResult représente le résultat détaillé d'un participant
type PvPResult struct {
	CharacterID     uuid.UUID `json:"character_id"`
	Team            int       `json:"team"`
	IsWinner        bool      `json:"is_winner"`
	Kills           int       `json:"kills"`
	Deaths          int       `json:"deaths"`
	Assists         int       `json:"assists"`
	DamageDealt     int64     `json:"damage_dealt"`
	DamageTaken     int64     `json:"damage_taken"`
	HealingDone     int64     `json:"healing_done"`
	RatingChange    int       `json:"rating_change"`
}

// PvPRules règles spécifiques au PvP
type PvPRules struct {
	MaxParticipants int           `json:"max_participants"`
	TeamSize        int           `json:"team_size"`
	TimeLimit       time.Duration `json:"time_limit"`
	KillLimit       int           `json:"kill_limit"`
	AllowItems      bool          `json:"allow_items"`
	AllowHealing    bool          `json:"allow_healing"`
	RespawnAllowed  bool          `json:"respawn_allowed"`
	RespawnTime     time.Duration `json:"respawn_time"`
	Suddendeath     bool          `json:"sudden_death"`
	LevelCap        int           `json:"level_cap"`
	EquipmentLocked bool          `json:"equipment_locked"`
}

// PvPRanking représente le classement PvP d'un joueur
type PvPRanking struct {
	ID              uuid.UUID   `json:"id" db:"id"`
	CharacterID     uuid.UUID   `json:"character_id" db:"character_id"`
	PlayerID        uuid.UUID   `json:"player_id" db:"player_id"`
	Season          string      `json:"season" db:"season"`
	
	// Rating et classement
	Rating          int         `json:"rating" db:"rating"`
	Rank            int         `json:"rank" db:"rank"`
	Tier            string      `json:"tier" db:"tier"` // bronze, silver, gold, platinum, diamond, master, grandmaster
	Division        int         `json:"division" db:"division"` // 1-5 dans chaque tier
	
	// Statistiques
	Wins            int         `json:"wins" db:"wins"`
	Losses          int         `json:"losses"`
	WinRate         float64     `json:"win_rate" db:"win_rate"`
	Streak          int         `json:"streak" db:"streak"` // série de victoires/défaites
	BestStreak      int         `json:"best_streak" db:"best_streak"`
	
	// Historique
	PeakRating      int         `json:"peak_rating" db:"peak_rating"`
	PeakRank        int         `json:"peak_rank" db:"peak_rank"`
	
	// Activité
	LastMatchAt     *time.Time  `json:"last_match_at" db:"last_match_at"`
	IsActive        bool        `json:"is_active" db:"is_active"`
	
	UpdatedAt       time.Time   `json:"updated_at" db:"updated_at"`
	
	// Relations
	Character       *Character  `json:"character,omitempty" db:"-"`
}

// PvPChallenge représente un défi PvP
type PvPChallenge struct {
	ID              uuid.UUID   `json:"id" db:"id"`
	ChallengerID    uuid.UUID   `json:"challenger_id" db:"challenger_id"`
	ChallengedID    uuid.UUID   `json:"challenged_id" db:"challenged_id"`
	
	// Configuration du défi
	Type            string      `json:"type" db:"type"` // duel, team_match
	Rules           PvPRules    `json:"rules" db:"rules"`
	Message         string      `json:"message" db:"message"`
	
	// État
	Status          string      `json:"status" db:"status"` // pending, accepted, declined, expired, cancelled
	
	// Réponse
	ResponseMessage string      `json:"response_message" db:"response_message"`
	RespondedAt     *time.Time  `json:"responded_at" db:"responded_at"`
	
	// Match résultant
	MatchID         *uuid.UUID  `json:"match_id" db:"match_id"`
	
	// Timing
	ExpiresAt       time.Time   `json:"expires_at" db:"expires_at"`
	CreatedAt       time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at" db:"updated_at"`
	
	// Relations
	Challenger      *Character  `json:"challenger,omitempty" db:"-"`
	Challenged      *Character  `json:"challenged,omitempty" db:"-"`
}

// DTOs pour les requêtes PvP

// ChallengePvPRequest requête pour défier en PvP
type ChallengePvPRequest struct {
	ChallengedCharacterID uuid.UUID `json:"challenged_character_id" binding:"required"`
	Type                  string    `json:"type" binding:"required,oneof=duel team_match"`
	Rules                 PvPRules  `json:"rules"`
	Message               string    `json:"message" binding:"max=200"`
}

// RespondToChallengeRequest réponse à un défi PvP
type RespondToChallengeRequest struct {
	Accept          bool   `json:"accept" binding:"required"`
	ResponseMessage string `json:"response_message" binding:"max=200"`
}

// PvPLeaderboardResponse réponse du classement PvP
type PvPLeaderboardResponse struct {
	Season      string        `json:"season"`
	Rankings    []PvPRanking  `json:"rankings"`
	PlayerRank  *PvPRanking   `json:"player_rank,omitempty"`
	TotalPages  int           `json:"total_pages"`
	CurrentPage int           `json:"current_page"`
}

// PvPStatsResponse statistiques PvP d'un joueur
type PvPStatsResponse struct {
	CharacterID     uuid.UUID           `json:"character_id"`
	CurrentRanking  *PvPRanking         `json:"current_ranking"`
	RecentMatches   []PvPMatchSummary   `json:"recent_matches"`
	SeasonStats     PvPSeasonStats      `json:"season_stats"`
	AllTimeStats    PvPAllTimeStats     `json:"all_time_stats"`
}

// PvPMatchSummary résumé d'un match PvP
type PvPMatchSummary struct {
	MatchID         uuid.UUID   `json:"match_id"`
	Type            string      `json:"type"`
	IsWinner        bool        `json:"is_winner"`
	RatingChange    int         `json:"rating_change"`
	Duration        time.Duration `json:"duration"`
	Opponents       []string    `json:"opponents"`
	PlayedAt        time.Time   `json:"played_at"`
}

// PvPSeasonStats statistiques d'une saison
type PvPSeasonStats struct {
	Season          string  `json:"season"`
	Wins            int     `json:"wins"`
	Losses          int     `json:"losses"`
	WinRate         float64 `json:"win_rate"`
	RatingGained    int     `json:"rating_gained"`
	BestStreak      int     `json:"best_streak"`
	CurrentStreak   int     `json:"current_streak"`
	MatchesPlayed   int     `json:"matches_played"`
}

// PvPAllTimeStats statistiques de tous les temps
type PvPAllTimeStats struct {
	TotalWins       int     `json:"total_wins"`
	TotalLosses     int     `json:"total_losses"`
	TotalMatches    int     `json:"total_matches"`
	OverallWinRate  float64 `json:"overall_win_rate"`
	TotalKills      int     `json:"total_kills"`
	TotalDeaths     int     `json:"total_deaths"`
	KDRatio         float64 `json:"kd_ratio"`
	TotalDamage     int64   `json:"total_damage"`
	TotalHealing    int64   `json:"total_healing"`
	BestRating      int     `json:"best_rating"`
	BestRank        int     `json:"best_rank"`
	SeasonsPlayed   int     `json:"seasons_played"`
}