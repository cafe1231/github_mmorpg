package models

import (
	"time"

	"combat/internal/config"

	"github.com/google/uuid"
)

// CombatType définit les types de combat possibles
type CombatType string

const (
	CombatTypePvE     CombatType = "pve"
	CombatTypePvP     CombatType = "pvp"
	CombatTypeDungeon CombatType = "dungeon"
	CombatTypeRaid    CombatType = "raid"
)

// CombatStatus définit les status d'un combat
type CombatStatus string

const (
	CombatStatusWaiting   CombatStatus = "waiting"
	CombatStatusActive    CombatStatus = "active"
	CombatStatusPaused    CombatStatus = "paused"
	CombatStatusFinished  CombatStatus = "finished"
	CombatStatusCancelled CombatStatus = "canceled"
)

// CombatInstance représente une instance de combat
type CombatInstance struct {
	ID              uuid.UUID      `json:"id" db:"id"`
	CombatType      CombatType     `json:"combat_type" db:"combat_type"`
	Status          CombatStatus   `json:"status" db:"status"`
	ZoneID          *string        `json:"zone_id" db:"zone_id"`
	MaxParticipants int            `json:"max_participants" db:"max_participants"`
	CurrentTurn     int            `json:"current_turn" db:"current_turn"`
	TurnTimeLimit   int            `json:"turn_time_limit" db:"turn_time_limit"`
	MaxDuration     int            `json:"max_duration" db:"max_duration"`
	Settings        CombatSettings `json:"settings" db:"settings"`
	CreatedAt       time.Time      `json:"created_at" db:"created_at"`
	StartedAt       *time.Time     `json:"started_at" db:"started_at"`
	EndedAt         *time.Time     `json:"ended_at" db:"ended_at"`
	UpdatedAt       time.Time      `json:"updated_at" db:"updated_at"`

	// Relations (chargées séparément)
	Participants []*CombatParticipant `json:"participants,omitempty" db:"-"`
	Actions      []*CombatAction      `json:"actions,omitempty" db:"-"`
	Effects      []*CombatEffect      `json:"effects,omitempty" db:"-"`
	Logs         []*CombatLog         `json:"logs,omitempty" db:"-"`
}

// CombatSettings représente les paramètres d'un combat
type CombatSettings struct {
	AllowItems     bool                   `json:"allow_items"`
	AllowFlee      bool                   `json:"allow_flee"`
	AllowSurrender bool                   `json:"allow_surrender"`
	TeamDamage     bool                   `json:"team_damage"`
	RespawnEnabled bool                   `json:"respawn_enabled"`
	ExperienceGain bool                   `json:"experience_gain"`
	LootEnabled    bool                   `json:"loot_enabled"`
	CustomRules    map[string]interface{} `json:"custom_rules,omitempty"`
}

// CombatParticipant représente un participant dans un combat
type CombatParticipant struct {
	ID          uuid.UUID `json:"id" db:"id"`
	CombatID    uuid.UUID `json:"combat_id" db:"combat_id"`
	CharacterID uuid.UUID `json:"character_id" db:"character_id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Team        int       `json:"team" db:"team"`
	Position    int       `json:"position" db:"position"`

	// Stats de combat
	Health    int `json:"health" db:"health"`
	MaxHealth int `json:"max_health" db:"max_health"`
	Mana      int `json:"mana" db:"mana"`
	MaxMana   int `json:"max_mana" db:"max_mana"`

	// Stats calculées
	PhysicalDamage  int     `json:"physical_damage" db:"physical_damage"`
	MagicalDamage   int     `json:"magical_damage" db:"magical_damage"`
	PhysicalDefense int     `json:"physical_defense" db:"physical_defense"`
	MagicalDefense  int     `json:"magical_defense" db:"magical_defense"`
	CriticalChance  float64 `json:"critical_chance" db:"critical_chance"`
	AttackSpeed     float64 `json:"attack_speed" db:"attack_speed"`

	// État
	IsAlive      bool       `json:"is_alive" db:"is_alive"`
	IsReady      bool       `json:"is_ready" db:"is_ready"`
	LastActionAt *time.Time `json:"last_action_at" db:"last_action_at"`

	// Résultats
	DamageDealt int `json:"damage_dealt" db:"damage_dealt"`
	DamageTaken int `json:"damage_taken" db:"damage_taken"`
	HealingDone int `json:"healing_done" db:"healing_done"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// Relations (chargées séparément)
	Character     *CharacterSummary `json:"character,omitempty" db:"-"`
	ActiveEffects []*CombatEffect   `json:"active_effects,omitempty" db:"-"`
}

// CharacterSummary représente un résumé des informations d'un personnage
type CharacterSummary struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Level  int       `json:"level"`
	Class  string    `json:"class"`
	Avatar string    `json:"avatar,omitempty"`
}

// CombatStatistics représente les statistiques de combat d'un personnage
type CombatStatistics struct {
	ID          uuid.UUID `json:"id" db:"id"`
	CharacterID uuid.UUID `json:"character_id" db:"character_id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`

	// Statistiques PvE
	PvEBattlesWon  int `json:"pve_battles_won" db:"pve_battles_won"`
	PvEBattlesLost int `json:"pve_battles_lost" db:"pve_battles_lost"`
	MonstersKilled int `json:"monsters_killed" db:"monsters_killed"`
	BossesKilled   int `json:"bosses_killed" db:"bosses_killed"`

	// Statistiques PvP
	PvPBattlesWon  int `json:"pvp_battles_won" db:"pvp_battles_won"`
	PvPBattlesLost int `json:"pvp_battles_lost" db:"pvp_battles_lost"`
	PvPDraws       int `json:"pvp_draws" db:"pvp_draws"`
	PvPRating      int `json:"pvp_rating" db:"pvp_rating"`

	// Statistiques générales
	TotalDamageDealt int64 `json:"total_damage_dealt" db:"total_damage_dealt"`
	TotalDamageTaken int64 `json:"total_damage_taken" db:"total_damage_taken"`
	TotalHealingDone int64 `json:"total_healing_done" db:"total_healing_done"`
	TotalDeaths      int   `json:"total_deaths" db:"total_deaths"`

	// Records
	HighestDamageDealt    int `json:"highest_damage_dealt" db:"highest_damage_dealt"`
	LongestCombatDuration int `json:"longest_combat_duration" db:"longest_combat_duration"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CombatResult représente le résultat d'un combat
type CombatResult struct {
	CombatID     uuid.UUID                   `json:"combat_id"`
	Status       CombatStatus                `json:"status"`
	WinningTeam  *int                        `json:"winning_team,omitempty"`
	Duration     time.Duration               `json:"duration"`
	Participants []*CombatParticipant        `json:"participants"`
	Rewards      map[uuid.UUID]*CombatReward `json:"rewards,omitempty"`
	EndReason    string                      `json:"end_reason"`
	Summary      *CombatSummary              `json:"summary"`
}

// CombatReward représente les récompenses d'un combat
type CombatReward struct {
	Experience int            `json:"experience"`
	Gold       int            `json:"gold"`
	Items      []RewardItem   `json:"items,omitempty"`
	Reputation map[string]int `json:"reputation,omitempty"`
	PvPRating  int            `json:"pvp_rating,omitempty"`
}

// RewardItem représente un objet récompense
type RewardItem struct {
	ItemID   string `json:"item_id"`
	Quantity int    `json:"quantity"`
	Quality  string `json:"quality,omitempty"`
}

// CombatSummary représente un résumé des statistiques du combat
type CombatSummary struct {
	TotalTurns   int                             `json:"total_turns"`
	TotalDamage  int64                           `json:"total_damage"`
	TotalHealing int64                           `json:"total_healing"`
	ActionsCount int                             `json:"actions_count"`
	PlayerStats  map[uuid.UUID]*ParticipantStats `json:"player_stats"`
}

// ParticipantStats représente les statistiques d'un participant
type ParticipantStats struct {
	DamageDealt  int     `json:"damage_dealt"`
	DamageTaken  int     `json:"damage_taken"`
	HealingDone  int     `json:"healing_done"`
	ActionsUsed  int     `json:"actions_used"`
	CriticalHits int     `json:"critical_hits"`
	AccuracyRate float64 `json:"accuracy_rate"`
}

// IsActive vérifie si le combat est actif
func (c *CombatInstance) IsActive() bool {
	return c.Status == CombatStatusActive || c.Status == CombatStatusWaiting
}

// IsFinished vérifie si le combat est terminé
func (c *CombatInstance) IsFinished() bool {
	return c.Status == CombatStatusFinished || c.Status == CombatStatusCancelled
}

// GetDuration retourne la durée du combat
func (c *CombatInstance) GetDuration() time.Duration {
	if c.StartedAt == nil {
		return 0
	}

	endTime := time.Now()
	if c.EndedAt != nil {
		endTime = *c.EndedAt
	}

	return endTime.Sub(*c.StartedAt)
}

// GetAliveParticipants retourne les participants encore en vie
func (c *CombatInstance) GetAliveParticipants() []*CombatParticipant {
	var alive []*CombatParticipant
	for _, p := range c.Participants {
		if p.IsAlive {
			alive = append(alive, p)
		}
	}
	return alive
}

// GetTeamParticipants retourne les participants d'une équipe
func (c *CombatInstance) GetTeamParticipants(team int) []*CombatParticipant {
	var teamParticipants []*CombatParticipant
	for _, p := range c.Participants {
		if p.Team == team {
			teamParticipants = append(teamParticipants, p)
		}
	}
	return teamParticipants
}

// IsTeamAlive vérifie si une équipe a encore des participants en vie
func (c *CombatInstance) IsTeamAlive(team int) bool {
	for _, p := range c.Participants {
		if p.Team == team && p.IsAlive {
			return true
		}
	}
	return false
}

// GetHealthPercentage retourne le pourcentage de vie d'un participant
func (p *CombatParticipant) GetHealthPercentage() float64 {
	if p.MaxHealth == 0 {
		return 0
	}
	return float64(p.Health) / float64(p.MaxHealth) * config.DefaultHealthPercentage
}

// GetManaPercentage retourne le pourcentage de mana d'un participant
func (p *CombatParticipant) GetManaPercentage() float64 {
	if p.MaxMana == 0 {
		return 0
	}
	return float64(p.Mana) / float64(p.MaxMana) * config.DefaultManaPercentage
}

// GetDefaultCombatSettings retourne les paramètres par défaut d'un combat
func GetDefaultCombatSettings() CombatSettings {
	return CombatSettings{
		AllowItems:     true,
		AllowFlee:      true,
		AllowSurrender: true,
		TeamDamage:     false,
		RespawnEnabled: false,
		ExperienceGain: true,
		LootEnabled:    true,
		CustomRules:    make(map[string]interface{}),
	}
}

// CombatLog représente un log d'événement de combat
type CombatLog struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	CombatID   uuid.UUID  `json:"combat_id" db:"combat_id"`
	LogType    string     `json:"log_type" db:"log_type"` // "action", "effect", "damage", "healing", "death", etc.
	ActorID    *uuid.UUID `json:"actor_id" db:"actor_id"`
	ActorName  string     `json:"actor_name" db:"actor_name"`
	TargetID   *uuid.UUID `json:"target_id" db:"target_id"`
	TargetName string     `json:"target_name" db:"target_name"`
	Message    string     `json:"message" db:"message"`
	TurnNumber *int       `json:"turn_number" db:"turn_number"`
	Timestamp  time.Time  `json:"timestamp" db:"timestamp"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}
