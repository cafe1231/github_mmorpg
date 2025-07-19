package models

import (
	"time"

	"github.com/google/uuid"
)

// CraftingRecipe représente une recette d'artisanat
type CraftingRecipe struct {
	ID             uuid.UUID             `json:"id"`
	Name           string                `json:"name"`
	Description    string                `json:"description"`
	ResultItemID   uuid.UUID             `json:"result_item_id"`
	ResultQuantity int                   `json:"result_quantity"`
	Requirements   []CraftingRequirement `json:"requirements"`
	Category       string                `json:"category"`
	Skill          string                `json:"skill"`
	MinLevel       int                   `json:"min_level"`
	CreatedAt      time.Time             `json:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at"`
}

type CraftingRequirement struct {
	ItemID   uuid.UUID `json:"item_id"`
	Quantity int       `json:"quantity"`
}

// Rarity est un alias pour ItemRarity pour la compatibilité
type Rarity = ItemRarity

// Standard API response wrapper
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

type Meta struct {
	Page       int `json:"page,omitempty"`
	Limit      int `json:"limit,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

// Inventory responses
type InventoryResponse struct {
	Inventory *Inventory     `json:"inventory"`
	Stats     InventoryStats `json:"stats"`
}

type InventoryListResponse struct {
	Items []InventoryItem `json:"items"`
	Stats InventoryStats  `json:"stats"`
	Meta  *Meta           `json:"meta,omitempty"`
}

type ItemResponse struct {
	Item *InventoryItem `json:"item"`
}

// Equipment responses
type EquipmentResponse struct {
	Equipment *EquipmentSet  `json:"equipment"`
	Stats     EquipmentStats `json:"stats"`
}

type EquipmentStats struct {
	TotalStats    map[string]int `json:"total_stats"`
	SetBonuses    []SetBonus     `json:"set_bonuses"`
	EquippedSlots int            `json:"equipped_slots"`
	TotalSlots    int            `json:"total_slots"`
}

type EquipItemResponse struct {
	Success        bool           `json:"success"`
	EquippedItem   *Equipment     `json:"equipped_item,omitempty"`
	UnequippedItem *InventoryItem `json:"unequipped_item,omitempty"`
	Message        string         `json:"message,omitempty"`
}

// Trade responses
type TradeResponse struct {
	Trade *Trade `json:"trade"`
}

type TradeListResponse struct {
	Trades []Trade `json:"trades"`
	Meta   *Meta   `json:"meta,omitempty"`
}

type TradeOfferResponse struct {
	Trade   *Trade `json:"trade"`
	Updated bool   `json:"updated"`
	Message string `json:"message,omitempty"`
}

type TradeStatusResponse struct {
	TradeID uuid.UUID   `json:"trade_id"`
	Status  TradeStatus `json:"status"`
	Message string      `json:"message"`
}

// Crafting responses
type CraftingRecipeResponse struct {
	Recipe *CraftingRecipe `json:"recipe"`
}

type CraftingRecipeListResponse struct {
	Recipes []CraftingRecipe `json:"recipes"`
	Meta    *Meta            `json:"meta,omitempty"`
}

type CraftItemResponse struct {
	Success       bool           `json:"success"`
	CraftedItem   *InventoryItem `json:"crafted_item,omitempty"`
	UsedMaterials []UsedMaterial `json:"used_materials,omitempty"`
	Message       string         `json:"message,omitempty"`
}

type UsedMaterial struct {
	ItemID   uuid.UUID `json:"item_id"`
	Name     string    `json:"name"`
	Quantity int       `json:"quantity"`
}

// Operation responses
type BulkOperationResponse struct {
	Success    bool                  `json:"success"`
	Results    []BulkOperationResult `json:"results"`
	TotalItems int                   `json:"total_items"`
	Successful int                   `json:"successful"`
	Failed     int                   `json:"failed"`
}

type BulkOperationResult struct {
	ItemID  uuid.UUID `json:"item_id"`
	Success bool      `json:"success"`
	Error   string    `json:"error,omitempty"`
}

type MoveItemResponse struct {
	Success  bool   `json:"success"`
	FromSlot int    `json:"from_slot"`
	ToSlot   int    `json:"to_slot"`
	Message  string `json:"message,omitempty"`
}

type SplitStackResponse struct {
	Success      bool           `json:"success"`
	OriginalItem *InventoryItem `json:"original_item"`
	NewItem      *InventoryItem `json:"new_item"`
	Message      string         `json:"message,omitempty"`
}

// Storage responses
type StorageExpansionResponse struct {
	Success    bool   `json:"success"`
	NewSlots   int    `json:"new_slots"`
	TotalSlots int    `json:"total_slots"`
	Cost       int64  `json:"cost"`
	Message    string `json:"message,omitempty"`
}

// Analytics responses
type InventoryAnalyticsResponse struct {
	CharacterID uuid.UUID          `json:"character_id"`
	Analytics   InventoryAnalytics `json:"analytics"`
	GeneratedAt time.Time          `json:"generated_at"`
}

type InventoryAnalytics struct {
	ItemsByType       map[ItemType]int `json:"items_by_type"`
	ItemsByRarity     map[Rarity]int   `json:"items_by_rarity"`
	MostValuableItems []ValuedItem     `json:"most_valuable_items"`
	RecentActivity    []ActivityRecord `json:"recent_activity"`
	SpaceUtilization  SpaceUtilization `json:"space_utilization"`
}

type ValuedItem struct {
	Item  *Item `json:"item"`
	Value int64 `json:"value"`
}

type ActivityRecord struct {
	Action    string    `json:"action"`
	ItemName  string    `json:"item_name"`
	Quantity  int       `json:"quantity"`
	Timestamp time.Time `json:"timestamp"`
}

type SpaceUtilization struct {
	UsedSlots     int     `json:"used_slots"`
	TotalSlots    int     `json:"total_slots"`
	UsagePercent  float64 `json:"usage_percent"`
	WeightUsed    float64 `json:"weight_used"`
	MaxWeight     float64 `json:"max_weight"`
	WeightPercent float64 `json:"weight_percent"`
}

// Helper functions for creating responses
func NewSuccessResponse(data interface{}) *APIResponse {
	return &APIResponse{
		Success: true,
		Data:    data,
	}
}

func NewErrorResponse(code, message, details string) *APIResponse {
	return &APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

func NewPaginatedResponse(data interface{}, page, limit, total int) *APIResponse {
	totalPages := (total + limit - 1) / limit

	return &APIResponse{
		Success: true,
		Data:    data,
		Meta: &Meta{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}
}
