// internal/repository/effect.go
package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"combat/internal/database"
	"combat/internal/models"
)

// EffectRepositoryInterface définit les méthodes du repository effect
type EffectRepositoryInterface interface {
	ApplyEffect(effect *models.StatusEffect) error
	GetActiveEffects(characterID uuid.UUID) ([]*models.StatusEffect, error)
	GetEffectByID(id uuid.UUID) (*models.StatusEffect, error)
	UpdateEffect(effect *models.StatusEffect) error
	RemoveEffect(effectID uuid.UUID) error
	CleanupExpiredEffects() (int, error)
	ProcessPeriodicEffects() (int, error)
	DispelEffects(characterID uuid.UUID, dispelType string, maxCount int) ([]*models.StatusEffect, error)
	
	// Nouvelles méthodes manquantes
	GetByCharacterAndName(characterID uuid.UUID, effectName string) ([]*models.StatusEffect, error)
	Create(effect *models.StatusEffect) error
	Update(effect *models.StatusEffect) error
	Delete(effectID uuid.UUID) error
	GetByID(effectID uuid.UUID) (*models.StatusEffect, error)
	RemoveSessionEffects(sessionID uuid.UUID) (int, error) // Ajouté la méthode manquante
}

// EffectRepository implémente l'interface EffectRepositoryInterface
type EffectRepository struct {
	db *database.DB
}

// NewEffectRepository crée une nouvelle instance du repository effect
func NewEffectRepository(db *database.DB) EffectRepositoryInterface {
	return &EffectRepository{db: db}
}

// Create crée un nouvel effet
func (r *EffectRepository) Create(effect *models.StatusEffect) error {
	statModifiersJSON, err := json.Marshal(effect.StatModifiers)
	if err != nil {
		return fmt.Errorf("failed to marshal stat modifiers: %w", err)
	}

	var periodicEffectJSON []byte
	if effect.PeriodicEffect != nil {
		periodicEffectJSON, err = json.Marshal(effect.PeriodicEffect)
		if err != nil {
			return fmt.Errorf("failed to marshal periodic effect: %w", err)
		}
	}

	query := `
		INSERT INTO status_effects (
			id, character_id, session_id, type, source, source_id, caster_id,
			name, description, stat_modifiers, periodic_effect, 
			started_at, ends_at, duration, stack_count, max_stacks,
			is_dispellable, dispel_type, priority, is_active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22
		)`

	_, err = r.db.Exec(query,
		effect.ID, effect.CharacterID, effect.SessionID, effect.Type, effect.Source,
		effect.SourceID, effect.CasterID, effect.Name, effect.Description,
		statModifiersJSON, periodicEffectJSON, effect.AppliedAt, effect.ExpiresAt,
		effect.Duration, effect.Stacks, effect.MaxStacks, effect.IsDispellable,
		effect.DispelType, effect.Priority, effect.IsActive, time.Now(), time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to create effect: %w", err)
	}

	return nil
}

// GetByCharacterAndName récupère les effets d'un personnage par nom
func (r *EffectRepository) GetByCharacterAndName(characterID uuid.UUID, effectName string) ([]*models.StatusEffect, error) {
	var effects []*models.StatusEffect

	query := `
		SELECT id, character_id, session_id, type, source, source_id, caster_id,
		       name, description, stat_modifiers, periodic_effect,
		       started_at, ends_at, duration, stack_count, max_stacks,
		       is_dispellable, dispel_type, priority, is_active, created_at, updated_at
		FROM status_effects 
		WHERE character_id = $1 AND name = $2 AND is_active = TRUE
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, characterID, effectName)
	if err != nil {
		return nil, fmt.Errorf("failed to query effects by name: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		effect, err := r.scanStatusEffect(rows)
		if err != nil {
			return nil, err
		}
		effects = append(effects, effect)
	}

	return effects, nil
}

// Update met à jour un effet existant
func (r *EffectRepository) Update(effect *models.StatusEffect) error {
	statModifiersJSON, err := json.Marshal(effect.StatModifiers)
	if err != nil {
		return fmt.Errorf("failed to marshal stat modifiers: %w", err)
	}

	var periodicEffectJSON []byte
	if effect.PeriodicEffect != nil {
		periodicEffectJSON, err = json.Marshal(effect.PeriodicEffect)
		if err != nil {
			return fmt.Errorf("failed to marshal periodic effect: %w", err)
		}
	}

	query := `
		UPDATE status_effects 
		SET type = $2, source = $3, source_id = $4, caster_id = $5,
		    name = $6, description = $7, stat_modifiers = $8, periodic_effect = $9,
		    started_at = $10, ends_at = $11, duration = $12, stack_count = $13, max_stacks = $14,
		    is_dispellable = $15, dispel_type = $16, priority = $17, is_active = $18, updated_at = $19
		WHERE id = $1`

	_, err = r.db.Exec(query,
		effect.ID, effect.Type, effect.Source, effect.SourceID, effect.CasterID,
		effect.Name, effect.Description, statModifiersJSON, periodicEffectJSON,
		effect.AppliedAt, effect.ExpiresAt, effect.Duration, effect.Stacks, effect.MaxStacks,
		effect.IsDispellable, effect.DispelType, effect.Priority, effect.IsActive, time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to update effect: %w", err)
	}

	return nil
}

// Delete supprime un effet
func (r *EffectRepository) Delete(effectID uuid.UUID) error {
	query := `DELETE FROM status_effects WHERE id = $1`

	result, err := r.db.Exec(query, effectID)
	if err != nil {
		return fmt.Errorf("failed to delete effect: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("effect not found")
	}

	return nil
}

// GetByID récupère un effet par son ID
func (r *EffectRepository) GetByID(effectID uuid.UUID) (*models.StatusEffect, error) {
	query := `
		SELECT id, character_id, session_id, type, source, source_id, caster_id,
		       name, description, stat_modifiers, periodic_effect,
		       started_at, ends_at, duration, stack_count, max_stacks,
		       is_dispellable, dispel_type, priority, is_active, created_at, updated_at
		FROM status_effects 
		WHERE id = $1`

	row := r.db.QueryRow(query, effectID)
	effect, err := r.scanStatusEffect(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("effect not found")
		}
		return nil, err
	}

	return effect, nil
}

// ApplyEffect applique un effet de statut
func (r *EffectRepository) ApplyEffect(effect *models.StatusEffect) error {
	return r.Create(effect)
}

// GetActiveEffects récupère tous les effets actifs d'un personnage
func (r *EffectRepository) GetActiveEffects(characterID uuid.UUID) ([]*models.StatusEffect, error) {
	var effects []*models.StatusEffect

	query := `
		SELECT id, character_id, session_id, type, source, source_id, caster_id,
		       name, description, stat_modifiers, periodic_effect,
		       started_at, ends_at, duration, stack_count, max_stacks,
		       is_dispellable, dispel_type, priority, is_active, created_at, updated_at
		FROM status_effects 
		WHERE character_id = $1 AND is_active = TRUE
		ORDER BY priority DESC, created_at ASC`

	rows, err := r.db.Query(query, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active effects: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		effect, err := r.scanStatusEffect(rows)
		if err != nil {
			return nil, err
		}
		effects = append(effects, effect)
	}

	return effects, nil
}

// GetEffectByID récupère un effet par son ID (alias pour GetByID)
func (r *EffectRepository) GetEffectByID(id uuid.UUID) (*models.StatusEffect, error) {
	return r.GetByID(id)
}

// UpdateEffect met à jour un effet (alias pour Update)
func (r *EffectRepository) UpdateEffect(effect *models.StatusEffect) error {
	return r.Update(effect)
}

// RemoveEffect supprime un effet (alias pour Delete)
func (r *EffectRepository) RemoveEffect(effectID uuid.UUID) error {
	return r.Delete(effectID)
}

// CleanupExpiredEffects nettoie les effets expirés
func (r *EffectRepository) CleanupExpiredEffects() (int, error) {
	query := `
		UPDATE status_effects 
		SET is_active = FALSE, updated_at = NOW()
		WHERE is_active = TRUE 
		AND ends_at IS NOT NULL 
		AND ends_at < NOW()`

	result, err := r.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired effects: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	return int(rowsAffected), nil
}

// ProcessPeriodicEffects traite les effets périodiques
func (r *EffectRepository) ProcessPeriodicEffects() (int, error) {
	query := `
		SELECT id, character_id, session_id, type, source, source_id, caster_id,
		       name, description, stat_modifiers, periodic_effect,
		       started_at, ends_at, duration, stack_count, max_stacks,
		       is_dispellable, dispel_type, priority, is_active, created_at, updated_at
		FROM status_effects 
		WHERE is_active = TRUE 
		AND periodic_effect IS NOT NULL 
		AND (next_tick_at IS NULL OR next_tick_at <= NOW())`

	rows, err := r.db.Query(query)
	if err != nil {
		return 0, fmt.Errorf("failed to get periodic effects: %w", err)
	}
	defer rows.Close()

	processedCount := 0
	for rows.Next() {
		effect, err := r.scanStatusEffect(rows)
		if err != nil {
			continue
		}

		// TODO: Traiter l'effet périodique
		processedCount++

		// Mettre à jour le prochain tick
		nextTick := time.Now().Add(3 * time.Second) // Intervalle par défaut
		updateQuery := `UPDATE status_effects SET next_tick_at = $1 WHERE id = $2`
		r.db.Exec(updateQuery, nextTick, effect.ID)
	}

	return processedCount, nil
}

// DispelEffects dissipe des effets selon leur type
func (r *EffectRepository) DispelEffects(characterID uuid.UUID, dispelType string, maxCount int) ([]*models.StatusEffect, error) {
	var dispelCondition string
	switch dispelType {
	case "magic":
		dispelCondition = "dispel_type IN ('magic', 'all')"
	case "poison":
		dispelCondition = "dispel_type IN ('poison', 'all')"
	case "disease":
		dispelCondition = "dispel_type IN ('disease', 'all')"
	case "curse":
		dispelCondition = "dispel_type IN ('curse', 'all')"
	case "all":
		dispelCondition = "TRUE"
	default:
		dispelCondition = "FALSE"
	}

	query := fmt.Sprintf(`
		SELECT id, character_id, session_id, type, source, source_id, caster_id,
		       name, description, stat_modifiers, periodic_effect,
		       started_at, ends_at, duration, stack_count, max_stacks,
		       is_dispellable, dispel_type, priority, is_active, created_at, updated_at
		FROM status_effects 
		WHERE character_id = $1 AND is_active = TRUE AND is_dispellable = TRUE AND %s
		ORDER BY priority ASC
		LIMIT $2`, dispelCondition)

	rows, err := r.db.Query(query, characterID, maxCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get dispellable effects: %w", err)
	}
	defer rows.Close()

	var dispelledEffects []*models.StatusEffect
	var effectIDs []uuid.UUID

	for rows.Next() {
		effect, err := r.scanStatusEffect(rows)
		if err != nil {
			continue
		}
		dispelledEffects = append(dispelledEffects, effect)
		effectIDs = append(effectIDs, effect.ID)
	}

	// Supprimer les effets dissipés
	if len(effectIDs) > 0 {
		deleteQuery := `DELETE FROM status_effects WHERE id = ANY($1)`
		_, err = r.db.Exec(deleteQuery, pq.Array(effectIDs))
		if err != nil {
			return nil, fmt.Errorf("failed to delete dispelled effects: %w", err)
		}
	}

	return dispelledEffects, nil
}

// RemoveSessionEffects supprime tous les effets d'une session
func (r *EffectRepository) RemoveSessionEffects(sessionID uuid.UUID) (int, error) {
	query := `DELETE FROM status_effects WHERE session_id = $1`

	result, err := r.db.Exec(query, sessionID)
	if err != nil {
		return 0, fmt.Errorf("failed to remove session effects: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	return int(rowsAffected), nil
}

// scanStatusEffect scanne une ligne de résultat en StatusEffect
func (r *EffectRepository) scanStatusEffect(scanner interface {
	Scan(dest ...interface{}) error
}) (*models.StatusEffect, error) {
	var effect models.StatusEffect
	var statModifiersJSON, periodicEffectJSON []byte
	var sessionID sql.NullString
	var sourceID, casterID sql.NullString
	var expiresAt sql.NullTime

	err := scanner.Scan(
		&effect.ID, &effect.CharacterID, &sessionID, &effect.Type, &effect.Source,
		&sourceID, &casterID, &effect.Name, &effect.Description,
		&statModifiersJSON, &periodicEffectJSON, &effect.AppliedAt, &expiresAt,
		&effect.Duration, &effect.Stacks, &effect.MaxStacks, &effect.IsDispellable,
		&effect.DispelType, &effect.Priority, &effect.IsActive, &effect.CreatedAt, &effect.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to scan status effect: %w", err)
	}

	// Convertir les champs nullable
	if sessionID.Valid {
		sessionUUID, _ := uuid.Parse(sessionID.String)
		effect.SessionID = &sessionUUID
	}
	if sourceID.Valid {
		sourceUUID, _ := uuid.Parse(sourceID.String)
		effect.SourceID = &sourceUUID
	}
	if casterID.Valid {
		casterUUID, _ := uuid.Parse(casterID.String)
		effect.CasterID = &casterUUID
	}
	if expiresAt.Valid {
		effect.ExpiresAt = expiresAt.Time
	}

	// Désérialiser JSON
	if len(statModifiersJSON) > 0 {
		if err := json.Unmarshal(statModifiersJSON, &effect.StatModifiers); err != nil {
			return nil, fmt.Errorf("failed to unmarshal stat modifiers: %w", err)
		}
	}

	if len(periodicEffectJSON) > 0 {
		if err := json.Unmarshal(periodicEffectJSON, &effect.PeriodicEffect); err != nil {
			return nil, fmt.Errorf("failed to unmarshal periodic effect: %w", err)
		}
	}

	return &effect, nil
}