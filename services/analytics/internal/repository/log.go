package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/dan-2/github_mmorpg/services/analytics/internal/models"
)

type logRepository struct {
	db *sql.DB
}

func NewLogRepository(db *sql.DB) LogRepository {
	return &logRepository{db: db}
}

func (r *logRepository) Create(ctx context.Context, log *models.LogEntry) error {
	query := `
		INSERT INTO analytics_logs (id, level, message, context, timestamp)
		VALUES ($1, $2, $3, $4, $5)
	`

	log.ID = uuid.New()
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now()
	}

	_, err := r.db.ExecContext(ctx, query,
		log.ID, log.Level, log.Message, log.Context, log.Timestamp,
	)
	return err
}

func (r *logRepository) List(
	ctx context.Context,
	level *string,
	from, to *time.Time,
	contextFilter *string,
	page, limit int,
) ([]*models.LogEntry, int, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	if level != nil && *level != "" {
		conditions = append(conditions, fmt.Sprintf("level = $%d", argIndex))
		args = append(args, *level)
		argIndex++
	}

	if from != nil {
		conditions = append(conditions, fmt.Sprintf("timestamp >= $%d", argIndex))
		args = append(args, *from)
		argIndex++
	}

	if to != nil {
		conditions = append(conditions, fmt.Sprintf("timestamp <= $%d", argIndex))
		args = append(args, *to)
		argIndex++
	}

	if contextFilter != nil && *contextFilter != "" {
		conditions = append(conditions, fmt.Sprintf("context ILIKE $%d", argIndex))
		args = append(args, "%"+*contextFilter+"%")
		argIndex++
	}

	var whereClause string
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Compter le total
	countQuery := "SELECT COUNT(*) FROM analytics_logs " + whereClause
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Récupérer les données
	baseQuery := "SELECT id, level, message, context, timestamp FROM analytics_logs " + whereClause + " ORDER BY timestamp DESC"
	query := baseQuery + " LIMIT $" + fmt.Sprintf("%d", argIndex) + " OFFSET $" + fmt.Sprintf("%d", argIndex+1)
	args = append(args, limit, (page-1)*limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []*models.LogEntry
	for rows.Next() {
		log := &models.LogEntry{}
		err := rows.Scan(&log.ID, &log.Level, &log.Message, &log.Context, &log.Timestamp)
		if err != nil {
			return nil, 0, err
		}
		logs = append(logs, log)
	}

	return logs, total, nil
}
