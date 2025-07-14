package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"player/internal/database"
	"player/internal/models"
)

// PlayerRepositoryInterface définit les méthodes du repository player
type PlayerRepositoryInterface interface {
	Create(player *models.Player) error
	GetByID(id uuid.UUID) (*models.Player, error)
	GetByUserID(userID uuid.UUID) (*models.Player, error)
	GetByDisplayName(displayName string) (*models.Player, error)
	Update(player *models.Player) error
	Delete(id uuid.UUID) error
	List(limit, offset int) ([]*models.Player, error)
	GetStats(playerID uuid.UUID) (*models.PlayerStats, error)
	UpdateStats(stats *models.PlayerStats) error
	GetCharacterCount(playerID uuid.UUID) (int, error)
}

// PlayerRepository implémente l'interface PlayerRepositoryInterface
type PlayerRepository struct {
	db *database.DB
}

// NewPlayerRepository crée une nouvelle instance du repository player
func NewPlayerRepository(db *database.DB) PlayerRepositoryInterface {
	return &PlayerRepository{db: db}
}

// Create crée un nouveau joueur
func (r *PlayerRepository) Create(player *models.Player) error {
	// Sérialiser les préférences en JSON
	preferencesJSON, err := json.Marshal(player.Preferences)
	if err != nil {
		return fmt.Errorf("failed to marshal preferences: %w", err)
	}

	query := `
		INSERT INTO players (
			id, user_id, display_name, avatar, title, guild_id,
			total_play_time, last_seen, preferences, created_at, updated_at
		) VALUES (
			:id, :user_id, :display_name, :avatar, :title, :guild_id,
			:total_play_time, :last_seen, :preferences, :created_at, :updated_at
		)`

	// Préparer les données avec JSON sérialisé
	data := map[string]interface{}{
		"id":              player.ID,
		"user_id":         player.UserID,
		"display_name":    player.DisplayName,
		"avatar":          player.Avatar,
		"title":           player.Title,
		"guild_id":        player.GuildID,
		"total_play_time": player.TotalPlayTime,
		"last_seen":       player.LastSeen,
		"preferences":     string(preferencesJSON),
		"created_at":      player.CreatedAt,
		"updated_at":      player.UpdatedAt,
	}

	_, err = r.db.NamedExec(query, data)
	if err != nil {
		return fmt.Errorf("failed to create player: %w", err)
	}

	return nil
}

// GetByID récupère un joueur par son ID
func (r *PlayerRepository) GetByID(id uuid.UUID) (*models.Player, error) {
	var player models.Player
	var preferencesJSON string

	query := `
		SELECT id, user_id, display_name, avatar, title, guild_id,
		       total_play_time, last_seen, preferences,
		       created_at, updated_at
		FROM players 
		WHERE id = $1`

	row := r.db.QueryRow(query, id)
	err := row.Scan(
		&player.ID, &player.UserID, &player.DisplayName,
		&player.Avatar, &player.Title, &player.GuildID,
		&player.TotalPlayTime, &player.LastSeen, &preferencesJSON,
		&player.CreatedAt, &player.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("player not found")
		}
		return nil, fmt.Errorf("failed to get player by ID: %w", err)
	}

	// Désérialiser les préférences
	if err := json.Unmarshal([]byte(preferencesJSON), &player.Preferences); err != nil {
		player.Preferences = models.GetDefaultPreferences()
	}

	return &player, nil
}

// GetByUserID récupère un joueur par son ID utilisateur (service Auth)
func (r *PlayerRepository) GetByUserID(userID uuid.UUID) (*models.Player, error) {
	var player models.Player
	var preferencesJSON string

	query := `
		SELECT id, user_id, display_name, avatar, title, guild_id,
		       total_play_time, last_seen, preferences,
		       created_at, updated_at
		FROM players 
		WHERE user_id = $1`

	row := r.db.QueryRow(query, userID)
	err := row.Scan(
		&player.ID, &player.UserID, &player.DisplayName,
		&player.Avatar, &player.Title, &player.GuildID,
		&player.TotalPlayTime, &player.LastSeen, &preferencesJSON,
		&player.CreatedAt, &player.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("player not found")
		}
		return nil, fmt.Errorf("failed to get player by user ID: %w", err)
	}

	// Désérialiser les préférences
	if err := json.Unmarshal([]byte(preferencesJSON), &player.Preferences); err != nil {
		player.Preferences = models.GetDefaultPreferences()
	}

	return &player, nil
}

// GetByDisplayName récupère un joueur par son nom d'affichage
func (r *PlayerRepository) GetByDisplayName(displayName string) (*models.Player, error) {
	var player models.Player
	var preferencesJSON string

	query := `
		SELECT id, user_id, display_name, avatar, title, guild_id,
		       total_play_time, last_seen, preferences,
		       created_at, updated_at
		FROM players 
		WHERE display_name = $1`

	row := r.db.QueryRow(query, displayName)
	err := row.Scan(
		&player.ID, &player.UserID, &player.DisplayName,
		&player.Avatar, &player.Title, &player.GuildID,
		&player.TotalPlayTime, &player.LastSeen, &preferencesJSON,
		&player.CreatedAt, &player.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("player not found")
		}
		return nil, fmt.Errorf("failed to get player by display name: %w", err)
	}

	// Désérialiser les préférences
	if err := json.Unmarshal([]byte(preferencesJSON), &player.Preferences); err != nil {
		player.Preferences = models.GetDefaultPreferences()
	}

	return &player, nil
}

// Update met à jour un joueur
func (r *PlayerRepository) Update(player *models.Player) error {
	// Sérialiser les préférences en JSON
	preferencesJSON, err := json.Marshal(player.Preferences)
	if err != nil {
		return fmt.Errorf("failed to marshal preferences: %w", err)
	}

	query := `
		UPDATE players SET
			display_name = :display_name,
			avatar = :avatar,
			title = :title,
			guild_id = :guild_id,
			total_play_time = :total_play_time,
			last_seen = :last_seen,
			preferences = :preferences,
			updated_at = :updated_at
		WHERE id = :id`

	data := map[string]interface{}{
		"id":              player.ID,
		"display_name":    player.DisplayName,
		"avatar":          player.Avatar,
		"title":           player.Title,
		"guild_id":        player.GuildID,
		"total_play_time": player.TotalPlayTime,
		"last_seen":       player.LastSeen,
		"preferences":     string(preferencesJSON),
		"updated_at":      player.UpdatedAt,
	}

	_, err = r.db.NamedExec(query, data)
	if err != nil {
		return fmt.Errorf("failed to update player: %w", err)
	}

	return nil
}

// Delete supprime un joueur
func (r *PlayerRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM players WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete player: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("player not found")
	}

	return nil
}

// List récupère une liste de joueurs avec pagination
func (r *PlayerRepository) List(limit, offset int) ([]*models.Player, error) {
	var players []*models.Player

	query := `
		SELECT id, user_id, display_name, avatar, title, guild_id,
		       total_play_time, last_seen, preferences,
		       created_at, updated_at
		FROM players 
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list players: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var player models.Player
		var preferencesJSON string

		err := rows.Scan(
			&player.ID, &player.UserID, &player.DisplayName,
			&player.Avatar, &player.Title, &player.GuildID,
			&player.TotalPlayTime, &player.LastSeen, &preferencesJSON,
			&player.CreatedAt, &player.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan player: %w", err)
		}

		// Désérialiser les préférences
		if err := json.Unmarshal([]byte(preferencesJSON), &player.Preferences); err != nil {
			player.Preferences = models.GetDefaultPreferences()
		}

		players = append(players, &player)
	}

	return players, nil
}

// GetStats récupère les statistiques d'un joueur
func (r *PlayerRepository) GetStats(playerID uuid.UUID) (*models.PlayerStats, error) {
	var stats models.PlayerStats

	query := `
		SELECT 
			p.id as player_id,
			COALESCE(COUNT(c.id), 0) as total_characters,
			COALESCE(MAX(c.level), 0) as highest_level,
			COALESCE(SUM(c.experience), 0) as total_experience,
			0 as achievement_count, -- TODO: implémenter les achievements
			COALESCE(SUM(cs.monsters_killed + cs.bosses_killed), 0) as total_kills,
			COALESCE(SUM(cs.deaths + cs.pvp_deaths), 0) as total_deaths,
			COALESCE(SUM(cs.quests_completed), 0) as total_quests,
			COALESCE(SUM(cs.gold_earned), 0) as total_gold,
			CURRENT_TIMESTAMP as updated_at
		FROM players p
		LEFT JOIN characters c ON c.player_id = p.id AND c.status = 'active'
		LEFT JOIN combat_stats cs ON cs.character_id = c.id
		WHERE p.id = $1
		GROUP BY p.id`

	err := r.db.Get(&stats, query, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get player stats: %w", err)
	}

	return &stats, nil
}

// UpdateStats met à jour les statistiques d'un joueur
func (r *PlayerRepository) UpdateStats(stats *models.PlayerStats) error {
	// Note: Les stats de joueur sont généralement calculées automatiquement
	// Cette méthode pourrait être utilisée pour des ajustements manuels
	query := `
		UPDATE players SET
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	_, err := r.db.Exec(query, stats.PlayerID)
	if err != nil {
		return fmt.Errorf("failed to update player stats: %w", err)
	}

	return nil
}

// GetCharacterCount récupère le nombre de personnages d'un joueur
func (r *PlayerRepository) GetCharacterCount(playerID uuid.UUID) (int, error) {
	var count int

	query := `
		SELECT COUNT(*) 
		FROM characters 
		WHERE player_id = $1 AND status != 'deleted'`

	err := r.db.Get(&count, query, playerID)
	if err != nil {
		return 0, fmt.Errorf("failed to get character count: %w", err)
	}

	return count, nil
}