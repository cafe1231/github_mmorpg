package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"inventory/internal/models"
)

type tradeRepository struct {
	db *sqlx.DB
}

func NewTradeRepository(db *sqlx.DB) TradeRepository {
	return &tradeRepository{db: db}
}

// Create creates a new trade
func (r *tradeRepository) Create(ctx context.Context, trade *models.Trade) error {
	query := `
		INSERT INTO trades (id, initiator_id, recipient_id, status, initiator_ready, recipient_ready,
			initiator_gold, recipient_gold, expires_at, created_at, updated_at)
		VALUES (:id, :initiator_id, :recipient_id, :status, :initiator_ready, :recipient_ready,
			:initiator_gold, :recipient_gold, :expires_at, :created_at, :updated_at)
	`

	_, err := r.db.NamedExecContext(ctx, query, trade)
	if err != nil {
		return fmt.Errorf("failed to create trade: %w", err)
	}

	return nil
}

// GetByID retrieves a trade by ID
func (r *tradeRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Trade, error) {
	query := `
		SELECT id, initiator_id, recipient_id, status, initiator_ready, recipient_ready,
			initiator_gold, recipient_gold, expires_at, created_at, updated_at, completed_at
		FROM trades 
		WHERE id = $1
	`

	var trade models.Trade
	err := r.db.GetContext(ctx, &trade, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.NewNotFoundError("trade", id.String())
		}
		return nil, fmt.Errorf("failed to get trade: %w", err)
	}

	// Load trade items
	err = r.loadTradeItems(ctx, &trade)
	if err != nil {
		return nil, fmt.Errorf("failed to load trade items: %w", err)
	}

	return &trade, nil
}

// Update updates a trade
func (r *tradeRepository) Update(ctx context.Context, trade *models.Trade) error {
	query := `
		UPDATE trades SET
			status = :status,
			initiator_ready = :initiator_ready,
			recipient_ready = :recipient_ready,
			initiator_gold = :initiator_gold,
			recipient_gold = :recipient_gold,
			updated_at = :updated_at,
			completed_at = :completed_at
		WHERE id = :id
	`

	result, err := r.db.NamedExecContext(ctx, query, trade)
	if err != nil {
		return fmt.Errorf("failed to update trade: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return models.NewNotFoundError("trade", trade.ID.String())
	}

	return nil
}

// Delete deletes a trade
func (r *tradeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			logrus.WithError(err).Warn("Erreur lors du rollback")
		}
	}()

	// Delete trade items first
	_, err = tx.ExecContext(ctx, "DELETE FROM trade_items WHERE trade_id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete trade items: %w", err)
	}

	// Delete trade
	result, err := tx.ExecContext(ctx, "DELETE FROM trades WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete trade: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return models.NewNotFoundError("trade", id.String())
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetByPlayer retrieves trades for a specific player with status filter
func (r *tradeRepository) GetByPlayer(ctx context.Context, playerID uuid.UUID, status []models.TradeStatus) ([]models.Trade, error) {
	query := `
		SELECT id, initiator_id, recipient_id, status, initiator_ready, recipient_ready,
			initiator_gold, recipient_gold, expires_at, created_at, updated_at, completed_at
		FROM trades 
		WHERE (initiator_id = $1 OR recipient_id = $1)
	`

	args := []interface{}{playerID}

	if len(status) > 0 {
		query += " AND status = ANY($2)"
		args = append(args, status)
	}

	query += " ORDER BY created_at DESC"

	var trades []models.Trade
	err := r.db.SelectContext(ctx, &trades, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get trades by player: %w", err)
	}

	// Load trade items for each trade
	for i := range trades {
		err = r.loadTradeItems(ctx, &trades[i])
		if err != nil {
			return nil, fmt.Errorf("failed to load trade items: %w", err)
		}
	}

	return trades, nil
}

// GetPendingTrades retrieves pending trades for a player
func (r *tradeRepository) GetPendingTrades(ctx context.Context, playerID uuid.UUID) ([]models.Trade, error) {
	return r.GetByPlayer(ctx, playerID, []models.TradeStatus{models.TradeStatusPending})
}

// GetExpiredTrades retrieves expired trades
func (r *tradeRepository) GetExpiredTrades(ctx context.Context) ([]models.Trade, error) {
	query := `
		SELECT id, initiator_id, recipient_id, status, initiator_ready, recipient_ready,
			initiator_gold, recipient_gold, expires_at, created_at, updated_at, completed_at
		FROM trades 
		WHERE status = $1 AND expires_at < NOW()
		ORDER BY expires_at ASC
	`

	var trades []models.Trade
	err := r.db.SelectContext(ctx, &trades, query, models.TradeStatusPending)
	if err != nil {
		return nil, fmt.Errorf("failed to get expired trades: %w", err)
	}

	return trades, nil
}

// AddTradeItem adds an item to a trade
func (r *tradeRepository) AddTradeItem(ctx context.Context, tradeID uuid.UUID, item *models.TradeItem) error {
	query := `
		INSERT INTO trade_items (id, trade_id, item_id, quantity, owner_id, created_at)
		VALUES (:id, :trade_id, :item_id, :quantity, :owner_id, :created_at)
	`

	_, err := r.db.NamedExecContext(ctx, query, item)
	if err != nil {
		return fmt.Errorf("failed to add trade item: %w", err)
	}

	return nil
}

// RemoveTradeItem removes an item from a trade
func (r *tradeRepository) RemoveTradeItem(ctx context.Context, tradeID uuid.UUID, itemID uuid.UUID) error {
	query := `DELETE FROM trade_items WHERE trade_id = $1 AND item_id = $2`

	result, err := r.db.ExecContext(ctx, query, tradeID, itemID)
	if err != nil {
		return fmt.Errorf("failed to remove trade item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return models.NewNotFoundError("trade item", itemID.String())
	}

	return nil
}

// GetTradeItems retrieves trade items for a specific owner
func (r *tradeRepository) GetTradeItems(ctx context.Context, tradeID uuid.UUID, ownerID uuid.UUID) ([]models.TradeItem, error) {
	query := `
		SELECT ti.id, ti.trade_id, ti.item_id, ti.quantity, ti.owner_id, ti.created_at,
			i.name, i.description, i.type, i.rarity, i.level, i.stats, i.requirements,
			i.metadata, i.weight, i.max_stack_size, i.tradeable, i.sellable, i.destroyable, 
			i.value, i.image_url, i.created_at as item_created_at, i.updated_at as item_updated_at
		FROM trade_items ti
		JOIN items i ON ti.item_id = i.id
		WHERE ti.trade_id = $1 AND ti.owner_id = $2
		ORDER BY ti.created_at ASC
	`

	var items []models.TradeItem
	err := r.db.SelectContext(ctx, &items, query, tradeID, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get trade items: %w", err)
	}

	return items, nil
}

// UpdateOffer updates a player's offer in a trade
func (r *tradeRepository) UpdateOffer(ctx context.Context, tradeID, playerID uuid.UUID, offer *models.TradeOffer) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			logrus.WithError(err).Warn("Erreur lors du rollback")
		}
	}()

	// Get trade to determine player role
	var trade models.Trade
	err = tx.GetContext(ctx, &trade,
		"SELECT initiator_id, recipient_id FROM trades WHERE id = $1",
		tradeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.NewNotFoundError("trade", tradeID.String())
		}
		return fmt.Errorf("failed to get trade: %w", err)
	}

	// Verify player is part of trade
	if trade.InitiatorID != playerID && trade.RecipientID != playerID {
		return models.NewValidationError("player is not part of this trade")
	}

	// Remove existing items for this player
	_, err = tx.ExecContext(ctx,
		"DELETE FROM trade_items WHERE trade_id = $1 AND owner_id = $2",
		tradeID, playerID)
	if err != nil {
		return fmt.Errorf("failed to remove existing trade items: %w", err)
	}

	// Add new items
	for _, item := range offer.Items {
		tradeItem := &models.TradeItem{
			ID:        uuid.New(),
			TradeID:   tradeID,
			ItemID:    item.ItemID,
			Quantity:  item.Quantity,
			OwnerID:   playerID,
			CreatedAt: time.Now(),
		}

		_, err = tx.NamedExecContext(ctx,
			"INSERT INTO trade_items (id, trade_id, item_id, quantity, owner_id, created_at) VALUES (:id, :trade_id, :item_id, :quantity, :owner_id, :created_at)",
			tradeItem)
		if err != nil {
			return fmt.Errorf("failed to add trade item: %w", err)
		}
	}

	// Update gold offer
	if trade.InitiatorID == playerID {
		_, err = tx.ExecContext(ctx,
			"UPDATE trades SET initiator_gold = $2, updated_at = NOW() WHERE id = $1",
			tradeID, offer.Gold)
	} else {
		_, err = tx.ExecContext(ctx,
			"UPDATE trades SET recipient_gold = $2, updated_at = NOW() WHERE id = $1",
			tradeID, offer.Gold)
	}

	if err != nil {
		return fmt.Errorf("failed to update gold offer: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// SetPlayerReady sets the ready status for a player
func (r *tradeRepository) SetPlayerReady(ctx context.Context, tradeID, playerID uuid.UUID, ready bool) error {
	// Get trade to determine player role
	var trade models.Trade
	err := r.db.GetContext(ctx, &trade,
		"SELECT initiator_id, recipient_id FROM trades WHERE id = $1",
		tradeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.NewNotFoundError("trade", tradeID.String())
		}
		return fmt.Errorf("failed to get trade: %w", err)
	}

	// Update ready status based on player role
	if trade.InitiatorID == playerID {
		_, err = r.db.ExecContext(ctx,
			"UPDATE trades SET initiator_ready = $2, updated_at = NOW() WHERE id = $1",
			tradeID, ready)
	} else if trade.RecipientID == playerID {
		_, err = r.db.ExecContext(ctx,
			"UPDATE trades SET recipient_ready = $2, updated_at = NOW() WHERE id = $1",
			tradeID, ready)
	} else {
		return models.NewValidationError("player is not part of this trade")
	}

	if err != nil {
		return fmt.Errorf("failed to set player ready status: %w", err)
	}

	return nil
}

// CompleteTrade marks a trade as completed
func (r *tradeRepository) CompleteTrade(ctx context.Context, tradeID uuid.UUID) error {
	now := time.Now()
	query := `
		UPDATE trades SET 
			status = $2, 
			completed_at = $3, 
			updated_at = $3 
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, tradeID, models.TradeStatusCompleted, now)
	if err != nil {
		return fmt.Errorf("failed to complete trade: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return models.NewNotFoundError("trade", tradeID.String())
	}

	return nil
}

// CancelTrade marks a trade as canceled
func (r *tradeRepository) CancelTrade(ctx context.Context, tradeID uuid.UUID, reason string) error {
	query := `
		UPDATE trades SET 
			status = $2, 
			updated_at = NOW() 
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, tradeID, models.TradeStatusCancelled)
	if err != nil {
		return fmt.Errorf("failed to cancel trade: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return models.NewNotFoundError("trade", tradeID.String())
	}

	return nil
}

// loadTradeItems loads items for both players in a trade
func (r *tradeRepository) loadTradeItems(ctx context.Context, trade *models.Trade) error {
	query := `
		SELECT ti.id, ti.trade_id, ti.item_id, ti.quantity, ti.owner_id, ti.created_at,
			i.name, i.description, i.type, i.rarity, i.level, i.stats, i.requirements,
			i.metadata, i.weight, i.max_stack_size, i.tradeable, i.sellable, i.destroyable, 
			i.value, i.image_url, i.created_at as item_created_at, i.updated_at as item_updated_at
		FROM trade_items ti
		JOIN items i ON ti.item_id = i.id
		WHERE ti.trade_id = $1
		ORDER BY ti.owner_id, ti.created_at ASC
	`

	var allItems []models.TradeItem
	err := r.db.SelectContext(ctx, &allItems, query, trade.ID)
	if err != nil {
		return fmt.Errorf("failed to load trade items: %w", err)
	}

	// Separate items by owner
	trade.InitiatorItems = []models.TradeItem{}
	trade.RecipientItems = []models.TradeItem{}

	for _, item := range allItems {
		if item.OwnerID == trade.InitiatorID {
			trade.InitiatorItems = append(trade.InitiatorItems, item)
		} else if item.OwnerID == trade.RecipientID {
			trade.RecipientItems = append(trade.RecipientItems, item)
		}
	}

	return nil
}
