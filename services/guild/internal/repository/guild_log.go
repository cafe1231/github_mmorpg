package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/dan-2/github_mmorpg/services/guild/internal/models"
	"github.com/google/uuid"
)

// guildLogRepository implémente GuildLogRepository
type guildLogRepository struct {
	db *sql.DB
}

// NewGuildLogRepository crée une nouvelle instance de GuildLogRepository
func NewGuildLogRepository(db *sql.DB) GuildLogRepository {
	return &guildLogRepository{db: db}
}

// Create crée un nouveau log
func (r *guildLogRepository) Create(ctx context.Context, log *models.GuildLog) error {
	query := `
		INSERT INTO guild_logs (id, guild_id, player_id, action, details, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	log.ID = uuid.New()
	log.CreatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query, log.ID, log.GuildID, log.PlayerID, log.Action, log.Details, log.CreatedAt)
	return err
}

// GetByGuild récupère les logs d'une guilde avec pagination et filtres
func (r *guildLogRepository) GetByGuild(ctx context.Context, guildID uuid.UUID, action *string, page, limit int) ([]*models.GuildLog, int, error) {
	// Construire la requête avec filtres
	baseQuery := `
		SELECT id, guild_id, player_id, action, details, created_at
		FROM guild_logs 
		WHERE guild_id = $1
	`

	var args []interface{}
	args = append(args, guildID)
	argIndex := 2

	if action != nil && *action != "" {
		baseQuery += fmt.Sprintf(" AND action = $%d", argIndex)
		args = append(args, *action)
		argIndex++
	}

	// Compter le total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM (%s) as count_query", baseQuery)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("erreur lors du comptage des logs: %w", err)
	}

	// Récupérer les logs avec pagination
	baseQuery += " ORDER BY created_at DESC LIMIT $%d OFFSET $%d"
	args = append(args, limit, (page-1)*limit)

	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("erreur lors de la récupération des logs: %w", err)
	}
	defer rows.Close()

	var logs []*models.GuildLog
	for rows.Next() {
		log := &models.GuildLog{}
		err := rows.Scan(&log.ID, &log.GuildID, &log.PlayerID, &log.Action, &log.Details, &log.CreatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("erreur lors du scan des logs: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, total, nil
}

// GetCountByGuildID compte le nombre de logs d'une guilde
func (r *guildLogRepository) GetCountByGuildID(ctx context.Context, guildID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM guild_logs WHERE guild_id = $1`

	var count int
	err := r.db.QueryRowContext(ctx, query, guildID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("erreur lors du comptage des logs: %w", err)
	}

	return count, nil
}

// DeleteOldLogs supprime les logs anciens
func (r *guildLogRepository) DeleteOldLogs(ctx context.Context, days int) error {
	query := `
		DELETE FROM guild_logs 
		WHERE created_at < NOW() - INTERVAL '%d days'
	`

	_, err := r.db.ExecContext(ctx, fmt.Sprintf(query, days))
	if err != nil {
		return fmt.Errorf("erreur lors du nettoyage des logs: %w", err)
	}

	return nil
}
