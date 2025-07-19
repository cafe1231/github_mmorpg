package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"chat/internal/models"
)

type channelRepository struct {
	db *sql.DB
}

// NewChannelRepository crée une nouvelle instance du repository des channels
func NewChannelRepository(db *sql.DB) ChannelRepository {
	return &channelRepository{db: db}
}

// Create crée un nouveau channel
func (r *channelRepository) Create(ctx context.Context, channel *models.Channel) error {
	if channel.ID == uuid.Nil {
		channel.ID = uuid.New()
	}

	settingsJSON, err := json.Marshal(channel.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	query := `
		INSERT INTO channels (
			id, name, type, description, owner_id, is_moderated, is_private, 
			max_members, zone_id, guild_id, party_id, settings, is_active, 
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
	`

	now := time.Now()
	channel.CreatedAt = now
	channel.UpdatedAt = now

	_, err = r.db.ExecContext(ctx, query,
		channel.ID, channel.Name, channel.Type, channel.Description,
		channel.OwnerID, channel.IsModerated, channel.IsPrivate,
		channel.MaxMembers, channel.ZoneID, channel.GuildID, channel.PartyID,
		settingsJSON, channel.IsActive, channel.CreatedAt, channel.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create channel: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"channel_id":   channel.ID,
		"channel_name": channel.Name,
		"type":         channel.Type,
		"owner_id":     channel.OwnerID,
	}).Info("Channel created successfully")

	return nil
}

// GetByID récupère un channel par son ID
func (r *channelRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Channel, error) {
	query := `
		SELECT 
			id, name, type, description, owner_id, is_moderated, is_private,
			max_members, zone_id, guild_id, party_id, settings, is_active,
			created_at, updated_at,
			(SELECT COUNT(*) FROM channel_members WHERE channel_id = c.id AND is_active = true) as member_count
		FROM channels c
		WHERE id = $1 AND is_active = true
	`

	var channel models.Channel
	var settingsJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&channel.ID, &channel.Name, &channel.Type, &channel.Description,
		&channel.OwnerID, &channel.IsModerated, &channel.IsPrivate,
		&channel.MaxMembers, &channel.ZoneID, &channel.GuildID, &channel.PartyID,
		&settingsJSON, &channel.IsActive, &channel.CreatedAt, &channel.UpdatedAt,
		&channel.MemberCount,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}

	if err := json.Unmarshal(settingsJSON, &channel.Settings); err != nil {
		logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")
		return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
	}

	return &channel, nil
}

// Update met à jour un channel
func (r *channelRepository) Update(ctx context.Context, channel *models.Channel) error {
	settingsJSON, err := json.Marshal(channel.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	query := `
		UPDATE channels SET
			name = $2, description = $3, is_moderated = $4, is_private = $5,
			max_members = $6, settings = $7, updated_at = $8
		WHERE id = $1 AND is_active = true
	`

	channel.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		channel.ID, channel.Name, channel.Description, channel.IsModerated,
		channel.IsPrivate, channel.MaxMembers, settingsJSON, channel.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update channel: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("channel not found or inactive")
	}

	return nil
}

// Delete supprime un channel (soft delete)
func (r *channelRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE channels SET is_active = false, updated_at = $2 WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete channel: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("channel not found")
	}

	return nil
}

// GetByType récupère les channels par type
func (r *channelRepository) GetByType(ctx context.Context, channelType models.ChannelType, limit, offset int) ([]*models.Channel, error) {
	query := `
		SELECT 
			id, name, type, description, owner_id, is_moderated, is_private,
			max_members, zone_id, guild_id, party_id, settings, is_active,
			created_at, updated_at,
			(SELECT COUNT(*) FROM channel_members WHERE channel_id = c.id AND is_active = true) as member_count
		FROM channels c
		WHERE type = $1 AND is_active = true
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, channelType, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get channels by type: %w", err)
	}
	defer rows.Close()

	return r.scanChannels(rows)
}

// GetByZone récupère les channels d'une zone
func (r *channelRepository) GetByZone(ctx context.Context, zoneID string, limit, offset int) ([]*models.Channel, error) {
	query := `
		SELECT 
			id, name, type, description, owner_id, is_moderated, is_private,
			max_members, zone_id, guild_id, party_id, settings, is_active,
			created_at, updated_at,
			(SELECT COUNT(*) FROM channel_members WHERE channel_id = c.id AND is_active = true) as member_count
		FROM channels c
		WHERE zone_id = $1 AND is_active = true
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, zoneID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get channels by zone: %w", err)
	}
	defer rows.Close()

	return r.scanChannels(rows)
}

// GetByGuild récupère les channels d'une guild
func (r *channelRepository) GetByGuild(ctx context.Context, guildID uuid.UUID, limit, offset int) ([]*models.Channel, error) {
	query := `
		SELECT 
			id, name, type, description, owner_id, is_moderated, is_private,
			max_members, zone_id, guild_id, party_id, settings, is_active,
			created_at, updated_at,
			(SELECT COUNT(*) FROM channel_members WHERE channel_id = c.id AND is_active = true) as member_count
		FROM channels c
		WHERE guild_id = $1 AND is_active = true
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, guildID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get channels by guild: %w", err)
	}
	defer rows.Close()

	return r.scanChannels(rows)
}

// GetUserChannels récupère les channels d'un utilisateur
func (r *channelRepository) GetUserChannels(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Channel, error) {
	query := `
		SELECT 
			c.id, c.name, c.type, c.description, c.owner_id, c.is_moderated, c.is_private,
			c.max_members, c.zone_id, c.guild_id, c.party_id, c.settings, c.is_active,
			c.created_at, c.updated_at,
			(SELECT COUNT(*) FROM channel_members WHERE channel_id = c.id AND is_active = true) as member_count
		FROM channels c
		INNER JOIN channel_members cm ON c.id = cm.channel_id
		WHERE cm.user_id = $1 AND c.is_active = true AND cm.is_active = true
		ORDER BY cm.last_seen_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user channels: %w", err)
	}
	defer rows.Close()

	return r.scanChannels(rows)
}

// Search recherche des channels par nom
func (r *channelRepository) Search(ctx context.Context, query string, limit, offset int) ([]*models.Channel, error) {
	searchQuery := `
		SELECT 
			id, name, type, description, owner_id, is_moderated, is_private,
			max_members, zone_id, guild_id, party_id, settings, is_active,
			created_at, updated_at,
			(SELECT COUNT(*) FROM channel_members WHERE channel_id = c.id AND is_active = true) as member_count
		FROM channels c
		WHERE (LOWER(name) LIKE LOWER($1) OR LOWER(description) LIKE LOWER($1))
		AND is_active = true AND is_private = false
		ORDER BY 
			CASE WHEN LOWER(name) = LOWER($2) THEN 1 ELSE 2 END,
			member_count DESC
		LIMIT $3 OFFSET $4
	`

	searchTerm := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, searchQuery, searchTerm, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search channels: %w", err)
	}
	defer rows.Close()

	return r.scanChannels(rows)
}

// AddMember ajoute un membre à un channel
func (r *channelRepository) AddMember(ctx context.Context, member *models.ChannelMember) error {
	if member.ID == uuid.Nil {
		member.ID = uuid.New()
	}

	query := `
		INSERT INTO channel_members (
			id, channel_id, user_id, display_name, role, can_moderate, can_invite,
			is_online, is_muted, muted_until, joined_at, last_seen_at, is_active
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
		ON CONFLICT (channel_id, user_id) 
		DO UPDATE SET is_active = true, joined_at = $11
	`

	now := time.Now()
	if member.JoinedAt.IsZero() {
		member.JoinedAt = now
	}
	if member.LastSeenAt.IsZero() {
		member.LastSeenAt = now
	}

	_, err := r.db.ExecContext(ctx, query,
		member.ID, member.ChannelID, member.UserID, member.DisplayName,
		member.Role, member.CanModerate, member.CanInvite, member.IsOnline,
		member.IsMuted, member.MutedUntil, member.JoinedAt, member.LastSeenAt, true,
	)

	if err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}

	return nil
}

// RemoveMember supprime un membre d'un channel
func (r *channelRepository) RemoveMember(ctx context.Context, channelID, userID uuid.UUID) error {
	query := `UPDATE channel_members SET is_active = false WHERE channel_id = $1 AND user_id = $2`

	result, err := r.db.ExecContext(ctx, query, channelID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("member not found")
	}

	return nil
}

// GetMembers récupère les membres d'un channel
func (r *channelRepository) GetMembers(ctx context.Context, channelID uuid.UUID, limit, offset int) ([]*models.ChannelMember, error) {
	query := `
		SELECT 
			id, channel_id, user_id, display_name, role, can_moderate, can_invite,
			is_online, is_muted, muted_until, joined_at, last_seen_at
		FROM channel_members
		WHERE channel_id = $1 AND is_active = true
		ORDER BY 
			CASE role 
				WHEN 'admin' THEN 1 
				WHEN 'moderator' THEN 2 
				ELSE 3 
			END,
			joined_at ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, channelID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get members: %w", err)
	}
	defer rows.Close()

	var members []*models.ChannelMember
	for rows.Next() {
		var member models.ChannelMember
		err := rows.Scan(
			&member.ID, &member.ChannelID, &member.UserID, &member.DisplayName,
			&member.Role, &member.CanModerate, &member.CanInvite, &member.IsOnline,
			&member.IsMuted, &member.MutedUntil, &member.JoinedAt, &member.LastSeenAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan member: %w", err)
		}
		members = append(members, &member)
	}

	return members, nil
}

// GetMember récupère un membre spécifique
func (r *channelRepository) GetMember(ctx context.Context, channelID, userID uuid.UUID) (*models.ChannelMember, error) {
	query := `
		SELECT 
			id, channel_id, user_id, display_name, role, can_moderate, can_invite,
			is_online, is_muted, muted_until, joined_at, last_seen_at
		FROM channel_members
		WHERE channel_id = $1 AND user_id = $2 AND is_active = true
	`

	var member models.ChannelMember
	err := r.db.QueryRowContext(ctx, query, channelID, userID).Scan(
		&member.ID, &member.ChannelID, &member.UserID, &member.DisplayName,
		&member.Role, &member.CanModerate, &member.CanInvite, &member.IsOnline,
		&member.IsMuted, &member.MutedUntil, &member.JoinedAt, &member.LastSeenAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get member: %w", err)
	}

	return &member, nil
}

// UpdateMember met à jour un membre
func (r *channelRepository) UpdateMember(ctx context.Context, member *models.ChannelMember) error {
	query := `
		UPDATE channel_members SET
			display_name = $3, role = $4, can_moderate = $5, can_invite = $6,
			is_online = $7, is_muted = $8, muted_until = $9, last_seen_at = $10
		WHERE channel_id = $1 AND user_id = $2 AND is_active = true
	`

	member.LastSeenAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		member.ChannelID, member.UserID, member.DisplayName, member.Role,
		member.CanModerate, member.CanInvite, member.IsOnline, member.IsMuted,
		member.MutedUntil, member.LastSeenAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update member: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("member not found")
	}

	return nil
}

// IsMember vérifie si un utilisateur est membre d'un channel
func (r *channelRepository) IsMember(ctx context.Context, channelID, userID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM channel_members WHERE channel_id = $1 AND user_id = $2 AND is_active = true)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, channelID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check membership: %w", err)
	}

	return exists, nil
}

// GetMemberCount retourne le nombre de membres d'un channel
func (r *channelRepository) GetMemberCount(ctx context.Context, channelID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM channel_members WHERE channel_id = $1 AND is_active = true`

	var count int
	err := r.db.QueryRowContext(ctx, query, channelID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get member count: %w", err)
	}

	return count, nil
}

// GetChannelStats retourne les statistiques d'un channel
func (r *channelRepository) GetChannelStats(ctx context.Context, channelID uuid.UUID) (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(DISTINCT cm.user_id) as member_count,
			COUNT(DISTINCT CASE WHEN cm.is_online THEN cm.user_id END) as online_count,
			COUNT(DISTINCT m.id) as message_count,
			MAX(m.created_at) as last_message_at
		FROM channel_members cm
		LEFT JOIN messages m ON m.channel_id = cm.channel_id AND m.is_deleted = false
		WHERE cm.channel_id = $1 AND cm.is_active = true
	`

	var memberCount, onlineCount, messageCount int
	var lastMessageAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, channelID).Scan(
		&memberCount, &onlineCount, &messageCount, &lastMessageAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get channel stats: %w", err)
	}

	stats := map[string]interface{}{
		"member_count":  memberCount,
		"online_count":  onlineCount,
		"message_count": messageCount,
	}

	if lastMessageAt.Valid {
		stats["last_message_at"] = lastMessageAt.Time
	}

	return stats, nil
}

// scanChannels aide à scanner les résultats des requêtes de channels
func (r *channelRepository) scanChannels(rows *sql.Rows) ([]*models.Channel, error) {
	var channels []*models.Channel

	for rows.Next() {
		var channel models.Channel
		var settingsJSON []byte

		err := rows.Scan(
			&channel.ID, &channel.Name, &channel.Type, &channel.Description,
			&channel.OwnerID, &channel.IsModerated, &channel.IsPrivate,
			&channel.MaxMembers, &channel.ZoneID, &channel.GuildID, &channel.PartyID,
			&settingsJSON, &channel.IsActive, &channel.CreatedAt, &channel.UpdatedAt,
			&channel.MemberCount,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan channel: %w", err)
		}

		if err := json.Unmarshal(settingsJSON, &channel.Settings); err != nil {
			logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")
			return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
		}

		channels = append(channels, &channel)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return channels, nil
}
