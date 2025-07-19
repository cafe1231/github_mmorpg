package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"world/internal/database"
	"world/internal/models"
)

// ZoneRepositoryInterface définit les méthodes du repository zone
type ZoneRepositoryInterface interface {
	Create(zone *models.Zone) error
	GetByID(id string) (*models.Zone, error)
	GetAll() ([]*models.Zone, error)
	GetByType(zoneType string) ([]*models.Zone, error)
	GetByLevel(minLevel, maxLevel int) ([]*models.Zone, error)
	Update(zone *models.Zone) error
	Delete(id string) error

	// Gestion des joueurs dans les zones
	GetPlayerCount(zoneID string) (int, error)
	GetPlayersInZone(zoneID string) ([]*models.PlayerPosition, error)

	// Transitions
	GetTransitions(fromZoneID string) ([]*models.ZoneTransition, error)
	CreateTransition(transition *models.ZoneTransition) error
	UpdateTransition(transition *models.ZoneTransition) error
	DeleteTransition(id uuid.UUID) error
}

// ZoneRepository implémente l'interface ZoneRepositoryInterface
type ZoneRepository struct {
	db *database.DB
}

// NewZoneRepository crée une nouvelle instance du repository zone
func NewZoneRepository(db *database.DB) ZoneRepositoryInterface {
	return &ZoneRepository{db: db}
}

// Create crée une nouvelle zone
func (r *ZoneRepository) Create(zone *models.Zone) error {
	// Sérialiser les settings en JSON
	settingsJSON, err := json.Marshal(zone.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	query := `
		INSERT INTO zones (
			id, name, display_name, description, type, level,
			min_x, min_y, min_z, max_x, max_y, max_z,
			spawn_x, spawn_y, spawn_z, max_players,
			is_pvp, is_safe_zone, settings, status
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11, $12,
			$13, $14, $15, $16,
			$17, $18, $19, $20
		)`

	_, err = r.db.Exec(query,
		zone.ID, zone.Name, zone.DisplayName, zone.Description, zone.Type, zone.Level,
		zone.MinX, zone.MinY, zone.MinZ, zone.MaxX, zone.MaxY, zone.MaxZ,
		zone.SpawnX, zone.SpawnY, zone.SpawnZ, zone.MaxPlayers,
		zone.IsPvP, zone.IsSafeZone, settingsJSON, zone.Status,
	)
	if err != nil {
		return fmt.Errorf("failed to create zone: %w", err)
	}

	return nil
}

// GetByID récupère une zone par son ID
func (r *ZoneRepository) GetByID(id string) (*models.Zone, error) {
	var zone models.Zone
	var settingsJSON string

	query := `
		SELECT id, name, display_name, description, type, level,
		       min_x, min_y, min_z, max_x, max_y, max_z,
		       spawn_x, spawn_y, spawn_z, max_players,
		       is_pvp, is_safe_zone, settings, status,
		       created_at, updated_at
		FROM zones 
		WHERE id = $1`

	row := r.db.QueryRow(query, id)
	err := row.Scan(
		&zone.ID, &zone.Name, &zone.DisplayName, &zone.Description, &zone.Type, &zone.Level,
		&zone.MinX, &zone.MinY, &zone.MinZ, &zone.MaxX, &zone.MaxY, &zone.MaxZ,
		&zone.SpawnX, &zone.SpawnY, &zone.SpawnZ, &zone.MaxPlayers,
		&zone.IsPvP, &zone.IsSafeZone, &settingsJSON, &zone.Status,
		&zone.CreatedAt, &zone.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("zone not found")
		}
		return nil, fmt.Errorf("failed to get zone by ID: %w", err)
	}

	// Désérialiser les settings
	if err := json.Unmarshal([]byte(settingsJSON), &zone.Settings); err != nil {
		logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")

		logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")
	}

	// Récupérer le nombre de joueurs
	zone.PlayerCount, _ = r.GetPlayerCount(zone.ID)

	return &zone, nil
}

// GetAll récupère toutes les zones
func (r *ZoneRepository) GetAll() ([]*models.Zone, error) {
	var zones []*models.Zone

	query := `
		SELECT id, name, display_name, description, type, level,
		       min_x, min_y, min_z, max_x, max_y, max_z,
		       spawn_x, spawn_y, spawn_z, max_players,
		       is_pvp, is_safe_zone, settings, status,
		       created_at, updated_at
		FROM zones 
		WHERE status != 'disabled'
		ORDER BY level ASC, name ASC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all zones: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var zone models.Zone
		var settingsJSON string

		err := rows.Scan(
			&zone.ID, &zone.Name, &zone.DisplayName, &zone.Description, &zone.Type, &zone.Level,
			&zone.MinX, &zone.MinY, &zone.MinZ, &zone.MaxX, &zone.MaxY, &zone.MaxZ,
			&zone.SpawnX, &zone.SpawnY, &zone.SpawnZ, &zone.MaxPlayers,
			&zone.IsPvP, &zone.IsSafeZone, &settingsJSON, &zone.Status,
			&zone.CreatedAt, &zone.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan zone: %w", err)
		}

		// Désérialiser les settings
		if err := json.Unmarshal([]byte(settingsJSON), &zone.Settings); err != nil {
			logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")

			logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")
		}

		// Récupérer le nombre de joueurs
		zone.PlayerCount, _ = r.GetPlayerCount(zone.ID)

		zones = append(zones, &zone)
	}

	return zones, nil
}

// GetByType récupère les zones par type
func (r *ZoneRepository) GetByType(zoneType string) ([]*models.Zone, error) {
	var zones []*models.Zone

	query := `
		SELECT id, name, display_name, description, type, level,
		       min_x, min_y, min_z, max_x, max_y, max_z,
		       spawn_x, spawn_y, spawn_z, max_players,
		       is_pvp, is_safe_zone, settings, status,
		       created_at, updated_at
		FROM zones 
		WHERE type = $1 AND status != 'disabled'
		ORDER BY level ASC, name ASC`

	rows, err := r.db.Query(query, zoneType)
	if err != nil {
		return nil, fmt.Errorf("failed to get zones by type: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var zone models.Zone
		var settingsJSON string

		err := rows.Scan(
			&zone.ID, &zone.Name, &zone.DisplayName, &zone.Description, &zone.Type, &zone.Level,
			&zone.MinX, &zone.MinY, &zone.MinZ, &zone.MaxX, &zone.MaxY, &zone.MaxZ,
			&zone.SpawnX, &zone.SpawnY, &zone.SpawnZ, &zone.MaxPlayers,
			&zone.IsPvP, &zone.IsSafeZone, &settingsJSON, &zone.Status,
			&zone.CreatedAt, &zone.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan zone: %w", err)
		}

		// Désérialiser les settings
		if err := json.Unmarshal([]byte(settingsJSON), &zone.Settings); err != nil {
			logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")

			logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")
		}

		zone.PlayerCount, _ = r.GetPlayerCount(zone.ID)
		zones = append(zones, &zone)
	}

	return zones, nil
}

// GetByLevel récupère les zones par niveau
func (r *ZoneRepository) GetByLevel(minLevel, maxLevel int) ([]*models.Zone, error) {
	var zones []*models.Zone

	query := `
		SELECT id, name, display_name, description, type, level,
		       min_x, min_y, min_z, max_x, max_y, max_z,
		       spawn_x, spawn_y, spawn_z, max_players,
		       is_pvp, is_safe_zone, settings, status,
		       created_at, updated_at
		FROM zones 
		WHERE level BETWEEN $1 AND $2 AND status != 'disabled'
		ORDER BY level ASC, name ASC`

	rows, err := r.db.Query(query, minLevel, maxLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to get zones by level: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var zone models.Zone
		var settingsJSON string

		err := rows.Scan(
			&zone.ID, &zone.Name, &zone.DisplayName, &zone.Description, &zone.Type, &zone.Level,
			&zone.MinX, &zone.MinY, &zone.MinZ, &zone.MaxX, &zone.MaxY, &zone.MaxZ,
			&zone.SpawnX, &zone.SpawnY, &zone.SpawnZ, &zone.MaxPlayers,
			&zone.IsPvP, &zone.IsSafeZone, &settingsJSON, &zone.Status,
			&zone.CreatedAt, &zone.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan zone: %w", err)
		}

		// Désérialiser les settings
		if err := json.Unmarshal([]byte(settingsJSON), &zone.Settings); err != nil {
			logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")

			logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")
		}

		zone.PlayerCount, _ = r.GetPlayerCount(zone.ID)
		zones = append(zones, &zone)
	}

	return zones, nil
}

// Update met à jour une zone
func (r *ZoneRepository) Update(zone *models.Zone) error {
	// Sérialiser les settings en JSON
	settingsJSON, err := json.Marshal(zone.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	query := `
		UPDATE zones SET 
			name = $2, display_name = $3, description = $4, type = $5, level = $6,
			min_x = $7, min_y = $8, min_z = $9, max_x = $10, max_y = $11, max_z = $12,
			spawn_x = $13, spawn_y = $14, spawn_z = $15, max_players = $16,
			is_pvp = $17, is_safe_zone = $18, settings = $19, status = $20
		WHERE id = $1`

	result, err := r.db.Exec(query,
		zone.ID, zone.Name, zone.DisplayName, zone.Description, zone.Type, zone.Level,
		zone.MinX, zone.MinY, zone.MinZ, zone.MaxX, zone.MaxY, zone.MaxZ,
		zone.SpawnX, zone.SpawnY, zone.SpawnZ, zone.MaxPlayers,
		zone.IsPvP, zone.IsSafeZone, settingsJSON, zone.Status,
	)
	if err != nil {
		return fmt.Errorf("failed to update zone: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("zone not found")
	}

	return nil
}

// Delete supprime une zone (soft delete)
func (r *ZoneRepository) Delete(id string) error {
	query := `UPDATE zones SET status = 'disabled' WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete zone: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("zone not found")
	}

	return nil
}

// GetPlayerCount récupère le nombre de joueurs dans une zone
func (r *ZoneRepository) GetPlayerCount(zoneID string) (int, error) {
	var count int

	query := `SELECT COUNT(*) FROM player_positions WHERE zone_id = $1 AND is_online = true`

	err := r.db.QueryRow(query, zoneID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get player count: %w", err)
	}

	return count, nil
}

// GetPlayersInZone récupère tous les joueurs dans une zone
func (r *ZoneRepository) GetPlayersInZone(zoneID string) ([]*models.PlayerPosition, error) {
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
		return nil, fmt.Errorf("failed to get players in zone: %w", err)
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

// GetTransitions récupère les transitions depuis une zone
func (r *ZoneRepository) GetTransitions(fromZoneID string) ([]*models.ZoneTransition, error) {
	var transitions []*models.ZoneTransition

	query := `
		SELECT id, from_zone_id, to_zone_id,
		       trigger_x, trigger_y, trigger_z, trigger_radius,
		       destination_x, destination_y, destination_z,
		       required_level, required_quest, is_active,
		       created_at, updated_at
		FROM zone_transitions 
		WHERE from_zone_id = $1 AND is_active = true
		ORDER BY created_at ASC`

	rows, err := r.db.Query(query, fromZoneID)
	if err != nil {
		return nil, fmt.Errorf("failed to get zone transitions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var transition models.ZoneTransition
		var requiredQuest sql.NullString

		err := rows.Scan(
			&transition.ID, &transition.FromZoneID, &transition.ToZoneID,
			&transition.TriggerX, &transition.TriggerY, &transition.TriggerZ, &transition.TriggerRadius,
			&transition.DestinationX, &transition.DestinationY, &transition.DestinationZ,
			&transition.RequiredLevel, &requiredQuest, &transition.IsActive,
			&transition.CreatedAt, &transition.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan zone transition: %w", err)
		}

		if requiredQuest.Valid {
			transition.RequiredQuest = requiredQuest.String
		}

		transitions = append(transitions, &transition)
	}

	return transitions, nil
}

// CreateTransition crée une nouvelle transition entre zones
func (r *ZoneRepository) CreateTransition(transition *models.ZoneTransition) error {
	query := `
		INSERT INTO zone_transitions (
			id, from_zone_id, to_zone_id,
			trigger_x, trigger_y, trigger_z, trigger_radius,
			destination_x, destination_y, destination_z,
			required_level, required_quest, is_active
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)`

	var requiredQuest interface{}
	if transition.RequiredQuest != "" {
		requiredQuest = transition.RequiredQuest
	}

	_, err := r.db.Exec(query,
		transition.ID, transition.FromZoneID, transition.ToZoneID,
		transition.TriggerX, transition.TriggerY, transition.TriggerZ, transition.TriggerRadius,
		transition.DestinationX, transition.DestinationY, transition.DestinationZ,
		transition.RequiredLevel, requiredQuest, transition.IsActive,
	)
	if err != nil {
		return fmt.Errorf("failed to create zone transition: %w", err)
	}

	return nil
}

// UpdateTransition met à jour une transition
func (r *ZoneRepository) UpdateTransition(transition *models.ZoneTransition) error {
	query := `
		UPDATE zone_transitions SET 
			from_zone_id = $2, to_zone_id = $3,
			trigger_x = $4, trigger_y = $5, trigger_z = $6, trigger_radius = $7,
			destination_x = $8, destination_y = $9, destination_z = $10,
			required_level = $11, required_quest = $12, is_active = $13
		WHERE id = $1`

	var requiredQuest interface{}
	if transition.RequiredQuest != "" {
		requiredQuest = transition.RequiredQuest
	}

	result, err := r.db.Exec(query,
		transition.ID, transition.FromZoneID, transition.ToZoneID,
		transition.TriggerX, transition.TriggerY, transition.TriggerZ, transition.TriggerRadius,
		transition.DestinationX, transition.DestinationY, transition.DestinationZ,
		transition.RequiredLevel, requiredQuest, transition.IsActive,
	)
	if err != nil {
		return fmt.Errorf("failed to update zone transition: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("zone transition not found")
	}

	return nil
}

// DeleteTransition supprime une transition
func (r *ZoneRepository) DeleteTransition(id uuid.UUID) error {
	query := `DELETE FROM zone_transitions WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete zone transition: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("zone transition not found")
	}

	return nil
}
