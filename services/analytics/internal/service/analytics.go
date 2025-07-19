package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/dan-2/github_mmorpg/services/analytics/internal/models"
	"github.com/dan-2/github_mmorpg/services/analytics/internal/repository"
)

type analyticsService struct {
	eventRepo  repository.EventRepository
	metricRepo repository.MetricRepository
	logRepo    repository.LogRepository
}

func NewAnalyticsService(
	eventRepo repository.EventRepository,
	metricRepo repository.MetricRepository,
	logRepo repository.LogRepository,
) AnalyticsService {
	return &analyticsService{
		eventRepo:  eventRepo,
		metricRepo: metricRepo,
		logRepo:    logRepo,
	}
}

func (s *analyticsService) TrackEvent(ctx context.Context, req *models.CreateEventRequest) error {
	event := &models.Event{
		Type:      req.Type,
		PlayerID:  req.PlayerID,
		GuildID:   req.GuildID,
		Payload:   req.Payload,
		Timestamp: time.Now(),
	}

	if req.Timestamp != nil {
		event.Timestamp = *req.Timestamp
	}

	return s.eventRepo.Create(ctx, event)
}

func (s *analyticsService) GetEvents(
	ctx context.Context,
	eventType *string,
	from, to *time.Time,
	playerID, guildID *uuid.UUID,
	page, limit int,
) ([]*models.Event, int, error) {
	return s.eventRepo.List(ctx, eventType, from, to, playerID, guildID, page, limit)
}

func (s *analyticsService) GetEvent(ctx context.Context, eventID uuid.UUID) (*models.Event, error) {
	return s.eventRepo.GetByID(ctx, eventID)
}
