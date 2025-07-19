package repository

import (
	"context"
	"database/sql"
	"math"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/dan-2/github_mmorpg/services/guild/internal/models"
	"github.com/google/uuid"
)

const (
	// DefaultMaxMembers est le nombre maximum de membres par défaut
	DefaultMaxMembers = 50
	// DefaultPageSize est la taille de page par défaut
	DefaultPageSize = 10
	// MaxPageSize est la taille de page maximale
	MaxPageSize = 100
	// Pagination limits
	MaxLimit  = 1000
	MaxOffset = 100000
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

// buildSearchQuery construit la requête de recherche avec les filtres
func (r *guildRepository) buildSearchQuery(name, tag *string, minLevel, maxLevel *int) (
	selectBuilder squirrel.SelectBuilder, countBuilder squirrel.SelectBuilder) {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	selectBuilder = psql.Select(
		"id", "name", "description", "tag", "level", "experience", "max_members", "created_at", "updated_at",
	).From("guilds")

	countBuilder = psql.Select("COUNT(*)").From("guilds")

	if name != nil && *name != "" {
		selectBuilder = selectBuilder.Where("name ILIKE ?", "%"+*name+"%")
		countBuilder = countBuilder.Where("name ILIKE ?", "%"+*name+"%")
	}
	if tag != nil && *tag != "" {
		selectBuilder = selectBuilder.Where("tag ILIKE ?", "%"+*tag+"%")
		countBuilder = countBuilder.Where("tag ILIKE ?", "%"+*tag+"%")
	}
	if minLevel != nil {
		selectBuilder = selectBuilder.Where("level >= ?", *minLevel)
		countBuilder = countBuilder.Where("level >= ?", *minLevel)
	}
	if maxLevel != nil {
		selectBuilder = selectBuilder.Where("level <= ?", *maxLevel)
		countBuilder = countBuilder.Where("level <= ?", *maxLevel)
	}

	return selectBuilder, countBuilder
}

// Search recherche des guildes avec filtres
func (r *guildRepository) Search(ctx context.Context, name, tag *string, minLevel, maxLevel *int, page, limit int) (
	[]*models.Guild, int, error) {
	selectBuilder, countBuilder := r.buildSearchQuery(name, tag, minLevel, maxLevel)

	// Pagination sécurisée
	var safeLimit, safeOffset uint64
	if limit > 0 {
		safeLimit = uint64(math.Min(float64(limit), float64(MaxLimit)))
	}
	if page > 1 && limit > 0 {
		offset := (page - 1) * limit
		if offset > 0 {
			safeOffset = uint64(math.Min(float64(offset), float64(MaxOffset)))
		}
	}

	selectBuilder = selectBuilder.OrderBy("level DESC", "experience DESC").Limit(safeLimit).Offset(safeOffset)

	// Compter le total
	countSql, countArgs, err := countBuilder.ToSql()
	if err != nil {
		return nil, 0, err
	}
	var total int
	err = r.db.QueryRowContext(ctx, countSql, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Récupérer les résultats
	query, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, 0, err
	}

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
