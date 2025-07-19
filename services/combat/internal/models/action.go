package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ActionType définit les types d'actions possibles
type ActionType string

const (
	ActionTypeAttack ActionType = "attack"
	ActionTypeSkill  ActionType = "skill"
	ActionTypeItem   ActionType = "item"
	ActionTypeDefend ActionType = "defend"
	ActionTypeFlee   ActionType = "flee"
	ActionTypeWait   ActionType = "wait"
)

// ActionStatistics représente les statistiques d'actions
type ActionStatistics struct {
	TotalActions      int     `json:"total_actions" db:"total_actions"`
	CriticalHits      int     `json:"critical_hits" db:"critical_hits"`
	Misses            int     `json:"misses" db:"misses"`
	Blocks            int     `json:"blocks" db:"blocks"`
	AvgDamage         float64 `json:"avg_damage" db:"avg_damage"`
	MaxDamage         int     `json:"max_damage" db:"max_damage"`
	AvgHealing        float64 `json:"avg_healing" db:"avg_healing"`
	AvgProcessingTime float64 `json:"avg_processing_time" db:"avg_processing_time"`
	CriticalRate      float64 `json:"critical_rate" db:"-"`
	MissRate          float64 `json:"miss_rate" db:"-"`
	BlockRate         float64 `json:"block_rate" db:"-"`
	AccuracyRate      float64 `json:"accuracy_rate" db:"-"`
}

// CombatAction représente une action dans un combat
type CombatAction struct {
	ID       uuid.UUID  `json:"id" db:"id"`
	CombatID uuid.UUID  `json:"combat_id" db:"combat_id"`
	ActorID  uuid.UUID  `json:"actor_id" db:"actor_id"`
	TargetID *uuid.UUID `json:"target_id" db:"target_id"`

	// Détails de l'action
	ActionType ActionType `json:"action_type" db:"action_type"`
	SkillID    *string    `json:"skill_id" db:"skill_id"`
	ItemID     *string    `json:"item_id" db:"item_id"`

	// Résultats de l'action
	DamageDealt int  `json:"damage_dealt" db:"damage_dealt"`
	HealingDone int  `json:"healing_done" db:"healing_done"`
	ManaUsed    int  `json:"mana_used" db:"mana_used"`
	IsCritical  bool `json:"is_critical" db:"is_critical"`
	IsMiss      bool `json:"is_miss" db:"is_miss"`
	IsBlocked   bool `json:"is_blocked" db:"is_blocked"`

	// Métadonnées
	TurnNumber       int  `json:"turn_number" db:"turn_number"`
	ActionOrder      int  `json:"action_order" db:"action_order"`
	ProcessingTimeMs *int `json:"processing_time_ms" db:"processing_time_ms"`

	// Validation anti-cheat
	ClientTimestamp *time.Time `json:"client_timestamp" db:"client_timestamp"`
	ServerTimestamp time.Time  `json:"server_timestamp" db:"server_timestamp"`
	IsValidated     bool       `json:"is_validated" db:"is_validated"`
	ValidationNotes *string    `json:"validation_notes" db:"validation_notes"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// Relations (chargées séparément)
	Actor   *CombatParticipant `json:"actor,omitempty" db:"-"`
	Target  *CombatParticipant `json:"target,omitempty" db:"-"`
	Skill   *SkillInfo         `json:"skill,omitempty" db:"-"`
	Item    *ItemInfo          `json:"item,omitempty" db:"-"`
	Effects []*CombatEffect    `json:"effects,omitempty" db:"-"`
}

// SkillInfo représente les informations d'une compétence
type SkillInfo struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	Description  string             `json:"description"`
	Type         string             `json:"type"` // "physical", "magical", "hybrid"
	ManaCost     int                `json:"mana_cost"`
	Cooldown     int                `json:"cooldown"`
	Range        int                `json:"range"`
	AreaOfEffect bool               `json:"area_of_effect"`
	TargetType   string             `json:"target_type"` // "self", "ally", "enemy", "any"
	BaseDamage   int                `json:"base_damage"`
	BaseHealing  int                `json:"base_healing"`
	Effects      []SkillEffect      `json:"effects,omitempty"`
	Requirements map[string]int     `json:"requirements,omitempty"`
	Modifiers    map[string]float64 `json:"modifiers,omitempty"`
	Icon         string             `json:"icon,omitempty"`
	Animation    string             `json:"animation,omitempty"`
	SoundEffect  string             `json:"sound_effect,omitempty"`
}

// SkillEffect représente un effet d'une compétence
type SkillEffect struct {
	Type         string  `json:"type"`
	Value        int     `json:"value"`
	Duration     int     `json:"duration"`
	Probability  float64 `json:"probability"`
	Target       string  `json:"target"`
	StatAffected string  `json:"stat_affected,omitempty"`
	ModifierType string  `json:"modifier_type,omitempty"`
}

// ItemInfo représente les informations d'un objet
type ItemInfo struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	Description    string         `json:"description"`
	Type           string         `json:"type"` // "consumable", "equipment", "scroll"
	Rarity         string         `json:"rarity"`
	UsableInCombat bool           `json:"usable_in_combat"`
	Consumable     bool           `json:"consumable"`
	Cooldown       int            `json:"cooldown"`
	TargetType     string         `json:"target_type"`
	Effects        []ItemEffect   `json:"effects,omitempty"`
	Requirements   map[string]int `json:"requirements,omitempty"`
	Icon           string         `json:"icon,omitempty"`
	Animation      string         `json:"animation,omitempty"`
}

// ItemEffect représente un effet d'un objet
type ItemEffect struct {
	Type         string `json:"type"`
	Value        int    `json:"value"`
	Duration     int    `json:"duration"`
	Target       string `json:"target"`
	StatAffected string `json:"stat_affected,omitempty"`
}

// ActionRequest représente une demande d'action de combat
type ActionRequest struct {
	ActionType      ActionType `json:"action_type" binding:"required"`
	TargetID        *uuid.UUID `json:"target_id"`
	SkillID         *string    `json:"skill_id"`
	ItemID          *string    `json:"item_id"`
	ClientTimestamp time.Time  `json:"client_timestamp"`

	// Données additionnelles pour validation
	Position *Position              `json:"position,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Position représente une position dans l'espace
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

// ActionResult représente le résultat d'une action
type ActionResult struct {
	Success      bool            `json:"success"`
	Action       *CombatAction   `json:"action,omitempty"`
	Error        string          `json:"error,omitempty"`
	Effects      []*CombatEffect `json:"effects,omitempty"`
	StateChanges *StateChanges   `json:"state_changes,omitempty"`
	Logs         []*CombatLog    `json:"logs,omitempty"`
	Message      string          `json:"message,omitempty"`
	Warnings     []string        `json:"warnings,omitempty"`
}

// StateChanges représente les changements d'état après une action
type StateChanges struct {
	ParticipantChanges map[uuid.UUID]*ParticipantChange `json:"participant_changes,omitempty"`
	CombatChange       *CombatChange                    `json:"combat_change,omitempty"`
}

// ParticipantChange représente les changements d'un participant
type ParticipantChange struct {
	HealthChange   int                    `json:"health_change,omitempty"`
	ManaChange     int                    `json:"mana_change,omitempty"`
	StatusChange   string                 `json:"status_change,omitempty"`
	EffectsAdded   []*CombatEffect        `json:"effects_added,omitempty"`
	EffectsRemoved []uuid.UUID            `json:"effects_removed,omitempty"`
	StatsModified  map[string]interface{} `json:"stats_modified,omitempty"`
}

// CombatChange représente les changements du combat
type CombatChange struct {
	TurnAdvanced   bool          `json:"turn_advanced,omitempty"`
	StatusChanged  *CombatStatus `json:"status_changed,omitempty"`
	Winner         *int          `json:"winner,omitempty"`
	NextActionTime *time.Time    `json:"next_action_time,omitempty"`
}

// ActionValidation représente la validation d'une action
type ActionValidation struct {
	IsValid         bool     `json:"is_valid"`
	Errors          []string `json:"errors,omitempty"`
	Warnings        []string `json:"warnings,omitempty"`
	SuspiciousFlags []string `json:"suspicious_flags,omitempty"`
}

// ActionTemplate représente un modèle d'action
type ActionTemplate struct {
	Type            ActionType     `json:"type"`
	Name            string         `json:"name"`
	Description     string         `json:"description"`
	Icon            string         `json:"icon,omitempty"`
	RequiredTargets int            `json:"required_targets"`
	TargetType      string         `json:"target_type"`
	Range           int            `json:"range"`
	Cooldown        int            `json:"cooldown"`
	ManaCost        int            `json:"mana_cost"`
	Requirements    map[string]int `json:"requirements,omitempty"`
	Available       bool           `json:"available"`
	Damage          *DamageInfo    `json:"damage,omitempty"`
	Healing         *HealingInfo   `json:"healing,omitempty"`
}

// DamageInfo représente les informations de dégâts
type DamageInfo struct {
	BaseDamage         int     `json:"base_damage"`
	DamageType         string  `json:"damage_type"` // "physical", "magical", "true"
	ScaleWithStat      string  `json:"scale_with_stat,omitempty"`
	ScaleMultiplier    float64 `json:"scale_multiplier,omitempty"`
	CriticalChance     float64 `json:"critical_chance,omitempty"`
	CriticalMultiplier float64 `json:"critical_multiplier,omitempty"`
}

// HealingInfo représente les informations de soins
type HealingInfo struct {
	BaseHealing     int     `json:"base_healing"`
	ScaleWithStat   string  `json:"scale_with_stat,omitempty"`
	ScaleMultiplier float64 `json:"scale_multiplier,omitempty"`
	CanCritical     bool    `json:"can_critical"`
}

// ActionCooldown représente un cooldown d'action
type ActionCooldown struct {
	ID         string        `json:"id"`
	ActorID    uuid.UUID     `json:"actor_id"`
	ActionType ActionType    `json:"action_type"`
	SkillID    string        `json:"skill_id,omitempty"`
	ExpiresAt  time.Time     `json:"expires_at"`
	Duration   time.Duration `json:"duration"`
}

// GetActionTemplates retourne les modèles d'actions disponibles
func GetActionTemplates() []*ActionTemplate {
	return []*ActionTemplate{
		{
			Type:            ActionTypeAttack,
			Name:            "Attaque de base",
			Description:     "Une attaque physique simple",
			Icon:            "sword",
			RequiredTargets: 1,
			TargetType:      "enemy",
			Range:           1,
			Cooldown:        0,
			ManaCost:        0,
			Available:       true,
			Damage: &DamageInfo{
				BaseDamage:         10,
				DamageType:         "physical",
				ScaleWithStat:      "physical_damage",
				ScaleMultiplier:    1.0,
				CriticalChance:     0.05,
				CriticalMultiplier: 1.5,
			},
		},
		{
			Type:            ActionTypeDefend,
			Name:            "Défense",
			Description:     "Réduit les dégâts reçus ce tour de 50%",
			Icon:            "shield",
			RequiredTargets: 0,
			TargetType:      "self",
			Range:           0,
			Cooldown:        0,
			ManaCost:        0,
			Available:       true,
		},
		{
			Type:            ActionTypeWait,
			Name:            "Attendre",
			Description:     "Passe le tour et récupère 10% de mana",
			Icon:            "clock",
			RequiredTargets: 0,
			TargetType:      "self",
			Range:           0,
			Cooldown:        0,
			ManaCost:        -10, // Récupère de la mana
			Available:       true,
		},
		{
			Type:            ActionTypeFlee,
			Name:            "Fuir",
			Description:     "Tente de fuir le combat (50% de réussite)",
			Icon:            "run",
			RequiredTargets: 0,
			TargetType:      "self",
			Range:           0,
			Cooldown:        0,
			ManaCost:        0,
			Available:       true,
		},
	}
}

// GetSkillTemplates retourne les modèles de compétences prédéfinies
func GetSkillTemplates() map[string]*SkillInfo {
	return map[string]*SkillInfo{
		"fireball": {
			ID:           "fireball",
			Name:         "Boule de feu",
			Description:  "Lance une boule de feu qui inflige des dégâts magiques",
			Type:         "magical",
			ManaCost:     25,
			Cooldown:     2,
			Range:        3,
			AreaOfEffect: false,
			TargetType:   "enemy",
			BaseDamage:   35,
			BaseHealing:  0,
			Effects: []SkillEffect{
				{
					Type:        "damage_over_time",
					Value:       5,
					Duration:    3,
					Probability: 0.3,
					Target:      "target",
				},
			},
			Icon:        "fire",
			Animation:   "fireball_cast",
			SoundEffect: "fire_whoosh",
		},
		"heal": {
			ID:           "heal",
			Name:         "Soin",
			Description:  "Restaure les points de vie d'un allié",
			Type:         "magical",
			ManaCost:     20,
			Cooldown:     1,
			Range:        2,
			AreaOfEffect: false,
			TargetType:   "ally",
			BaseDamage:   0,
			BaseHealing:  30,
			Icon:         "heart",
			Animation:    "heal_cast",
			SoundEffect:  "heal_chime",
		},
		"lightning_bolt": {
			ID:           "lightning_bolt",
			Name:         "Éclair",
			Description:  "Frappe l'ennemi avec un éclair rapide",
			Type:         "magical",
			ManaCost:     30,
			Cooldown:     3,
			Range:        4,
			AreaOfEffect: false,
			TargetType:   "enemy",
			BaseDamage:   40,
			BaseHealing:  0,
			Effects: []SkillEffect{
				{
					Type:        "stun",
					Value:       1,
					Duration:    1,
					Probability: 0.2,
					Target:      "target",
				},
			},
			Icon:        "lightning",
			Animation:   "lightning_cast",
			SoundEffect: "thunder",
		},
		"shield_bash": {
			ID:           "shield_bash",
			Name:         "Coup de bouclier",
			Description:  "Frappe avec le bouclier et étourdit l'ennemi",
			Type:         "physical",
			ManaCost:     15,
			Cooldown:     2,
			Range:        1,
			AreaOfEffect: false,
			TargetType:   "enemy",
			BaseDamage:   20,
			BaseHealing:  0,
			Effects: []SkillEffect{
				{
					Type:        "stun",
					Value:       1,
					Duration:    1,
					Probability: 0.5,
					Target:      "target",
				},
			},
			Requirements: map[string]int{
				"shield_equipped": 1,
			},
			Icon:        "shield",
			Animation:   "shield_bash",
			SoundEffect: "metal_clang",
		},
		"backstab": {
			ID:           "backstab",
			Name:         "Attaque sournoise",
			Description:  "Attaque critique garantie si l'ennemi ne vous a pas ciblé",
			Type:         "physical",
			ManaCost:     20,
			Cooldown:     4,
			Range:        1,
			AreaOfEffect: false,
			TargetType:   "enemy",
			BaseDamage:   25,
			BaseHealing:  0,
			Modifiers: map[string]float64{
				"critical_chance_bonus": 0.5,
				"critical_multiplier":   2.0,
			},
			Icon:        "dagger",
			Animation:   "backstab",
			SoundEffect: "blade_slice",
		},
	}
}

// Validate valide une demande d'action
func (ar *ActionRequest) Validate() *ActionValidation {
	validation := &ActionValidation{
		IsValid:         true,
		Errors:          []string{},
		Warnings:        []string{},
		SuspiciousFlags: []string{},
	}

	// Validation du type d'action
	validTypes := []ActionType{ActionTypeAttack, ActionTypeSkill, ActionTypeItem, ActionTypeDefend, ActionTypeFlee, ActionTypeWait}
	isValidType := false
	for _, t := range validTypes {
		if ar.ActionType == t {
			isValidType = true
			break
		}
	}
	if !isValidType {
		validation.IsValid = false
		validation.Errors = append(validation.Errors, "Type d'action invalide")
	}

	// Validation des paramètres spécifiques
	switch ar.ActionType {
	case ActionTypeAttack:
		if ar.TargetID == nil {
			validation.IsValid = false
			validation.Errors = append(validation.Errors, "Cible requise pour une attaque")
		}
	case ActionTypeSkill:
		if ar.SkillID == nil || *ar.SkillID == "" {
			validation.IsValid = false
			validation.Errors = append(validation.Errors, "ID de compétence requis")
		}
	case ActionTypeItem:
		if ar.ItemID == nil || *ar.ItemID == "" {
			validation.IsValid = false
			validation.Errors = append(validation.Errors, "ID d'objet requis")
		}
	}

	// Validation du timestamp
	if ar.ClientTimestamp.IsZero() {
		validation.Warnings = append(validation.Warnings, "Timestamp client manquant")
	} else {
		timeDiff := time.Since(ar.ClientTimestamp).Abs()
		if timeDiff > 5*time.Second {
			validation.SuspiciousFlags = append(validation.SuspiciousFlags, "Timestamp client suspect")
		}
	}

	return validation
}

// CalculateDamage calcule les dégâts d'une action
func (ca *CombatAction) CalculateDamage(actor, target *CombatParticipant, skill *SkillInfo) int {
	baseDamage := 0
	damageType := "physical"

	if skill != nil {
		baseDamage = skill.BaseDamage
		damageType = skill.Type
	} else {
		// Dégâts d'attaque de base
		baseDamage = actor.PhysicalDamage
	}

	// Application des modificateurs d'acteur
	damage := float64(baseDamage)

	if damageType == "magical" {
		damage += float64(actor.MagicalDamage) * 0.8
	} else {
		damage += float64(actor.PhysicalDamage) * 0.8
	}

	// Facteur de défense
	defense := float64(target.PhysicalDefense)
	if damageType == "magical" {
		defense = float64(target.MagicalDefense)
	}

	// Formule de réduction des dégâts
	damageReduction := defense / (defense + 100)
	damage = damage * (1 - damageReduction)

	// Critique
	if ca.IsCritical {
		damage *= 1.5
		if skill != nil {
			// Vérifier si la compétence a un multiplicateur de critique spécial
			if critMultiplier, exists := skill.Modifiers["critical_multiplier"]; exists {
				damage = float64(baseDamage) * critMultiplier
			}
		}
	}

	// Variabilité (±15%)
	variance := 0.85 + (0.3 * (float64(time.Now().UnixNano()%100) / 100))
	damage *= variance

	// S'assurer que les dégâts ne sont jamais négatifs
	if damage < 0 {
		damage = 0
	}

	return int(damage)
}

// CalculateHealing calcule les soins d'une action
func (ca *CombatAction) CalculateHealing(actor *CombatParticipant, skill *SkillInfo) int {
	baseHealing := 0

	if skill != nil {
		baseHealing = skill.BaseHealing
	}

	if baseHealing == 0 {
		return 0
	}

	// Application des modificateurs
	healing := float64(baseHealing)

	// Modificateur d'intelligence pour les soins magiques
	if skill != nil && skill.Type == "magical" {
		healing += float64(actor.MagicalDamage) * 0.6
	}

	// Critique
	if ca.IsCritical {
		healing *= 1.3
	}

	// Variabilité (±10%)
	variance := 0.9 + (0.2 * (float64(time.Now().UnixNano()%100) / 100))
	healing *= variance

	return int(healing)
}

// CalculateCriticalChance calcule la chance de critique
func CalculateCriticalChance(actor *CombatParticipant, skill *SkillInfo) float64 {
	baseCritChance := actor.CriticalChance

	if skill != nil {
		// Ajouter la chance de critique de la compétence
		if critBonus, exists := skill.Modifiers["critical_chance_bonus"]; exists {
			baseCritChance += critBonus
		}
	}

	// Limiter à 95% maximum
	if baseCritChance > 0.95 {
		baseCritChance = 0.95
	}

	return baseCritChance
}

// CalculateHitChance calcule la chance de toucher
func CalculateHitChance(actor, target *CombatParticipant, skill *SkillInfo) float64 {
	// Chance de base de 85%
	baseHitChance := 0.85

	// Facteur d'agilité
	agilityDiff := float64(actor.PhysicalDamage-target.PhysicalDefense) / 100.0
	hitChance := baseHitChance + (agilityDiff * 0.1)

	// Limiter entre 5% et 95%
	if hitChance < 0.05 {
		hitChance = 0.05
	}
	if hitChance > 0.95 {
		hitChance = 0.95
	}

	return hitChance
}

// GetDescription retourne une description textuelle de l'action
func (ca *CombatAction) GetDescription() string {
	switch ca.ActionType {
	case ActionTypeAttack:
		if ca.IsMiss {
			return "rate son attaque"
		}
		if ca.IsBlocked {
			return "voit son attaque bloquée"
		}
		if ca.IsCritical {
			return fmt.Sprintf("effectue une attaque critique (%d dégâts)", ca.DamageDealt)
		}
		return fmt.Sprintf("attaque (%d dégâts)", ca.DamageDealt)
	case ActionTypeSkill:
		skillName := "une compétence"
		if ca.Skill != nil {
			skillName = ca.Skill.Name
		}
		if ca.DamageDealt > 0 && ca.HealingDone > 0 {
			return fmt.Sprintf("utilize %s (%d dégâts, %d soins)", skillName, ca.DamageDealt, ca.HealingDone)
		} else if ca.DamageDealt > 0 {
			return fmt.Sprintf("utilize %s (%d dégâts)", skillName, ca.DamageDealt)
		} else if ca.HealingDone > 0 {
			return fmt.Sprintf("utilize %s (%d soins)", skillName, ca.HealingDone)
		}
		return fmt.Sprintf("utilize %s", skillName)
	case ActionTypeItem:
		itemName := "un objet"
		if ca.Item != nil {
			itemName = ca.Item.Name
		}
		return fmt.Sprintf("utilize %s", itemName)
	case ActionTypeDefend:
		return "se défend"
	case ActionTypeFlee:
		return "tente de fuir"
	case ActionTypeWait:
		return "attend et récupère de la mana"
	default:
		return "effectue une action"
	}
}

// IsSuccessful vérifie si l'action a réussi
func (ca *CombatAction) IsSuccessful() bool {
	return !ca.IsMiss && ca.IsValidated
}

// GetEffectiveness retourne l'efficacité de l'action (0-100)
func (ca *CombatAction) GetEffectiveness() float64 {
	if ca.IsMiss {
		return 0.0
	}

	effectiveness := 50.0 // Base

	if ca.DamageDealt > 0 {
		effectiveness += float64(ca.DamageDealt) / 10.0
	}

	if ca.HealingDone > 0 {
		effectiveness += float64(ca.HealingDone) / 8.0
	}

	if ca.IsCritical {
		effectiveness += 25.0
	}

	if ca.IsBlocked {
		effectiveness -= 20.0
	}

	// Limiter entre 0 et 100
	if effectiveness < 0 {
		effectiveness = 0
	}
	if effectiveness > 100 {
		effectiveness = 100
	}

	return effectiveness
}

// CreateAction crée une nouvelle action avec des valeurs par défaut
func CreateAction(combatID, actorID uuid.UUID, request *ActionRequest) *CombatAction {
	return &CombatAction{
		ID:              uuid.New(),
		CombatID:        combatID,
		ActorID:         actorID,
		TargetID:        request.TargetID,
		ActionType:      request.ActionType,
		SkillID:         request.SkillID,
		ItemID:          request.ItemID,
		DamageDealt:     0,
		HealingDone:     0,
		ManaUsed:        0,
		IsCritical:      false,
		IsMiss:          false,
		IsBlocked:       false,
		ClientTimestamp: &request.ClientTimestamp,
		ServerTimestamp: time.Now(),
		IsValidated:     true,
		CreatedAt:       time.Now(),
	}
}
