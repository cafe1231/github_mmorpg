package models

import "github.com/google/uuid"

// Inventory requests
type AddItemRequest struct {
	ItemID   uuid.UUID `json:"item_id" binding:"required"`
	Quantity int       `json:"quantity" binding:"required,min=1"`
	Slot     *int      `json:"slot,omitempty"`
}

type UpdateItemRequest struct {
	Quantity *int `json:"quantity,omitempty" binding:"omitempty,min=0"`
	Slot     *int `json:"slot,omitempty"`
}

type MoveItemRequest struct {
	FromSlot int `json:"from_slot" binding:"required,min=0"`
	ToSlot   int `json:"to_slot" binding:"required,min=0"`
}

type SplitStackRequest struct {
	FromSlot int `json:"from_slot" binding:"required,min=0"`
	ToSlot   int `json:"to_slot" binding:"required,min=0"`
	Quantity int `json:"quantity" binding:"required,min=1"`
}

// Equipment requests
type EquipItemRequest struct {
	ItemID uuid.UUID     `json:"item_id" binding:"required"`
	Slot   EquipmentSlot `json:"slot" binding:"required"`
}

type UnequipItemRequest struct {
	Slot EquipmentSlot `json:"slot" binding:"required"`
}

// Trade requests
type CreateTradeRequest struct {
	RecipientID uuid.UUID `json:"recipient_id" binding:"required"`
	Message     string    `json:"message,omitempty"`
}

type UpdateTradeOfferRequest struct {
	Items []TradeItemOffer `json:"items"`
	Gold  int64            `json:"gold" binding:"min=0"`
}

type AcceptTradeRequest struct {
	TradeID uuid.UUID `json:"trade_id" binding:"required"`
}

type CancelTradeRequest struct {
	TradeID uuid.UUID `json:"trade_id" binding:"required"`
	Reason  string    `json:"reason,omitempty"`
}

type SetTradeReadyRequest struct {
	Ready bool `json:"ready"`
}

// Crafting requests
type CraftItemRequest struct {
	RecipeID  uuid.UUID               `json:"recipe_id" binding:"required"`
	Quantity  int                     `json:"quantity" binding:"required,min=1"`
	Materials []CraftingMaterialInput `json:"materials,omitempty"`
}

type CraftingMaterialInput struct {
	ItemID   uuid.UUID `json:"item_id" binding:"required"`
	Quantity int       `json:"quantity" binding:"required,min=1"`
}

// Search and filter requests
type InventoryFilterRequest struct {
	ItemType  *ItemType `json:"item_type,omitempty"`
	Rarity    *Rarity   `json:"rarity,omitempty"`
	MinLevel  *int      `json:"min_level,omitempty"`
	MaxLevel  *int      `json:"max_level,omitempty"`
	Search    string    `json:"search,omitempty"`
	SortBy    string    `json:"sort_by,omitempty"`    // name, level, rarity, created_at
	SortOrder string    `json:"sort_order,omitempty"` // asc, desc
	Page      int       `json:"page,omitempty"`
	Limit     int       `json:"limit,omitempty"`
}

// Bulk operations
type BulkAddItemsRequest struct {
	Items []AddItemRequest `json:"items" binding:"required,min=1"`
}

type BulkRemoveItemsRequest struct {
	Items []BulkRemoveItem `json:"items" binding:"required,min=1"`
}

type BulkRemoveItem struct {
	ItemID   uuid.UUID `json:"item_id" binding:"required"`
	Quantity int       `json:"quantity" binding:"required,min=1"`
}

// Storage requests
type ExpandStorageRequest struct {
	AdditionalSlots int    `json:"additional_slots" binding:"required,min=1"`
	PaymentMethod   string `json:"payment_method" binding:"required"` // gold, premium_currency
}

// Validation helpers
func (r *InventoryFilterRequest) SetDefaults() {
	if r.Page <= 0 {
		r.Page = 1
	}
	if r.Limit <= 0 {
		r.Limit = 50
	}
	if r.Limit > 100 {
		r.Limit = 100
	}
	if r.SortBy == "" {
		r.SortBy = "created_at"
	}
	if r.SortOrder == "" {
		r.SortOrder = "desc"
	}
}

func (r *InventoryFilterRequest) Validate() error {
	if r.MinLevel != nil && r.MaxLevel != nil && *r.MinLevel > *r.MaxLevel {
		return NewValidationError("min_level cannot be greater than max_level")
	}

	validSortBy := map[string]bool{
		"name": true, "level": true, "rarity": true,
		"created_at": true, "updated_at": true,
	}
	if !validSortBy[r.SortBy] {
		return NewValidationError("invalid sort_by field")
	}

	if r.SortOrder != "asc" && r.SortOrder != "desc" {
		return NewValidationError("sort_order must be 'asc' or 'desc'")
	}

	return nil
}
