package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/dan-2/github_mmorpg/services/guild/internal/models"
	"github.com/google/uuid"
)

// guildMemberRepository implémente GuildMemberRepository
type guildMemberRepository struct {
	db *sql.DB
}

// NewGuildMemberRepository crée une nouvelle instance de GuildMemberRepository
func NewGuildMemberRepository(db *sql.DB) GuildMemberRepository {
	return &guildMemberRepository{db: db}
}

// Create crée un nouveau membre de guilde
func (r *guildMemberRepository) Create(ctx context.Context, member *models.GuildMember) error {
	query := `
		INSERT INTO guild_members (id, guild_id, player_id, role, joined_at, last_seen, contribution)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	now := time.Now()
	member.ID = uuid.New()
	member.JoinedAt = now
	member.LastSeen = now

	_, err := r.db.ExecContext(ctx, query,
		member.ID, member.GuildID, member.PlayerID, member.Role,
		member.JoinedAt, member.LastSeen, member.Contribution,
	)
	return err
}

// GetByID récupère un membre par son ID
func (r *guildMemberRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.GuildMember, error) {
	query := `
		SELECT id, guild_id, player_id, role, joined_at, last_seen, contribution
		FROM guild_members WHERE id = $1
	`

	member := &models.GuildMember{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&member.ID, &member.GuildID, &member.PlayerID, &member.Role,
		&member.JoinedAt, &member.LastSeen, &member.Contribution,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrMemberNotFound
		}
		return nil, err
	}

	return member, nil
}

// GetByGuildAndPlayer récupère un membre par guilde et joueur
func (r *guildMemberRepository) GetByGuildAndPlayer(ctx context.Context, guildID, playerID uuid.UUID) (*models.GuildMember, error) {
	query := `
		SELECT id, guild_id, player_id, role, joined_at, last_seen, contribution
		FROM guild_members WHERE guild_id = $1 AND player_id = $2
	`

	member := &models.GuildMember{}
	err := r.db.QueryRowContext(ctx, query, guildID, playerID).Scan(
		&member.ID, &member.GuildID, &member.PlayerID, &member.Role,
		&member.JoinedAt, &member.LastSeen, &member.Contribution,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrMemberNotFound
		}
		return nil, err
	}

	return member, nil
}

// GetByGuild récupère les membres d'une guilde
func (r *guildMemberRepository) GetByGuild(ctx context.Context, guildID uuid.UUID, page, limit int) ([]*models.GuildMember, int, error) {
	offset := (page - 1) * limit

	// Compter le total
	var total int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM guild_members WHERE guild_id = $1", guildID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Récupérer les résultats
	query := `
		SELECT id, guild_id, player_id, role, joined_at, last_seen, contribution
		FROM guild_members 
		WHERE guild_id = $1
		ORDER BY role ASC, joined_at ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, guildID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var members []*models.GuildMember
	for rows.Next() {
		member := &models.GuildMember{}
		err := rows.Scan(
			&member.ID, &member.GuildID, &member.PlayerID, &member.Role,
			&member.JoinedAt, &member.LastSeen, &member.Contribution,
		)
		if err != nil {
			return nil, 0, err
		}
		members = append(members, member)
	}

	return members, total, nil
}

// GetByPlayer récupère le membre par joueur
func (r *guildMemberRepository) GetByPlayer(ctx context.Context, playerID uuid.UUID) (*models.GuildMember, error) {
	query := `
		SELECT id, guild_id, player_id, role, joined_at, last_seen, contribution
		FROM guild_members WHERE player_id = $1
	`

	member := &models.GuildMember{}
	err := r.db.QueryRowContext(ctx, query, playerID).Scan(
		&member.ID, &member.GuildID, &member.PlayerID, &member.Role,
		&member.JoinedAt, &member.LastSeen, &member.Contribution,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrMemberNotFound
		}
		return nil, err
	}

	return member, nil
}

// Update met à jour un membre
func (r *guildMemberRepository) Update(ctx context.Context, member *models.GuildMember) error {
	query := `
		UPDATE guild_members 
		SET role = $3, last_seen = $4, contribution = $5
		WHERE id = $1 AND guild_id = $2
	`

	member.LastSeen = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		member.ID, member.GuildID, member.Role,
		member.LastSeen, member.Contribution,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return models.ErrMemberNotFound
	}

	return nil
}

// Delete supprime un membre
func (r *guildMemberRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM guild_members WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return models.ErrMemberNotFound
	}

	return nil
}

// GetMemberCount compte les membres d'une guilde
func (r *guildMemberRepository) GetMemberCount(ctx context.Context, guildID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM guild_members WHERE guild_id = $1`

	var count int
	err := r.db.QueryRowContext(ctx, query, guildID).Scan(&count)
	return count, err
}

// UpdateLastSeen met à jour la dernière connection
func (r *guildMemberRepository) UpdateLastSeen(ctx context.Context, memberID uuid.UUID) error {
	query := `UPDATE guild_members SET last_seen = $2 WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, memberID, time.Now())
	return err
}

// UpdateContribution met à jour la contribution
func (r *guildMemberRepository) UpdateContribution(ctx context.Context, memberID uuid.UUID, contribution int64) error {
	query := `UPDATE guild_members SET contribution = $2 WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, memberID, contribution)
	return err
}
