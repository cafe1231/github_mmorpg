// internal/models/spell.go
package models

import (
	"encoding/json"
	"fmt"
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

// StatusEffect représente un effet de statut actif sur un personnage (MODÈLE COMPLET)
type StatusEffect struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	CharacterID     uuid.UUID       `json:"character_id" db:"character_id"`
	SessionID       *uuid.UUID      `json:"session_id" db:"session_id"` // null si hors combat
	
	// Type et source
	Type            string          `json:"type" db:"type"` // buff, debuff, poison, burn, etc.
	Source          string          `json:"source" db:"source"` // spell, item, environment, ability, consumable
	SourceID        *uuid.UUID      `json:"source_id" db:"source_id"` // ID du sort/item qui a causé l'effet
	CasterID        *uuid.UUID      `json:"caster_id" db:"caster_id"` // qui a lancé l'effet
	
	// Propriétés de base
	Name            string          `json:"name" db:"name"`
	Description     string          `json:"description" db:"description"`
	Icon            string          `json:"icon" db:"icon"`
	
	// Effets sur les stats
	StatModifiers   map[string]interface{} `json:"stat_modifiers" db:"stat_modifiers"`
	
	// Effets périodiques
	PeriodicEffect  *PeriodicEffect `json:"periodic_effect" db:"periodic_effect"`
	LastTick        *time.Time      `json:"last_tick" db:"last_tick"`
	
	// Timing
	AppliedAt       time.Time       `json:"applied_at" db:"started_at"` // Champ manquant ajouté
	ExpiresAt       time.Time       `json:"expires_at" db:"ends_at"`    // Champ manquant ajouté
	Duration        time.Duration   `json:"duration" db:"duration"`     // Corrigé de *time.Duration
	
	// Stack
	Stacks          int             `json:"stacks" db:"stack_count"`    // Champ manquant ajouté
	MaxStacks       int             `json:"max_stacks" db:"max_stacks"`
	
	// Propriétés spéciales
	IsDispellable   bool            `json:"is_dispellable" db:"is_dispellable"`
	DispelType      string          `json:"dispel_type" db:"dispel_type"`
	Priority        int             `json:"priority" db:"priority"`
	IsActive        bool            `json:"is_active" db:"is_active"`
	
	// Métadonnées
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

// PeriodicEffect représente un effet qui se déclenche périodiquement (MODÈLE COMPLET)
type PeriodicEffect struct {
	Type            string          `json:"type"` // damage, healing, mana
	Amount          int             `json:"amount"`
	Interval        time.Duration   `json:"interval"` // intervalle entre les ticks
	Conditions      []string        `json:"conditions,omitempty"` // conditions pour le déclenchement
	ScalesWithLevel bool            `json:"scales_with_level"`
	ScalesWithStats map[string]float64 `json:"scales_with_stats,omitempty"`
}

// Méthodes utilitaires pour StatusEffect

// IsExpired vérifie si l'effet a expiré
func (se *StatusEffect) IsExpired() bool {
	return se.ExpiresAt.Before(time.Now())
}

// IsStackable vérifie si l'effet peut être stacké
func (se *StatusEffect) IsStackable() bool {
	return se.MaxStacks > 1
}

// CanAddStack vérifie si on peut ajouter un stack
func (se *StatusEffect) CanAddStack() bool {
	return se.IsStackable() && se.Stacks < se.MaxStacks
}

// GetRemainingDuration retourne la durée restante
func (se *StatusEffect) GetRemainingDuration() time.Duration {
	if se.ExpiresAt.IsZero() {
		return time.Duration(0)
	}
	
	remaining := time.Until(se.ExpiresAt)
	if remaining < 0 {
		return time.Duration(0)
	}
	
	return remaining
}

// GetEffectiveness retourne l'efficacité de l'effet basée sur les stacks
func (se *StatusEffect) GetEffectiveness() float64 {
	if se.MaxStacks <= 1 {
		return 1.0
	}
	
	return float64(se.Stacks) / float64(se.MaxStacks)
}

// ApplyModifier applique un modificateur à une valeur de base
func (se *StatusEffect) ApplyModifier(baseStat string, baseValue interface{}) interface{} {
	if se.StatModifiers == nil {
		return baseValue
	}
	
	// Chercher les modificateurs pour cette stat
	for modifierKey, modifierValue := range se.StatModifiers {
		if modifierKey == baseStat+"_bonus" {
			// Bonus additif
			if baseInt, ok := baseValue.(int); ok {
				if bonusInt, ok := modifierValue.(int); ok {
					return baseInt + (bonusInt * se.Stacks)
				}
				if bonusFloat, ok := modifierValue.(float64); ok {
					return baseInt + int(bonusFloat*float64(se.Stacks))
				}
			}
		} else if modifierKey == baseStat+"_multiplier" {
			// Multiplicateur
			if baseInt, ok := baseValue.(int); ok {
				if multiplier, ok := modifierValue.(float64); ok {
					return int(float64(baseInt) * multiplier)
				}
			}
			if baseFloat, ok := baseValue.(float64); ok {
				if multiplier, ok := modifierValue.(float64); ok {
					return baseFloat * multiplier
				}
			}
		}
	}
	
	return baseValue
}

// ShouldTick vérifie si l'effet doit déclencher un tick périodique
func (se *StatusEffect) ShouldTick() bool {
	if se.PeriodicEffect == nil {
		return false
	}
	
	if se.LastTick == nil {
		return true
	}
	
	return time.Since(*se.LastTick) >= se.PeriodicEffect.Interval
}

// Validate vérifie la validité de l'effet
func (se *StatusEffect) Validate() error {
	if se.Name == "" {
		return fmt.Errorf("effect name is required")
	}
	
	if se.Type == "" {
		return fmt.Errorf("effect type is required")
	}
	
	if se.CharacterID == uuid.Nil {
		return fmt.Errorf("character ID is required")
	}
	
	if se.MaxStacks < 1 {
		return fmt.Errorf("max stacks must be at least 1")
	}
	
	if se.Stacks < 1 || se.Stacks > se.MaxStacks {
		return fmt.Errorf("stacks must be between 1 and %d", se.MaxStacks)
	}
	
	if se.Duration < 0 {
		return fmt.Errorf("duration cannot be negative")
	}
	
	return nil
}

// GetDisplayInfo retourne les informations d'affichage de l'effet
func (se *StatusEffect) GetDisplayInfo() map[string]interface{} {
	info := map[string]interface{}{
		"name":        se.Name,
		"description": se.Description,
		"type":        se.Type,
		"stacks":      se.Stacks,
		"max_stacks":  se.MaxStacks,
		"remaining":   se.GetRemainingDuration().String(),
		"icon":        se.Icon,
	}
	
	if se.IsStackable() && se.Stacks > 1 {
		info["display_name"] = fmt.Sprintf("%s (x%d)", se.Name, se.Stacks)
	} else {
		info["display_name"] = se.Name
	}
	
	return info
}

// Structures pour les requêtes/réponses

// SpellBookEntry représente un sort dans un livre de sorts
type SpellBookEntry struct {
	Spell           Spell           `json:"spell"`
	CharacterSpell  CharacterSpell  `json:"character_spell"`
	IsOnCooldown    bool            `json:"is_on_cooldown"`
	CooldownRemaining time.Duration `json:"cooldown_remaining"`
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

// Constants pour les types d'effets
const (
	EffectTypeBuff         = "buff"
	EffectTypeDebuff       = "debuff"
	EffectTypePoison       = "poison"
	EffectTypeBurn         = "burn"
	EffectTypeFreeze       = "freeze"
	EffectTypeStun         = "stun"
	EffectTypeSlow         = "slow"
	EffectTypeHaste        = "haste"
	EffectTypeShield       = "shield"
	EffectTypeRegeneration = "regeneration"
	EffectTypeBleed        = "bleed"
	EffectTypeSilence      = "silence"
	EffectTypeImmunity     = "immunity"
)

// Constants pour les sources d'effets
const (
	EffectSourceSpell       = "spell"
	EffectSourceItem        = "item"
	EffectSourceEnvironment = "environment"
	EffectSourceAbility     = "ability"
	EffectSourceConsumable  = "consumable"
	EffectSourcePassive     = "passive"
)

// Constants pour les types de dispel
const (
	DispelTypeMagic    = "magic"
	DispelTypePoison   = "poison"
	DispelTypeDisease  = "disease"
	DispelTypeCurse    = "curse"
	DispelTypePhysical = "physical"
	DispelTypeNone     = "none"
	DispelTypeAll      = "all"
)