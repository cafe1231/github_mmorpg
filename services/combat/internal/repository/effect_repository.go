package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"combat/internal/database"
	"combat/internal/models"
)

// EffectRepositoryInterface définit les méthodes du repository effect
type EffectRepositoryInterface interface {
	// CRUD de base
	Create(effect *models.CombatEffect) error
	GetByID(id uuid.UUID) (*models.CombatEffect, error)
	Update(effect *models.CombatEffect) error
	Delete(id uuid.UUID) error
	
	// Récupération par cible et combat
	GetActiveByTarget(targetID uuid.UUID) ([]*models.CombatEffect, error)
	GetActiveByCombat(combatID uuid.UUID) ([]*models.CombatEffect, error)
	GetByTargetAndType(targetID uuid.UUID, effectType models.EffectType) ([]*models.CombatEffect, error)
	
	// Gestion des effets actifs
	GetExpiredEffects() ([]*models.CombatEffect, error)
	DeactivateEffect(effectID uuid.UUID) error
	CleanupExpiredEffects(olderThan time.Duration) error
	
	// Recherche avancée
	GetEffectsByCaster(casterID uuid.UUID) ([]*models.CombatEffect, error)
	GetEffectsByName(effectName string) ([]*models.CombatEffect, error)
	GetStackableEffects(targetID uuid.UUID, effectName string) ([]*models.CombatEffect, error)
	
	// Statistiques
	GetEffectCount(combatID uuid.UUID) (int, error)
	GetEffectDuration(effectID uuid.UUID) (time.Duration, error)
}

// EffectRepository implémente l'interface EffectRepositoryInterface
type EffectRepository struct {
	db *database.DB
}

// NewEffectRepository crée une nouvelle instance du repository effect
func NewEffectRepository(db *database.DB) EffectRepositoryInterface {
	return &EffectRepository{db: db}
}

// Create crée un nouvel effet de combat
func (r *EffectRepository) Create(effect *models.CombatEffect) error {
	query := `
		INSERT INTO combat_effects (
			id, combat_id, target_id, caster_id, effect_type, effect_name, effect_description,
			stat_affected, modifier_value, modifier_type, duration_turns, remaining_turns,
			max_stacks, current_stacks, is_active, is_dispellable,
			applied_at, expires_at, created_at, updated_at
		) VALUES (
			:id, :combat_id, :target_id, :caster_id, :effect_type, :effect_name, :effect_description,
			:stat_affected, :modifier_value, :modifier_type, :duration_turns, :remaining_turns,
			:max_stacks, :current_stacks, :is_active, :is_dispellable,
			:applied_at, :expires_at, :created_at, :updated_at
		)`

	_, err := r.db.NamedExec(query, effect)
	if err != nil {
		return fmt.Errorf("failed to create effect: %w", err)
	}

	return nil
}

// GetByID récupère un effet par son ID
func (r *EffectRepository) GetByID(id uuid.UUID) (*models.CombatEffect, error) {
	var effect models.CombatEffect

	query := `
		SELECT id, combat_id, target_id, caster_id, effect_type, effect_name, effect_description,
		       stat_affected, modifier_value, modifier_type, duration_turns, remaining_turns,
		       max_stacks, current_stacks, is_active, is_dispellable,
		       applied_at, expires_at, created_at, updated_at
		FROM combat_effects 
		WHERE id = $1`

	err := r.db.Get(&effect, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("effect not found")
		}
		return nil, fmt.Errorf("failed to get effect: %w", err)
	}

	return &effect, nil
}

// Update met à jour un effet
func (r *EffectRepository) Update(effect *models.CombatEffect) error {
	effect.UpdatedAt = time.Now()

	query := `
		UPDATE combat_effects SET
			remaining_turns = :remaining_turns,
			current_stacks = :current_stacks,
			is_active = :is_active,
			modifier_value = :modifier_value,
			expires_at = :expires_at,
			updated_at = :updated_at
		WHERE id = :id`

	result, err := r.db.NamedExec(query, effect)
	if err != nil {
		return fmt.Errorf("failed to update effect: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("effect not found")
	}

	return nil
}

// Delete supprime un effet
func (r *EffectRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM combat_effects WHERE id = $1`
	
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete effect: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("effect not found")
	}

	return nil
}

// GetActiveByTarget récupère tous les effets actifs d'une cible
func (r *EffectRepository) GetActiveByTarget(targetID uuid.UUID) ([]*models.CombatEffect, error) {
	var effects []*models.CombatEffect

	query := `
		SELECT id, combat_id, target_id, caster_id, effect_type, effect_name, effect_description,
		       stat_affected, modifier_value, modifier_type, duration_turns, remaining_turns,
		       max_stacks, current_stacks, is_active, is_dispellable,
		       applied_at, expires_at, created_at, updated_at
		FROM combat_effects 
		WHERE target_id = $1 AND is_active = true 
		ORDER BY applied_at DESC`

	err := r.db.Select(&effects, query, targetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active effects by target: %w", err)
	}

	return effects, nil
}

// GetActiveByCombat récupère tous les effets actifs d'un combat
func (r *EffectRepository) GetActiveByCombat(combatID uuid.UUID) ([]*models.CombatEffect, error) {
	var effects []*models.CombatEffect

	query := `
		SELECT id, combat_id, target_id, caster_id, effect_type, effect_name, effect_description,
		       stat_affected, modifier_value, modifier_type, duration_turns, remaining_turns,
		       max_stacks, current_stacks, is_active, is_dispellable,
		       applied_at, expires_at, created_at, updated_at
		FROM combat_effects 
		WHERE combat_id = $1 AND is_active = true 
		ORDER BY applied_at DESC`

	err := r.db.Select(&effects, query, combatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active effects by combat: %w", err)
	}

	return effects, nil
}

// GetByTargetAndType récupère les effets d'un type spécifique pour une cible
func (r *EffectRepository) GetByTargetAndType(targetID uuid.UUID, effectType models.EffectType) ([]*models.CombatEffect, error) {
	var effects []*models.CombatEffect

	query := `
		SELECT id, combat_id, target_id, caster_id, effect_type, effect_name, effect_description,
		       stat_affected, modifier_value, modifier_type, duration_turns, remaining_turns,
		       max_stacks, current_stacks, is_active, is_dispellable,
		       applied_at, expires_at, created_at, updated_at
		FROM combat_effects 
		WHERE target_id = $1 AND effect_type = $2 AND is_active = true 
		ORDER BY applied_at DESC`

	err := r.db.Select(&effects, query, targetID, effectType)
	if err != nil {
		return nil, fmt.Errorf("failed to get effects by target and type: %w", err)
	}

	return effects, nil
}

// GetExpiredEffects récupère tous les effets expirés
func (r *EffectRepository) GetExpiredEffects() ([]*models.CombatEffect, error) {
	var effects []*models.CombatEffect

	query := `
		SELECT id, combat_id, target_id, caster_id, effect_type, effect_name, effect_description,
		       stat_affected, modifier_value, modifier_type, duration_turns, remaining_turns,
		       max_stacks, current_stacks, is_active, is_dispellable,
		       applied_at, expires_at, created_at, updated_at
		FROM combat_effects 
		WHERE (remaining_turns <= 0 OR expires_at < CURRENT_TIMESTAMP) AND is_active = true`

	err := r.db.Select(&effects, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get expired effects: %w", err)
	}

	return effects, nil
}

// DeactivateEffect désactive un effet
func (r *EffectRepository) DeactivateEffect(effectID uuid.UUID) error {
	query := `
		UPDATE combat_effects 
		SET is_active = false, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $1`

	result, err := r.db.Exec(query, effectID)
	if err != nil {
		return fmt.Errorf("failed to deactivate effect: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("effect not found")
	}

	return nil
}

// CleanupExpiredEffects nettoie les effets expirés anciens
func (r *EffectRepository) CleanupExpiredEffects(olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	
	query := `
		DELETE FROM combat_effects 
		WHERE is_active = false AND updated_at < $1`

	result, err := r.db.Exec(query, cutoff)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired effects: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected > 0 {
		fmt.Printf("Cleaned up %d expired effects\n", rowsAffected)
	}

	return nil
}

// GetEffectsByCaster récupère tous les effets lancés par un caster
func (r *EffectRepository) GetEffectsByCaster(casterID uuid.UUID) ([]*models.CombatEffect, error) {
	var effects []*models.CombatEffect

	query := `
		SELECT id, combat_id, target_id, caster_id, effect_type, effect_name, effect_description,
		       stat_affected, modifier_value, modifier_type, duration_turns, remaining_turns,
		       max_stacks, current_stacks, is_active, is_dispellable,
		       applied_at, expires_at, created_at, updated_at
		FROM combat_effects 
		WHERE caster_id = $1 AND is_active = true 
		ORDER BY applied_at DESC`

	err := r.db.Select(&effects, query, casterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get effects by caster: %w", err)
	}

	return effects, nil
}

// GetEffectsByName récupère tous les effets avec un nom spécifique
func (r *EffectRepository) GetEffectsByName(effectName string) ([]*models.CombatEffect, error) {
	var effects []*models.CombatEffect

	query := `
		SELECT id, combat_id, target_id, caster_id, effect_type, effect_name, effect_description,
		       stat_affected, modifier_value, modifier_type, duration_turns, remaining_turns,
		       max_stacks, current_stacks, is_active, is_dispellable,
		       applied_at, expires_at, created_at, updated_at
		FROM combat_effects 
		WHERE effect_name = $1 AND is_active = true 
		ORDER BY applied_at DESC`

	err := r.db.Select(&effects, query, effectName)
	if err != nil {
		return nil, fmt.Errorf("failed to get effects by name: %w", err)
	}

	return effects, nil
}

// GetStackableEffects récupère les effets empilables du même nom pour une cible
func (r *EffectRepository) GetStackableEffects(targetID uuid.UUID, effectName string) ([]*models.CombatEffect, error) {
	var effects []*models.CombatEffect

	query := `
		SELECT id, combat_id, target_id, caster_id, effect_type, effect_name, effect_description,
		       stat_affected, modifier_value, modifier_type, duration_turns, remaining_turns,
		       max_stacks, current_stacks, is_active, is_dispellable,
		       applied_at, expires_at, created_at, updated_at
		FROM combat_effects 
		WHERE target_id = $1 AND effect_name = $2 AND is_active = true AND max_stacks > 1
		ORDER BY applied_at DESC`

	err := r.db.Select(&effects, query, targetID, effectName)
	if err != nil {
		return nil, fmt.Errorf("failed to get stackable effects: %w", err)
	}

	return effects, nil
}

// GetEffectCount retourne le nombre d'effets actifs dans un combat
func (r *EffectRepository) GetEffectCount(combatID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM combat_effects WHERE combat_id = $1 AND is_active = true`
	
	err := r.db.Get(&count, query, combatID)
	if err != nil {
		return 0, fmt.Errorf("failed to get effect count: %w", err)
	}

	return count, nil
}

// GetEffectDuration calcule la durée restante d'un effet
func (r *EffectRepository) GetEffectDuration(effectID uuid.UUID) (time.Duration, error) {
	var effect models.CombatEffect
	
	query := `
		SELECT remaining_turns, expires_at 
		FROM combat_effects 
		WHERE id = $1`

	err := r.db.Get(&effect, query, effectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("effect not found")
		}
		return 0, fmt.Errorf("failed to get effect duration: %w", err)
	}

	// Si il y a une expiration absolue
	if effect.ExpiresAt != nil {
		return time.Until(*effect.ExpiresAt), nil
	}

	// Sinon, estimer basé sur les tours restants (30 secondes par tour)
	estimatedDuration := time.Duration(effect.RemainingTurns) * 30 * time.Second
	return estimatedDuration, nil
}

// GetEffectsByTimeRange récupère les effets dans une plage de temps
func (r *EffectRepository) GetEffectsByTimeRange(start, end time.Time) ([]*models.CombatEffect, error) {
	var effects []*models.CombatEffect

	query := `
		SELECT id, combat_id, target_id, caster_id, effect_type, effect_name, effect_description,
		       stat_affected, modifier_value, modifier_type, duration_turns, remaining_turns,
		       max_stacks, current_stacks, is_active, is_dispellable,
		       applied_at, expires_at, created_at, updated_at
		FROM combat_effects 
		WHERE applied_at BETWEEN $1 AND $2 
		ORDER BY applied_at DESC`

	err := r.db.Select(&effects, query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get effects by time range: %w", err)
	}

	return effects, nil
}

// GetEffectStatistics récupère des statistiques sur les effets
func (r *EffectRepository) GetEffectStatistics(combatID uuid.UUID) (*EffectStatistics, error) {
	query := `
		SELECT 
			COUNT(*) as total_effects,
			COUNT(CASE WHEN effect_type = 'buff' THEN 1 END) as buffs,
			COUNT(CASE WHEN effect_type = 'debuff' THEN 1 END) as debuffs,
			COUNT(CASE WHEN effect_type = 'dot' THEN 1 END) as dots,
			COUNT(CASE WHEN effect_type = 'hot' THEN 1 END) as hots,
			COUNT(CASE WHEN is_active = true THEN 1 END) as active_effects,
			AVG(duration_turns) as avg_duration,
			MAX(current_stacks) as max_stacks_seen
		FROM combat_effects 
		WHERE combat_id = $1`

	var stats EffectStatistics
	err := r.db.Get(&stats, query, combatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get effect statistics: %w", err)
	}

	return &stats, nil
}

// EffectStatistics représente les statistiques d'effets
type EffectStatistics struct {
	TotalEffects   int     `json:"total_effects" db:"total_effects"`
	Buffs          int     `json:"buffs" db:"buffs"`
	Debuffs        int     `json:"debuffs" db:"debuffs"`
	DoTs           int     `json:"dots" db:"dots"`
	HoTs           int     `json:"hots" db:"hots"`
	ActiveEffects  int     `json:"active_effects" db:"active_effects"`
	AvgDuration    float64 `json:"avg_duration" db:"avg_duration"`
	MaxStacksSeen  int     `json:"max_stacks_seen" db:"max_stacks_seen"`
}

// BatchUpdate met à jour plusieurs effets en une seule transaction
func (r *EffectRepository) BatchUpdate(effects []*models.CombatEffect) error {
	if len(effects) == 0 {
		return nil
	}

	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		UPDATE combat_effects SET
			remaining_turns = :remaining_turns,
			current_stacks = :current_stacks,
			is_active = :is_active,
			modifier_value = :modifier_value,
			expires_at = :expires_at,
			updated_at = :updated_at
		WHERE id = :id`

	for _, effect := range effects {
		effect.UpdatedAt = time.Now()
		_, err := tx.NamedExec(query, effect)
		if err != nil {
			return fmt.Errorf("failed to update effect %s: %w", effect.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit batch update: %w", err)
	}

	return nil
}

// BatchDelete supprime plusieurs effets en une seule transaction
func (r *EffectRepository) BatchDelete(effectIDs []uuid.UUID) error {
	if len(effectIDs) == 0 {
		return nil
	}

	// Construire la requête avec les placeholders
	query := `DELETE FROM combat_effects WHERE id = ANY($1)`
	
	// Convertir les UUIDs en string array pour PostgreSQL
	uuidStrings := make([]string, len(effectIDs))
	for i, id := range effectIDs {
		uuidStrings[i] = id.String()
	}

	result, err := r.db.Exec(query, pq.Array(uuidStrings))
	if err != nil {
		return fmt.Errorf("failed to batch delete effects: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if int(rowsAffected) != len(effectIDs) {
		return fmt.Errorf("expected to delete %d effects, but deleted %d", len(effectIDs), rowsAffected)
	}

	return nil
}