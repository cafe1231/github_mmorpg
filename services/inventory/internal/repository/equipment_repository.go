package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"inventory/internal/models"
	"github.com/jmoiron/sqlx"
)

type equipmentRepository struct {
	db *sqlx.DB
}

func NewEquipmentRepository(db *sqlx.DB) EquipmentRepository {
	return &equipmentRepository{db: db}
}

// GetByCharacterID retrieves all equipment for a character
func (r *equipmentRepository) GetByCharacterID(ctx context.Context, characterID uuid.UUID) (*models.EquipmentSet, error) {
	query := `
		SELECT e.id, e.character_id, e.slot, e.item_id, e.equipped_at, e.created_at, e.updated_at,
			i.name, i.description, i.type, i.rarity, i.level, i.stats, i.requirements,
			i.metadata, i.weight, i.max_stack_size, i.tradeable, i.sellable, i.destroyable, 
			i.value, i.image_url, i.created_at as item_created_at, i.updated_at as item_updated_at
		FROM equipment e
		LEFT JOIN items i ON e.item_id = i.id
		WHERE e.character_id = $1
		ORDER BY e.slot
	`

	var equipmentItems []struct {
		models.Equipment
		Item *models.Item `db:"item"`
	}

	err := r.db.SelectContext(ctx, &equipmentItems, query, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get equipment: %w", err)
	}

	equipmentSet := &models.EquipmentSet{
		CharacterID: characterID,
		Equipment:   make(map[models.EquipmentSlot]*models.Equipment),
		TotalStats:  make(map[string]int),
		SetBonuses:  []models.SetBonus{},
	}

	for _, eq := range equipmentItems {
		eq.Equipment.Item = eq.Item
		equipmentSet.Equipment[eq.Equipment.Slot] = &eq.Equipment
	}

	// Calculate stats and set bonuses
	equipmentSet.CalculateTotalStats()
	equipmentSet.CalculateSetBonuses()

	return equipmentSet, nil
}

// EquipItem equips an item to a specific slot
func (r *equipmentRepository) EquipItem(ctx context.Context, characterID uuid.UUID, itemID uuid.UUID, slot models.EquipmentSlot) (*models.Equipment, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check if slot is already occupied
	var existingID uuid.UUID
	err = tx.GetContext(ctx, &existingID,
		"SELECT id FROM equipment WHERE character_id = $1 AND slot = $2",
		characterID, slot)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check existing equipment: %w", err)
	}

	equipmentID := uuid.New()

	if err == sql.ErrNoRows {
		// Insert new equipment
		_, err = tx.ExecContext(ctx,
			"INSERT INTO equipment (id, character_id, slot, item_id, equipped_at, created_at, updated_at) VALUES ($1, $2, $3, $4, NOW(), NOW(), NOW())",
			equipmentID, characterID, slot, itemID)
	} else {
		// Update existing equipment
		_, err = tx.ExecContext(ctx,
			"UPDATE equipment SET item_id = $4, equipped_at = NOW(), updated_at = NOW() WHERE character_id = $1 AND slot = $2",
			characterID, slot, itemID)
		equipmentID = existingID
	}

	if err != nil {
		return nil, fmt.Errorf("failed to equip item: %w", err)
	}

	// Get the equipped item details
	var equipment models.Equipment
	err = tx.GetContext(ctx, &equipment,
		"SELECT id, character_id, slot, item_id, equipped_at, created_at, updated_at FROM equipment WHERE id = $1",
		equipmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get equipped item: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &equipment, nil
}

// UnequipItem removes an item from a specific slot
func (r *equipmentRepository) UnequipItem(ctx context.Context, characterID uuid.UUID, slot models.EquipmentSlot) (*models.Equipment, error) {
	// Get current equipment
	var equipment models.Equipment
	err := r.db.GetContext(ctx, &equipment,
		"SELECT id, character_id, slot, item_id, equipped_at, created_at, updated_at FROM equipment WHERE character_id = $1 AND slot = $2",
		characterID, slot)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.NewNotFoundError("equipment", fmt.Sprintf("character %s slot %s", characterID, slot))
		}
		return nil, fmt.Errorf("failed to get equipment: %w", err)
	}

	// Remove item from slot
	_, err = r.db.ExecContext(ctx,
		"UPDATE equipment SET item_id = NULL, updated_at = NOW() WHERE character_id = $1 AND slot = $2",
		characterID, slot)
	if err != nil {
		return nil, fmt.Errorf("failed to unequip item: %w", err)
	}

	return &equipment, nil
}

// GetEquippedItem retrieves the item equipped in a specific slot
func (r *equipmentRepository) GetEquippedItem(ctx context.Context, characterID uuid.UUID, slot models.EquipmentSlot) (*models.Equipment, error) {
	query := `
		SELECT e.id, e.character_id, e.slot, e.item_id, e.equipped_at, e.created_at, e.updated_at,
			i.name, i.description, i.type, i.rarity, i.level, i.stats, i.requirements,
			i.metadata, i.weight, i.max_stack_size, i.tradeable, i.sellable, i.destroyable, 
			i.value, i.image_url, i.created_at as item_created_at, i.updated_at as item_updated_at
		FROM equipment e
		LEFT JOIN items i ON e.item_id = i.id
		WHERE e.character_id = $1 AND e.slot = $2 AND e.item_id IS NOT NULL
	`

	var equipment models.Equipment
	err := r.db.GetContext(ctx, &equipment, query, characterID, slot)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.NewNotFoundError("equipment", fmt.Sprintf("character %s slot %s", characterID, slot))
		}
		return nil, fmt.Errorf("failed to get equipped item: %w", err)
	}

	return &equipment, nil
}

// ListEquipped retrieves all equipped items for a character
func (r *equipmentRepository) ListEquipped(ctx context.Context, characterID uuid.UUID) ([]models.Equipment, error) {
	query := `
		SELECT e.id, e.character_id, e.slot, e.item_id, e.equipped_at, e.created_at, e.updated_at,
			i.name, i.description, i.type, i.rarity, i.level, i.stats, i.requirements,
			i.metadata, i.weight, i.max_stack_size, i.tradeable, i.sellable, i.destroyable, 
			i.value, i.image_url, i.created_at as item_created_at, i.updated_at as item_updated_at
		FROM equipment e
		JOIN items i ON e.item_id = i.id
		WHERE e.character_id = $1 AND e.item_id IS NOT NULL
		ORDER BY e.slot
	`

	var equipment []models.Equipment
	err := r.db.SelectContext(ctx, &equipment, query, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to list equipped items: %w", err)
	}

	return equipment, nil
}

// CalculateStats calculates total stats from all equipped items
func (r *equipmentRepository) CalculateStats(ctx context.Context, characterID uuid.UUID) (map[string]int, error) {
	equipmentSet, err := r.GetByCharacterID(ctx, characterID)
	if err != nil {
		return nil, err
	}

	equipmentSet.CalculateTotalStats()
	return equipmentSet.TotalStats, nil
}

// GetSetBonuses calculates set bonuses for equipped items
func (r *equipmentRepository) GetSetBonuses(ctx context.Context, characterID uuid.UUID) ([]models.SetBonus, error) {
	equipmentSet, err := r.GetByCharacterID(ctx, characterID)
	if err != nil {
		return nil, err
	}

	equipmentSet.CalculateSetBonuses()
	return equipmentSet.SetBonuses, nil
}
