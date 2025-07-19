package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/dan-2/github_mmorpg/services/analytics/internal/models"
)

type metricRepository struct {
	db *sql.DB
}

func NewMetricRepository(db *sql.DB) MetricRepository {
	return &metricRepository{db: db}
}

func (r *metricRepository) Create(ctx context.Context, metric *models.Metric) error {
	query := `
		INSERT INTO analytics_metrics (id, name, value, date, tags)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (name, date, tags) DO UPDATE SET
		value = EXCLUDED.value
	`

	metric.ID = uuid.New()

	_, err := r.db.ExecContext(ctx, query,
		metric.ID, metric.Name, metric.Value, metric.Date, metric.Tags,
	)
	return err
}

func (r *metricRepository) GetByNameAndDate(ctx context.Context, name string, date time.Time, tags string) (*models.Metric, error) {
	query := `
		SELECT id, name, value, date, tags
		FROM analytics_metrics 
		WHERE name = $1 AND date = $2 AND tags = $3
	`

	metric := &models.Metric{}
	err := r.db.QueryRowContext(ctx, query, name, date, tags).Scan(
		&metric.ID, &metric.Name, &metric.Value, &metric.Date, &metric.Tags,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("métrique non trouvée")
		}
		return nil, err
	}

	return metric, nil
}

func (r *metricRepository) Aggregate(ctx context.Context, name string, from, to time.Time, tags string) ([]*models.Metric, error) {
	query := `
		SELECT id, name, value, date, tags
		FROM analytics_metrics 
		WHERE name = $1 AND date >= $2 AND date <= $3 AND tags = $4
		ORDER BY date ASC
	`

	rows, err := r.db.QueryContext(ctx, query, name, from, to, tags)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []*models.Metric
	for rows.Next() {
		metric := &models.Metric{}
		err := rows.Scan(
			&metric.ID, &metric.Name, &metric.Value, &metric.Date, &metric.Tags,
		)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, metric)
	}

	return metrics, nil
}
