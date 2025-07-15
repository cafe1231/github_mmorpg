// internal/models/combat_log.go
package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// CombatLog représente une entrée dans les logs de combat
type CombatLog struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	SessionID   uuid.UUID       `json:"session_id" db:"session_id"`
	ActionID    *uuid.UUID      `json:"action_id" db:"action_id"` // référence à l'action qui a généré ce log

	// Participants
	ActorID     *uuid.UUID      `json:"actor_id" db:"actor_id"`   // qui a fait l'action
	TargetID    *uuid.UUID      `json:"target_id" db:"target_id"` // qui a reçu l'action

	// Type d'événement
	EventType   string          `json:"event_type" db:"event_type"` // damage, heal, death, spell_cast, effect_applied, etc.

	// Messages
	Message     string          `json:"message" db:"message"`          // message formaté pour l'affichage
	RawData     json.RawMessage `json:"raw_data" db:"raw_data"`      // données brutes de l'événement

	// Valeurs numériques
	Value       int             `json:"value" db:"value"`               // dégâts, soins, etc.
	OldValue    int             `json:"old_value" db:"old_value"`     // valeur avant l'action
	NewValue    int             `json:"new_value" db:"new_value"`      // valeur après l'action

	// Contexte
	IsCritical  bool            `json:"is_critical" db:"is_critical"`
	IsResisted  bool            `json:"is_resisted" db:"is_resisted"`
	IsAbsorbed  bool            `json:"is_absorbed" db:"is_absorbed"`

	// Métadonnées
	Color       string          `json:"color" db:"color"`               // couleur pour l'affichage
	Icon        string          `json:"icon" db:"icon"`
	Priority    int             `json:"priority" db:"priority"`       // pour le tri d'affichage

	Timestamp   time.Time       `json:"timestamp" db:"timestamp"`
}

// CombatSummary représente un résumé de combat
type CombatSummary struct {
	SessionID       uuid.UUID             `json:"session_id"`
	Type            string                `json:"type"`
	Status          string                `json:"status"`
	Duration        time.Duration         `json:"duration"`

	// Participants
	ParticipantCount int                   `json:"participant_count"`
	Winners         []uuid.UUID           `json:"winners"`
	Losers          []uuid.UUID           `json:"losers"`

	// Statistiques générales
	TotalDamage     int64                 `json:"total_damage"`
	TotalHealing    int64                 `json:"total_healing"`
	ActionsCount    int                   `json:"actions_count"`
	SpellsCast      int                   `json:"spells_cast"`

	// Statistiques par participant
	ParticipantStats []ParticipantSummary `json:"participant_stats"`

	StartedAt       time.Time             `json:"started_at"`
	EndedAt         time.Time             `json:"ended_at"`
}

// ParticipantSummary statistiques d'un participant
type ParticipantSummary struct {
	CharacterID     uuid.UUID            `json:"character_id"`
	CharacterName   string               `json:"character_name"`
	Team            int                  `json:"team"`

	// Statistiques de combat
	DamageDealt     int64                `json:"damage_dealt"`
	DamageTaken     int64                `json:"damage_taken"`
	HealingDone     int64                `json:"healing_done"`
	HealingReceived int64                `json:"healing_received"`

	// Actions
	ActionsPerformed int                   `json:"actions_performed"`
	SpellsCast       int                   `json:"spells_cast"`
	CriticalHits     int                   `json:"critical_hits"`

	// Résultat
	Survived        bool                 `json:"survived"`
	KillCount       int                  `json:"kill_count"`
	DeathCount      int                  `json:"death_count"`

	// Temps
	TimeInCombat    time.Duration        `json:"time_in_combat"`
}

// LogEvent représente un événement spécifique pour les logs
type LogEvent struct {
	Type        string      `json:"type"`
	ActorName   string      `json:"actor_name,omitempty"`
	TargetName  string      `json:"target_name,omitempty"`
	SpellName   string      `json:"spell_name,omitempty"`
	ItemName    string      `json:"item_name,omitempty"`
	Value       int         `json:"value,omitempty"`
	IsCritical  bool        `json:"is_critical,omitempty"`
	Timestamp   time.Time   `json:"timestamp"`
}
