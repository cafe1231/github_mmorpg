package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/dan-2/github_mmorpg/services/analytics/internal/models"
	"github.com/dan-2/github_mmorpg/services/analytics/internal/repository"
)

type metricsService struct {
	metricRepo repository.MetricRepository
}

func NewMetricsService(metricRepo repository.MetricRepository) MetricsService {
	return &metricsService{
		metricRepo: metricRepo,
	}
}

func (s *metricsService) RecordMetric(ctx context.Context, name string, value float64, tags map[string]string) error {
	tagsJSON := "{}"
	if tags != nil {
		if tagsBytes, err := json.Marshal(tags); err == nil {
			tagsJSON = string(tagsBytes)
		}
	}

	metric := &models.Metric{
		Name:  name,
		Value: value,
		Date:  time.Now().Truncate(24 * time.Hour), // Agrégation par jour
		Tags:  tagsJSON,
	}

	return s.metricRepo.Create(ctx, metric)
}

func (s *metricsService) GetMetrics(ctx context.Context, req *models.GetMetricsRequest) ([]*models.Metric, error) {
	tagsJSON := "{}"
	if req.Tags != nil {
		if tagsBytes, err := json.Marshal(req.Tags); err == nil {
			tagsJSON = string(tagsBytes)
		}
	}

	return s.metricRepo.Aggregate(ctx, req.Name, req.From, req.To, tagsJSON)
}

func (s *metricsService) GetMetric(ctx context.Context, name string, date time.Time, tags map[string]string) (*models.Metric, error) {
	tagsJSON := "{}"
	if tags != nil {
		if tagsBytes, err := json.Marshal(tags); err == nil {
			tagsJSON = string(tagsBytes)
		}
	}

	return s.metricRepo.GetByNameAndDate(ctx, name, date, tagsJSON)
}
