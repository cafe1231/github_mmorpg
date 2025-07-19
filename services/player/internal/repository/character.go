package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"player/internal/database"
	"player/internal/models"
)

// CharacterRepositoryInterface définit les méthodes du repository character
type CharacterRepositoryInterface interface {
	Create(character *models.Character) error
	GetByID(id uuid.UUID) (*models.Character, error)
	GetByName(name string) (*models.Character, error)
	GetByPlayerID(playerID uuid.UUID) ([]*models.Character, error)
	Update(character *models.Character) error
	Delete(id uuid.UUID) error

	// Stats management
	GetStats(characterID uuid.UUID) (*models.CharacterStats, error)
	UpdateStats(stats *models.CharacterStats) error
	GetCombatStats(characterID uuid.UUID) (*models.CombatStats, error)
	UpdateCombatStats(stats *models.CombatStats) error

	// Modifiers
	AddStatModifier(modifier *models.StatModifier) error
	GetActiveModifiers(characterID uuid.UUID) ([]*models.StatModifier, error)
	RemoveModifier(modifierID uuid.UUID) error
	CleanupExpiredModifiers() error
}

// CharacterRepository implémente l'interface CharacterRepositoryInterface
type CharacterRepository struct {
	db *database.DB
}

// NewCharacterRepository crée une nouvelle instance du repository character
func NewCharacterRepository(db *database.DB) CharacterRepositoryInterface {
	return &CharacterRepository{db: db}
}

// Create crée un nouveau personnage
func (r *CharacterRepository) Create(character *models.Character) error {
	// Sérialiser l'apparence en JSON
	appearanceJSON, err := json.Marshal(character.Appearance)
	if err != nil {
		return fmt.Errorf("failed to marshal appearance: %w", err)
	}

	query := `
		INSERT INTO characters (
			id, player_id, name, class, race, gender, appearance,
			level, experience, zone_id, position_x, position_y, position_z,
			status, last_played, created_at, updated_at
		) VALUES (
			:id, :player_id, :name, :class, :race, :gender, :appearance,
			:level, :experience, :zone_id, :position_x, :position_y, :position_z,
			:status, :last_played, :created_at, :updated_at
		)`

	data := map[string]interface{}{
		"id":          character.ID,
		"player_id":   character.PlayerID,
		"name":        character.Name,
		"class":       character.Class,
		"race":        character.Race,
		"gender":      character.Gender,
		"appearance":  string(appearanceJSON),
		"level":       character.Level,
		"experience":  character.Experience,
		"zone_id":     character.ZoneID,
		"position_x":  character.PositionX,
		"position_y":  character.PositionY,
		"position_z":  character.PositionZ,
		"status":      character.Status,
		"last_played": character.LastPlayed,
		"created_at":  character.CreatedAt,
		"updated_at":  character.UpdatedAt,
	}

	_, err = r.db.NamedExec(query, data)
	if err != nil {
		return fmt.Errorf("failed to create character: %w", err)
	}

	return nil
}

// GetByID récupère un personnage par son ID
func (r *CharacterRepository) GetByID(id uuid.UUID) (*models.Character, error) {
	var character models.Character
	var appearanceJSON string

	query := `
		SELECT id, player_id, name, class, race, gender, appearance,
		       level, experience, zone_id, position_x, position_y, position_z,
		       status, last_played, created_at, updated_at
		FROM characters 
		WHERE id = $1`

	row := r.db.QueryRow(query, id)
	err := row.Scan(
		&character.ID, &character.PlayerID, &character.Name,
		&character.Class, &character.Race, &character.Gender, &appearanceJSON,
		&character.Level, &character.Experience, &character.ZoneID,
		&character.PositionX, &character.PositionY, &character.PositionZ,
		&character.Status, &character.LastPlayed, &character.CreatedAt, &character.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("character not found")
		}
		return nil, fmt.Errorf("failed to get character by ID: %w", err)
	}

	// Désérialiser l'apparence
	if err := if err := json.Unmarshal([]byte(appearanceJSON), &character.Appearance); err != nil {
		logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")
	}; err != nil {
		character.Appearance = models.GetDefaultAppearance(character.Race)
	}

	return &character, nil
}

// GetByName récupère un personnage par son nom
func (r *CharacterRepository) GetByName(name string) (*models.Character, error) {
	var character models.Character
	var appearanceJSON string

	query := `
		SELECT id, player_id, name, class, race, gender, appearance,
		       level, experience, zone_id, position_x, position_y, position_z,
		       status, last_played, created_at, updated_at
		FROM characters 
		WHERE name = $1`

	row := r.db.QueryRow(query, name)
	err := row.Scan(
		&character.ID, &character.PlayerID, &character.Name,
		&character.Class, &character.Race, &character.Gender, &appearanceJSON,
		&character.Level, &character.Experience, &character.ZoneID,
		&character.PositionX, &character.PositionY, &character.PositionZ,
		&character.Status, &character.LastPlayed, &character.CreatedAt, &character.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("character not found")
		}
		return nil, fmt.Errorf("failed to get character by name: %w", err)
	}

	// Désérialiser l'apparence
	if err := if err := json.Unmarshal([]byte(appearanceJSON), &character.Appearance); err != nil {
		logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")
	}; err != nil {
		character.Appearance = models.GetDefaultAppearance(character.Race)
	}

	return &character, nil
}

// GetByPlayerID récupère tous les personnages d'un joueur
func (r *CharacterRepository) GetByPlayerID(playerID uuid.UUID) ([]*models.Character, error) {
	var characters []*models.Character

	query := `
		SELECT id, player_id, name, class, race, gender, appearance,
		       level, experience, zone_id, position_x, position_y, position_z,
		       status, last_played, created_at, updated_at
		FROM characters 
		WHERE player_id = $1 AND status != 'deleted'
		ORDER BY last_played DESC`

	rows, err := r.db.Query(query, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get characters by player ID: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var character models.Character
		var appearanceJSON string

		err := rows.Scan(
			&character.ID, &character.PlayerID, &character.Name,
			&character.Class, &character.Race, &character.Gender, &appearanceJSON,
			&character.Level, &character.Experience, &character.ZoneID,
			&character.PositionX, &character.PositionY, &character.PositionZ,
			&character.Status, &character.LastPlayed, &character.CreatedAt, &character.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan character: %w", err)
		}

		// Désérialiser l'apparence
		if err := if err := json.Unmarshal([]byte(appearanceJSON), &character.Appearance); err != nil {
		logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")
	}; err != nil {
			character.Appearance = models.GetDefaultAppearance(character.Race)
		}

		characters = append(characters, &character)
	}

	return characters, nil
}

// Update met à jour un personnage
func (r *CharacterRepository) Update(character *models.Character) error {
	// Sérialiser l'apparence en JSON
	appearanceJSON, err := json.Marshal(character.Appearance)
	if err != nil {
		return fmt.Errorf("failed to marshal appearance: %w", err)
	}

	query := `
		UPDATE characters SET
			name = :name,
			appearance = :appearance,
			level = :level,
			experience = :experience,
			zone_id = :zone_id,
			position_x = :position_x,
			position_y = :position_y,
			position_z = :position_z,
			status = :status,
			last_played = :last_played,
			updated_at = :updated_at
		WHERE id = :id`

	data := map[string]interface{}{
		"id":          character.ID,
		"name":        character.Name,
		"appearance":  string(appearanceJSON),
		"level":       character.Level,
		"experience":  character.Experience,
		"zone_id":     character.ZoneID,
		"position_x":  character.PositionX,
		"position_y":  character.PositionY,
		"position_z":  character.PositionZ,
		"status":      character.Status,
		"last_played": character.LastPlayed,
		"updated_at":  character.UpdatedAt,
	}

	_, err = r.db.NamedExec(query, data)
	if err != nil {
		return fmt.Errorf("failed to update character: %w", err)
	}

	return nil
}

// Delete supprime un personnage (soft delete)
func (r *CharacterRepository) Delete(id uuid.UUID) error {
	query := `UPDATE characters SET status = 'deleted', updated_at = CURRENT_TIMESTAMP WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete character: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("character not found")
	}

	return nil
}

// GetStats récupère les statistiques d'un personnage
func (r *CharacterRepository) GetStats(characterID uuid.UUID) (*models.CharacterStats, error) {
	var stats models.CharacterStats

	query := `
		SELECT character_id, health, max_health, mana, max_mana,
		       strength, agility, intelligence, vitality,
		       stat_points, skill_points,
		       physical_damage, magical_damage, physical_defense, magical_defense,
		       critical_chance, attack_speed, movement_speed, updated_at
		FROM character_stats 
		WHERE character_id = $1`

	err := r.db.Get(&stats, query, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get character stats: %w", err)
	}

	return &stats, nil
}

// UpdateStats met à jour les statistiques d'un personnage
func (r *CharacterRepository) UpdateStats(stats *models.CharacterStats) error {
	query := `
		UPDATE character_stats SET
			health = :health,
			max_health = :max_health,
			mana = :mana,
			max_mana = :max_mana,
			strength = :strength,
			agility = :agility,
			intelligence = :intelligence,
			vitality = :vitality,
			stat_points = :stat_points,
			skill_points = :skill_points,
			physical_damage = :physical_damage,
			magical_damage = :magical_damage,
			physical_defense = :physical_defense,
			magical_defense = :magical_defense,
			critical_chance = :critical_chance,
			attack_speed = :attack_speed,
			movement_speed = :movement_speed,
			updated_at = :updated_at
		WHERE character_id = :character_id`

	_, err := r.db.NamedExec(query, stats)
	if err != nil {
		return fmt.Errorf("failed to update character stats: %w", err)
	}

	return nil
}

// GetCombatStats récupère les statistiques de combat d'un personnage
func (r *CharacterRepository) GetCombatStats(characterID uuid.UUID) (*models.CombatStats, error) {
	var stats models.CombatStats

	query := `
		SELECT character_id, monsters_killed, bosses_killed, deaths,
		       damage_dealt, damage_taken, healing_done,
		       pvp_kills, pvp_deaths, pvp_damage_dealt, pvp_damage_taken,
		       quests_completed, items_looted, gold_earned, distance_traveled,
		       updated_at
		FROM combat_stats 
		WHERE character_id = $1`

	err := r.db.Get(&stats, query, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get combat stats: %w", err)
	}

	return &stats, nil
}

// UpdateCombatStats met à jour les statistiques de combat
func (r *CharacterRepository) UpdateCombatStats(stats *models.CombatStats) error {
	query := `
		UPDATE combat_stats SET
			monsters_killed = :monsters_killed,
			bosses_killed = :bosses_killed,
			deaths = :deaths,
			damage_dealt = :damage_dealt,
			damage_taken = :damage_taken,
			healing_done = :healing_done,
			pvp_kills = :pvp_kills,
			pvp_deaths = :pvp_deaths,
			pvp_damage_dealt = :pvp_damage_dealt,
			pvp_damage_taken = :pvp_damage_taken,
			quests_completed = :quests_completed,
			items_looted = :items_looted,
			gold_earned = :gold_earned,
			distance_traveled = :distance_traveled,
			updated_at = :updated_at
		WHERE character_id = :character_id`

	_, err := r.db.NamedExec(query, stats)
	if err != nil {
		return fmt.Errorf("failed to update combat stats: %w", err)
	}

	return nil
}

// AddStatModifier ajoute un modificateur de statistiques
func (r *CharacterRepository) AddStatModifier(modifier *models.StatModifier) error {
	query := `
		INSERT INTO stat_modifiers (
			id, character_id, type, source, stat_name,
			value, duration, expires_at, created_at
		) VALUES (
			:id, :character_id, :type, :source, :stat_name,
			:value, :duration, :expires_at, :created_at
		)`

	_, err := r.db.NamedExec(query, modifier)
	if err != nil {
		return fmt.Errorf("failed to add stat modifier: %w", err)
	}

	return nil
}

// GetActiveModifiers récupère tous les modificateurs actifs d'un personnage
func (r *CharacterRepository) GetActiveModifiers(characterID uuid.UUID) ([]*models.StatModifier, error) {
	var modifiers []*models.StatModifier

	query := `
		SELECT id, character_id, type, source, stat_name,
		       value, duration, expires_at, created_at
		FROM stat_modifiers 
		WHERE character_id = $1 
		AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
		ORDER BY created_at DESC`

	err := r.db.Select(&modifiers, query, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active modifiers: %w", err)
	}

	return modifiers, nil
}

// RemoveModifier supprime un modificateur
func (r *CharacterRepository) RemoveModifier(modifierID uuid.UUID) error {
	query := `DELETE FROM stat_modifiers WHERE id = $1`

	result, err := r.db.Exec(query, modifierID)
	if err != nil {
		return fmt.Errorf("failed to remove modifier: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("modifier not found")
	}

	return nil
}

// CleanupExpiredModifiers supprime les modificateurs expirés
func (r *CharacterRepository) CleanupExpiredModifiers() error {
	query := `DELETE FROM stat_modifiers WHERE expires_at IS NOT NULL AND expires_at < CURRENT_TIMESTAMP`

	result, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired modifiers: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		// Log du nettoyage si nécessaire
	}

	return nil
}

