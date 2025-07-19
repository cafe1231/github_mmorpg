package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"world/internal/database"
	"world/internal/models"
)

// PlayerPositionRepositoryInterface définit les méthodes du repository player position
type PlayerPositionRepositoryInterface interface {
	Upsert(position *models.PlayerPosition) error
	GetByCharacterID(characterID uuid.UUID) (*models.PlayerPosition, error)
	GetByUserID(userID uuid.UUID) ([]*models.PlayerPosition, error)
	GetByZoneID(zoneID string) ([]*models.PlayerPosition, error)
	GetNearbyPlayers(zoneID string, x, y, z, radius float64) ([]*models.PlayerPosition, error)
	SetOffline(characterID uuid.UUID) error
	SetOnline(characterID uuid.UUID) error
	CleanupOfflinePlayers(timeout time.Duration) error
	UpdateZone(characterID uuid.UUID, newZoneID string, x, y, z float64) error

	// Statistiques
	GetOnlinePlayerCount() (int, error)
	GetZonePlayerCounts() (map[string]int, error)
}

// PlayerPositionRepository implémente l'interface PlayerPositionRepositoryInterface
type PlayerPositionRepository struct {
	db *database.DB
}

// NewPlayerPositionRepository crée une nouvelle instance du repository player position
func NewPlayerPositionRepository(db *database.DB) PlayerPositionRepositoryInterface {
	return &PlayerPositionRepository{db: db}
}

// Upsert insère ou met à jour une position de joueur
func (r *PlayerPositionRepository) Upsert(position *models.PlayerPosition) error {
	query := `
		INSERT INTO player_positions (
			character_id, user_id, zone_id, x, y, z, rotation,
			velocity_x, velocity_y, velocity_z, is_moving, is_online
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
		ON CONFLICT (character_id) 
		DO UPDATE SET 
			user_id = EXCLUDED.user_id,
			zone_id = EXCLUDED.zone_id,
			x = EXCLUDED.x,
			y = EXCLUDED.y,
			z = EXCLUDED.z,
			rotation = EXCLUDED.rotation,
			velocity_x = EXCLUDED.velocity_x,
			velocity_y = EXCLUDED.velocity_y,
			velocity_z = EXCLUDED.velocity_z,
			is_moving = EXCLUDED.is_moving,
			is_online = EXCLUDED.is_online,
			last_update = CURRENT_TIMESTAMP`

	_, err := r.db.Exec(query,
		position.CharacterID, position.UserID, position.ZoneID,
		position.X, position.Y, position.Z, position.Rotation,
		position.VelocityX, position.VelocityY, position.VelocityZ,
		position.IsMoving, position.IsOnline,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert player position: %w", err)
	}

	return nil
}

// GetByCharacterID récupère la position d'un personnage
func (r *PlayerPositionRepository) GetByCharacterID(characterID uuid.UUID) (*models.PlayerPosition, error) {
	var position models.PlayerPosition

	query := `
		SELECT character_id, user_id, zone_id, x, y, z, rotation,
		       velocity_x, velocity_y, velocity_z, is_moving,
		       is_online, last_update
		FROM player_positions 
		WHERE character_id = $1`

	row := r.db.QueryRow(query, characterID)
	err := row.Scan(
		&position.CharacterID, &position.UserID, &position.ZoneID,
		&position.X, &position.Y, &position.Z, &position.Rotation,
		&position.VelocityX, &position.VelocityY, &position.VelocityZ, &position.IsMoving,
		&position.IsOnline, &position.LastUpdate,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("player position not found")
		}
		return nil, fmt.Errorf("failed to get player position: %w", err)
	}

	return &position, nil
}

// GetByUserID récupère toutes les positions des personnages d'un utilisateur
func (r *PlayerPositionRepository) GetByUserID(userID uuid.UUID) ([]*models.PlayerPosition, error) {
	var positions []*models.PlayerPosition

	query := `
		SELECT character_id, user_id, zone_id, x, y, z, rotation,
		       velocity_x, velocity_y, velocity_z, is_moving,
		       is_online, last_update
		FROM player_positions 
		WHERE user_id = $1
		ORDER BY last_update DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user positions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var position models.PlayerPosition

		err := rows.Scan(
			&position.CharacterID, &position.UserID, &position.ZoneID,
			&position.X, &position.Y, &position.Z, &position.Rotation,
			&position.VelocityX, &position.VelocityY, &position.VelocityZ, &position.IsMoving,
			&position.IsOnline, &position.LastUpdate,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan player position: %w", err)
		}

		positions = append(positions, &position)
	}

	return positions, nil
}

// GetByZoneID récupère toutes les positions dans une zone
func (r *PlayerPositionRepository) GetByZoneID(zoneID string) ([]*models.PlayerPosition, error) {
	var positions []*models.PlayerPosition

	query := `
		SELECT character_id, user_id, zone_id, x, y, z, rotation,
		       velocity_x, velocity_y, velocity_z, is_moving,
		       is_online, last_update
		FROM player_positions 
		WHERE zone_id = $1 AND is_online = true
		ORDER BY last_update DESC`

	rows, err := r.db.Query(query, zoneID)
	if err != nil {
		return nil, fmt.Errorf("failed to get zone positions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var position models.PlayerPosition

		err := rows.Scan(
			&position.CharacterID, &position.UserID, &position.ZoneID,
			&position.X, &position.Y, &position.Z, &position.Rotation,
			&position.VelocityX, &position.VelocityY, &position.VelocityZ, &position.IsMoving,
			&position.IsOnline, &position.LastUpdate,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan player position: %w", err)
		}

		positions = append(positions, &position)
	}

	return positions, nil
}

// GetNearbyPlayers récupère les joueurs proches d'une position
func (r *PlayerPositionRepository) GetNearbyPlayers(zoneID string, x, y, z, radius float64) ([]*models.PlayerPosition, error) {
	var positions []*models.PlayerPosition

	// Requête utilisant la distance euclidienne en 3D
	query := `
		SELECT character_id, user_id, zone_id, x, y, z, rotation,
		       velocity_x, velocity_y, velocity_z, is_moving,
		       is_online, last_update,
		       SQRT(POWER(x - $2, 2) + POWER(y - $3, 2) + POWER(z - $4, 2)) as distance
		FROM player_positions 
		WHERE zone_id = $1 
		  AND is_online = true
		  AND SQRT(POWER(x - $2, 2) + POWER(y - $3, 2) + POWER(z - $4, 2)) <= $5
		ORDER BY distance ASC`

	rows, err := r.db.Query(query, zoneID, x, y, z, radius)
	if err != nil {
		return nil, fmt.Errorf("failed to get nearby players: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var position models.PlayerPosition
		var distance float64

		err := rows.Scan(
			&position.CharacterID, &position.UserID, &position.ZoneID,
			&position.X, &position.Y, &position.Z, &position.Rotation,
			&position.VelocityX, &position.VelocityY, &position.VelocityZ, &position.IsMoving,
			&position.IsOnline, &position.LastUpdate, &distance,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan nearby player: %w", err)
		}

		positions = append(positions, &position)
	}

	return positions, nil
}

// SetOffline marque un joueur comme hors ligne
func (r *PlayerPositionRepository) SetOffline(characterID uuid.UUID) error {
	query := `
		UPDATE player_positions 
		SET is_online = false, last_update = CURRENT_TIMESTAMP
		WHERE character_id = $1`

	result, err := r.db.Exec(query, characterID)
	if err != nil {
		return fmt.Errorf("failed to set player offline: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("player position not found")
	}

	return nil
}

// SetOnline marque un joueur comme en ligne
func (r *PlayerPositionRepository) SetOnline(characterID uuid.UUID) error {
	query := `
		UPDATE player_positions 
		SET is_online = true, last_update = CURRENT_TIMESTAMP
		WHERE character_id = $1`

	result, err := r.db.Exec(query, characterID)
	if err != nil {
		return fmt.Errorf("failed to set player online: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("player position not found")
	}

	return nil
}

// CleanupOfflinePlayers nettoie les joueurs hors ligne depuis trop longtemps
func (r *PlayerPositionRepository) CleanupOfflinePlayers(timeout time.Duration) error {
	query := `
		UPDATE player_positions 
		SET is_online = false
		WHERE is_online = true 
		  AND last_update < $1`

	cutoffTime := time.Now().Add(-timeout)

	result, err := r.db.Exec(query, cutoffTime)
	if err != nil {
		return fmt.Errorf("failed to cleanup offline players: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected > 0 {
		fmt.Printf("Marked %d players as offline due to timeout\n", rowsAffected)
	}

	return nil
}

// UpdateZone met à jour la zone d'un joueur avec nouvelle position
func (r *PlayerPositionRepository) UpdateZone(characterID uuid.UUID, newZoneID string, x, y, z float64) error {
	query := `
		UPDATE player_positions 
		SET zone_id = $2, x = $3, y = $4, z = $5, 
		    velocity_x = 0, velocity_y = 0, velocity_z = 0, 
		    is_moving = false, last_update = CURRENT_TIMESTAMP
		WHERE character_id = $1`

	result, err := r.db.Exec(query, characterID, newZoneID, x, y, z)
	if err != nil {
		return fmt.Errorf("failed to update player zone: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("player position not found")
	}

	return nil
}

// GetOnlinePlayerCount récupère le nombre total de joueurs en ligne
func (r *PlayerPositionRepository) GetOnlinePlayerCount() (int, error) {
	var count int

	query := `SELECT COUNT(*) FROM player_positions WHERE is_online = true`

	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get online player count: %w", err)
	}

	return count, nil
}

// GetZonePlayerCounts récupère le nombre de joueurs par zone
func (r *PlayerPositionRepository) GetZonePlayerCounts() (map[string]int, error) {
	counts := make(map[string]int)

	query := `
		SELECT zone_id, COUNT(*) as player_count
		FROM player_positions 
		WHERE is_online = true
		GROUP BY zone_id`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get zone player counts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var zoneID string
		var count int

		err := rows.Scan(&zoneID, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan zone player count: %w", err)
		}

		counts[zoneID] = count
	}

	return counts, nil
}
