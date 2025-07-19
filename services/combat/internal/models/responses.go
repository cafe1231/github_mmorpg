package models

import (
	"time"

	"github.com/google/uuid"
)

// CombatResponse représente la réponse de création/mise à jour de combat
type CombatResponse struct {
	Success bool            `json:"success"`
	Combat  *CombatInstance `json:"combat,omitempty"`
	Message string          `json:"message,omitempty"`
	Error   string          `json:"error,omitempty"`
}

// CombatStatusResponse représente la réponse de statut de combat
type CombatStatusResponse struct {
	Combat           *CombatInstance      `json:"combat"`
	Participants     []*CombatParticipant `json:"participants,omitempty"`
	CurrentTurn      *TurnInfo            `json:"current_turn,omitempty"`
	AvailableActions []*ActionTemplate    `json:"available_actions,omitempty"`
	RecentActions    []*CombatAction      `json:"recent_actions,omitempty"`
	ActiveEffects    []*CombatEffect      `json:"active_effects,omitempty"`
	Logs             []*CombatLog         `json:"logs,omitempty"`
	NextUpdate       *time.Time           `json:"next_update,omitempty"`
}

// TurnInfo représente les informations du tour actuel
type TurnInfo struct {
	TurnNumber      int        `json:"turn_number"`
	CurrentPlayer   *uuid.UUID `json:"current_player,omitempty"`
	TimeRemaining   int        `json:"time_remaining_seconds"`
	TurnStartTime   time.Time  `json:"turn_start_time"`
	ActionsThisTurn int        `json:"actions_this_turn"`
	CanAct          bool       `json:"can_act"`
}

// ActionResponse représente la réponse d'une action de combat
type ActionResponse struct {
	Success      bool            `json:"success"`
	Action       *CombatAction   `json:"action,omitempty"`
	Effects      []*CombatEffect `json:"effects,omitempty"`
	StateChanges *StateChanges   `json:"state_changes,omitempty"`
	Logs         []*CombatLog    `json:"logs,omitempty"`
	TurnAdvanced bool            `json:"turn_advanced,omitempty"`
	CombatEnded  bool            `json:"combat_ended,omitempty"`
	Winner       *int            `json:"winner,omitempty"`
	Message      string          `json:"message,omitempty"`
	Error        string          `json:"error,omitempty"`
	Warnings     []string        `json:"warnings,omitempty"`
}

// EffectResponse représente la réponse d'application d'effet
type EffectResponse struct {
	Success        bool          `json:"success"`
	Effect         *CombatEffect `json:"effect,omitempty"`
	ExistingEffect *CombatEffect `json:"existing_effect,omitempty"`
	Action         string        `json:"action"` // "applied", "stacked", "refreshed", "resisted"
	Message        string        `json:"message,omitempty"`
	Error          string        `json:"error,omitempty"`
}

// PvPChallengeResponse représente la réponse de défi PvP
type PvPChallengeResponse struct {
	Success   bool          `json:"success"`
	Challenge *PvPChallenge `json:"challenge,omitempty"`
	Message   string        `json:"message,omitempty"`
	Error     string        `json:"error,omitempty"`
}

// CombatListResponse représente la réponse de liste de combats
type CombatListResponse struct {
	Combats  []*CombatListItem `json:"combats"` // <- Changer le type
	Total    int               `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
	HasMore  bool              `json:"has_more"`
	Filters  *SearchFilters    `json:"filters,omitempty"`
}

// SearchFilters représente les filtres de recherche appliqués
type SearchFilters struct {
	CombatType    *CombatType   `json:"combat_type,omitempty"`
	Status        *CombatStatus `json:"status,omitempty"`
	ZoneID        *string       `json:"zone_id,omitempty"`
	ParticipantID *uuid.UUID    `json:"participant_id,omitempty"`
	DateFrom      *time.Time    `json:"date_from,omitempty"`
	DateTo        *time.Time    `json:"date_to,omitempty"`
}

// CombatHistoryResponse représente la réponse d'historique de combat
type CombatHistoryResponse struct {
	History  []*CombatHistoryEntry `json:"history"`
	Total    int                   `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
	HasMore  bool                  `json:"has_more"`
	Summary  *HistorySummary       `json:"summary,omitempty"`
}

// CombatHistoryEntry représente une entrée d'historique
type CombatHistoryEntry struct {
	CombatID      uuid.UUID     `json:"combat_id"`
	CombatType    CombatType    `json:"combat_type"`
	Result        string        `json:"result"` // "win", "loss", "draw"
	Duration      time.Duration `json:"duration"`
	DamageDealt   int           `json:"damage_dealt"`
	DamageTaken   int           `json:"damage_taken"`
	HealingDone   int           `json:"healing_done"`
	OpponentCount int           `json:"opponent_count"`
	Opponents     []string      `json:"opponents,omitempty"`
	Rewards       *CombatReward `json:"rewards,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
}

// Renommer CombatSummary en CombatListItem
type CombatListItem struct {
	ID               uuid.UUID     `json:"id"`
	CombatType       CombatType    `json:"combat_type"`
	Status           CombatStatus  `json:"status"`
	ParticipantCount int           `json:"participant_count"`
	MaxParticipants  int           `json:"max_participants"`
	Duration         time.Duration `json:"duration"`
	CreatedAt        time.Time     `json:"created_at"`
	StartedAt        *time.Time    `json:"started_at,omitempty"`
	EndedAt          *time.Time    `json:"ended_at,omitempty"`
	WinnerTeam       *int          `json:"winner_team,omitempty"`
	ZoneID           *string       `json:"zone_id,omitempty"`
}

// HistorySummary représente un résumé de l'historique
type HistorySummary struct {
	TotalCombats    int     `json:"total_combats"`
	Wins            int     `json:"wins"`
	Losses          int     `json:"losses"`
	Draws           int     `json:"draws"`
	WinRate         float64 `json:"win_rate"`
	AverageDuration float64 `json:"average_duration_seconds"`
	TotalDamage     int64   `json:"total_damage"`
	TotalHealing    int64   `json:"total_healing"`
	BestStreak      int     `json:"best_streak"`
	CurrentStreak   int     `json:"current_streak"`
}

// StatisticsResponse représente la réponse de statistiques
type StatisticsResponse struct {
	CharacterID  uuid.UUID            `json:"character_id,omitempty"`
	UserID       uuid.UUID            `json:"user_id,omitempty"`
	Period       string               `json:"period"`
	General      *GeneralStats        `json:"general"`
	PvE          *PvEStats            `json:"pve,omitempty"`
	PvP          *PvPStats            `json:"pvp,omitempty"`
	Combat       *DetailedCombatStats `json:"combat,omitempty"`
	Trends       *StatsTrends         `json:"trends,omitempty"`
	Achievements []*Achievement       `json:"achievements,omitempty"`
	GeneratedAt  time.Time            `json:"generated_at"`
}

// GeneralStats représente les statistiques générales
type GeneralStats struct {
	TotalCombats       int           `json:"total_combats"`
	TotalDuration      time.Duration `json:"total_duration"`
	AverageDuration    time.Duration `json:"average_duration"`
	TotalDamageDealt   int64         `json:"total_damage_dealt"`
	TotalDamageTaken   int64         `json:"total_damage_taken"`
	TotalHealingDone   int64         `json:"total_healing_done"`
	TotalDeaths        int           `json:"total_deaths"`
	HighestDamageDealt int           `json:"highest_damage_dealt"`
	LongestCombat      time.Duration `json:"longest_combat"`
	FavoriteTime       string        `json:"favorite_time_of_day,omitempty"`
}

// PvEStats représente les statistiques PvE
type PvEStats struct {
	BattlesWon       int     `json:"battles_won"`
	BattlesLost      int     `json:"battles_lost"`
	WinRate          float64 `json:"win_rate"`
	MonstersKilled   int     `json:"monsters_killed"`
	BossesKilled     int     `json:"bosses_killed"`
	DungeonsCleared  int     `json:"dungeons_cleared"`
	RaidsCompleted   int     `json:"raids_completed"`
	ExperienceGained int64   `json:"experience_gained"`
	GoldEarned       int64   `json:"gold_earned"`
	ItemsLooted      int     `json:"items_looted"`
}

// PvPStats représente les statistiques PvP
type PvPStats struct {
	BattlesWon     int     `json:"battles_won"`
	BattlesLost    int     `json:"battles_lost"`
	Draws          int     `json:"draws"`
	WinRate        float64 `json:"win_rate"`
	CurrentRating  int     `json:"current_rating"`
	HighestRating  int     `json:"highest_rating"`
	RankPosition   int     `json:"rank_position,omitempty"`
	RankName       string  `json:"rank_name"`
	CurrentStreak  int     `json:"current_streak"`
	BestStreak     int     `json:"best_streak"`
	TournamentWins int     `json:"tournament_wins"`
}

// DetailedCombatStats représente des statistiques de combat détaillées
type DetailedCombatStats struct {
	ActionsPerformed int           `json:"actions_performed"`
	SkillsUsed       int           `json:"skills_used"`
	ItemsUsed        int           `json:"items_used"`
	CriticalHits     int           `json:"critical_hits"`
	CriticalRate     float64       `json:"critical_rate"`
	AccuracyRate     float64       `json:"accuracy_rate"`
	BlockRate        float64       `json:"block_rate"`
	EffectsApplied   int           `json:"effects_applied"`
	EffectsReceived  int           `json:"effects_received"`
	TurnsPlayed      int           `json:"turns_played"`
	AverageTurnTime  time.Duration `json:"average_turn_time"`
	FavoriteActions  []ActionStats `json:"favorite_actions,omitempty"`
	FavoriteSkills   []SkillStats  `json:"favorite_skills,omitempty"`
}

// ActionStats représente les statistiques d'une action
type ActionStats struct {
	ActionType  ActionType `json:"action_type"`
	Count       int        `json:"count"`
	SuccessRate float64    `json:"success_rate"`
	AvgDamage   float64    `json:"avg_damage,omitempty"`
}

// SkillStats représente les statistiques d'une compétence
type SkillStats struct {
	SkillID     string  `json:"skill_id"`
	SkillName   string  `json:"skill_name"`
	Count       int     `json:"count"`
	SuccessRate float64 `json:"success_rate"`
	AvgDamage   float64 `json:"avg_damage,omitempty"`
	AvgHealing  float64 `json:"avg_healing,omitempty"`
}

// StatsTrends représente les tendances statistiques
type StatsTrends struct {
	WinRateTrend     string   `json:"win_rate_trend"` // "up", "down", "stable"
	DamageTrend      string   `json:"damage_trend"`
	PerformanceTrend string   `json:"performance_trend"`
	ActivityTrend    string   `json:"activity_trend"`
	ImprovementScore float64  `json:"improvement_score"`
	RecommendedFocus []string `json:"recommended_focus,omitempty"`
}

// Achievement représente un succès/achievement
type Achievement struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Icon        string              `json:"icon,omitempty"`
	Category    string              `json:"category"`
	Progress    int                 `json:"progress"`
	Target      int                 `json:"target"`
	Completed   bool                `json:"completed"`
	CompletedAt *time.Time          `json:"completed_at,omitempty"`
	Rewards     *AchievementRewards `json:"rewards,omitempty"`
}

// AchievementRewards représente les récompenses d'un succès
type AchievementRewards struct {
	Experience int          `json:"experience,omitempty"`
	Gold       int          `json:"gold,omitempty"`
	Items      []RewardItem `json:"items,omitempty"`
	Title      string       `json:"title,omitempty"`
}

// ReplayResponse représente la réponse de rejeu
type ReplayResponse struct {
	Success  bool            `json:"success"`
	ReplayID uuid.UUID       `json:"replay_id,omitempty"`
	Events   []*ReplayEvent  `json:"events,omitempty"`
	Duration time.Duration   `json:"duration"`
	Metadata *ReplayMetadata `json:"metadata,omitempty"`
	Error    string          `json:"error,omitempty"`
}

// ReplayEvent représente un événement de rejeu
type ReplayEvent struct {
	Timestamp   time.Time   `json:"timestamp"`
	TurnNumber  int         `json:"turn_number"`
	EventType   string      `json:"event_type"`
	ActorID     *uuid.UUID  `json:"actor_id,omitempty"`
	TargetID    *uuid.UUID  `json:"target_id,omitempty"`
	Data        interface{} `json:"data"`
	Description string      `json:"description"`
}

// ReplayMetadata représente les métadonnées de rejeu
type ReplayMetadata struct {
	CombatID  uuid.UUID     `json:"combat_id"`
	Version   string        `json:"version"`
	Duration  time.Duration `json:"duration"`
	Players   []string      `json:"players"`
	Winner    string        `json:"winner,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
}

// ValidationResponse représente la réponse de validation d'action
// ValidationResponse représente la réponse de validation d'action
type ValidationResponse struct {
	Valid        bool             `json:"valid"`
	Action       *CombatAction    `json:"action,omitempty"`
	Warnings     []string         `json:"warnings,omitempty"`
	Errors       []string         `json:"errors,omitempty"`
	Suggestions  []string         `json:"suggestions,omitempty"`
	CanExecute   bool             `json:"can_execute"`
	Cooldown     *time.Duration   `json:"cooldown,omitempty"`
	RequiredMana int              `json:"required_mana,omitempty"`
	Message      string           `json:"message,omitempty"`
	AntiCheat    *AntiCheatResult `json:"anti_cheat,omitempty"` // <- AJOUTER CE CHAMP
}

// AntiCheatResult représente le résultat de la validation anti-triche
type AntiCheatResult struct {
	Suspicious      bool     `json:"suspicious"`
	Score           float64  `json:"score"` // 0-100, plus élevé = plus suspect
	Flags           []string `json:"flags,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
	Action          string   `json:"action"` // "allow", "warn", "block"
}

// ErrorResponse représente une réponse d'erreur standardisée
type ErrorResponse struct {
	Error     string    `json:"error"`
	Code      string    `json:"code,omitempty"`
	Details   string    `json:"details,omitempty"`
	RequestID string    `json:"request_id,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// SuccessResponse représente une réponse de succès simple
type SuccessResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// CreateSuccessResponse crée une réponse de succès
func CreateSuccessResponse(message string, data interface{}) *SuccessResponse {
	return &SuccessResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// CreateErrorResponse crée une réponse d'erreur
func CreateErrorResponse(errorMsg, code, details, requestID string) *ErrorResponse {
	return &ErrorResponse{
		Error:     errorMsg,
		Code:      code,
		Details:   details,
		RequestID: requestID,
		Timestamp: time.Now(),
	}
}
