package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/dan-2/github_mmorpg/services/analytics/internal/models"
	"github.com/google/uuid"
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

func (r *logRepository) List(ctx context.Context, level *string, from, to *time.Time, contextFilter *string, page, limit int) ([]*models.LogEntry, int, error) {
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

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Compter le total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM analytics_logs %s", whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Récupérer les résultats
	offset := (page - 1) * limit
	query := fmt.Sprintf(`
		SELECT id, level, message, context, timestamp
		FROM analytics_logs %s
		ORDER BY timestamp DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []*models.LogEntry
	for rows.Next() {
		log := &models.LogEntry{}
		err := rows.Scan(
			&log.ID, &log.Level, &log.Message, &log.Context, &log.Timestamp,
		)
		if err != nil {
			return nil, 0, err
		}
		logs = append(logs, log)
	}

	return logs, total, nil
}
