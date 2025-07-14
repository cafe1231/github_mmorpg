// internal/models/spell.go
package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Spell représente un sort dans le jeu
type Spell struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	Name            string          `json:"name" db:"name"`
	Description     string          `json:"description" db:"description"`
	
	// Classification
	School          string          `json:"school" db:"school"` // fire, water, earth, air, dark, light, nature
	Type            string          `json:"type" db:"type"` // damage, heal, buff, debuff, utility
	TargetType      string          `json:"target_type" db:"target_type"` // self, single, aoe, line
	
	// Pré-requis
	RequiredLevel   int             `json:"required_level" db:"required_level"`
	RequiredClass   []string        `json:"required_class" db:"required_class"`
	RequiredStats   StatRequirement `json:"required_stats" db:"required_stats"`
	
	// Coûts
	ManaCost        int             `json:"mana_cost" db:"mana_cost"`
	HealthCost      int             `json:"health_cost" db:"health_cost"`
	MaterialCost    []MaterialCost  `json:"material_cost" db:"material_cost"`
	
	// Timing
	CastTime        time.Duration   `json:"cast_time" db:"cast_time"`
	Cooldown        time.Duration   `json:"cooldown" db:"cooldown"`
	Duration        time.Duration   `json:"duration" db:"duration"` // pour les sorts persistants
	
	// Effets
	Effects         []SpellEffect   `json:"effects" db:"effects"`
	
	// Propriétés
	Range           float64         `json:"range" db:"range"` // portée en mètres
	Radius          float64         `json:"radius" db:"radius"` // rayon d'effet pour AoE
	Accuracy        float64         `json:"accuracy" db:"accuracy"` // chance de réussir (0-1)
	CanCrit         bool            `json:"can_crit" db:"can_crit"`
	IsChanneled     bool            `json:"is_channeled" db:"is_channeled"`
	RequiresTarget  bool            `json:"requires_target" db:"requires_target"`
	
	// Métadonnées
	Icon            string          `json:"icon" db:"icon"`
	SoundEffect     string          `json:"sound_effect" db:"sound_effect"`
	VisualEffect    string          `json:"visual_effect" db:"visual_effect"`
	
	// État
	IsActive        bool            `json:"is_active" db:"is_active"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

// SpellEffect représente un effet d'un sort
type SpellEffect struct {
	Type            string          `json:"type"` // damage, heal, buff, debuff, summon, teleport
	Value           int             `json:"value"` // valeur de base
	Scaling         StatScaling     `json:"scaling"` // échelle avec les stats
	DamageType      string          `json:"damage_type,omitempty"` // physical, magical, true
	Element         string          `json:"element,omitempty"` // fire, ice, lightning, etc.
	StatusEffect    string          `json:"status_effect,omitempty"` // poison, burn, freeze
	Duration        time.Duration   `json:"duration,omitempty"`
	Probability     float64         `json:"probability"` // chance d'appliquer l'effet (0-1)
}

// StatScaling représente comment un effet scale avec les statistiques
type StatScaling struct {
	Strength     float64 `json:"strength"`
	Agility      float64 `json:"agility"`
	Intelligence float64 `json:"intelligence"`
	Vitality     float64 `json:"vitality"`
}

// StatRequirement représente les stats requises pour un sort
type StatRequirement struct {
	MinStrength     int `json:"min_strength"`
	MinAgility      int `json:"min_agility"`
	MinIntelligence int `json:"min_intelligence"`
	MinVitality     int `json:"min_vitality"`
}

// MaterialCost représente le coût en matériaux d'un sort
type MaterialCost struct {
	ItemID   uuid.UUID `json:"item_id"`
	Quantity int       `json:"quantity"`
}

// CharacterSpell représente un sort appris par un personnage
type CharacterSpell struct {
	ID          uuid.UUID `json:"id" db:"id"`
	CharacterID uuid.UUID `json:"character_id" db:"character_id"`
	SpellID     uuid.UUID `json:"spell_id" db:"spell_id"`
	
	// Progression
	Level       int       `json:"level" db:"level"` // niveau de maîtrise du sort
	Experience  int       `json:"experience" db:"experience"`
	
	// Personnalisation
	Customizations json.RawMessage `json:"customizations" db:"customizations"`
	
	// Raccourcis
	Hotkey      string    `json:"hotkey" db:"hotkey"`
	SlotNumber  int       `json:"slot_number" db:"slot_number"`
	
	LearnedAt   time.Time `json:"learned_at" db:"learned_at"`
	
	// Relations
	Spell       *Spell    `json:"spell,omitempty" db:"-"`
}

// SpellCooldown représente le cooldown d'un sort pour un personnage
type SpellCooldown struct {
	ID          uuid.UUID `json:"id" db:"id"`
	CharacterID uuid.UUID `json:"character_id" db:"character_id"`
	SpellID     uuid.UUID `json:"spell_id" db:"spell_id"`
	
	// Timing
	StartedAt   time.Time `json:"started_at" db:"started_at"`
	EndsAt      time.Time `json:"ends_at" db:"ends_at"`
	Duration    time.Duration `json:"duration" db:"duration"`
	
	// État
	IsActive    bool      `json:"is_active" db:"is_active"`
}

// internal/models/effect.go
// StatusEffect représente un effet de statut actif sur un personnage
type StatusEffect struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	CharacterID     uuid.UUID       `json:"character_id" db:"character_id"`
	SessionID       *uuid.UUID      `json:"session_id" db:"session_id"` // null si hors combat
	
	// Type et source
	Type            string          `json:"type" db:"type"` // buff, debuff, poison, burn, etc.
	Source          string          `json:"source" db:"source"` // spell, item, environment
	SourceID        *uuid.UUID      `json:"source_id" db:"source_id"` // ID du sort/item qui a causé l'effet
	CasterID        *uuid.UUID      `json:"caster_id" db:"caster_id"` // qui a lancé l'effet
	
	// Propriétés
	Name            string          `json:"name" db:"name"`
	Description     string          `json:"description" db:"description"`
	Icon            string          `json:"icon" db:"icon"`
	
	// Effets sur les stats
	StatModifiers   StatModifiers   `json:"stat_modifiers" db:"stat_modifiers"`
	
	// Effets périodiques
	PeriodicEffect  *PeriodicEffect `json:"periodic_effect" db:"periodic_effect"`
	
	// Timing
	StartedAt       time.Time       `json:"started_at" db:"started_at"`
	EndsAt          *time.Time      `json:"ends_at" db:"ends_at"` // null pour les effets permanents
	Duration        *time.Duration  `json:"duration" db:"duration"`
	
	// Stack
	StackCount      int             `json:"stack_count" db:"stack_count"`
	MaxStacks       int             `json:"max_stacks" db:"max_stacks"`
	
	// Propriétés spéciales
	IsDispellable   bool            `json:"is_dispellable" db:"is_dispellable"`
	DispelType      string          `json:"dispel_type" db:"dispel_type"` // magic, curse, poison, disease
	Priority        int             `json:"priority" db:"priority"` // pour l'ordre d'application
	
	// État
	IsActive        bool            `json:"is_active" db:"is_active"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

// StatModifiers représente les modifications de stats
type StatModifiers struct {
	// Stats de base (modificateurs additifs)
	HealthBonus      int     `json:"health_bonus"`
	ManaBonus        int     `json:"mana_bonus"`
	StrengthBonus    int     `json:"strength_bonus"`
	AgilityBonus     int     `json:"agility_bonus"`
	IntelligenceBonus int    `json:"intelligence_bonus"`
	VitalityBonus    int     `json:"vitality_bonus"`
	
	// Stats de combat (modificateurs additifs)
	DamageBonus      int     `json:"damage_bonus"`
	DefenseBonus     int     `json:"defense_bonus"`
	
	// Multiplicateurs (modificateurs multiplicatifs en pourcentage)
	HealthMultiplier    float64 `json:"health_multiplier"`
	ManaMultiplier      float64 `json:"mana_multiplier"`
	DamageMultiplier    float64 `json:"damage_multiplier"`
	DefenseMultiplier   float64 `json:"defense_multiplier"`
	SpeedMultiplier     float64 `json:"speed_multiplier"`
	CritChanceBonus     float64 `json:"crit_chance_bonus"`
	CritDamageBonus     float64 `json:"crit_damage_bonus"`
	
	// Résistances (en pourcentage)
	PhysicalResistance  float64 `json:"physical_resistance"`
	MagicalResistance   float64 `json:"magical_resistance"`
	FireResistance      float64 `json:"fire_resistance"`
	IceResistance       float64 `json:"ice_resistance"`
	LightningResistance float64 `json:"lightning_resistance"`
	PoisonResistance    float64 `json:"poison_resistance"`
	
	// Immunités
	Immunities          []string `json:"immunities"` // types d'effets auxquels le personnage est immunisé
}

// PeriodicEffect représente un effet qui se déclenche périodiquement
type PeriodicEffect struct {
	Type        string        `json:"type"` // damage, heal, mana_drain, mana_restore
	Value       int           `json:"value"` // valeur par tick
	Interval    time.Duration `json:"interval"` // intervalle entre les ticks
	TicksLeft   int           `json:"ticks_left"` // nombre de ticks restants
	DamageType  string        `json:"damage_type,omitempty"` // physical, magical, true
	Element     string        `json:"element,omitempty"`
}

// DTOs pour les requêtes

// CreateSpellRequest requête pour créer un sort
type CreateSpellRequest struct {
	Name            string          `json:"name" binding:"required,min=3,max=50"`
	Description     string          `json:"description" binding:"max=500"`
	School          string          `json:"school" binding:"required"`
	Type            string          `json:"type" binding:"required"`
	TargetType      string          `json:"target_type" binding:"required"`
	RequiredLevel   int             `json:"required_level" binding:"min=1,max=100"`
	RequiredClass   []string        `json:"required_class"`
	RequiredStats   StatRequirement `json:"required_stats"`
	ManaCost        int             `json:"mana_cost" binding:"min=0"`
	HealthCost      int             `json:"health_cost" binding:"min=0"`
	CastTime        time.Duration   `json:"cast_time"`
	Cooldown        time.Duration   `json:"cooldown"`
	Duration        time.Duration   `json:"duration"`
	Effects         []SpellEffect   `json:"effects" binding:"required,min=1"`
	Range           float64         `json:"range" binding:"min=0"`
	Radius          float64         `json:"radius" binding:"min=0"`
	Accuracy        float64         `json:"accuracy" binding:"min=0,max=1"`
	CanCrit         bool            `json:"can_crit"`
	IsChanneled     bool            `json:"is_channeled"`
	RequiresTarget  bool            `json:"requires_target"`
}

// LearnSpellRequest requête pour apprendre un sort
type LearnSpellRequest struct {
	SpellID     uuid.UUID `json:"spell_id" binding:"required"`
	SlotNumber  int       `json:"slot_number" binding:"min=1,max=20"`
}

// CastSpellRequest requête pour lancer un sort
type CastSpellRequest struct {
	SessionID   uuid.UUID   `json:"session_id" binding:"required"`
	CharacterID uuid.UUID   `json:"character_id" binding:"required"`
	SpellID     uuid.UUID   `json:"spell_id" binding:"required"`
	Targets     []uuid.UUID `json:"targets"`
	Position    *Position   `json:"position,omitempty"`
	PowerLevel  float64     `json:"power_level" binding:"min=0.1,max=1.0"`
}

// SpellbookResponse réponse contenant les sorts d'un personnage
type SpellbookResponse struct {
	CharacterID     uuid.UUID         `json:"character_id"`
	LearnedSpells   []CharacterSpell  `json:"learned_spells"`
	ActiveCooldowns []SpellCooldown   `json:"active_cooldowns"`
	AvailableSlots  int               `json:"available_slots"`
	TotalSlots      int               `json:"total_slots"`
}

// EffectSummary résumé d'un effet pour l'interface
type EffectSummary struct {
	ID          uuid.UUID     `json:"id"`
	Name        string        `json:"name"`
	Type        string        `json:"type"`
	Icon        string        `json:"icon"`
	StackCount  int           `json:"stack_count"`
	TimeLeft    *time.Duration `json:"time_left"`
	IsPositive  bool          `json:"is_positive"`
}