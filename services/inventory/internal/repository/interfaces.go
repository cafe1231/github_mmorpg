package repository

import (
	"context"

	"inventory/internal/models"

	"github.com/google/uuid"
)

// ItemRepository defines methods for item data access
type ItemRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, item *models.Item) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Item, error)
	Update(ctx context.Context, item *models.Item) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Query operations
	List(ctx context.Context, filter *models.InventoryFilterRequest) ([]models.Item, int, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Item, error)
	GetByType(ctx context.Context, itemType models.ItemType) ([]models.Item, error)
	Search(ctx context.Context, query string) ([]models.Item, error)

	// Batch operations
	CreateBatch(ctx context.Context, items []models.Item) error
	UpdateBatch(ctx context.Context, items []models.Item) error
}

// InventoryRepository defines methods for inventory data access
type InventoryRepository interface {
	// Inventory management
	Create(ctx context.Context, inventory *models.Inventory) error
	GetByCharacterID(ctx context.Context, characterID uuid.UUID) (*models.Inventory, error)
	Update(ctx context.Context, inventory *models.Inventory) error
	Delete(ctx context.Context, characterID uuid.UUID) error

	// Item management within inventory
	AddItem(ctx context.Context, characterID uuid.UUID, item *models.InventoryItem) error
	RemoveItem(ctx context.Context, characterID uuid.UUID, itemID uuid.UUID, quantity int) error
	UpdateItem(ctx context.Context, characterID uuid.UUID, item *models.InventoryItem) error
	GetItem(ctx context.Context, characterID uuid.UUID, itemID uuid.UUID) (*models.InventoryItem, error)

	// Slot management
	MoveItem(ctx context.Context, characterID uuid.UUID, fromSlot, toSlot int) error
	SwapItems(ctx context.Context, characterID uuid.UUID, slot1, slot2 int) error
	SplitStack(ctx context.Context, characterID uuid.UUID, fromSlot, toSlot int, quantity int) error

	// Queries
	ListItems(ctx context.Context, characterID uuid.UUID, filter *models.InventoryFilterRequest) ([]models.InventoryItem, error)
	GetInventoryStats(ctx context.Context, characterID uuid.UUID) (*models.InventoryStats, error)

	// Bulk operations
	AddItems(ctx context.Context, characterID uuid.UUID, items []models.InventoryItem) error
	RemoveItems(ctx context.Context, characterID uuid.UUID, items []models.BulkRemoveItem) error
}

// EquipmentRepository defines methods for equipment data access
type EquipmentRepository interface {
	// Equipment management
	GetByCharacterID(ctx context.Context, characterID uuid.UUID) (*models.EquipmentSet, error)
	EquipItem(ctx context.Context, characterID uuid.UUID, itemID uuid.UUID, slot models.EquipmentSlot) (*models.Equipment, error)
	UnequipItem(ctx context.Context, characterID uuid.UUID, slot models.EquipmentSlot) (*models.Equipment, error)

	// Equipment queries
	GetEquippedItem(ctx context.Context, characterID uuid.UUID, slot models.EquipmentSlot) (*models.Equipment, error)
	ListEquipped(ctx context.Context, characterID uuid.UUID) ([]models.Equipment, error)

	// Equipment stats
	CalculateStats(ctx context.Context, characterID uuid.UUID) (map[string]int, error)
	GetSetBonuses(ctx context.Context, characterID uuid.UUID) ([]models.SetBonus, error)
}

// TradeRepository defines methods for trade data access
type TradeRepository interface {
	// Trade lifecycle
	Create(ctx context.Context, trade *models.Trade) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Trade, error)
	Update(ctx context.Context, trade *models.Trade) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Trade queries
	GetByPlayer(ctx context.Context, playerID uuid.UUID, status []models.TradeStatus) ([]models.Trade, error)
	GetPendingTrades(ctx context.Context, playerID uuid.UUID) ([]models.Trade, error)
	GetExpiredTrades(ctx context.Context) ([]models.Trade, error)

	// Trade items
	AddTradeItem(ctx context.Context, tradeID uuid.UUID, item *models.TradeItem) error
	RemoveTradeItem(ctx context.Context, tradeID uuid.UUID, itemID uuid.UUID) error
	GetTradeItems(ctx context.Context, tradeID uuid.UUID, ownerID uuid.UUID) ([]models.TradeItem, error)

	// Trade operations
	UpdateOffer(ctx context.Context, tradeID, playerID uuid.UUID, offer *models.TradeOffer) error
	SetPlayerReady(ctx context.Context, tradeID, playerID uuid.UUID, ready bool) error
	CompleteTrade(ctx context.Context, tradeID uuid.UUID) error
	CancelTrade(ctx context.Context, tradeID uuid.UUID, reason string) error
}

// CraftingRepository defines methods for crafting data access
type CraftingRepository interface {
	// Recipe management
	GetRecipeByID(ctx context.Context, id uuid.UUID) (*models.CraftingRecipe, error)
	ListRecipes(ctx context.Context, filter *models.InventoryFilterRequest) ([]models.CraftingRecipe, error)
	GetRecipesByMaterial(ctx context.Context, materialID uuid.UUID) ([]models.CraftingRecipe, error)

	// Crafting operations
	CraftItem(ctx context.Context, characterID uuid.UUID, recipe *models.CraftingRecipe,
		materials []models.CraftingMaterialInput) (*models.InventoryItem, error)
	ValidateMaterials(ctx context.Context, characterID uuid.UUID, recipe *models.CraftingRecipe,
		materials []models.CraftingMaterialInput) error
	ConsumeMaterials(ctx context.Context, characterID uuid.UUID, materials []models.CraftingMaterialInput) error
}

// Repository aggregates all repositories
type Repository struct {
	Item      ItemRepository
	Inventory InventoryRepository
	Equipment EquipmentRepository
	Trade     TradeRepository
	Crafting  CraftingRepository
}
