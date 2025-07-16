package repository

import (
	"context"
	"database/sql"
	"fmt"

	"inventory/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type inventoryRepository struct {
	db *sqlx.DB
}

func NewInventoryRepository(db *sqlx.DB) InventoryRepository {
	return &inventoryRepository{db: db}
}

// Create creates a new inventory
func (r *inventoryRepository) Create(ctx context.Context, inventory *models.Inventory) error {
	query := `
		INSERT INTO inventories (character_id, slots, max_weight, created_at, updated_at)
		VALUES (:character_id, :slots, :max_weight, :created_at, :updated_at)
	`

	_, err := r.db.NamedExecContext(ctx, query, inventory)
	if err != nil {
		return fmt.Errorf("failed to create inventory: %w", err)
	}

	return nil
}

// GetByCharacterID retrieves an inventory by character ID
func (r *inventoryRepository) GetByCharacterID(ctx context.Context, characterID uuid.UUID) (*models.Inventory, error) {
	// Get inventory
	inventoryQuery := `
		SELECT character_id, slots, max_weight, created_at, updated_at
		FROM inventories 
		WHERE character_id = $1
	`

	var inventory models.Inventory
	err := r.db.GetContext(ctx, &inventory, inventoryQuery, characterID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.NewNotFoundError("inventory", characterID.String())
		}
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}

	// Get inventory items
	itemsQuery := `
		SELECT ii.id, ii.character_id, ii.item_id, ii.quantity, ii.slot, ii.created_at, ii.updated_at,
			i.name, i.description, i.type, i.rarity, i.level, i.stats, i.requirements,
			i.metadata, i.weight, i.max_stack_size, i.tradeable, i.sellable, i.destroyable, 
			i.value, i.image_url, i.created_at as item_created_at, i.updated_at as item_updated_at
		FROM inventory_items ii
		JOIN items i ON ii.item_id = i.id
		WHERE ii.character_id = $1
		ORDER BY ii.slot ASC
	`

	var items []models.InventoryItem
	err = r.db.SelectContext(ctx, &items, itemsQuery, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory items: %w", err)
	}

	inventory.Slots = items
	return &inventory, nil
}

// AddItem adds an item to inventory
func (r *inventoryRepository) AddItem(ctx context.Context, characterID uuid.UUID, item *models.InventoryItem) error {
	query := `
		INSERT INTO inventory_items (id, character_id, item_id, quantity, slot, created_at, updated_at)
		VALUES (:id, :character_id, :item_id, :quantity, :slot, :created_at, :updated_at)
	`

	_, err := r.db.NamedExecContext(ctx, query, item)
	if err != nil {
		return fmt.Errorf("failed to add item to inventory: %w", err)
	}

	return nil
}

// Update updates an inventory
func (r *inventoryRepository) Update(ctx context.Context, inventory *models.Inventory) error {
	query := `
		UPDATE inventories SET
			slots = :slots,
			max_weight = :max_weight,
			updated_at = :updated_at
		WHERE character_id = :character_id
	`

	result, err := r.db.NamedExecContext(ctx, query, inventory)
	if err != nil {
		return fmt.Errorf("failed to update inventory: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return models.NewNotFoundError("inventory", inventory.CharacterID.String())
	}

	return nil
}

// Delete deletes an inventory
func (r *inventoryRepository) Delete(ctx context.Context, characterID uuid.UUID) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete inventory items first
	_, err = tx.ExecContext(ctx, "DELETE FROM inventory_items WHERE character_id = $1", characterID)
	if err != nil {
		return fmt.Errorf("failed to delete inventory items: %w", err)
	}

	// Delete inventory
	result, err := tx.ExecContext(ctx, "DELETE FROM inventories WHERE character_id = $1", characterID)
	if err != nil {
		return fmt.Errorf("failed to delete inventory: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return models.NewNotFoundError("inventory", characterID.String())
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// RemoveItem removes an item from inventory
func (r *inventoryRepository) RemoveItem(ctx context.Context, characterID uuid.UUID, itemID uuid.UUID, quantity int) error {
	// Get current item
	var currentQuantity int
	err := r.db.GetContext(ctx, &currentQuantity,
		"SELECT quantity FROM inventory_items WHERE character_id = $1 AND item_id = $2",
		characterID, itemID)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.NewNotFoundError("inventory item", itemID.String())
		}
		return fmt.Errorf("failed to get current quantity: %w", err)
	}

	if currentQuantity < quantity {
		return models.NewInsufficientResourcesError("item quantity", quantity, currentQuantity)
	}

	if currentQuantity == quantity {
		// Remove item completely
		_, err = r.db.ExecContext(ctx,
			"DELETE FROM inventory_items WHERE character_id = $1 AND item_id = $2",
			characterID, itemID)
	} else {
		// Update quantity
		_, err = r.db.ExecContext(ctx,
			"UPDATE inventory_items SET quantity = quantity - $3, updated_at = NOW() WHERE character_id = $1 AND item_id = $2",
			characterID, itemID, quantity)
	}

	if err != nil {
		return fmt.Errorf("failed to remove item: %w", err)
	}

	return nil
}

// UpdateItem updates an inventory item
func (r *inventoryRepository) UpdateItem(ctx context.Context, characterID uuid.UUID, item *models.InventoryItem) error {
	query := `
		UPDATE inventory_items SET
			quantity = :quantity,
			slot = :slot,
			updated_at = :updated_at
		WHERE character_id = :character_id AND item_id = :item_id
	`

	result, err := r.db.NamedExecContext(ctx, query, item)
	if err != nil {
		return fmt.Errorf("failed to update inventory item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return models.NewNotFoundError("inventory item", item.ItemID.String())
	}

	return nil
}

// GetItem retrieves a specific item from inventory
func (r *inventoryRepository) GetItem(ctx context.Context, characterID uuid.UUID, itemID uuid.UUID) (*models.InventoryItem, error) {
	query := `
		SELECT ii.id, ii.character_id, ii.item_id, ii.quantity, ii.slot, ii.created_at, ii.updated_at,
			i.name, i.description, i.type, i.rarity, i.level, i.stats, i.requirements,
			i.metadata, i.weight, i.max_stack_size, i.tradeable, i.sellable, i.destroyable, 
			i.value, i.image_url, i.created_at as item_created_at, i.updated_at as item_updated_at
		FROM inventory_items ii
		JOIN items i ON ii.item_id = i.id
		WHERE ii.character_id = $1 AND ii.item_id = $2
	`

	var item models.InventoryItem
	err := r.db.GetContext(ctx, &item, query, characterID, itemID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.NewNotFoundError("inventory item", itemID.String())
		}
		return nil, fmt.Errorf("failed to get inventory item: %w", err)
	}

	return &item, nil
}

// MoveItem moves an item from one slot to another
func (r *inventoryRepository) MoveItem(ctx context.Context, characterID uuid.UUID, fromSlot, toSlot int) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update item in from slot
	_, err = tx.ExecContext(ctx,
		"UPDATE inventory_items SET slot = $3, updated_at = NOW() WHERE character_id = $1 AND slot = $2",
		characterID, fromSlot, toSlot)
	if err != nil {
		return fmt.Errorf("failed to move item: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// SwapItems swaps items between two slots
func (r *inventoryRepository) SwapItems(ctx context.Context, characterID uuid.UUID, slot1, slot2 int) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Use a temporary slot to avoid conflicts
	tempSlot := -1

	// Move item from slot1 to temp
	_, err = tx.ExecContext(ctx,
		"UPDATE inventory_items SET slot = $3, updated_at = NOW() WHERE character_id = $1 AND slot = $2",
		characterID, slot1, tempSlot)
	if err != nil {
		return fmt.Errorf("failed to move item from slot1 to temp: %w", err)
	}

	// Move item from slot2 to slot1
	_, err = tx.ExecContext(ctx,
		"UPDATE inventory_items SET slot = $3, updated_at = NOW() WHERE character_id = $1 AND slot = $2",
		characterID, slot2, slot1)
	if err != nil {
		return fmt.Errorf("failed to move item from slot2 to slot1: %w", err)
	}

	// Move item from temp to slot2
	_, err = tx.ExecContext(ctx,
		"UPDATE inventory_items SET slot = $3, updated_at = NOW() WHERE character_id = $1 AND slot = $2",
		characterID, tempSlot, slot2)
	if err != nil {
		return fmt.Errorf("failed to move item from temp to slot2: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// SplitStack splits a stack of items into two slots
func (r *inventoryRepository) SplitStack(ctx context.Context, characterID uuid.UUID, fromSlot, toSlot int, quantity int) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get item from source slot
	var itemID uuid.UUID
	var currentQuantity int
	err = tx.GetContext(ctx, &struct {
		ItemID   uuid.UUID `db:"item_id"`
		Quantity int       `db:"quantity"`
	}{ItemID: itemID, Quantity: currentQuantity},
		"SELECT item_id, quantity FROM inventory_items WHERE character_id = $1 AND slot = $2",
		characterID, fromSlot)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.NewNotFoundError("inventory item", fmt.Sprintf("slot %d", fromSlot))
		}
		return fmt.Errorf("failed to get source item: %w", err)
	}

	if currentQuantity < quantity {
		return models.NewInsufficientResourcesError("item quantity", quantity, currentQuantity)
	}

	// Update source slot quantity
	_, err = tx.ExecContext(ctx,
		"UPDATE inventory_items SET quantity = quantity - $3, updated_at = NOW() WHERE character_id = $1 AND slot = $2",
		characterID, fromSlot, quantity)
	if err != nil {
		return fmt.Errorf("failed to update source slot: %w", err)
	}

	// Create new item in destination slot
	newItemID := uuid.New()
	_, err = tx.ExecContext(ctx,
		"INSERT INTO inventory_items (id, character_id, item_id, quantity, slot, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, NOW(), NOW())",
		newItemID, characterID, itemID, quantity, toSlot)
	if err != nil {
		return fmt.Errorf("failed to create new stack: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ListItems lists items in inventory with filtering
func (r *inventoryRepository) ListItems(ctx context.Context, characterID uuid.UUID, filter *models.InventoryFilterRequest) ([]models.InventoryItem, error) {
	baseQuery := `
		SELECT ii.id, ii.character_id, ii.item_id, ii.quantity, ii.slot, ii.created_at, ii.updated_at,
			i.name, i.description, i.type, i.rarity, i.level, i.stats, i.requirements,
			i.metadata, i.weight, i.max_stack_size, i.tradeable, i.sellable, i.destroyable, 
			i.value, i.image_url, i.created_at as item_created_at, i.updated_at as item_updated_at
		FROM inventory_items ii
		JOIN items i ON ii.item_id = i.id
		WHERE ii.character_id = $1
	`

	args := []interface{}{characterID}
	argCount := 1

	// Add filters
	if filter.ItemType != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND i.type = $%d", argCount)
		args = append(args, *filter.ItemType)
	}

	if filter.Rarity != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND i.rarity = $%d", argCount)
		args = append(args, *filter.Rarity)
	}

	if filter.Search != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND (i.name ILIKE $%d OR i.description ILIKE $%d)", argCount, argCount)
		args = append(args, "%"+filter.Search+"%")
	}

	baseQuery += " ORDER BY ii.slot ASC"

	var items []models.InventoryItem
	err := r.db.SelectContext(ctx, &items, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list inventory items: %w", err)
	}

	return items, nil
}

// GetInventoryStats calculates inventory statistics
func (r *inventoryRepository) GetInventoryStats(ctx context.Context, characterID uuid.UUID) (*models.InventoryStats, error) {
	query := `
		SELECT 
			COUNT(ii.id) as item_count,
			COALESCE(SUM(ii.quantity), 0) as total_quantity,
			COALESCE(SUM(ii.quantity * i.value), 0) as total_value,
			COALESCE(SUM(ii.quantity * i.weight), 0) as weight_used,
			inv.slots as total_slots,
			inv.max_weight
		FROM inventories inv
		LEFT JOIN inventory_items ii ON inv.character_id = ii.character_id
		LEFT JOIN items i ON ii.item_id = i.id
		WHERE inv.character_id = $1
		GROUP BY inv.character_id, inv.slots, inv.max_weight
	`

	var stats models.InventoryStats
	err := r.db.GetContext(ctx, &stats, query, characterID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.NewNotFoundError("inventory", characterID.String())
		}
		return nil, fmt.Errorf("failed to get inventory stats: %w", err)
	}

	return &stats, nil
}

// AddItems adds multiple items to inventory in a transaction
func (r *inventoryRepository) AddItems(ctx context.Context, characterID uuid.UUID, items []models.InventoryItem) error {
	if len(items) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO inventory_items (id, character_id, item_id, quantity, slot, created_at, updated_at)
		VALUES (:id, :character_id, :item_id, :quantity, :slot, :created_at, :updated_at)
	`

	for _, item := range items {
		_, err := tx.NamedExecContext(ctx, query, item)
		if err != nil {
			return fmt.Errorf("failed to add item %s: %w", item.ItemID, err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// RemoveItems removes multiple items from inventory in a transaction
func (r *inventoryRepository) RemoveItems(ctx context.Context, characterID uuid.UUID, items []models.BulkRemoveItem) error {
	if len(items) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, item := range items {
		err := r.RemoveItem(ctx, characterID, item.ItemID, item.Quantity)
		if err != nil {
			return fmt.Errorf("failed to remove item %s: %w", item.ItemID, err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
