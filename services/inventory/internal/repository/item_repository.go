package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"inventory/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type itemRepository struct {
	db *sqlx.DB
}

func NewItemRepository(db *sqlx.DB) ItemRepository {
	return &itemRepository{db: db}
}

// Create creates a new item
func (r *itemRepository) Create(ctx context.Context, item *models.Item) error {
	query := `
		INSERT INTO items (id, name, description, type, rarity, level, stats, requirements, 
			metadata, weight, max_stack_size, tradeable, sellable, destroyable, value, 
			image_url, created_at, updated_at)
		VALUES (:id, :name, :description, :type, :rarity, :level, :stats, :requirements,
			:metadata, :weight, :max_stack_size, :tradeable, :sellable, :destroyable, :value,
			:image_url, :created_at, :updated_at)
	`

	_, err := r.db.NamedExecContext(ctx, query, item)
	if err != nil {
		return fmt.Errorf("failed to create item: %w", err)
	}

	return nil
}

// GetByID retrieves an item by ID
func (r *itemRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Item, error) {
	query := `
		SELECT id, name, description, type, rarity, level, stats, requirements,
			metadata, weight, max_stack_size, tradeable, sellable, destroyable, value,
			image_url, created_at, updated_at
		FROM items 
		WHERE id = $1
	`

	var item models.Item
	err := r.db.GetContext(ctx, &item, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.NewNotFoundError("item", id.String())
		}
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	return &item, nil
}

// Update updates an existing item
func (r *itemRepository) Update(ctx context.Context, item *models.Item) error {
	query := `
		UPDATE items SET
			name = :name,
			description = :description,
			type = :type,
			rarity = :rarity,
			level = :level,
			stats = :stats,
			requirements = :requirements,
			metadata = :metadata,
			weight = :weight,
			max_stack_size = :max_stack_size,
			tradeable = :tradeable,
			sellable = :sellable,
			destroyable = :destroyable,
			value = :value,
			image_url = :image_url,
			updated_at = :updated_at
		WHERE id = :id
	`

	result, err := r.db.NamedExecContext(ctx, query, item)
	if err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return models.NewNotFoundError("item", item.ID.String())
	}

	return nil
}

// Delete deletes an item
func (r *itemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM items WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return models.NewNotFoundError("item", id.String())
	}

	return nil
}

// List retrieves items with filtering and pagination
func (r *itemRepository) List(ctx context.Context, filter *models.InventoryFilterRequest) ([]models.Item, int, error) {
	// Build WHERE clause
	var conditions []string
	var args []interface{}
	argCount := 0

	if filter.ItemType != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("type = $%d", argCount))
		args = append(args, *filter.ItemType)
	}

	if filter.Rarity != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("rarity = $%d", argCount))
		args = append(args, *filter.Rarity)
	}

	if filter.MinLevel != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("level >= $%d", argCount))
		args = append(args, *filter.MinLevel)
	}

	if filter.MaxLevel != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("level <= $%d", argCount))
		args = append(args, *filter.MaxLevel)
	}

	if filter.Search != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", argCount, argCount))
		args = append(args, "%"+filter.Search+"%")
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total items
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM items %s", whereClause)
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count items: %w", err)
	}

	// Build ORDER BY clause
	orderBy := "ORDER BY created_at DESC"
	if filter.SortBy != "" {
		direction := "ASC"
		if filter.SortOrder == "desc" {
			direction = "DESC"
		}
		orderBy = fmt.Sprintf("ORDER BY %s %s", filter.SortBy, direction)
	}

	// Build pagination
	offset := (filter.Page - 1) * filter.Limit
	argCount++
	limitClause := fmt.Sprintf("LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, filter.Limit, offset)

	// Execute query
	query := fmt.Sprintf(`
		SELECT id, name, description, type, rarity, level, stats, requirements,
			metadata, weight, max_stack_size, tradeable, sellable, destroyable, value,
			image_url, created_at, updated_at
		FROM items 
		%s %s %s
	`, whereClause, orderBy, limitClause)

	var items []models.Item
	err = r.db.SelectContext(ctx, &items, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list items: %w", err)
	}

	return items, total, nil
}

// GetByIDs retrieves multiple items by their IDs
func (r *itemRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Item, error) {
	if len(ids) == 0 {
		return []models.Item{}, nil
	}

	query, args, err := sqlx.In(`
		SELECT id, name, description, type, rarity, level, stats, requirements,
			metadata, weight, max_stack_size, tradeable, sellable, destroyable, value,
			image_url, created_at, updated_at
		FROM items 
		WHERE id IN (?)
	`, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to build IN query: %w", err)
	}

	query = r.db.Rebind(query)

	var items []models.Item
	err = r.db.SelectContext(ctx, &items, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get items by IDs: %w", err)
	}

	return items, nil
}

// GetByType retrieves items by type
func (r *itemRepository) GetByType(ctx context.Context, itemType models.ItemType) ([]models.Item, error) {
	query := `
		SELECT id, name, description, type, rarity, level, stats, requirements,
			metadata, weight, max_stack_size, tradeable, sellable, destroyable, value,
			image_url, created_at, updated_at
		FROM items 
		WHERE type = $1
		ORDER BY level ASC, name ASC
	`

	var items []models.Item
	err := r.db.SelectContext(ctx, &items, query, itemType)
	if err != nil {
		return nil, fmt.Errorf("failed to get items by type: %w", err)
	}

	return items, nil
}

// Search searches for items by name or description
func (r *itemRepository) Search(ctx context.Context, query string) ([]models.Item, error) {
	searchQuery := `
		SELECT id, name, description, type, rarity, level, stats, requirements,
			metadata, weight, max_stack_size, tradeable, sellable, destroyable, value,
			image_url, created_at, updated_at
		FROM items 
		WHERE name ILIKE $1 OR description ILIKE $1
		ORDER BY 
			CASE WHEN name ILIKE $1 THEN 1 ELSE 2 END,
			name ASC
		LIMIT 100
	`

	searchTerm := "%" + query + "%"

	var items []models.Item
	err := r.db.SelectContext(ctx, &items, searchQuery, searchTerm)
	if err != nil {
		return nil, fmt.Errorf("failed to search items: %w", err)
	}

	return items, nil
}

// CreateBatch creates multiple items in a single transaction
func (r *itemRepository) CreateBatch(ctx context.Context, items []models.Item) error {
	if len(items) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			logrus.WithError(err).Warn("Erreur lors du rollback")
		}
	}()

	query := `
		INSERT INTO items (id, name, description, type, rarity, level, stats, requirements, 
			metadata, weight, max_stack_size, tradeable, sellable, destroyable, value, 
			image_url, created_at, updated_at)
		VALUES (:id, :name, :description, :type, :rarity, :level, :stats, :requirements,
			:metadata, :weight, :max_stack_size, :tradeable, :sellable, :destroyable, :value,
			:image_url, :created_at, :updated_at)
	`

	for i := range items {
		item := &items[i]
		_, err := tx.NamedExecContext(ctx, query, item)
		if err != nil {
			return fmt.Errorf("failed to create item %s: %w", item.ID, err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// UpdateBatch updates multiple items in a single transaction
func (r *itemRepository) UpdateBatch(ctx context.Context, items []models.Item) error {
	if len(items) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			logrus.WithError(err).Warn("Erreur lors du rollback")
		}
	}()

	query := `
		UPDATE items SET
			name = :name,
			description = :description,
			type = :type,
			rarity = :rarity,
			level = :level,
			stats = :stats,
			requirements = :requirements,
			metadata = :metadata,
			weight = :weight,
			max_stack_size = :max_stack_size,
			tradeable = :tradeable,
			sellable = :sellable,
			destroyable = :destroyable,
			value = :value,
			image_url = :image_url,
			updated_at = :updated_at
		WHERE id = :id
	`

	for i := range items {
		item := &items[i]
		_, err := tx.NamedExecContext(ctx, query, item)
		if err != nil {
			return fmt.Errorf("failed to update item %s: %w", item.ID, err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
