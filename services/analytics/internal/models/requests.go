package models

import (
	"time"

	"github.com/google/uuid"
)

// Enregistrement d'un événement
type CreateEventRequest struct {
	Type      string     `json:"type" binding:"required"`
	PlayerID  *uuid.UUID `json:"player_id,omitempty"`
	GuildID   *uuid.UUID `json:"guild_id,omitempty"`
	Payload   string     `json:"payload" binding:"required"`
	Timestamp *time.Time `json:"timestamp,omitempty"`
}

// Requête de métriques agrégées
type GetMetricsRequest struct {
	Name string            `json:"name" binding:"required"`
	From time.Time         `json:"from" binding:"required"`
	To   time.Time         `json:"to" binding:"required"`
	Tags map[string]string `json:"tags,omitempty"`
}

// Requête de logs
type GetLogsRequest struct {
	Level   *string    `json:"level,omitempty"`
	From    *time.Time `json:"from,omitempty"`
	To      *time.Time `json:"to,omitempty"`
	Context *string    `json:"context,omitempty"`
	Limit   int        `json:"limit" binding:"min=1,max=100"`
	Page    int        `json:"page" binding:"min=1"`
}
