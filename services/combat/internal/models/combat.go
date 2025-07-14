// internal/models/combat.go
package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// CombatSession représente une session de combat active
type CombatSession struct {
	ID              uuid.UUID            `json:"id" db:"id"`
	Type            string               `json:"type" db:"type"` // pve, pvp, raid
	Status          string               `json:"status" db:"status"` // waiting, active, ended, cancelled
	ZoneID          string               `json:"zone_id" db:"zone_id"`
	CreatedBy       uuid.UUID            `json:"created_by" db:"created_by"`
	
	// Configuration
	MaxParticipants int                  `json:"max_participants" db:"max_participants"`
	IsPrivate       bool                 `json:"is_private" db:"is_private"`
	LevelRange      LevelRange           `json:"level_range" db:"level_range"`
	Rules           CombatRules          `json:"rules" db:"rules"`
	
	// Timing
	StartedAt       *time.Time           `json:"started_at" db:"started_at"`
	EndedAt         *time.Time           `json:"ended_at" db:"ended_at"`
	LastActionAt    time.Time            `json:"last_action_at" db:"last_action_at"`
	
	// Relations (chargées séparément)
	Participants    []CombatParticipant  `json:"participants,omitempty" db:"-"`
	Actions         []CombatAction       `json:"actions,omitempty" db:"-"`
	
	CreatedAt       time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at" db:"updated_at"`
}

// CombatParticipant représente un participant dans un combat
type CombatParticipant struct {
	ID              uuid.UUID    `json:"id" db:"id"`
	SessionID       uuid.UUID    `json:"session_id" db:"session_id"`
	CharacterID     uuid.UUID    `json:"character_id" db:"character_id"`
	PlayerID        uuid.UUID    `json:"player_id" db:"player_id"`
	
	// État du participant
	Team            int          `json:"team" db:"team"` // 0=neutral, 1=team1, 2=team2
	Position        Position     `json:"position" db:"position"`
	Status          string       `json:"status" db:"status"` // alive, dead, fled, disconnected
	
	// Stats de combat actuelles
	CurrentHealth   int          `json:"current_health" db:"current_health"`
	MaxHealth       int          `json:"max_health" db:"max_health"`
	CurrentMana     int          `json:"current_mana" db:"current_mana"`
	MaxMana         int          `json:"max_mana" db:"max_mana"`
	
	// Stats calculées avec les effets
	Damage          int          `json:"damage" db:"damage"`
	Defense         int          `json:"defense" db:"defense"`
	CritChance      float64      `json:"crit_chance" db:"crit_chance"`
	AttackSpeed     float64      `json:"attack_speed" db:"attack_speed"`
	
	// Timing
	JoinedAt        time.Time    `json:"joined_at" db:"joined_at"`
	LastActionAt    *time.Time   `json:"last_action_at" db:"last_action_at"`
	
	// Relations
	Character       *Character   `json:"character,omitempty" db:"-"`
	ActiveEffects   []StatusEffect `json:"active_effects,omitempty" db:"-"`
}

// CombatAction représente une action de combat
type CombatAction struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	SessionID       uuid.UUID       `json:"session_id" db:"session_id"`
	ActorID         uuid.UUID       `json:"actor_id" db:"actor_id"` // CharacterID qui fait l'action
	
	// Type d'action
	Type            string          `json:"type" db:"type"` // attack, spell, move, item, defend
	ActionData      json.RawMessage `json:"action_data" db:"action_data"`
	
	// Cibles
	Targets         []uuid.UUID     `json:"targets" db:"targets"`
	
	// Résultats
	Results         []ActionResult  `json:"results" db:"results"`
	Success         bool            `json:"success" db:"success"`
	CriticalHit     bool            `json:"critical_hit" db:"critical_hit"`
	
	// Timing
	ExecutedAt      time.Time       `json:"executed_at" db:"executed_at"`
	Duration        time.Duration   `json:"duration" db:"duration"`
}

// ActionResult représente le résultat d'une action sur une cible
type ActionResult struct {
	TargetID        uuid.UUID `json:"target_id"`
	DamageDealt     int       `json:"damage_dealt"`
	HealingDone     int       `json:"healing_done"`
	EffectsApplied  []string  `json:"effects_applied"`
	EffectsRemoved  []string  `json:"effects_removed"`
	Absorbed        int       `json:"absorbed"` // dégâts absorbés par des boucliers
	Reflected       int       `json:"reflected"` // dégâts renvoyés
}

// Position représente une position dans l'espace de combat
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

// LevelRange représente une plage de niveaux
type LevelRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

// CombatRules représente les règles spécifiques d'un combat
type CombatRules struct {
	FriendlyFire    bool          `json:"friendly_fire"`
	AllowItems      bool          `json:"allow_items"`
	AllowFleeing    bool          `json:"allow_fleeing"`
	TurnBased       bool          `json:"turn_based"`
	TimeLimit       time.Duration `json:"time_limit"`
	RespawnAllowed  bool          `json:"respawn_allowed"`
	ExperienceGain  float64       `json:"experience_gain"` // multiplicateur d'XP
}

// Character informations de base du personnage (depuis Player service)
type Character struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Level       int       `json:"level"`
	Class       string    `json:"class"`
	Race        string    `json:"race"`
	
	// Stats de base
	Strength     int `json:"strength"`
	Agility      int `json:"agility"`
	Intelligence int `json:"intelligence"`
	Vitality     int `json:"vitality"`
	
	// Stats calculées
	MaxHealth    int `json:"max_health"`
	MaxMana      int `json:"max_mana"`
	PhysicalDamage int `json:"physical_damage"`
	MagicalDamage  int `json:"magical_damage"`
	PhysicalDefense int `json:"physical_defense"`
	MagicalDefense  int `json:"magical_defense"`
	CriticalChance  int `json:"critical_chance"`
	AttackSpeed     int `json:"attack_speed"`
	MovementSpeed   int `json:"movement_speed"`
}

// DTOs pour les requêtes

// StartCombatRequest requête pour démarrer un combat
type StartCombatRequest struct {
	Type            string      `json:"type" binding:"required,oneof=pve pvp raid"`
	ZoneID          string      `json:"zone_id" binding:"required"`
	MaxParticipants int         `json:"max_participants" binding:"min=2,max=20"`
	IsPrivate       bool        `json:"is_private"`
	LevelRange      LevelRange  `json:"level_range"`
	Rules           CombatRules `json:"rules"`
	Password        string      `json:"password,omitempty"`
}

// JoinCombatRequest requête pour rejoindre un combat
type JoinCombatRequest struct {
	CharacterID uuid.UUID `json:"character_id" binding:"required"`
	Team        int       `json:"team" binding:"min=0,max=2"`
	Password    string    `json:"password,omitempty"`
}

// PerformActionRequest requête pour effectuer une action
type PerformActionRequest struct {
	SessionID   uuid.UUID       `json:"session_id" binding:"required"`
	CharacterID uuid.UUID       `json:"character_id" binding:"required"`
	Type        string          `json:"type" binding:"required,oneof=attack spell move item defend flee"`
	Targets     []uuid.UUID     `json:"targets"`
	Position    *Position       `json:"position,omitempty"`
	ActionData  json.RawMessage `json:"action_data,omitempty"`
}

// AttackActionData données spécifiques à une attaque
type AttackActionData struct {
	WeaponType   string  `json:"weapon_type"`
	AttackType   string  `json:"attack_type"` // melee, ranged, special
	PowerLevel   float64 `json:"power_level"` // 0.1 à 1.0 pour ajuster la puissance
	AimingTime   int     `json:"aiming_time"` // temps de visée en ms
}

// SpellActionData données spécifiques à un sort
type SpellActionData struct {
	SpellID     uuid.UUID `json:"spell_id" binding:"required"`
	PowerLevel  float64   `json:"power_level"` // 0.1 à 1.0
	Concentration bool    `json:"concentration"` // sort à concentration
}

// MoveActionData données spécifiques à un déplacement
type MoveActionData struct {
	TargetPosition Position `json:"target_position" binding:"required"`
	MovementType   string   `json:"movement_type"` // walk, run, teleport
}

// ItemActionData données spécifiques à l'utilisation d'un objet
type ItemActionData struct {
	ItemID   uuid.UUID `json:"item_id" binding:"required"`
	Quantity int       `json:"quantity" binding:"min=1"`
}

// CombatStatusResponse réponse du statut d'un combat
type CombatStatusResponse struct {
	Session      CombatSession         `json:"session"`
	Participants []CombatParticipant   `json:"participants"`
	RecentActions []CombatAction       `json:"recent_actions"`
	TurnOrder    []uuid.UUID           `json:"turn_order,omitempty"`
	CurrentTurn  *uuid.UUID            `json:"current_turn,omitempty"`
	TimeLeft     *time.Duration        `json:"time_left,omitempty"`
}

// CombatStatistics statistiques de combat
type CombatStatistics struct {
	TotalCombats        int           `json:"total_combats"`
	ActiveCombats       int           `json:"active_combats"`
	AverageSessionTime  time.Duration `json:"average_session_time"`
	MostUsedSpells      []SpellUsage  `json:"most_used_spells"`
	TopDamageDealer     string        `json:"top_damage_dealer"`
	PvPMatches          int           `json:"pvp_matches"`
	PvEEncounters       int           `json:"pve_encounters"`
}

// SpellUsage statistiques d'utilisation de sorts
type SpellUsage struct {
	SpellID   uuid.UUID `json:"spell_id"`
	SpellName string    `json:"spell_name"`
	UseCount  int       `json:"use_count"`
}