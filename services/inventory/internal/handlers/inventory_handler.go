package handlers

import (
	"net/http"
	"strconv"

	"inventory/internal/models"
	"inventory/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type InventoryHandler struct {
	inventoryService service.InventoryService
}

func NewInventoryHandler(inventoryService service.InventoryService) *InventoryHandler {
	return &InventoryHandler{
		inventoryService: inventoryService,
	}
}

// GetInventory retrieves a character's inventory
// GET /:characterId
func (h *InventoryHandler) GetInventory(c *gin.Context) {
	characterID, err := uuid.Parse(c.Param("characterId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("invalid_character_id", "Invalid character ID format", ""))
		return
	}

	inventory, err := h.inventoryService.GetInventory(c.Request.Context(), characterID)
	if err != nil {
		if _, ok := err.(*models.NotFoundError); ok {
			c.JSON(http.StatusNotFound, models.NewErrorResponse("inventory_not_found", "Inventory not found", err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("internal_error", "Failed to get inventory", err.Error()))
		return
	}

	stats, err := h.inventoryService.GetStats(c.Request.Context(), characterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("internal_error", "Failed to get inventory stats", err.Error()))
		return
	}

	response := models.InventoryResponse{
		Inventory: inventory,
		Stats:     *stats,
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(response))
}

// AddItem adds an item to inventory
// POST /:characterId/items
func (h *InventoryHandler) AddItem(c *gin.Context) {
	characterID, err := uuid.Parse(c.Param("characterId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("invalid_character_id", "Invalid character ID format", ""))
		return
	}

	var request models.AddItemRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("invalid_request", "Invalid request body", err.Error()))
		return
	}

	item, err := h.inventoryService.AddItem(c.Request.Context(), characterID, request.ItemID, request.Quantity)
	if err != nil {
		if _, ok := err.(*models.ValidationError); ok {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse("validation_error", err.Error(), ""))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("internal_error", "Failed to add item", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, models.NewSuccessResponse(models.ItemResponse{Item: item}))
}

// UpdateItem updates an inventory item
// PUT /:characterId/items/:itemId
func (h *InventoryHandler) UpdateItem(c *gin.Context) {
	characterID, err := uuid.Parse(c.Param("characterId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("invalid_character_id", "Invalid character ID format", ""))
		return
	}

	itemID, err := uuid.Parse(c.Param("itemId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("invalid_item_id", "Invalid item ID format", ""))
		return
	}

	var request models.UpdateItemRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("invalid_request", "Invalid request body", err.Error()))
		return
	}

	item, err := h.inventoryService.UpdateItem(c.Request.Context(), characterID, itemID, &request)
	if err != nil {
		if _, ok := err.(*models.NotFoundError); ok {
			c.JSON(http.StatusNotFound, models.NewErrorResponse("item_not_found", "Item not found", err.Error()))
			return
		}
		if _, ok := err.(*models.ValidationError); ok {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse("validation_error", err.Error(), ""))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("internal_error", "Failed to update item", err.Error()))
		return
	}

	if item == nil {
		c.JSON(http.StatusOK, models.NewSuccessResponse(gin.H{"message": "Item removed successfully"}))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(models.ItemResponse{Item: item}))
}

// RemoveItem removes an item from inventory
// DELETE /:characterId/items/:itemId
func (h *InventoryHandler) RemoveItem(c *gin.Context) {
	characterID, err := uuid.Parse(c.Param("characterId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("invalid_character_id", "Invalid character ID format", ""))
		return
	}

	itemID, err := uuid.Parse(c.Param("itemId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("invalid_item_id", "Invalid item ID format", ""))
		return
	}

	// Get quantity from query parameter (default to 1)
	quantityStr := c.DefaultQuery("quantity", "1")
	quantity, err := strconv.Atoi(quantityStr)
	if err != nil || quantity <= 0 {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("invalid_quantity", "Invalid quantity", ""))
		return
	}

	err = h.inventoryService.RemoveItem(c.Request.Context(), characterID, itemID, quantity)
	if err != nil {
		if _, ok := err.(*models.NotFoundError); ok {
			c.JSON(http.StatusNotFound, models.NewErrorResponse("item_not_found", "Item not found", err.Error()))
			return
		}
		if _, ok := err.(*models.InsufficientResourcesError); ok {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse("insufficient_quantity", err.Error(), ""))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("internal_error", "Failed to remove item", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(gin.H{"message": "Item removed successfully"}))
}

// MoveItem moves an item between slots
// POST /:characterId/items/move
func (h *InventoryHandler) MoveItem(c *gin.Context) {
	characterID, err := uuid.Parse(c.Param("characterId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("invalid_character_id", "Invalid character ID format", ""))
		return
	}

	var request models.MoveItemRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("invalid_request", "Invalid request body", err.Error()))
		return
	}

	err = h.inventoryService.MoveItem(c.Request.Context(), characterID, &request)
	if err != nil {
		if _, ok := err.(*models.ValidationError); ok {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse("validation_error", err.Error(), ""))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("internal_error", "Failed to move item", err.Error()))
		return
	}

	response := models.MoveItemResponse{
		Success:  true,
		FromSlot: request.FromSlot,
		ToSlot:   request.ToSlot,
		Message:  "Item moved successfully",
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(response))
}

// SplitStack splits a stack of items
// POST /:characterId/items/split
func (h *InventoryHandler) SplitStack(c *gin.Context) {
	characterID, err := uuid.Parse(c.Param("characterId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("invalid_character_id", "Invalid character ID format", ""))
		return
	}

	var request models.SplitStackRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("invalid_request", "Invalid request body", err.Error()))
		return
	}

	response, err := h.inventoryService.SplitStack(c.Request.Context(), characterID, &request)
	if err != nil {
		if _, ok := err.(*models.ValidationError); ok {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse("validation_error", err.Error(), ""))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("internal_error", "Failed to split stack", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(response))
}

// ListItems lists inventory items with filtering
// GET /:characterId/items
func (h *InventoryHandler) ListItems(c *gin.Context) {
	characterID, err := uuid.Parse(c.Param("characterId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("invalid_character_id", "Invalid character ID format", ""))
		return
	}

	// Parse query parameters
	filter := &models.InventoryFilterRequest{
		Search:    c.Query("search"),
		SortBy:    c.Query("sort_by"),
		SortOrder: c.Query("sort_order"),
	}

	if itemType := c.Query("item_type"); itemType != "" {
		it := models.ItemType(itemType)
		filter.ItemType = &it
	}

	if rarity := c.Query("rarity"); rarity != "" {
		r := models.Rarity(rarity)
		filter.Rarity = &r
	}

	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			filter.Page = p
		}
	}

	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filter.Limit = l
		}
	}

	response, err := h.inventoryService.ListItems(c.Request.Context(), characterID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("internal_error", "Failed to list items", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewPaginatedResponse(response, filter.Page, filter.Limit, response.Meta.Total))
}

// AddBulkItems adds multiple items to inventory
// POST /:characterId/items/bulk
func (h *InventoryHandler) AddBulkItems(c *gin.Context) {
	characterID, err := uuid.Parse(c.Param("characterId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("invalid_character_id", "Invalid character ID format", ""))
		return
	}

	var request models.BulkAddItemsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("invalid_request", "Invalid request body", err.Error()))
		return
	}

	response, err := h.inventoryService.AddItems(c.Request.Context(), characterID, &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("internal_error", "Failed to add items", err.Error()))
		return
	}

	statusCode := http.StatusOK
	if !response.Success {
		statusCode = http.StatusPartialContent
	}

	c.JSON(statusCode, models.NewSuccessResponse(response))
}

// RemoveBulkItems removes multiple items from inventory
// DELETE /:characterId/items/bulk
func (h *InventoryHandler) RemoveBulkItems(c *gin.Context) {
	characterID, err := uuid.Parse(c.Param("characterId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("invalid_character_id", "Invalid character ID format", ""))
		return
	}

	var request models.BulkRemoveItemsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("invalid_request", "Invalid request body", err.Error()))
		return
	}

	response, err := h.inventoryService.RemoveItems(c.Request.Context(), characterID, &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("internal_error", "Failed to remove items", err.Error()))
		return
	}

	statusCode := http.StatusOK
	if !response.Success {
		statusCode = http.StatusPartialContent
	}

	c.JSON(statusCode, models.NewSuccessResponse(response))
}
