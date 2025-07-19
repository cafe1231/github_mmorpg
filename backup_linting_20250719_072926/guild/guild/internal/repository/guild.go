package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/dan-2/github_mmorpg/services/guild/internal/models"
	"github.com/google/uuid"
)

// guildRepository implémente GuildRepository
type guildRepository struct {
	db *sql.DB
}

// NewGuildRepository crée une nouvelle instance de GuildRepository
func NewGuildRepository(db *sql.DB) GuildRepository {
	return &guildRepository{db: db}
}

// Create crée une nouvelle guilde
func (r *guildRepository) Create(ctx context.Context, guild *models.Guild) error {
	query := `
		INSERT INTO guilds (id, name, description, tag, level, experience, max_members, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	now := time.Now()
	guild.ID = uuid.New()
	guild.CreatedAt = now
	guild.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		guild.ID, guild.Name, guild.Description, guild.Tag,
		guild.Level, guild.Experience, guild.MaxMembers,
		guild.CreatedAt, guild.UpdatedAt,
	)
	return err
}

// GetByID récupère une guilde par son ID
func (r *guildRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Guild, error) {
	query := `
		SELECT id, name, description, tag, level, experience, max_members, created_at, updated_at
		FROM guilds WHERE id = $1
	`

	guild := &models.Guild{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&guild.ID, &guild.Name, &guild.Description, &guild.Tag,
		&guild.Level, &guild.Experience, &guild.MaxMembers,
		&guild.CreatedAt, &guild.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrGuildNotFound
		}
		return nil, err
	}

	return guild, nil
}

// GetByName récupère une guilde par son nom
func (r *guildRepository) GetByName(ctx context.Context, name string) (*models.Guild, error) {
	query := `
		SELECT id, name, description, tag, level, experience, max_members, created_at, updated_at
		FROM guilds WHERE name = $1
	`

	guild := &models.Guild{}
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&guild.ID, &guild.Name, &guild.Description, &guild.Tag,
		&guild.Level, &guild.Experience, &guild.MaxMembers,
		&guild.CreatedAt, &guild.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrGuildNotFound
		}
		return nil, err
	}

	return guild, nil
}

// GetByTag récupère une guilde par son tag
func (r *guildRepository) GetByTag(ctx context.Context, tag string) (*models.Guild, error) {
	query := `
		SELECT id, name, description, tag, level, experience, max_members, created_at, updated_at
		FROM guilds WHERE tag = $1
	`

	guild := &models.Guild{}
	err := r.db.QueryRowContext(ctx, query, tag).Scan(
		&guild.ID, &guild.Name, &guild.Description, &guild.Tag,
		&guild.Level, &guild.Experience, &guild.MaxMembers,
		&guild.CreatedAt, &guild.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrGuildNotFound
		}
		return nil, err
	}

	return guild, nil
}

// Update met à jour une guilde
func (r *guildRepository) Update(ctx context.Context, guild *models.Guild) error {
	query := `
		UPDATE guilds 
		SET name = $2, description = $3, tag = $4, level = $5, 
		    experience = $6, max_members = $7, updated_at = $8
		WHERE id = $1
	`

	guild.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		guild.ID, guild.Name, guild.Description, guild.Tag,
		guild.Level, guild.Experience, guild.MaxMembers, guild.UpdatedAt,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return models.ErrGuildNotFound
	}

	return nil
}

// Delete supprime une guilde
func (r *guildRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM guilds WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return models.ErrGuildNotFound
	}

	return nil
}

// Search recherche des guildes avec filtres
func (r *guildRepository) Search(ctx context.Context, name, tag *string, minLevel, maxLevel *int, page, limit int) ([]*models.Guild, int, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	if name != nil && *name != "" {
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argIndex))
		args = append(args, "%"+*name+"%")
		argIndex++
	}

	if tag != nil && *tag != "" {
		conditions = append(conditions, fmt.Sprintf("tag ILIKE $%d", argIndex))
		args = append(args, "%"+*tag+"%")
		argIndex++
	}

	if minLevel != nil {
		conditions = append(conditions, fmt.Sprintf("level >= $%d", argIndex))
		args = append(args, *minLevel)
		argIndex++
	}

	if maxLevel != nil {
		conditions = append(conditions, fmt.Sprintf("level <= $%d", argIndex))
		args = append(args, *maxLevel)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Compter le total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM guilds %s", whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Récupérer les résultats
	offset := (page - 1) * limit
	query := fmt.Sprintf(`
		SELECT id, name, description, tag, level, experience, max_members, created_at, updated_at
		FROM guilds %s
		ORDER BY level DESC, experience DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var guilds []*models.Guild
	for rows.Next() {
		guild := &models.Guild{}
		err := rows.Scan(
			&guild.ID, &guild.Name, &guild.Description, &guild.Tag,
			&guild.Level, &guild.Experience, &guild.MaxMembers,
			&guild.CreatedAt, &guild.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		guilds = append(guilds, guild)
	}

	return guilds, total, nil
}

// List récupère une liste de guildes
func (r *guildRepository) List(ctx context.Context, page, limit int) ([]*models.Guild, int, error) {
	offset := (page - 1) * limit

	// Compter le total
	var total int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM guilds").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Récupérer les résultats
	query := `
		SELECT id, name, description, tag, level, experience, max_members, created_at, updated_at
		FROM guilds
		ORDER BY level DESC, experience DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var guilds []*models.Guild
	for rows.Next() {
		guild := &models.Guild{}
		err := rows.Scan(
			&guild.ID, &guild.Name, &guild.Description, &guild.Tag,
			&guild.Level, &guild.Experience, &guild.MaxMembers,
			&guild.CreatedAt, &guild.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		guilds = append(guilds, guild)
	}

	return guilds, total, nil
}
