package models

import (
	"time"

	"github.com/google/uuid"
)

// CombatActionRequest représente une requête d'action de combat
type CombatActionRequest struct {
	SessionID  uuid.UUID              `json:"session_id"`
	ActorID    uuid.UUID              `json:"actor_id"`
	Type       string                 `json:"type"` // attack, spell, move, item, defend
	Targets    []uuid.UUID            `json:"targets"`
	ActionData map[string]interface{} `json:"action_data"`
}

// CombatActionResult représente le résultat d'une action de combat
type CombatActionResult struct {
	Results     []ActionResult `json:"results"`
	Success     bool           `json:"success"`
	CriticalHit bool           `json:"critical_hit"`
	Duration    time.Duration  `json:"duration"`
}

// NOTE: ActionResult est déjà défini dans combat.go, pas besoin de le redéfinir ici

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
	ID          uuid.UUID  `json:"id" db:"id"`
	CharacterID uuid.UUID  `json:"character_id" db:"character_id"`
	Reason      string     `json:"reason" db:"reason"`
	Duration    time.Duration `json:"duration" db:"duration"`
	ExpiresAt   *time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	IsActive    bool       `json:"is_active" db:"is_active"`
}

// SuspicionReport représente un rapport de suspicion
type SuspicionReport struct {
	CharacterID     uuid.UUID         `json:"character_id"`
	SuspicionScore  float64           `json:"suspicion_score"`
	RiskLevel       string            `json:"risk_level"` // low, medium, high, critical
	Factors         []SuspicionFactor `json:"factors"`
	Timestamp       time.Time         `json:"timestamp"`
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
	CharacterID       uuid.UUID `json:"character_id"`
	StartTime         time.Time `json:"start_time"`
	TotalActions      int       `json:"total_actions"`
	ActionsPerMinute  float64   `json:"actions_per_minute"`
	AttackActions     int       `json:"attack_actions"`
	SpellActions      int       `json:"spell_actions"`
	MoveActions       int       `json:"move_actions"`
	CriticalHitRate   float64   `json:"critical_hit_rate"`
	AccuracyRate      float64   `json:"accuracy_rate"`
	TotalDamageDealt  int       `json:"total_damage_dealt"`
	DamageEvents      int       `json:"damage_events"`
	AverageDamage     float64   `json:"average_damage"`
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