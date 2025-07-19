package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/dan-2/github_mmorpg/services/analytics/internal/models"
)

type AnalyticsService interface {
	TrackEvent(ctx context.Context, req *models.CreateEventRequest) error
	GetEvents(
		ctx context.Context,
		eventType *string,
		from, to *time.Time,
		playerID, guildID *uuid.UUID,
		page, limit int,
	) ([]*models.Event, int, error)
	GetEvent(ctx context.Context, eventID uuid.UUID) (*models.Event, error)
}

type MetricsService interface {
	RecordMetric(ctx context.Context, name string, value float64, tags map[string]string) error
	GetMetrics(ctx context.Context, req *models.GetMetricsRequest) ([]*models.Metric, error)
	GetMetric(ctx context.Context, name string, date time.Time, tags map[string]string) (*models.Metric, error)
}

type LoggingService interface {
	Log(ctx context.Context, level, message string, context map[string]interface{}) error
	GetLogs(ctx context.Context, req *models.GetLogsRequest) ([]*models.LogEntry, int, error)
}
