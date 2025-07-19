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

type eventRepository struct {
	db *sql.DB
}

func NewEventRepository(db *sql.DB) EventRepository {
	return &eventRepository{db: db}
}

func (r *eventRepository) Create(ctx context.Context, event *models.Event) error {
	query := `
		INSERT INTO analytics_events (id, type, player_id, guild_id, timestamp, payload)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	event.ID = uuid.New()
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	_, err := r.db.ExecContext(ctx, query,
		event.ID, event.Type, event.PlayerID, event.GuildID,
		event.Timestamp, event.Payload,
	)
	return err
}

func (r *eventRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Event, error) {
	query := `
		SELECT id, type, player_id, guild_id, timestamp, payload
		FROM analytics_events WHERE id = $1
	`

	event := &models.Event{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&event.ID, &event.Type, &event.PlayerID, &event.GuildID,
		&event.Timestamp, &event.Payload,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("événement non trouvé")
		}
		return nil, err
	}

	return event, nil
}

func (r *eventRepository) List(ctx context.Context, eventType *string, from, to *time.Time, playerID, guildID *uuid.UUID, page, limit int) ([]*models.Event, int, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	if eventType != nil {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argIndex))
		args = append(args, *eventType)
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

	if playerID != nil {
		conditions = append(conditions, fmt.Sprintf("player_id = $%d", argIndex))
		args = append(args, *playerID)
		argIndex++
	}

	if guildID != nil {
		conditions = append(conditions, fmt.Sprintf("guild_id = $%d", argIndex))
		args = append(args, *guildID)
		argIndex++
	}

	var whereClause string
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Compter le total
	countQuery := "SELECT COUNT(*) FROM analytics_events " + whereClause
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Récupérer les données
	baseQuery := "SELECT id, type, player_id, guild_id, timestamp, payload FROM analytics_events " + whereClause + " ORDER BY timestamp DESC"
	query := baseQuery + " LIMIT $" + fmt.Sprintf("%d", argIndex) + " OFFSET $" + fmt.Sprintf("%d", argIndex+1)
	args = append(args, limit, (page-1)*limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		event := &models.Event{}
		err := rows.Scan(&event.ID, &event.Type, &event.PlayerID, &event.GuildID, &event.Timestamp, &event.Payload)
		if err != nil {
			return nil, 0, err
		}
		events = append(events, event)
	}

	return events, total, nil
}
