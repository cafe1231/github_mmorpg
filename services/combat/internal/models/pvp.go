package models

import (
	"combat/internal/config"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
)

// ChallengeType définit les types de défis PvP
type ChallengeType string

const (
	ChallengeTypeDuel       ChallengeType = "duel"
	ChallengeTypeArena      ChallengeType = "arena"
	ChallengeTypeTournament ChallengeType = "tournament"
)

// ChallengeStatus définit les status d'un défi
type ChallengeStatus string

const (
	ChallengeStatusPending   ChallengeStatus = "pending"
	ChallengeStatusAccepted  ChallengeStatus = "accepted"
	ChallengeStatusDeclined  ChallengeStatus = "declined"
	ChallengeStatusCancelled ChallengeStatus = "canceled"
	ChallengeStatusExpired   ChallengeStatus = "expired"
	ChallengeStatusCompleted ChallengeStatus = "completed"
)

// ResultType définit les types de résultats
type ResultType string

const (
	ResultTypeVictory ResultType = "victory"
	ResultTypeDefeat  ResultType = "defeat"
	ResultTypeDraw    ResultType = "draw"
	ResultTypeForfeit ResultType = "forfeit"
	ResultTypeTimeout ResultType = "timeout"
)

// PvPChallenge représente un défi PvP
type PvPChallenge struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	ChallengerID uuid.UUID  `json:"challenger_id" db:"challenger_id"`
	ChallengedID uuid.UUID  `json:"challenged_id" db:"challenged_id"`
	CombatID     *uuid.UUID `json:"combat_id" db:"combat_id"`

	// Détails du défi
	ChallengeType ChallengeType `json:"challenge_type" db:"challenge_type"`
	Message       *string       `json:"message" db:"message"`
	Stakes        PvPStakes     `json:"stakes" db:"stakes"`

	// État du défi
	Status ChallengeStatus `json:"status" db:"status"`

	// Résultat
	WinnerID   *uuid.UUID  `json:"winner_id" db:"winner_id"`
	LoserID    *uuid.UUID  `json:"loser_id" db:"loser_id"`
	ResultType *ResultType `json:"result_type" db:"result_type"`

	// Timestamps
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	RespondedAt *time.Time `json:"responded_at" db:"responded_at"`
	ExpiresAt   time.Time  `json:"expires_at" db:"expires_at"`
	CompletedAt *time.Time `json:"completed_at" db:"completed_at"`

	// Relations (chargées séparément)
	Challenger *PlayerSummary  `json:"challenger,omitempty" db:"-"`
	Challenged *PlayerSummary  `json:"challenged,omitempty" db:"-"`
	Combat     *CombatInstance `json:"combat,omitempty" db:"-"`
}

// RespondToChallengeRequest représente une réponse à un défi
type RespondToChallengeRequest struct {
	PlayerID uuid.UUID `json:"player_id" binding:"required"`
	Accept   bool      `json:"accept" binding:"required"`
	Message  string    `json:"message,omitempty"`
}

// PvPStakes représente les enjeux d'un défi PvP
type PvPStakes struct {
	Gold          int                    `json:"gold,omitempty"`
	Items         []StakeItem            `json:"items,omitempty"`
	Experience    int                    `json:"experience,omitempty"`
	Reputation    int                    `json:"reputation,omitempty"`
	Title         string                 `json:"title,omitempty"`
	CustomRewards map[string]interface{} `json:"custom_rewards,omitempty"`
}

// StakeItem représente un objet mis en jeu
type StakeItem struct {
	ItemID   string `json:"item_id"`
	Quantity int    `json:"quantity"`
	Quality  string `json:"quality,omitempty"`
}

// ChallengeResponse représente la réponse à un défi
type ChallengeResponse struct {
	Success   bool          `json:"success"`
	Challenge *PvPChallenge `json:"challenge,omitempty"`
	Match     *PvPMatch     `json:"match,omitempty"` // Si accepté et match créé
	Message   string        `json:"message,omitempty"`
	Error     string        `json:"error,omitempty"`
}

// GetRankingsRequest représente une demande de classements
type GetRankingsRequest struct {
	Season   string        `json:"season,omitempty"`
	Type     ChallengeType `json:"type,omitempty"`
	Limit    int           `json:"limit,omitempty"`
	Offset   int           `json:"offset,omitempty"`
	PlayerID *uuid.UUID    `json:"player_id,omitempty"` // Pour obtenir le rang du joueur
}

// RankingsResponse représente la réponse de classements
type RankingsResponse struct {
	Rankings   []*PvPRanking `json:"rankings"`
	PlayerRank *PvPRanking   `json:"player_rank,omitempty"`
	Total      int           `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	Season     string        `json:"season"`
}

// MatchResult représente le résultat d'un match
type MatchResult struct {
	MatchID      uuid.UUID     `json:"match_id" binding:"required"`
	WinnerID     uuid.UUID     `json:"winner_id" binding:"required"`
	LoserID      uuid.UUID     `json:"loser_id" binding:"required"`
	ResultType   ResultType    `json:"result_type" binding:"required"`
	Duration     time.Duration `json:"duration"`
	WinnerRating int           `json:"winner_rating"`
	LoserRating  int           `json:"loser_rating"`
	RatingChange int           `json:"rating_change"`
}

// SeasonInfo représente les informations de saison
type SeasonInfo struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	StartDate   time.Time      `json:"start_date"`
	EndDate     time.Time      `json:"end_date"`
	IsActive    bool           `json:"is_active"`
	IsCurrent   bool           `json:"is_current"`
	Rewards     []SeasonReward `json:"rewards,omitempty"`
}

// SeasonReward représente une reward de saison
type SeasonReward struct {
	RankRequired string       `json:"rank_required"`
	Title        string       `json:"title,omitempty"`
	Items        []RewardItem `json:"items,omitempty"`
	Gold         int          `json:"gold,omitempty"`
	Experience   int          `json:"experience,omitempty"`
}

// JoinQueueRequest représente une demande d'entrée en file d'attente
type JoinQueueRequest struct {
	PlayerID    uuid.UUID         `json:"player_id" binding:"required"`
	QueueType   ChallengeType     `json:"queue_type" binding:"required"`
	Preferences *QueuePreferences `json:"preferences,omitempty"`
}

// QueueResponse représente la réponse d'entrée en file d'attente
type QueueResponse struct {
	Success       bool           `json:"success"`
	QueueEntry    *PvPQueueEntry `json:"queue_entry,omitempty"`
	Position      int            `json:"position"`
	EstimatedWait time.Duration  `json:"estimated_wait"`
	Message       string         `json:"message,omitempty"`
	Error         string         `json:"error,omitempty"`
}

// QueueStatus représente le statut de file d'attente
type QueueStatus struct {
	InQueue       bool           `json:"in_queue"`
	QueueEntry    *PvPQueueEntry `json:"queue_entry,omitempty"`
	Position      int            `json:"position"`
	EstimatedWait time.Duration  `json:"estimated_wait"`
	QueueSize     int            `json:"queue_size"`
}

// PlayerSummary représente un résumé d'informations joueur
type PlayerSummary struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Name      string    `json:"name"`
	Level     int       `json:"level"`
	Class     string    `json:"class"`
	PvPRating int       `json:"pvp_rating"`
	Avatar    string    `json:"avatar,omitempty"`
	Status    string    `json:"status,omitempty"`
}

// PvPMatch représente un match PvP
type PvPMatch struct {
	ID            uuid.UUID                `json:"id"`
	ChallengeID   uuid.UUID                `json:"challenge_id"`
	CombatID      uuid.UUID                `json:"combat_id"`
	MatchType     ChallengeType            `json:"match_type"`
	Players       []*PlayerSummary         `json:"players"`
	WinnerID      *uuid.UUID               `json:"winner_id,omitempty"`
	Duration      time.Duration            `json:"duration"`
	RatingChanges map[uuid.UUID]int        `json:"rating_changes,omitempty"`
	Rewards       map[uuid.UUID]*PvPReward `json:"rewards,omitempty"`
	CreatedAt     time.Time                `json:"created_at"`
	CompletedAt   *time.Time               `json:"completed_at,omitempty"`
}

// PvPReward représente les récompenses PvP
type PvPReward struct {
	RatingChange int                    `json:"rating_change"`
	Gold         int                    `json:"gold"`
	Experience   int                    `json:"experience"`
	Items        []RewardItem           `json:"items,omitempty"`
	Titles       []string               `json:"titles,omitempty"`
	Achievements []string               `json:"achievements,omitempty"`
	Custom       map[string]interface{} `json:"custom,omitempty"`
}

// PvPRanking représente le classement PvP
type PvPRanking struct {
	Rank       int        `json:"rank"`
	PlayerID   uuid.UUID  `json:"player_id"`
	PlayerName string     `json:"player_name"`
	Rating     int        `json:"rating"`
	Wins       int        `json:"wins"`
	Losses     int        `json:"losses"`
	Draws      int        `json:"draws"`
	WinRate    float64    `json:"win_rate"`
	Streak     int        `json:"streak"`
	LastMatch  *time.Time `json:"last_match,omitempty"`
}

// CreateChallengeRequest représente une demande de création de défi
type CreateChallengeRequest struct {
	ChallengedID  uuid.UUID     `json:"challenged_id" binding:"required"`
	ChallengeType ChallengeType `json:"challenge_type" binding:"required"`
	Message       string        `json:"message,omitempty"`
	Stakes        *PvPStakes    `json:"stakes,omitempty"`
	ExpiresIn     *int          `json:"expires_in,omitempty"` // en minutes
}

// RespondChallengeRequest représente une réponse à un défi
type RespondChallengeRequest struct {
	Accept  bool   `json:"accept" binding:"required"`
	Message string `json:"message,omitempty"`
}

// PvPQueueEntry représente une entrée dans la file d'attente PvP
type PvPQueueEntry struct {
	ID            uuid.UUID         `json:"id" db:"id"`
	PlayerID      uuid.UUID         `json:"player_id" db:"player_id"`
	QueueType     ChallengeType     `json:"queue_type" db:"queue_type"`
	ChallengeType ChallengeType     `json:"challenge_type" db:"challenge_type"` // Alias pour QueueType
	Rating        int               `json:"rating" db:"rating"`
	MinRating     int               `json:"min_rating" db:"min_rating"`
	MaxRating     int               `json:"max_rating" db:"max_rating"`
	QueuedAt      time.Time         `json:"queued_at" db:"queued_at"`
	JoinedAt      time.Time         `json:"joined_at" db:"joined_at"` // Alias pour QueuedAt
	UpdatedAt     time.Time         `json:"updated_at" db:"updated_at"`
	EstimatedWait time.Duration     `json:"estimated_wait" db:"-"` // Calculé
	Preferences   *QueuePreferences `json:"preferences,omitempty" db:"preferences"`
}

// GetChallengesRequest représente une demande de récupération de défis
type GetChallengesRequest struct {
	PlayerID uuid.UUID `json:"player_id" binding:"required"`
	Status   string    `json:"status,omitempty"` // "pending", "accepted", "completed", etc.
	Type     string    `json:"type,omitempty"`   // "sent", "received", "all"
	Limit    int       `json:"limit,omitempty"`
	Offset   int       `json:"offset,omitempty"`
}

// QueuePreferences représente les préférences de file d'attente
type QueuePreferences struct {
	RatingRange   *RatingRange `json:"rating_range,omitempty"`
	MaxWaitTime   *int         `json:"max_wait_time,omitempty"` // en minutes
	AllowBots     bool         `json:"allow_bots"`
	PreferredMaps []string     `json:"preferred_maps,omitempty"`
}

// RatingRange représente une plage de classement
type RatingRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

// PvPStatistics représente les statistiques PvP d'un joueur
type PvPStatistics struct {
	PlayerID         uuid.UUID  `json:"player_id" db:"player_id"`
	UserID           uuid.UUID  `json:"user_id" db:"user_id"`
	CurrentRating    int        `json:"current_rating" db:"current_rating"`
	HighestRating    int        `json:"highest_rating" db:"highest_rating"`
	BattlesWon       int        `json:"battles_won" db:"battles_won"`
	BattlesLost      int        `json:"battles_lost" db:"battles_lost"`
	Draws            int        `json:"draws" db:"draws"`
	TotalMatches     int        `json:"total_matches" db:"-"` // Calculé
	WinRate          float64    `json:"win_rate" db:"-"`      // Calculé
	CurrentStreak    int        `json:"current_streak" db:"current_streak"`
	BestStreak       int        `json:"best_streak" db:"best_streak"`
	TotalDamageDealt int64      `json:"total_damage_dealt" db:"total_damage_dealt"`
	TotalDamageTaken int64      `json:"total_damage_taken" db:"total_damage_taken"`
	TotalHealingDone int64      `json:"total_healing_done" db:"total_healing_done"`
	RankName         string     `json:"rank_name" db:"-"` // Calculé
	LastMatchAt      *time.Time `json:"last_match_at" db:"last_match_at"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// IsExpired vérifie si le défi a expiré
func (c *PvPChallenge) IsExpired() bool {
	return time.Now().After(c.ExpiresAt) && c.Status == ChallengeStatusPending
}

// CanAccept vérifie si le défi peut être accepté
func (c *PvPChallenge) CanAccept(playerID uuid.UUID) bool {
	return c.Status == ChallengeStatusPending &&
		c.ChallengedID == playerID &&
		!c.IsExpired()
}

// CanCancel vérifie si le défi peut être annulé
func (c *PvPChallenge) CanCancel(playerID uuid.UUID) bool {
	return c.Status == ChallengeStatusPending &&
		c.ChallengerID == playerID
}

// GetDuration retourne la durée du défi depuis sa création
func (c *PvPChallenge) GetDuration() time.Duration {
	if c.CompletedAt != nil {
		return c.CompletedAt.Sub(c.CreatedAt)
	}
	return time.Since(c.CreatedAt)
}

// CalculateRatingChange calcule le changement de classement après un match
func CalculateRatingChange(winnerRating, loserRating int, isWinner bool) int {
	// Constante K pour l'algorithme Elo
	k := config.DefaultEloK

	// Calcul de la probabilité de victoire attendue
	expectedScore := 1.0 / (1.0 + math.Pow(config.DefaultEloBase, float64(loserRating-winnerRating)/config.DefaultEloDivisor))

	if !isWinner {
		expectedScore = 1.0 - expectedScore
	}

	// Score réel (1 pour victoire, 0 pour défaite)
	actualScore := 0.0
	if isWinner {
		actualScore = 1.0
	}

	// Calcul du changement de classement
	change := k * (actualScore - expectedScore)

	return int(math.Round(change))
}

// GetRankFromRating retourne le nom du rang basé sur le rating
func GetRankFromRating(rating int) string {
	switch {
	case rating >= config.DefaultGrandMasterRating:
		return "Grand Master"
	case rating >= config.DefaultMasterRating:
		return "Master"
	case rating >= config.DefaultDiamondRating:
		return "Diamond"
	case rating >= config.DefaultPlatinumRating:
		return "Platinum"
	case rating >= config.DefaultGoldRating:
		return "Gold"
	case rating >= config.DefaultPvPRating:
		return "Silver"
	default:
		return "Bronze"
	}
}

// ValidateStakes valide les enjeux d'un défi
func (s *PvPStakes) Validate() error {
	// Vérification des limites
	if s.Gold < 0 {
		return fmt.Errorf("l'or misé ne peut pas être négatif")
	}

	if s.Gold > config.DefaultMaxGold {
		return fmt.Errorf("l'or misé ne peut pas dépasser %d", config.DefaultMaxGold)
	}

	if s.Experience < 0 {
		return fmt.Errorf("l'expérience misée ne peut pas être négative")
	}

	if len(s.Items) > config.DefaultMaxItems {
		return fmt.Errorf("impossible de miser plus de %d objets", config.DefaultMaxItems)
	}

	// Validation des objets
	for _, item := range s.Items {
		if item.ItemID == "" {
			return fmt.Errorf("ID d'objet invalide")
		}
		if item.Quantity <= 0 {
			return fmt.Errorf("la quantité d'objet doit être positive")
		}
	}

	return nil
}

// GetTotalValue calcule la valeur totale des enjeux
func (s *PvPStakes) GetTotalValue() int {
	total := s.Gold + s.Experience + s.Reputation

	// Estimation de la valeur des objets (à implémenter selon le système d'objets)
	for _, item := range s.Items {
		// Valeur estimée basée sur la rareté
		itemValue := 100 // Valeur par défaut
		switch item.Quality {
		case "common":
			itemValue = 50
		case "uncommon":
			itemValue = 100
		case "rare":
			itemValue = 250
		case "epic":
			itemValue = 500
		case "legendary":
			itemValue = 1000
		}
		total += itemValue * item.Quantity
	}

	return total
}

// CreateDefaultStakes crée des enjeux par défaut basés sur le niveau des joueurs
func CreateDefaultStakes(challengerLevel, challengedLevel int) *PvPStakes {
	avgLevel := (challengerLevel + challengedLevel) / config.DefaultLevelDivisor

	return &PvPStakes{
		Gold:       avgLevel * config.DefaultGoldMultiplier,
		Experience: avgLevel * config.DefaultExperienceMultiplier,
		Reputation: 1,
	}
}

// GetMatchmakingRange retourne la plage de classement pour le matchmaking
func GetMatchmakingRange(rating int, waitTime time.Duration) *RatingRange {
	// Base range de ±100 points
	baseRange := config.DefaultRatingRange

	// Élargir la plage en fonction du temps d'attente
	waitMinutes := int(waitTime.Minutes())
	expandedRange := baseRange + (waitMinutes * config.DefaultRangeMultiplier)

	// Limiter l'expansion maximale
	if expandedRange > config.DefaultExpandedRange {
		expandedRange = config.DefaultExpandedRange
	}

	return &RatingRange{
		Min: rating - expandedRange,
		Max: rating + expandedRange,
	}
}

// IsInRange vérifie si un classement est dans la plage
func (r *RatingRange) IsInRange(rating int) bool {
	return rating >= r.Min && rating <= r.Max
}

// GetSeasonInfo retourne les informations de la saison PvP actuelle
func GetSeasonInfo() *PvPSeasonInfo {
	return &PvPSeasonInfo{
		ID:          1,
		Name:        "Saison 1 - Les Premiers Combats",
		StartDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 3, 31, 23, 59, 59, 0, time.UTC),
		IsActive:    true,
		Rewards:     GetSeasonRewards(),
		Leaderboard: true,
	}
}

// PvPSeasonInfo représente les informations d'une saison PvP
type PvPSeasonInfo struct {
	ID          int                   `json:"id"`
	Name        string                `json:"name"`
	StartDate   time.Time             `json:"start_date"`
	EndDate     time.Time             `json:"end_date"`
	IsActive    bool                  `json:"is_active"`
	Rewards     map[string]*PvPReward `json:"rewards"`
	Leaderboard bool                  `json:"leaderboard"`
}

// GetSeasonRewards retourne les récompenses de fin de saison
func GetSeasonRewards() map[string]*PvPReward {
	return map[string]*PvPReward{
		"Grandmaster": {
			Gold:       config.DefaultMaxGoldReward,
			Experience: config.DefaultMaxExperienceReward,
			Items: []RewardItem{
				{ItemID: "legendary_pvp_mount", Quantity: 1, Quality: "legendary"},
				{ItemID: "grandmaster_title", Quantity: 1, Quality: "epic"},
			},
			Titles: []string{"Grandmaster"},
		},
		"Master": {
			Gold:       config.DefaultGoldReward2,
			Experience: config.DefaultExperienceReward2,
			Items: []RewardItem{
				{ItemID: "epic_pvp_weapon", Quantity: 1, Quality: "epic"},
				{ItemID: "master_title", Quantity: 1, Quality: "rare"},
			},
			Titles: []string{"Master"},
		},
		"Diamond": {
			Gold:       config.DefaultGoldReward3,
			Experience: config.DefaultExperienceReward3,
			Items: []RewardItem{
				{ItemID: "rare_pvp_armor", Quantity: 1, Quality: "rare"},
			},
			Titles: []string{"Diamond Fighter"},
		},
	}
}
