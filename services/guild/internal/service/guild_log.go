package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/dan-2/github_mmorpg/services/guild/internal/models"
	"github.com/google/uuid"
)

// guildLogService implémente GuildLogService
type guildLogService struct {
	db *sql.DB
}

// NewGuildLogService crée une nouvelle instance de GuildLogService
func NewGuildLogService(db *sql.DB) GuildLogService {
	return &guildLogService{db: db}
}

// AddLog ajoute un log à une guilde
func (s *guildLogService) AddLog(ctx context.Context, guildID, playerID uuid.UUID, action, details string) error {
	query := `
		INSERT INTO guild_logs (id, guild_id, player_id, action, details, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	logID := uuid.New()
	now := time.Now()

	_, err := s.db.ExecContext(ctx, query, logID, guildID, playerID, action, details, now)
	if err != nil {
		return fmt.Errorf("erreur lors de l'ajout du log: %w", err)
	}

	return nil
}

// GetLogs récupère les logs d'une guilde avec filtres
func (s *guildLogService) GetLogs(ctx context.Context, guildID uuid.UUID, action *string,
	page, limit int) ([]*models.GuildLogResponse, int, error) {
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

	// Compter le total avec requête préparée
	countQuery := "SELECT COUNT(*) FROM (" + baseQuery + ") as count_query"
	var total int
	err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("erreur lors du comptage des logs: %w", err)
	}

	// Récupérer les logs avec pagination
	baseQuery += " ORDER BY created_at DESC LIMIT $" + fmt.Sprintf("%d", argIndex) + " OFFSET $" + fmt.Sprintf("%d", argIndex+1)
	args = append(args, limit, (page-1)*limit)

	rows, err := s.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("erreur lors de la récupération des logs: %w", err)
	}
	defer rows.Close()

	var logs []*models.GuildLogResponse
	for rows.Next() {
		log := &models.GuildLogResponse{}
		err := rows.Scan(&log.ID, &log.PlayerID, &log.Action, &log.Details, &log.CreatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("erreur lors du scan des logs: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, total, nil
}

// CleanOldLogs supprime les logs anciens
func (s *guildLogService) CleanOldLogs(ctx context.Context, days int) error {
	query := `
		DELETE FROM guild_logs 
		WHERE created_at < NOW() - INTERVAL '%d days'
	`

	_, err := s.db.ExecContext(ctx, fmt.Sprintf(query, days))
	if err != nil {
		return fmt.Errorf("erreur lors du nettoyage des logs: %w", err)
	}

	return nil
}
