package models

import (
	"time"

	"github.com/google/uuid"
)

// Event représente un événement de jeu (connexion, achat, combat, etc.)
type Event struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	Type      string     `json:"type" db:"type"` // ex: login, purchase, combat, quest
	PlayerID  *uuid.UUID `json:"player_id,omitempty" db:"player_id"`
	GuildID   *uuid.UUID `json:"guild_id,omitempty" db:"guild_id"`
	Timestamp time.Time  `json:"timestamp" db:"timestamp"`
	Payload   string     `json:"payload" db:"payload"` // JSON (détails spécifiques)
}

// Metric représente une métrique agrégée (DAU, revenus, etc.)
type Metric struct {
	ID    uuid.UUID `json:"id" db:"id"`
	Name  string    `json:"name" db:"name"` // ex: dau, revenue, pvp_matches
	Value float64   `json:"value" db:"value"`
	Date  time.Time `json:"date" db:"date"`
	Tags  string    `json:"tags" db:"tags"` // JSON (ex: {"zone":"forest"})
}

// LogEntry représente un log d'activité ou d'erreur
type LogEntry struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Level     string    `json:"level" db:"level"` // info, warn, error
	Message   string    `json:"message" db:"message"`
	Context   string    `json:"context" db:"context"` // JSON (détails)
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
}
