package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"world/internal/database"
	"world/internal/models"
)

// NPCRepositoryInterface définit les méthodes du repository NPC
type NPCRepositoryInterface interface {
	// Méthodes de base CRUD
	Create(npc *models.NPC) error
	GetByID(id string) (*models.NPC, error)
	GetAll() ([]*models.NPC, error)
	Update(npc *models.NPC) error
	Delete(id string) error

	// Méthodes spécifiques
	GetByZone(zoneID string) ([]*models.NPC, error)
	GetByType(npcType string) ([]*models.NPC, error)
	GetActiveByZone(zoneID string) ([]*models.NPC, error)
	SetStatus(id string, status string) error
	GetNearbyNPCs(zoneID string, x, y, z, radius float64) ([]*models.NPC, error)
}

// NPCRepository implémente l'interface NPCRepositoryInterface
type NPCRepository struct {
	db *database.DB
}

func NewNPCRepository(db *database.DB) NPCRepositoryInterface {
	return &NPCRepository{db: db}
}

func (r *NPCRepository) Create(npc *models.NPC) error {
	// Sérialiser le comportement en JSON
	behaviorJSON, err := json.Marshal(npc.Behavior)
	if err != nil {
		return fmt.Errorf("failed to marshal behavior: %w", err)
	}

	query := `
		INSERT INTO npcs (
			id, zone_id, name, type, subtype,
			position_x, position_y, position_z, rotation,
			model, texture, scale, behavior,
			level, health, max_health, status
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		)`

	_, err = r.db.Exec(query,
		npc.ID, npc.ZoneID, npc.Name, npc.Type, npc.Subtype,
		npc.PositionX, npc.PositionY, npc.PositionZ, npc.Rotation,
		npc.Model, npc.Texture, npc.Scale, behaviorJSON,
		npc.Level, npc.Health, npc.MaxHealth, npc.Status,
	)
	if err != nil {
		return fmt.Errorf("failed to create NPC: %w", err)
	}

	return nil
}

func (r *NPCRepository) GetByID(id string) (*models.NPC, error) {
	var npc models.NPC
	var behaviorJSON string

	query := `
		SELECT id, zone_id, name, type, subtype,
		       position_x, position_y, position_z, rotation,
		       model, texture, scale, behavior,
		       level, health, max_health, status, last_seen,
		       created_at, updated_at
		FROM npcs 
		WHERE id = $1`

	row := r.db.QueryRow(query, id)
	err := row.Scan(
		&npc.ID, &npc.ZoneID, &npc.Name, &npc.Type, &npc.Subtype,
		&npc.PositionX, &npc.PositionY, &npc.PositionZ, &npc.Rotation,
		&npc.Model, &npc.Texture, &npc.Scale, &behaviorJSON,
		&npc.Level, &npc.Health, &npc.MaxHealth, &npc.Status, &npc.LastSeen,
		&npc.CreatedAt, &npc.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("NPC not found")
		}
		return nil, fmt.Errorf("failed to get NPC by ID: %w", err)
	}

	// Désérialiser le comportement
	if err := json.Unmarshal([]byte(behaviorJSON), &npc.Behavior); err != nil {
		logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")

		logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")
	}

	return &npc, nil
}

func (r *NPCRepository) GetAll() ([]*models.NPC, error) {
	var npcs []*models.NPC

	query := `
		SELECT id, zone_id, name, type, subtype,
		       position_x, position_y, position_z, rotation,
		       model, texture, scale, behavior,
		       level, health, max_health, status, last_seen,
		       created_at, updated_at
		FROM npcs 
		WHERE status != 'deleted'
		ORDER BY zone_id, type, name`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all NPCs: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var npc models.NPC
		var behaviorJSON string

		err := rows.Scan(
			&npc.ID, &npc.ZoneID, &npc.Name, &npc.Type, &npc.Subtype,
			&npc.PositionX, &npc.PositionY, &npc.PositionZ, &npc.Rotation,
			&npc.Model, &npc.Texture, &npc.Scale, &behaviorJSON,
			&npc.Level, &npc.Health, &npc.MaxHealth, &npc.Status, &npc.LastSeen,
			&npc.CreatedAt, &npc.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan NPC: %w", err)
		}

		// Désérialiser le comportement
		if err := json.Unmarshal([]byte(behaviorJSON), &npc.Behavior); err != nil {
			logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")

			logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")
		}

		npcs = append(npcs, &npc)
	}

	return npcs, nil
}

func (r *NPCRepository) Update(npc *models.NPC) error {
	// Sérialiser le comportement en JSON
	behaviorJSON, err := json.Marshal(npc.Behavior)
	if err != nil {
		return fmt.Errorf("failed to marshal behavior: %w", err)
	}

	query := `
		UPDATE npcs SET 
			zone_id = $2, name = $3, type = $4, subtype = $5,
			position_x = $6, position_y = $7, position_z = $8, rotation = $9,
			model = $10, texture = $11, scale = $12, behavior = $13,
			level = $14, health = $15, max_health = $16, status = $17, last_seen = $18
		WHERE id = $1`

	result, err := r.db.Exec(query,
		npc.ID, npc.ZoneID, npc.Name, npc.Type, npc.Subtype,
		npc.PositionX, npc.PositionY, npc.PositionZ, npc.Rotation,
		npc.Model, npc.Texture, npc.Scale, behaviorJSON,
		npc.Level, npc.Health, npc.MaxHealth, npc.Status, npc.LastSeen,
	)
	if err != nil {
		return fmt.Errorf("failed to update NPC: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("NPC not found")
	}

	return nil
}

func (r *NPCRepository) Delete(id string) error {
	query := `UPDATE npcs SET status = 'deleted' WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete NPC: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("NPC not found")
	}

	return nil
}

func (r *NPCRepository) GetByZone(zoneID string) ([]*models.NPC, error) {
	var npcs []*models.NPC

	query := `
		SELECT id, zone_id, name, type, subtype,
		       position_x, position_y, position_z, rotation,
		       model, texture, scale, behavior,
		       level, health, max_health, status, last_seen,
		       created_at, updated_at
		FROM npcs 
		WHERE zone_id = $1 AND status != 'deleted'
		ORDER BY type, name`

	rows, err := r.db.Query(query, zoneID)
	if err != nil {
		return nil, fmt.Errorf("failed to get NPCs by zone: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var npc models.NPC
		var behaviorJSON string

		err := rows.Scan(
			&npc.ID, &npc.ZoneID, &npc.Name, &npc.Type, &npc.Subtype,
			&npc.PositionX, &npc.PositionY, &npc.PositionZ, &npc.Rotation,
			&npc.Model, &npc.Texture, &npc.Scale, &behaviorJSON,
			&npc.Level, &npc.Health, &npc.MaxHealth, &npc.Status, &npc.LastSeen,
			&npc.CreatedAt, &npc.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan NPC: %w", err)
		}

		// Désérialiser le comportement
		if err := json.Unmarshal([]byte(behaviorJSON), &npc.Behavior); err != nil {
			logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")

			logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")
		}

		npcs = append(npcs, &npc)
	}

	return npcs, nil
}

func (r *NPCRepository) GetByType(npcType string) ([]*models.NPC, error) {
	var npcs []*models.NPC

	query := `
		SELECT id, zone_id, name, type, subtype,
		       position_x, position_y, position_z, rotation,
		       model, texture, scale, behavior,
		       level, health, max_health, status, last_seen,
		       created_at, updated_at
		FROM npcs 
		WHERE type = $1 AND status != 'deleted'
		ORDER BY zone_id, name`

	rows, err := r.db.Query(query, npcType)
	if err != nil {
		return nil, fmt.Errorf("failed to get NPCs by type: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var npc models.NPC
		var behaviorJSON string

		err := rows.Scan(
			&npc.ID, &npc.ZoneID, &npc.Name, &npc.Type, &npc.Subtype,
			&npc.PositionX, &npc.PositionY, &npc.PositionZ, &npc.Rotation,
			&npc.Model, &npc.Texture, &npc.Scale, &behaviorJSON,
			&npc.Level, &npc.Health, &npc.MaxHealth, &npc.Status, &npc.LastSeen,
			&npc.CreatedAt, &npc.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan NPC: %w", err)
		}

		// Désérialiser le comportement
		if err := json.Unmarshal([]byte(behaviorJSON), &npc.Behavior); err != nil {
			logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")

			logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")
		}

		npcs = append(npcs, &npc)
	}

	return npcs, nil
}

func (r *NPCRepository) GetActiveByZone(zoneID string) ([]*models.NPC, error) {
	var npcs []*models.NPC

	query := `
		SELECT id, zone_id, name, type, subtype,
		       position_x, position_y, position_z, rotation,
		       model, texture, scale, behavior,
		       level, health, max_health, status, last_seen,
		       created_at, updated_at
		FROM npcs 
		WHERE zone_id = $1 AND status = 'active'
		ORDER BY type, name`

	rows, err := r.db.Query(query, zoneID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active NPCs by zone: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var npc models.NPC
		var behaviorJSON string

		err := rows.Scan(
			&npc.ID, &npc.ZoneID, &npc.Name, &npc.Type, &npc.Subtype,
			&npc.PositionX, &npc.PositionY, &npc.PositionZ, &npc.Rotation,
			&npc.Model, &npc.Texture, &npc.Scale, &behaviorJSON,
			&npc.Level, &npc.Health, &npc.MaxHealth, &npc.Status, &npc.LastSeen,
			&npc.CreatedAt, &npc.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan NPC: %w", err)
		}

		// Désérialiser le comportement
		if err := json.Unmarshal([]byte(behaviorJSON), &npc.Behavior); err != nil {
			logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")

			logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")
		}

		npcs = append(npcs, &npc)
	}

	return npcs, nil
}

func (r *NPCRepository) SetStatus(id string, status string) error {
	query := `UPDATE npcs SET status = $2, last_seen = $3 WHERE id = $1`

	result, err := r.db.Exec(query, id, status, time.Now())
	if err != nil {
		return fmt.Errorf("failed to set NPC status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("NPC not found")
	}

	return nil
}

func (r *NPCRepository) GetNearbyNPCs(zoneID string, x, y, z, radius float64) ([]*models.NPC, error) {
	var npcs []*models.NPC

	// Requête utilisant la distance euclidienne en 3D
	query := `
		SELECT id, zone_id, name, type, subtype,
		       position_x, position_y, position_z, rotation,
		       model, texture, scale, behavior,
		       level, health, max_health, status, last_seen,
		       created_at, updated_at,
		       SQRT(POWER(position_x - $2, 2) + POWER(position_y - $3, 2) + POWER(position_z - $4, 2)) as distance
		FROM npcs 
		WHERE zone_id = $1 
		  AND status = 'active'
		  AND SQRT(POWER(position_x - $2, 2) + POWER(position_y - $3, 2) + POWER(position_z - $4, 2)) <= $5
		ORDER BY distance ASC`

	rows, err := r.db.Query(query, zoneID, x, y, z, radius)
	if err != nil {
		return nil, fmt.Errorf("failed to get nearby NPCs: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var npc models.NPC
		var behaviorJSON string
		var distance float64

		err := rows.Scan(
			&npc.ID, &npc.ZoneID, &npc.Name, &npc.Type, &npc.Subtype,
			&npc.PositionX, &npc.PositionY, &npc.PositionZ, &npc.Rotation,
			&npc.Model, &npc.Texture, &npc.Scale, &behaviorJSON,
			&npc.Level, &npc.Health, &npc.MaxHealth, &npc.Status, &npc.LastSeen,
			&npc.CreatedAt, &npc.UpdatedAt, &distance,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan nearby NPC: %w", err)
		}

		// Désérialiser le comportement
		if err := json.Unmarshal([]byte(behaviorJSON), &npc.Behavior); err != nil {
			logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")

			logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")
		}

		npcs = append(npcs, &npc)
	}

	return npcs, nil
}

// WorldEventRepositoryInterface définit les méthodes du repository WorldEvent
type WorldEventRepositoryInterface interface {
	Create(event *models.WorldEvent) error
	GetByID(id string) (*models.WorldEvent, error)
	GetAll() ([]*models.WorldEvent, error)
	Update(event *models.WorldEvent) error
	Delete(id string) error
	GetByZone(zoneID string) ([]*models.WorldEvent, error)
	GetActive() ([]*models.WorldEvent, error)
	GetUpcoming() ([]*models.WorldEvent, error)
	GetByStatus(status string) ([]*models.WorldEvent, error)
}

// WorldEventRepository implémente l'interface WorldEventRepositoryInterface
type WorldEventRepository struct {
	db *database.DB
}

func NewWorldEventRepository(db *database.DB) WorldEventRepositoryInterface {
	return &WorldEventRepository{db: db}
}

func (r *WorldEventRepository) Create(event *models.WorldEvent) error {
	// Sérialiser les récompenses en JSON
	rewardsJSON, err := json.Marshal(event.Rewards)
	if err != nil {
		return fmt.Errorf("failed to marshal rewards: %w", err)
	}

	query := `
		INSERT INTO world_events (
			id, zone_id, name, description, type,
			start_time, end_time, duration,
			max_participants, min_level, max_level, rewards, status
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)`

	_, err = r.db.Exec(query,
		event.ID, event.ZoneID, event.Name, event.Description, event.Type,
		event.StartTime, event.EndTime, event.Duration,
		event.MaxParticipants, event.MinLevel, event.MaxLevel, rewardsJSON, event.Status,
	)
	if err != nil {
		return fmt.Errorf("failed to create world event: %w", err)
	}

	return nil
}

func (r *WorldEventRepository) GetByID(id string) (*models.WorldEvent, error) {
	var event models.WorldEvent
	var rewardsJSON string
	var zoneID sql.NullString

	query := `
		SELECT id, zone_id, name, description, type,
		       start_time, end_time, duration,
		       max_participants, min_level, max_level, rewards, status,
		       created_at, updated_at
		FROM world_events 
		WHERE id = $1`

	row := r.db.QueryRow(query, id)
	err := row.Scan(
		&event.ID, &zoneID, &event.Name, &event.Description, &event.Type,
		&event.StartTime, &event.EndTime, &event.Duration,
		&event.MaxParticipants, &event.MinLevel, &event.MaxLevel, &rewardsJSON, &event.Status,
		&event.CreatedAt, &event.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("world event not found")
		}
		return nil, fmt.Errorf("failed to get world event by ID: %w", err)
	}

	if zoneID.Valid {
		event.ZoneID = zoneID.String
	}

	// Désérialiser les récompenses
	if err := json.Unmarshal([]byte(rewardsJSON), &event.Rewards); err != nil {
		logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")

		logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")
	}

	return &event, nil
}

func (r *WorldEventRepository) GetAll() ([]*models.WorldEvent, error) {
	var events []*models.WorldEvent

	query := `
		SELECT id, zone_id, name, description, type,
		       start_time, end_time, duration,
		       max_participants, min_level, max_level, rewards, status,
		       created_at, updated_at
		FROM world_events 
		WHERE status != 'canceled'
		ORDER BY start_time ASC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all world events: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var event models.WorldEvent
		var rewardsJSON string
		var zoneID sql.NullString

		err := rows.Scan(
			&event.ID, &zoneID, &event.Name, &event.Description, &event.Type,
			&event.StartTime, &event.EndTime, &event.Duration,
			&event.MaxParticipants, &event.MinLevel, &event.MaxLevel, &rewardsJSON, &event.Status,
			&event.CreatedAt, &event.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan world event: %w", err)
		}

		if zoneID.Valid {
			event.ZoneID = zoneID.String
		}

		// Désérialiser les récompenses
		if err := json.Unmarshal([]byte(rewardsJSON), &event.Rewards); err != nil {
			logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")

			logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")
		}

		events = append(events, &event)
	}

	return events, nil
}

func (r *WorldEventRepository) Update(event *models.WorldEvent) error {
	// Sérialiser les récompenses en JSON
	rewardsJSON, err := json.Marshal(event.Rewards)
	if err != nil {
		return fmt.Errorf("failed to marshal rewards: %w", err)
	}

	query := `
		UPDATE world_events SET 
			zone_id = $2, name = $3, description = $4, type = $5,
			start_time = $6, end_time = $7, duration = $8,
			max_participants = $9, min_level = $10, max_level = $11, rewards = $12, status = $13
		WHERE id = $1`

	result, err := r.db.Exec(query,
		event.ID, event.ZoneID, event.Name, event.Description, event.Type,
		event.StartTime, event.EndTime, event.Duration,
		event.MaxParticipants, event.MinLevel, event.MaxLevel, rewardsJSON, event.Status,
	)
	if err != nil {
		return fmt.Errorf("failed to update world event: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("world event not found")
	}

	return nil
}

func (r *WorldEventRepository) Delete(id string) error {
	query := `UPDATE world_events SET status = 'canceled' WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete world event: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("world event not found")
	}

	return nil
}

func (r *WorldEventRepository) GetByZone(zoneID string) ([]*models.WorldEvent, error) {
	var events []*models.WorldEvent

	query := `
		SELECT id, zone_id, name, description, type,
		       start_time, end_time, duration,
		       max_participants, min_level, max_level, rewards, status,
		       created_at, updated_at
		FROM world_events 
		WHERE zone_id = $1 AND status != 'canceled'
		ORDER BY start_time ASC`

	rows, err := r.db.Query(query, zoneID)
	if err != nil {
		return nil, fmt.Errorf("failed to get world events by zone: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var event models.WorldEvent
		var rewardsJSON string
		var eventZoneID sql.NullString

		err := rows.Scan(
			&event.ID, &eventZoneID, &event.Name, &event.Description, &event.Type,
			&event.StartTime, &event.EndTime, &event.Duration,
			&event.MaxParticipants, &event.MinLevel, &event.MaxLevel, &rewardsJSON, &event.Status,
			&event.CreatedAt, &event.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan world event: %w", err)
		}

		if eventZoneID.Valid {
			event.ZoneID = eventZoneID.String
		}

		// Désérialiser les récompenses
		if err := json.Unmarshal([]byte(rewardsJSON), &event.Rewards); err != nil {
			logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")

			logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")
		}

		events = append(events, &event)
	}

	return events, nil
}

func (r *WorldEventRepository) GetActive() ([]*models.WorldEvent, error) {
	var events []*models.WorldEvent

	query := `
		SELECT id, zone_id, name, description, type,
		       start_time, end_time, duration,
		       max_participants, min_level, max_level, rewards, status,
		       created_at, updated_at
		FROM world_events 
		WHERE status = 'active' 
		  AND start_time <= CURRENT_TIMESTAMP 
		  AND end_time >= CURRENT_TIMESTAMP
		ORDER BY start_time ASC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active world events: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var event models.WorldEvent
		var rewardsJSON string
		var zoneID sql.NullString

		err := rows.Scan(
			&event.ID, &zoneID, &event.Name, &event.Description, &event.Type,
			&event.StartTime, &event.EndTime, &event.Duration,
			&event.MaxParticipants, &event.MinLevel, &event.MaxLevel, &rewardsJSON, &event.Status,
			&event.CreatedAt, &event.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan world event: %w", err)
		}

		if zoneID.Valid {
			event.ZoneID = zoneID.String
		}

		// Désérialiser les récompenses
		if err := json.Unmarshal([]byte(rewardsJSON), &event.Rewards); err != nil {
			logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")

			logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")
		}

		events = append(events, &event)
	}

	return events, nil
}

func (r *WorldEventRepository) GetUpcoming() ([]*models.WorldEvent, error) {
	var events []*models.WorldEvent

	query := `
		SELECT id, zone_id, name, description, type,
		       start_time, end_time, duration,
		       max_participants, min_level, max_level, rewards, status,
		       created_at, updated_at
		FROM world_events 
		WHERE status = 'scheduled' 
		  AND start_time > CURRENT_TIMESTAMP
		ORDER BY start_time ASC
		LIMIT 10`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get upcoming world events: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var event models.WorldEvent
		var rewardsJSON string
		var zoneID sql.NullString

		err := rows.Scan(
			&event.ID, &zoneID, &event.Name, &event.Description, &event.Type,
			&event.StartTime, &event.EndTime, &event.Duration,
			&event.MaxParticipants, &event.MinLevel, &event.MaxLevel, &rewardsJSON, &event.Status,
			&event.CreatedAt, &event.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan world event: %w", err)
		}

		if zoneID.Valid {
			event.ZoneID = zoneID.String
		}

		// Désérialiser les récompenses
		if err := json.Unmarshal([]byte(rewardsJSON), &event.Rewards); err != nil {
			logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")

			logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")
		}

		events = append(events, &event)
	}

	return events, nil
}

func (r *WorldEventRepository) GetByStatus(status string) ([]*models.WorldEvent, error) {
	var events []*models.WorldEvent

	query := `
		SELECT id, zone_id, name, description, type,
		       start_time, end_time, duration,
		       max_participants, min_level, max_level, rewards, status,
		       created_at, updated_at
		FROM world_events 
		WHERE status = $1
		ORDER BY start_time ASC`

	rows, err := r.db.Query(query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get world events by status: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var event models.WorldEvent
		var rewardsJSON string
		var zoneID sql.NullString

		err := rows.Scan(
			&event.ID, &zoneID, &event.Name, &event.Description, &event.Type,
			&event.StartTime, &event.EndTime, &event.Duration,
			&event.MaxParticipants, &event.MinLevel, &event.MaxLevel, &rewardsJSON, &event.Status,
			&event.CreatedAt, &event.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan world event: %w", err)
		}

		if zoneID.Valid {
			event.ZoneID = zoneID.String
		}

		// Désérialiser les récompenses
		if err := json.Unmarshal([]byte(rewardsJSON), &event.Rewards); err != nil {
			logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")

			logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")
		}

		events = append(events, &event)
	}

	return events, nil
}

// WeatherRepositoryInterface définit les méthodes du repository Weather
type WeatherRepositoryInterface interface {
	Upsert(weather *models.Weather) error
	GetByZone(zoneID string) (*models.Weather, error)
	GetAll() ([]*models.Weather, error)
	Delete(zoneID string) error
	GetActiveWeather() ([]*models.Weather, error)
	UpdateWeatherStatus(zoneID string, isActive bool) error
}

// WeatherRepository implémente l'interface WeatherRepositoryInterface
type WeatherRepository struct {
	db *database.DB
}

func NewWeatherRepository(db *database.DB) WeatherRepositoryInterface {
	return &WeatherRepository{db: db}
}

func (r *WeatherRepository) Upsert(weather *models.Weather) error {
	query := `
		INSERT INTO weather (
			zone_id, type, intensity, temperature, wind_speed, wind_direction, visibility,
			start_time, end_time, is_active
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
		ON CONFLICT (zone_id) 
		DO UPDATE SET 
			type = EXCLUDED.type,
			intensity = EXCLUDED.intensity,
			temperature = EXCLUDED.temperature,
			wind_speed = EXCLUDED.wind_speed,
			wind_direction = EXCLUDED.wind_direction,
			visibility = EXCLUDED.visibility,
			start_time = EXCLUDED.start_time,
			end_time = EXCLUDED.end_time,
			is_active = EXCLUDED.is_active,
			updated_at = CURRENT_TIMESTAMP`

	_, err := r.db.Exec(query,
		weather.ZoneID, weather.Type, weather.Intensity, weather.Temperature,
		weather.WindSpeed, weather.WindDirection, weather.Visibility,
		weather.StartTime, weather.EndTime, weather.IsActive,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert weather: %w", err)
	}

	return nil
}

func (r *WeatherRepository) GetByZone(zoneID string) (*models.Weather, error) {
	var weather models.Weather

	query := `
		SELECT zone_id, type, intensity, temperature, wind_speed, wind_direction, visibility,
		       start_time, end_time, is_active, created_at, updated_at
		FROM weather 
		WHERE zone_id = $1`

	row := r.db.QueryRow(query, zoneID)
	err := row.Scan(
		&weather.ZoneID, &weather.Type, &weather.Intensity, &weather.Temperature,
		&weather.WindSpeed, &weather.WindDirection, &weather.Visibility,
		&weather.StartTime, &weather.EndTime, &weather.IsActive,
		&weather.CreatedAt, &weather.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Retourner une météo par défaut si aucune n'est trouvée
			return &models.Weather{
				ZoneID:        zoneID,
				Type:          "clear",
				Intensity:     0.3,
				Temperature:   22.0,
				WindSpeed:     5.0,
				WindDirection: 180.0,
				Visibility:    1000.0,
				IsActive:      true,
				StartTime:     time.Now(),
				EndTime:       time.Now().Add(24 * time.Hour),
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("failed to get weather by zone: %w", err)
	}

	return &weather, nil
}

func (r *WeatherRepository) GetAll() ([]*models.Weather, error) {
	var weathers []*models.Weather

	query := `
		SELECT zone_id, type, intensity, temperature, wind_speed, wind_direction, visibility,
		       start_time, end_time, is_active, created_at, updated_at
		FROM weather 
		ORDER BY zone_id ASC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all weather: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var weather models.Weather

		err := rows.Scan(
			&weather.ZoneID, &weather.Type, &weather.Intensity, &weather.Temperature,
			&weather.WindSpeed, &weather.WindDirection, &weather.Visibility,
			&weather.StartTime, &weather.EndTime, &weather.IsActive,
			&weather.CreatedAt, &weather.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan weather: %w", err)
		}

		weathers = append(weathers, &weather)
	}

	return weathers, nil
}

func (r *WeatherRepository) Delete(zoneID string) error {
	query := `DELETE FROM weather WHERE zone_id = $1`

	result, err := r.db.Exec(query, zoneID)
	if err != nil {
		return fmt.Errorf("failed to delete weather: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("weather not found")
	}

	return nil
}

func (r *WeatherRepository) GetActiveWeather() ([]*models.Weather, error) {
	var weathers []*models.Weather

	query := `
		SELECT zone_id, type, intensity, temperature, wind_speed, wind_direction, visibility,
		       start_time, end_time, is_active, created_at, updated_at
		FROM weather 
		WHERE is_active = true
		  AND start_time <= CURRENT_TIMESTAMP
		  AND end_time >= CURRENT_TIMESTAMP
		ORDER BY zone_id ASC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active weather: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var weather models.Weather

		err := rows.Scan(
			&weather.ZoneID, &weather.Type, &weather.Intensity, &weather.Temperature,
			&weather.WindSpeed, &weather.WindDirection, &weather.Visibility,
			&weather.StartTime, &weather.EndTime, &weather.IsActive,
			&weather.CreatedAt, &weather.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan weather: %w", err)
		}

		weathers = append(weathers, &weather)
	}

	return weathers, nil
}

func (r *WeatherRepository) UpdateWeatherStatus(zoneID string, isActive bool) error {
	query := `UPDATE weather SET is_active = $2, updated_at = CURRENT_TIMESTAMP WHERE zone_id = $1`

	result, err := r.db.Exec(query, zoneID, isActive)
	if err != nil {
		return fmt.Errorf("failed to update weather status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("weather not found")
	}

	return nil
}
