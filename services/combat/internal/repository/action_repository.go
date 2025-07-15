package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"combat/internal/database"
	"combat/internal/models"
)

// ActionRepositoryInterface définit les méthodes du repository action
type ActionRepositoryInterface interface {
	// CRUD de base
	Create(action *models.CombatAction) error
	GetByID(id uuid.UUID) (*models.CombatAction, error)
	Update(action *models.CombatAction) error
	Delete(id uuid.UUID) error
	
	// Récupération par combat
	GetByCombat(combatID uuid.UUID) ([]*models.CombatAction, error)
	GetByCombatAndTurn(combatID uuid.UUID, turnNumber int) ([]*models.CombatAction, error)
	GetRecentActions(combatID uuid.UUID, limit int) ([]*models.CombatAction, error)
	
	// Récupération par acteur
	GetByActor(actorID uuid.UUID) ([]*models.CombatAction, error)
	GetByActorAndCombat(actorID, combatID uuid.UUID) ([]*models.CombatAction, error)
	
	// Statistiques et analyse
	GetActionCount(combatID uuid.UUID) (int, error)
	GetActionsPerSecond(actorID uuid.UUID, timeWindow time.Duration) (int, error)
	GetSuspiciousActions(timeWindow time.Duration) ([]*models.CombatAction, error)
	
	// Validation et anti-cheat
	GetRecentActionsByActor(actorID uuid.UUID, timeWindow time.Duration) ([]*models.CombatAction, error)
	MarkAsInvalid(actionID uuid.UUID, reason string) error
	GetInvalidActions(combatID uuid.UUID) ([]*models.CombatAction, error)
	
	GetActionStatistics(actorID uuid.UUID, timeWindow time.Duration) (*models.ActionStatistics, error)
	
}

// ActionRepository implémente l'interface ActionRepositoryInterface
type ActionRepository struct {
	db *database.DB
}

// NewActionRepository crée une nouvelle instance du repository action
func NewActionRepository(db *database.DB) ActionRepositoryInterface {
	return &ActionRepository{db: db}
}

// Create crée une nouvelle action de combat
func (r *ActionRepository) Create(action *models.CombatAction) error {
	query := `
		INSERT INTO combat_actions (
			id, combat_id, actor_id, target_id, action_type, skill_id, item_id,
			damage_dealt, healing_done, mana_used, is_critical, is_miss, is_blocked,
			turn_number, action_order, processing_time_ms,
			client_timestamp, server_timestamp, is_validated, validation_notes,
			created_at
		) VALUES (
			:id, :combat_id, :actor_id, :target_id, :action_type, :skill_id, :item_id,
			:damage_dealt, :healing_done, :mana_used, :is_critical, :is_miss, :is_blocked,
			:turn_number, :action_order, :processing_time_ms,
			:client_timestamp, :server_timestamp, :is_validated, :validation_notes,
			:created_at
		)`

	_, err := r.db.NamedExec(query, action)
	if err != nil {
		return fmt.Errorf("failed to create action: %w", err)
	}

	return nil
}

// GetByID récupère une action par son ID
func (r *ActionRepository) GetByID(id uuid.UUID) (*models.CombatAction, error) {
	var action models.CombatAction

	query := `
		SELECT id, combat_id, actor_id, target_id, action_type, skill_id, item_id,
		       damage_dealt, healing_done, mana_used, is_critical, is_miss, is_blocked,
		       turn_number, action_order, processing_time_ms,
		       client_timestamp, server_timestamp, is_validated, validation_notes,
		       created_at
		FROM combat_actions 
		WHERE id = $1`

	err := r.db.Get(&action, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("action not found")
		}
		return nil, fmt.Errorf("failed to get action: %w", err)
	}

	return &action, nil
}

// Update met à jour une action
func (r *ActionRepository) Update(action *models.CombatAction) error {
	query := `
		UPDATE combat_actions SET
			damage_dealt = :damage_dealt,
			healing_done = :healing_done,
			mana_used = :mana_used,
			is_critical = :is_critical,
			is_miss = :is_miss,
			is_blocked = :is_blocked,
			processing_time_ms = :processing_time_ms,
			is_validated = :is_validated,
			validation_notes = :validation_notes
		WHERE id = :id`

	result, err := r.db.NamedExec(query, action)
	if err != nil {
		return fmt.Errorf("failed to update action: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("action not found")
	}

	return nil
}

// Delete supprime une action
func (r *ActionRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM combat_actions WHERE id = $1`
	
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete action: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("action not found")
	}

	return nil
}

// GetByCombat récupère toutes les actions d'un combat
func (r *ActionRepository) GetByCombat(combatID uuid.UUID) ([]*models.CombatAction, error) {
	var actions []*models.CombatAction

	query := `
		SELECT id, combat_id, actor_id, target_id, action_type, skill_id, item_id,
		       damage_dealt, healing_done, mana_used, is_critical, is_miss, is_blocked,
		       turn_number, action_order, processing_time_ms,
		       client_timestamp, server_timestamp, is_validated, validation_notes,
		       created_at
		FROM combat_actions 
		WHERE combat_id = $1 
		ORDER BY turn_number, action_order, created_at`

	err := r.db.Select(&actions, query, combatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get actions by combat: %w", err)
	}

	return actions, nil
}

// GetByCombatAndTurn récupère les actions d'un combat pour un tour spécifique
func (r *ActionRepository) GetByCombatAndTurn(combatID uuid.UUID, turnNumber int) ([]*models.CombatAction, error) {
	var actions []*models.CombatAction

	query := `
		SELECT id, combat_id, actor_id, target_id, action_type, skill_id, item_id,
		       damage_dealt, healing_done, mana_used, is_critical, is_miss, is_blocked,
		       turn_number, action_order, processing_time_ms,
		       client_timestamp, server_timestamp, is_validated, validation_notes,
		       created_at
		FROM combat_actions 
		WHERE combat_id = $1 AND turn_number = $2 
		ORDER BY action_order, created_at`

	err := r.db.Select(&actions, query, combatID, turnNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get actions by combat and turn: %w", err)
	}

	return actions, nil
}

// GetRecentActions récupère les actions récentes d'un combat
func (r *ActionRepository) GetRecentActions(combatID uuid.UUID, limit int) ([]*models.CombatAction, error) {
	var actions []*models.CombatAction

	query := `
		SELECT id, combat_id, actor_id, target_id, action_type, skill_id, item_id,
		       damage_dealt, healing_done, mana_used, is_critical, is_miss, is_blocked,
		       turn_number, action_order, processing_time_ms,
		       client_timestamp, server_timestamp, is_validated, validation_notes,
		       created_at
		FROM combat_actions 
		WHERE combat_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2`

	err := r.db.Select(&actions, query, combatID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent actions: %w", err)
	}

	return actions, nil
}



// GetByActor récupère toutes les actions d'un acteur
func (r *ActionRepository) GetByActor(actorID uuid.UUID) ([]*models.CombatAction, error) {
	var actions []*models.CombatAction

	query := `
		SELECT id, combat_id, actor_id, target_id, action_type, skill_id, item_id,
		       damage_dealt, healing_done, mana_used, is_critical, is_miss, is_blocked,
		       turn_number, action_order, processing_time_ms,
		       client_timestamp, server_timestamp, is_validated, validation_notes,
		       created_at
		FROM combat_actions 
		WHERE actor_id = $1 
		ORDER BY created_at DESC 
		LIMIT 100`

	err := r.db.Select(&actions, query, actorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get actions by actor: %w", err)
	}

	return actions, nil
}

// GetByActorAndCombat récupère les actions d'un acteur dans un combat spécifique
func (r *ActionRepository) GetByActorAndCombat(actorID, combatID uuid.UUID) ([]*models.CombatAction, error) {
	var actions []*models.CombatAction

	query := `
		SELECT id, combat_id, actor_id, target_id, action_type, skill_id, item_id,
		       damage_dealt, healing_done, mana_used, is_critical, is_miss, is_blocked,
		       turn_number, action_order, processing_time_ms,
		       client_timestamp, server_timestamp, is_validated, validation_notes,
		       created_at
		FROM combat_actions 
		WHERE actor_id = $1 AND combat_id = $2 
		ORDER BY turn_number, action_order, created_at`

	err := r.db.Select(&actions, query, actorID, combatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get actions by actor and combat: %w", err)
	}

	return actions, nil
}

// GetActionCount retourne le nombre d'actions dans un combat
func (r *ActionRepository) GetActionCount(combatID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM combat_actions WHERE combat_id = $1`
	
	err := r.db.Get(&count, query, combatID)
	if err != nil {
		return 0, fmt.Errorf("failed to get action count: %w", err)
	}

	return count, nil
}

// GetActionsPerSecond calcule le nombre d'actions par seconde pour un acteur
func (r *ActionRepository) GetActionsPerSecond(actorID uuid.UUID, timeWindow time.Duration) (int, error) {
	var count int
	since := time.Now().Add(-timeWindow)
	
	query := `
		SELECT COUNT(*) 
		FROM combat_actions 
		WHERE actor_id = $1 AND server_timestamp >= $2`
	
	err := r.db.Get(&count, query, actorID, since)
	if err != nil {
		return 0, fmt.Errorf("failed to get actions per second: %w", err)
	}

	return count, nil
}

// GetSuspiciousActions récupère les actions suspectes récentes
func (r *ActionRepository) GetSuspiciousActions(timeWindow time.Duration) ([]*models.CombatAction, error) {
	var actions []*models.CombatAction
	since := time.Now().Add(-timeWindow)

	query := `
		SELECT id, combat_id, actor_id, target_id, action_type, skill_id, item_id,
		       damage_dealt, healing_done, mana_used, is_critical, is_miss, is_blocked,
		       turn_number, action_order, processing_time_ms,
		       client_timestamp, server_timestamp, is_validated, validation_notes,
		       created_at
		FROM combat_actions 
		WHERE server_timestamp >= $1 
		AND (
			is_validated = false 
			OR validation_notes IS NOT NULL 
			OR ABS(EXTRACT(EPOCH FROM (server_timestamp - client_timestamp))) > 5
			OR damage_dealt > 9999
		)
		ORDER BY server_timestamp DESC 
		LIMIT 100`

	err := r.db.Select(&actions, query, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get suspicious actions: %w", err)
	}

	return actions, nil
}

// GetRecentActionsByActor récupère les actions récentes d'un acteur
func (r *ActionRepository) GetRecentActionsByActor(actorID uuid.UUID, timeWindow time.Duration) ([]*models.CombatAction, error) {
	var actions []*models.CombatAction
	since := time.Now().Add(-timeWindow)

	query := `
		SELECT id, combat_id, actor_id, target_id, action_type, skill_id, item_id,
		       damage_dealt, healing_done, mana_used, is_critical, is_miss, is_blocked,
		       turn_number, action_order, processing_time_ms,
		       client_timestamp, server_timestamp, is_validated, validation_notes,
		       created_at
		FROM combat_actions 
		WHERE actor_id = $1 AND server_timestamp >= $2 
		ORDER BY server_timestamp DESC`

	err := r.db.Select(&actions, query, actorID, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent actions by actor: %w", err)
	}

	return actions, nil
}

// MarkAsInvalid marque une action comme invalide
func (r *ActionRepository) MarkAsInvalid(actionID uuid.UUID, reason string) error {
	query := `
		UPDATE combat_actions 
		SET is_validated = false, validation_notes = $2 
		WHERE id = $1`

	result, err := r.db.Exec(query, actionID, reason)
	if err != nil {
		return fmt.Errorf("failed to mark action as invalid: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("action not found")
	}

	return nil
}

// GetInvalidActions récupère les actions invalides d'un combat
func (r *ActionRepository) GetInvalidActions(combatID uuid.UUID) ([]*models.CombatAction, error) {
	var actions []*models.CombatAction

	query := `
		SELECT id, combat_id, actor_id, target_id, action_type, skill_id, item_id,
		       damage_dealt, healing_done, mana_used, is_critical, is_miss, is_blocked,
		       turn_number, action_order, processing_time_ms,
		       client_timestamp, server_timestamp, is_validated, validation_notes,
		       created_at
		FROM combat_actions 
		WHERE combat_id = $1 AND is_validated = false 
		ORDER BY created_at DESC`

	err := r.db.Select(&actions, query, combatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invalid actions: %w", err)
	}

	return actions, nil
}

func (r *ActionRepository) GetActionStatistics(actorID uuid.UUID, timeWindow time.Duration) (*models.ActionStatistics, error) {
	since := time.Now().Add(-timeWindow)
	
	query := `
		SELECT 
			COUNT(*) as total_actions,
			COUNT(CASE WHEN is_critical = true THEN 1 END) as critical_hits,
			COUNT(CASE WHEN is_miss = true THEN 1 END) as misses,
			COUNT(CASE WHEN is_blocked = true THEN 1 END) as blocks,
			AVG(damage_dealt) as avg_damage,
			MAX(damage_dealt) as max_damage,
			AVG(healing_done) as avg_healing,
			AVG(processing_time_ms) as avg_processing_time
		FROM combat_actions 
		WHERE actor_id = $1 AND server_timestamp >= $2`

	var stats models.ActionStatistics // <- CHANGEMENT ICI
	err := r.db.Get(&stats, query, actorID, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get action statistics: %w", err)
	}

	// Calculer les taux
	if stats.TotalActions > 0 {
		stats.CriticalRate = float64(stats.CriticalHits) / float64(stats.TotalActions) * 100
		stats.MissRate = float64(stats.Misses) / float64(stats.TotalActions) * 100
		stats.BlockRate = float64(stats.Blocks) / float64(stats.TotalActions) * 100
		stats.AccuracyRate = 100 - stats.MissRate
	}

	return &stats, nil
}



// GetActionsByType récupère les actions par type pour un combat
func (r *ActionRepository) GetActionsByType(combatID uuid.UUID, actionType models.ActionType) ([]*models.CombatAction, error) {
	var actions []*models.CombatAction

	query := `
		SELECT id, combat_id, actor_id, target_id, action_type, skill_id, item_id,
		       damage_dealt, healing_done, mana_used, is_critical, is_miss, is_blocked,
		       turn_number, action_order, processing_time_ms,
		       client_timestamp, server_timestamp, is_validated, validation_notes,
		       created_at
		FROM combat_actions 
		WHERE combat_id = $1 AND action_type = $2 
		ORDER BY turn_number, action_order`

	err := r.db.Select(&actions, query, combatID, actionType)
	if err != nil {
		return nil, fmt.Errorf("failed to get actions by type: %w", err)
	}

	return actions, nil
}

// CleanupOldActions supprime les anciennes actions pour libérer de l'espace
func (r *ActionRepository) CleanupOldActions(olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	
	query := `
		DELETE FROM combat_actions 
		WHERE created_at < $1 
		AND combat_id IN (
			SELECT id FROM combat_instances 
			WHERE status IN ('finished', 'cancelled') 
			AND ended_at < $1
		)`

	result, err := r.db.Exec(query, cutoff)
	if err != nil {
		return fmt.Errorf("failed to cleanup old actions: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected > 0 {
		fmt.Printf("Cleaned up %d old actions\n", rowsAffected)
	}

	return nil
}