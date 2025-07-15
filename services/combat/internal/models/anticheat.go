// internal/models/anticheat.go
package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// CombatActionRequest utilise PerformActionRequest défini dans combat.go
// On crée un alias pour plus de clarté dans le contexte anti-cheat
type CombatActionRequest = PerformActionRequest

// CombatActionResult utilise ActionResult défini dans combat.go  
// et on ajoute les champs supplémentaires nécessaires
type CombatActionResult struct {
	Results     []ActionResult `json:"results"`
	Success     bool           `json:"success"`
	CriticalHit bool           `json:"critical_hit"`
	Duration    time.Duration  `json:"duration"`
}

// StartCombatRequest est défini dans combat.go - pas besoin de le redéclarer ici

// AntiCheatViolation représente une violation détectée par l'anti-cheat
type AntiCheatViolation struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	CharacterID uuid.UUID              `json:"character_id" db:"character_id"`
	Type        string                 `json:"type" db:"type"` // speed_hack, damage_hack, etc.
	Description string                 `json:"description" db:"description"`
	Severity    string                 `json:"severity" db:"severity"` // low, medium, high, critical
	Data        map[string]interface{} `json:"data" db:"data"`
	Timestamp   time.Time              `json:"timestamp" db:"timestamp"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
}

// AntiCheatBan représente un bannissement
type AntiCheatBan struct {
	ID          uuid.UUID     `json:"id" db:"id"`
	CharacterID uuid.UUID     `json:"character_id" db:"character_id"`
	Reason      string        `json:"reason" db:"reason"`
	Duration    time.Duration `json:"duration" db:"duration"`
	ExpiresAt   *time.Time    `json:"expires_at" db:"expires_at"`
	CreatedAt   time.Time     `json:"created_at" db:"created_at"`
	IsActive    bool          `json:"is_active" db:"is_active"`
}

// SuspicionReport représente un rapport de suspicion
type SuspicionReport struct {
	CharacterID    uuid.UUID         `json:"character_id"`
	SuspicionScore float64           `json:"suspicion_score"`
	RiskLevel      string            `json:"risk_level"` // low, medium, high, critical
	Factors        []SuspicionFactor `json:"factors"`
	Timestamp      time.Time         `json:"timestamp"`
}

// SuspicionFactor représente un facteur de suspicion
type SuspicionFactor struct {
	Type        string  `json:"type"`
	Severity    string  `json:"severity"`
	Description string  `json:"description"`
	Weight      float64 `json:"weight"`
}

// ActionEvent représente un événement d'action pour l'historique
type ActionEvent struct {
	CharacterID uuid.UUID              `json:"character_id"`
	ActionType  string                 `json:"action_type"`
	Timestamp   time.Time              `json:"timestamp"`
	Data        map[string]interface{} `json:"data"`
}

// PlayerBehaviorStats représente les statistiques de comportement d'un joueur
type PlayerBehaviorStats struct {
	CharacterID      uuid.UUID `json:"character_id"`
	StartTime        time.Time `json:"start_time"`
	TotalActions     int       `json:"total_actions"`
	ActionsPerMinute float64   `json:"actions_per_minute"`
	AttackActions    int       `json:"attack_actions"`
	SpellActions     int       `json:"spell_actions"`
	MoveActions      int       `json:"move_actions"`
	CriticalHitRate  float64   `json:"critical_hit_rate"`
	AccuracyRate     float64   `json:"accuracy_rate"`
	TotalDamageDealt int       `json:"total_damage_dealt"`
	DamageEvents     int       `json:"damage_events"`
	AverageDamage    float64   `json:"average_damage"`
}

// DamageResult représente le résultat d'un calcul de dégâts
type DamageResult struct {
	Amount      int    `json:"amount"`
	Type        string `json:"type"` // physical, magical
	IsCritical  bool   `json:"is_critical"`
	IsResisted  bool   `json:"is_resisted"`
	Penetration int    `json:"penetration"`
	Source      string `json:"source"`
	ElementType string `json:"element_type"`
}

// HealingResult représente le résultat d'un calcul de soins
type HealingResult struct {
	Amount       int    `json:"amount"`
	IsCritical   bool   `json:"is_critical"`
	Source       string `json:"source"`
	Overheal     int    `json:"overheal"`
	TargetHealth int    `json:"target_health"`
}

// AntiCheatStats représente les statistiques anti-cheat
type AntiCheatStats struct {
	TotalViolations      int            `json:"total_violations"`
	ViolationsByType     map[string]int `json:"violations_by_type"`
	ViolationsBySeverity map[string]int `json:"violations_by_severity"`
	ActiveBans           int            `json:"active_bans"`
	TotalBans            int            `json:"total_bans"`
}

// MovementValidation représente une validation de mouvement
type MovementValidation struct {
	CharacterID  uuid.UUID `json:"character_id"`
	OldPosition  Position  `json:"old_position"`
	NewPosition  Position  `json:"new_position"`
	Distance     float64   `json:"distance"`
	DeltaTime    float64   `json:"delta_time"` // en secondes
	Speed        float64   `json:"speed"`      // unités par seconde
	MaxSpeed     float64   `json:"max_speed"`
	IsValid      bool      `json:"is_valid"`
	Timestamp    time.Time `json:"timestamp"`
}

// DamageValidation représente une validation de dégâts
type DamageValidation struct {
	AttackerID     uuid.UUID `json:"attacker_id"`
	DefenderID     uuid.UUID `json:"defender_id"`
	DamageDealt    int       `json:"damage_dealt"`
	MaxPossible    int       `json:"max_possible"`
	DamageType     string    `json:"damage_type"`
	IsCritical     bool      `json:"is_critical"`
	IsValid        bool      `json:"is_valid"`
	ViolationType  string    `json:"violation_type,omitempty"`
	Timestamp      time.Time `json:"timestamp"`
}

// SpellValidation représente une validation de sort
type SpellValidation struct {
	CasterID       uuid.UUID     `json:"caster_id"`
	SpellID        uuid.UUID     `json:"spell_id"`
	CastTime       time.Duration `json:"cast_time"`
	ExpectedTime   time.Duration `json:"expected_time"`
	IsValid        bool          `json:"is_valid"`
	ViolationType  string        `json:"violation_type,omitempty"`
	Timestamp      time.Time     `json:"timestamp"`
}

// ActionRateCheck représente une vérification du taux d'actions
type ActionRateCheck struct {
	CharacterID       uuid.UUID `json:"character_id"`
	TimeWindow        int       `json:"time_window"` // en secondes
	ActionCount       int       `json:"action_count"`
	MaxAllowed        int       `json:"max_allowed"`
	ActionsPerSecond  float64   `json:"actions_per_second"`
	IsViolation       bool      `json:"is_violation"`
	Timestamp         time.Time `json:"timestamp"`
}

// Méthodes utilitaires

// ToJSON sérialise une violation en JSON
func (v *AntiCheatViolation) ToJSON() ([]byte, error) {
	return json.Marshal(v)
}

// FromJSON désérialise une violation depuis JSON
func (v *AntiCheatViolation) FromJSON(data []byte) error {
	return json.Unmarshal(data, v)
}

// IsExpired vérifie si un ban a expiré
func (b *AntiCheatBan) IsExpired() bool {
	if b.ExpiresAt == nil {
		return false // Ban permanent
	}
	return time.Now().After(*b.ExpiresAt)
}

// GetRemainingTime retourne le temps restant d'un ban
func (b *AntiCheatBan) GetRemainingTime() time.Duration {
	if b.ExpiresAt == nil {
		return 0 // Ban permanent
	}
	
	remaining := time.Until(*b.ExpiresAt)
	if remaining < 0 {
		return 0
	}
	
	return remaining
}

// GetRiskScore calcule un score de risque basé sur les facteurs
func (sr *SuspicionReport) GetRiskScore() float64 {
	if len(sr.Factors) == 0 {
		return 0.0
	}
	
	totalWeight := 0.0
	weightedSum := 0.0
	
	for _, factor := range sr.Factors {
		weight := factor.Weight
		if weight <= 0 {
			weight = 1.0 // Poids par défaut
		}
		
		totalWeight += weight
		
		// Convertir la sévérité en score
		severityScore := 0.0
		switch factor.Severity {
		case "low":
			severityScore = 0.25
		case "medium":
			severityScore = 0.5
		case "high":
			severityScore = 0.75
		case "critical":
			severityScore = 1.0
		}
		
		weightedSum += weight * severityScore
	}
	
	if totalWeight == 0 {
		return 0.0
	}
	
	return weightedSum / totalWeight
}

// Constants pour les types de violations
const (
	ViolationTypeSpeedHack     = "speed_hack"
	ViolationTypeDamageHack    = "damage_hack"
	ViolationTypePositionHack  = "position_hack"
	ViolationTypeActionSpam    = "action_spam"
	ViolationTypeInstantCast   = "instant_cast"
	ViolationTypeNoClip        = "no_clip"
	ViolationTypeFly           = "fly_hack"
	ViolationTypeTeleport      = "teleport_hack"
	ViolationTypeCooldownBypass = "cooldown_bypass"
	ViolationTypeResourceHack  = "resource_hack"
)

// Constants pour les niveaux de sévérité
const (
	SeverityLow      = "low"
	SeverityMedium   = "medium"
	SeverityHigh     = "high"
	SeverityCritical = "critical"
)

// Constants pour les niveaux de risque
const (
	RiskLevelLow      = "low"
	RiskLevelMedium   = "medium"
	RiskLevelHigh     = "high"
	RiskLevelCritical = "critical"
)