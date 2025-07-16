package service

import (
	"context"

	"github.com/google/uuid"
	"inventory/internal/models"
)

// InventoryService defines business logic for inventory management
type InventoryService interface {
	// Inventory management
	CreateInventory(ctx context.Context, characterID uuid.UUID, slots int, maxWeight float64) (*models.Inventory, error)
	GetInventory(ctx context.Context, characterID uuid.UUID) (*models.Inventory, error)
	DeleteInventory(ctx context.Context, characterID uuid.UUID) error

	// Item operations
	AddItem(ctx context.Context, characterID uuid.UUID, itemID uuid.UUID, quantity int) (*models.InventoryItem, error)
	RemoveItem(ctx context.Context, characterID uuid.UUID, itemID uuid.UUID, quantity int) error
	UpdateItem(ctx context.Context, characterID uuid.UUID, itemID uuid.UUID, updates *models.UpdateItemRequest) (*models.InventoryItem, error)
	GetItem(ctx context.Context, characterID uuid.UUID, itemID uuid.UUID) (*models.InventoryItem, error)

	// Slot operations
	MoveItem(ctx context.Context, characterID uuid.UUID, request *models.MoveItemRequest) error
	SplitStack(ctx context.Context, characterID uuid.UUID, request *models.SplitStackRequest) (*models.SplitStackResponse, error)

	// Queries
	ListItems(ctx context.Context, characterID uuid.UUID, filter *models.InventoryFilterRequest) (*models.InventoryListResponse, error)
	GetStats(ctx context.Context, characterID uuid.UUID) (*models.InventoryStats, error)

	// Bulk operations
	AddItems(ctx context.Context, characterID uuid.UUID, request *models.BulkAddItemsRequest) (*models.BulkOperationResponse, error)
	RemoveItems(ctx context.Context, characterID uuid.UUID, request *models.BulkRemoveItemsRequest) (*models.BulkOperationResponse, error)

	// Storage management
	ExpandStorage(ctx context.Context, characterID uuid.UUID, request *models.ExpandStorageRequest) (*models.StorageExpansionResponse, error)
}

// EquipmentService defines business logic for equipment management
type EquipmentService interface {
	// Equipment operations
	GetEquipment(ctx context.Context, characterID uuid.UUID) (*models.EquipmentSet, error)
	EquipItem(ctx context.Context, characterID uuid.UUID, request *models.EquipItemRequest) (*models.EquipItemResponse, error)
	UnequipItem(ctx context.Context, characterID uuid.UUID, request *models.UnequipItemRequest) (*models.EquipItemResponse, error)

	// Equipment queries
	GetEquippedItem(ctx context.Context, characterID uuid.UUID, slot models.EquipmentSlot) (*models.Equipment, error)
	ListEquipped(ctx context.Context, characterID uuid.UUID) ([]models.Equipment, error)

	// Stats and bonuses
	CalculateStats(ctx context.Context, characterID uuid.UUID) (map[string]int, error)
	GetSetBonuses(ctx context.Context, characterID uuid.UUID) ([]models.SetBonus, error)

	// Validation
	ValidateEquipment(ctx context.Context, characterID uuid.UUID, itemID uuid.UUID, slot models.EquipmentSlot) error
}

// TradeService defines business logic for trade management
type TradeService interface {
	// Trade lifecycle
	CreateTrade(ctx context.Context, initiatorID uuid.UUID, request *models.CreateTradeRequest) (*models.Trade, error)
	GetTrade(ctx context.Context, tradeID uuid.UUID) (*models.Trade, error)
	CancelTrade(ctx context.Context, tradeID uuid.UUID, playerID uuid.UUID, reason string) error
	CompleteTrade(ctx context.Context, tradeID uuid.UUID) error

	// Trade operations
	UpdateOffer(ctx context.Context, tradeID uuid.UUID, playerID uuid.UUID, request *models.UpdateTradeOfferRequest) (*models.TradeOfferResponse, error)
	SetReady(ctx context.Context, tradeID uuid.UUID, playerID uuid.UUID, request *models.SetTradeReadyRequest) (*models.TradeStatusResponse, error)

	// Trade queries
	GetPlayerTrades(ctx context.Context, playerID uuid.UUID, status []models.TradeStatus) ([]models.Trade, error)
	GetPendingTrades(ctx context.Context, playerID uuid.UUID) ([]models.Trade, error)

	// Validation
	ValidateTradeOffer(ctx context.Context, playerID uuid.UUID, offer *models.TradeOffer) error
	ValidateTradeCompletion(ctx context.Context, tradeID uuid.UUID) error

	// Background tasks
	ProcessExpiredTrades(ctx context.Context) (int, error)
}

// CraftingService defines business logic for crafting
type CraftingService interface {
	// Recipe queries
	GetRecipe(ctx context.Context, recipeID uuid.UUID) (*models.CraftingRecipe, error)
	ListRecipes(ctx context.Context, filter *models.InventoryFilterRequest) (*models.CraftingRecipeListResponse, error)
	GetRecipesByMaterial(ctx context.Context, materialID uuid.UUID) ([]models.CraftingRecipe, error)

	// Crafting operations
	CraftItem(ctx context.Context, characterID uuid.UUID, request *models.CraftItemRequest) (*models.CraftItemResponse, error)
	ValidateCrafting(ctx context.Context, characterID uuid.UUID, recipe *models.CraftingRecipe, materials []models.CraftingMaterialInput) error

	// Prerequisites
	CheckCraftingRequirements(ctx context.Context, characterID uuid.UUID, recipeID uuid.UUID) error
	GetCraftingCost(ctx context.Context, recipeID uuid.UUID, quantity int) (map[uuid.UUID]int, error)
}

// ItemService defines business logic for item management
type ItemService interface {
	// Item CRUD
	CreateItem(ctx context.Context, item *models.Item) error
	GetItem(ctx context.Context, itemID uuid.UUID) (*models.Item, error)
	UpdateItem(ctx context.Context, item *models.Item) error
	DeleteItem(ctx context.Context, itemID uuid.UUID) error

	// Item queries
	ListItems(ctx context.Context, filter *models.InventoryFilterRequest) ([]models.Item, int, error)
	SearchItems(ctx context.Context, query string) ([]models.Item, error)
	GetItemsByType(ctx context.Context, itemType models.ItemType) ([]models.Item, error)
	GetItemsByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Item, error)

	// Item validation
	ValidateItem(ctx context.Context, item *models.Item) error
	ValidateItemRequirements(ctx context.Context, characterID uuid.UUID, item *models.Item) error
}

// Service aggregates all business services
type Service struct {
	Inventory InventoryService
	Equipment EquipmentService
	Trade     TradeService
	Crafting  CraftingService
	Item      ItemService
}
