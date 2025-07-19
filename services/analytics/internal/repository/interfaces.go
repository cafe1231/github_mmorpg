package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/dan-2/github_mmorpg/services/analytics/internal/models"
)

type EventRepository interface {
	Create(ctx context.Context, event *models.Event) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Event, error)
	List(
		ctx context.Context,
		eventType *string,
		from, to *time.Time,
		playerID, guildID *uuid.UUID,
		page, limit int,
	) ([]*models.Event, int, error)
}

type MetricRepository interface {
	Create(ctx context.Context, metric *models.Metric) error
	GetByNameAndDate(ctx context.Context, name string, date time.Time, tags string) (*models.Metric, error)
	Aggregate(ctx context.Context, name string, from, to time.Time, tags string) ([]*models.Metric, error)
}

type LogRepository interface {
	Create(ctx context.Context, log *models.LogEntry) error
	List(ctx context.Context, level *string, from, to *time.Time, contextFilter *string, page, limit int) ([]*models.LogEntry, int, error)
}
