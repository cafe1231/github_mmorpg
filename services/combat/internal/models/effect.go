package models

import (
	"fmt"
	"time"

	"combat/internal/config"

	"github.com/google/uuid"
)

// EffectType définit les types d'effets possibles
type EffectType string

const (
	EffectTypeBuff    EffectType = "buff"
	EffectTypeDebuff  EffectType = "debuff"
	EffectTypeDot     EffectType = "dot" // Damage over time
	EffectTypeHot     EffectType = "hot" // Heal over time
	EffectTypeShield  EffectType = "shield"
	EffectTypeStun    EffectType = "stun"
	EffectTypeSilence EffectType = "silence"
)

// ModifierType définit le type de modificateur
type ModifierType string

const (
	ModifierTypeFlat       ModifierType = "flat"
	ModifierTypePercentage ModifierType = "percentage"
)

// CombatEffect représente un effet de combat
type CombatEffect struct {
	ID       uuid.UUID  `json:"id" db:"id"`
	CombatID uuid.UUID  `json:"combat_id" db:"combat_id"`
	TargetID uuid.UUID  `json:"target_id" db:"target_id"`
	CasterID *uuid.UUID `json:"caster_id" db:"caster_id"`

	// Détails de l'effet
	EffectType        EffectType `json:"effect_type" db:"effect_type"`
	EffectName        string     `json:"effect_name" db:"effect_name"`
	EffectDescription string     `json:"effect_description" db:"effect_description"`

	// Propriétés de l'effet
	StatAffected  *string      `json:"stat_affected" db:"stat_affected"`
	ModifierValue int          `json:"modifier_value" db:"modifier_value"`
	ModifierType  ModifierType `json:"modifier_type" db:"modifier_type"`

	// Durée et stacks
	DurationTurns  int `json:"duration_turns" db:"duration_turns"`
	RemainingTurns int `json:"remaining_turns" db:"remaining_turns"`
	MaxStacks      int `json:"max_stacks" db:"max_stacks"`
	CurrentStacks  int `json:"current_stacks" db:"current_stacks"`

	// État
	IsActive      bool `json:"is_active" db:"is_active"`
	IsDispellable bool `json:"is_dispellable" db:"is_dispellable"`

	AppliedAt time.Time  `json:"applied_at" db:"applied_at"`
	ExpiresAt *time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`

	// Relations (chargées séparément)
	Target *CombatParticipant `json:"target,omitempty" db:"-"`
	Caster *CombatParticipant `json:"caster,omitempty" db:"-"`
}

// EffectTemplate représente un modèle d'effet
type EffectTemplate struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Icon          string                 `json:"icon,omitempty"`
	EffectType    EffectType             `json:"effect_type"`
	StatAffected  string                 `json:"stat_affected,omitempty"`
	ModifierValue int                    `json:"modifier_value"`
	ModifierType  ModifierType           `json:"modifier_type"`
	BaseDuration  int                    `json:"base_duration"`
	MaxStacks     int                    `json:"max_stacks"`
	IsDispellable bool                   `json:"is_dispellable"`
	IsBeneficial  bool                   `json:"is_beneficial"`
	Tags          []string               `json:"tags,omitempty"`
	Conditions    map[string]interface{} `json:"conditions,omitempty"`
}

// EffectApplication représente l'application d'un effet
type EffectApplication struct {
	EffectTemplate *EffectTemplate        `json:"effect_template"`
	TargetID       uuid.UUID              `json:"target_id"`
	CasterID       *uuid.UUID             `json:"caster_id,omitempty"`
	Duration       int                    `json:"duration,omitempty"`
	Stacks         int                    `json:"stacks,omitempty"`
	CustomModifier *int                   `json:"custom_modifier,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// EffectResult représente le résultat de l'application d'un effet
type EffectResult struct {
	Success        bool          `json:"success"`
	Effect         *CombatEffect `json:"effect,omitempty"`
	ExistingEffect *CombatEffect `json:"existing_effect,omitempty"`
	Action         string        `json:"action"` // "applied", "stacked", "refreshed", "resisted"
	Message        string        `json:"message,omitempty"`
	Error          string        `json:"error,omitempty"`
}

// GetEffectTemplates retourne les modèles d'effets prédéfinis
func GetEffectTemplates() map[string]*EffectTemplate {
	return map[string]*EffectTemplate{
		"strength_buff": {
			ID:            "strength_buff",
			Name:          "Force accrue",
			Description:   "Augmente la force physique",
			Icon:          "muscle",
			EffectType:    EffectTypeBuff,
			StatAffected:  "physical_damage",
			ModifierValue: config.DefaultModifierValue,
			ModifierType:  ModifierTypePercentage,
			BaseDuration:  config.DefaultBaseDuration,
			MaxStacks:     config.DefaultMaxStacks,
			IsDispellable: true,
			IsBeneficial:  true,
			Tags:          []string{"buff", "physical"},
		},
		"poison": {
			ID:            "poison",
			Name:          "Poison",
			Description:   "Inflige des dégâts au fil du temps",
			Icon:          "poison",
			EffectType:    EffectTypeDot,
			ModifierValue: config.DefaultModifierValue2,
			ModifierType:  ModifierTypeFlat,
			BaseDuration:  config.DefaultBaseDuration2,
			MaxStacks:     config.DefaultMaxStacks2,
			IsDispellable: true,
			IsBeneficial:  false,
			Tags:          []string{"debuff", "dot", "poison"},
		},
		"regeneration": {
			ID:            "regeneration",
			Name:          "Régénération",
			Description:   "Restaure de la vie au fil du temps",
			Icon:          "heart",
			EffectType:    EffectTypeHot,
			ModifierValue: config.DefaultModifierValue3,
			ModifierType:  ModifierTypeFlat,
			BaseDuration:  config.DefaultBaseDuration3,
			MaxStacks:     config.DefaultMaxStacks3,
			IsDispellable: true,
			IsBeneficial:  true,
			Tags:          []string{"buff", "hot", "healing"},
		},
		"shield": {
			ID:            "shield",
			Name:          "Bouclier magique",
			Description:   "Absorbe les dégâts",
			Icon:          "shield",
			EffectType:    EffectTypeShield,
			ModifierValue: config.DefaultModifierValue4,
			ModifierType:  ModifierTypeFlat,
			BaseDuration:  config.DefaultBaseDuration4,
			MaxStacks:     1,
			IsDispellable: true,
			IsBeneficial:  true,
			Tags:          []string{"buff", "protection"},
		},
		"stun": {
			ID:            "stun",
			Name:          "Étourdissement",
			Description:   "Empêche d'agir",
			Icon:          "dizzy",
			EffectType:    EffectTypeStun,
			BaseDuration:  config.DefaultBaseDuration5,
			MaxStacks:     1,
			IsDispellable: true,
			IsBeneficial:  false,
			Tags:          []string{"debuff", "control"},
		},
		"silence": {
			ID:            "silence",
			Name:          "Silence",
			Description:   "Empêche l'utilisation de compétences",
			Icon:          "mute",
			EffectType:    EffectTypeSilence,
			BaseDuration:  config.DefaultBaseDuration6,
			MaxStacks:     1,
			IsDispellable: true,
			IsBeneficial:  false,
			Tags:          []string{"debuff", "control"},
		},
		"weakness": {
			ID:            "weakness",
			Name:          "Faiblesse",
			Description:   "Réduit les dégâts infligés",
			Icon:          "broken-heart",
			EffectType:    EffectTypeDebuff,
			StatAffected:  "physical_damage",
			ModifierValue: -30,
			ModifierType:  ModifierTypePercentage,
			BaseDuration:  config.DefaultBaseDuration7,
			MaxStacks:     config.DefaultMaxStacks4,
			IsDispellable: true,
			IsBeneficial:  false,
			Tags:          []string{"debuff", "physical"},
		},
		"haste": {
			ID:            "haste",
			Name:          "Hâte",
			Description:   "Augmente la vitesse d'attaque",
			Icon:          "fast-forward",
			EffectType:    EffectTypeBuff,
			StatAffected:  "attack_speed",
			ModifierValue: config.DefaultModifierValue5,
			ModifierType:  ModifierTypePercentage,
			BaseDuration:  config.DefaultBaseDuration8,
			MaxStacks:     config.DefaultMaxStacks5,
			IsDispellable: true,
			IsBeneficial:  true,
			Tags:          []string{"buff", "speed"},
		},
	}
}

// IsExpired vérifie si l'effet a expiré
func (e *CombatEffect) IsExpired() bool {
	return e.RemainingTurns <= 0 || (e.ExpiresAt != nil && time.Now().After(*e.ExpiresAt))
}

// IsBeneficial vérifie si l'effet est bénéfique
func (e *CombatEffect) IsBeneficial() bool {
	return e.EffectType == EffectTypeBuff || e.EffectType == EffectTypeHot || e.EffectType == EffectTypeShield
}

// IsHarmful vérifie si l'effet est néfaste
func (e *CombatEffect) IsHarmful() bool {
	return e.EffectType == EffectTypeDebuff || e.EffectType == EffectTypeDot ||
		e.EffectType == EffectTypeStun || e.EffectType == EffectTypeSilence
}

// CanStack vérifie si l'effet peut être empilé
func (e *CombatEffect) CanStack() bool {
	return e.MaxStacks > 1 && e.CurrentStacks < e.MaxStacks
}

// GetStatModifier retourne la modification de statistique
func (e *CombatEffect) GetStatModifier() (statName string, modifierValue int, modifierType ModifierType) {
	if e.StatAffected == nil {
		return "", 0, ModifierTypeFlat
	}

	// Multiplie par le nombre de stacks
	value := e.ModifierValue * e.CurrentStacks

	return *e.StatAffected, value, e.ModifierType
}

// ProcessTurn traite l'effet pour un tour
func (e *CombatEffect) ProcessTurn() *EffectProcessResult {
	result := &EffectProcessResult{
		EffectID: e.ID,
		Expired:  false,
	}

	// Appliquer l'effet selon son type
	switch e.EffectType {
	case EffectTypeDot:
		// Dégâts au fil du temps
		damage := e.ModifierValue * e.CurrentStacks
		result.DamageDealt = damage
		result.Message = fmt.Sprintf("Le %s inflige %d dégâts", e.EffectName, damage)

	case EffectTypeHot:
		// Soins au fil du temps
		healing := e.ModifierValue * e.CurrentStacks
		result.HealingDone = healing
		result.Message = fmt.Sprintf("La %s restaure %d points de vie", e.EffectName, healing)
	}

	// Décrémenter la durée
	e.RemainingTurns--

	// Vérifier l'expiration
	if e.RemainingTurns <= 0 {
		result.Expired = true
		e.IsActive = false
		result.Message += fmt.Sprintf(" (L'effet %s expire)", e.EffectName)
	}

	return result
}

// EffectProcessResult représente le résultat du traitement d'un effet
type EffectProcessResult struct {
	EffectID    uuid.UUID `json:"effect_id"`
	DamageDealt int       `json:"damage_dealt,omitempty"`
	HealingDone int       `json:"healing_done,omitempty"`
	Expired     bool      `json:"expired"`
	Message     string    `json:"message,omitempty"`
}

// CreateEffectFromTemplate crée un effet à partir d'un modèle
func CreateEffectFromTemplate(template *EffectTemplate, application *EffectApplication) *CombatEffect {
	effect := &CombatEffect{
		ID:                uuid.New(),
		CombatID:          application.TargetID, // Will be set properly by service
		TargetID:          application.TargetID,
		CasterID:          application.CasterID,
		EffectType:        template.EffectType,
		EffectName:        template.Name,
		EffectDescription: template.Description,
		ModifierValue:     template.ModifierValue,
		ModifierType:      template.ModifierType,
		DurationTurns:     template.BaseDuration,
		RemainingTurns:    template.BaseDuration,
		MaxStacks:         template.MaxStacks,
		CurrentStacks:     1,
		IsActive:          true,
		IsDispellable:     template.IsDispellable,
		AppliedAt:         time.Now(),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Appliquer les modifications personnalisées
	if application.Duration > 0 {
		effect.DurationTurns = application.Duration
		effect.RemainingTurns = application.Duration
	}

	if application.Stacks > 0 {
		effect.CurrentStacks = application.Stacks
		if effect.CurrentStacks > effect.MaxStacks {
			effect.CurrentStacks = effect.MaxStacks
		}
	}

	if application.CustomModifier != nil {
		effect.ModifierValue = *application.CustomModifier
	}

	// Définir la statistique affectée
	if template.StatAffected != "" {
		effect.StatAffected = &template.StatAffected
	}

	return effect
}
