package service

import (
	"context"
	"fmt"
	"time"

	"inventory/internal/models"
	"inventory/internal/repository"

	"github.com/google/uuid"
)

// Cost constants
const (
	SlotExpansionCost = 1000 // Cost in gold per expansion slot
)

type inventoryService struct {
	inventoryRepo repository.InventoryRepository
	itemRepo      repository.ItemRepository
}

func NewInventoryService(inventoryRepo repository.InventoryRepository, itemRepo repository.ItemRepository) InventoryService {
	return &inventoryService{
		inventoryRepo: inventoryRepo,
		itemRepo:      itemRepo,
	}
}

// CreateInventory creates a new inventory for a character
func (s *inventoryService) CreateInventory(ctx context.Context, characterID uuid.UUID,
	slots int, maxWeight float64) (*models.Inventory, error) {
	now := time.Now()
	inventory := &models.Inventory{
		ID:          uuid.New(),
		CharacterID: characterID,
		Slots:       make([]models.InventoryItem, 0, slots),
		MaxSlots:    slots,
		Gold:        0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err := s.inventoryRepo.Create(ctx, inventory)
	if err != nil {
		return nil, fmt.Errorf("failed to create inventory: %w", err)
	}

	return inventory, nil
}

// GetInventory retrieves a character's inventory
func (s *inventoryService) GetInventory(ctx context.Context, characterID uuid.UUID) (*models.Inventory, error) {
	inventory, err := s.inventoryRepo.GetByCharacterID(ctx, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}

	return inventory, nil
}

// AddItem adds an item to inventory
func (s *inventoryService) AddItem(ctx context.Context, characterID uuid.UUID,
	itemID uuid.UUID, quantity int) (*models.InventoryItem, error) {
	// Get inventory
	inventory, err := s.inventoryRepo.GetByCharacterID(ctx, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}

	// Get item details
	item, err := s.itemRepo.GetByID(ctx, itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	// Check if inventory has space
	if !inventory.HasSpace(item, quantity) {
		return nil, models.NewInsufficientResourcesError("inventory space", quantity, inventory.GetEmptySlotCount())
	}

	// Add item to inventory
	err = inventory.AddItem(item, quantity)
	if err != nil {
		return nil, fmt.Errorf("failed to add item to inventory: %w", err)
	}

	// Update inventory in database
	err = s.inventoryRepo.Update(ctx, inventory)
	if err != nil {
		return nil, fmt.Errorf("failed to update inventory: %w", err)
	}

	// Return the added item
	slotIndex, inventoryItem := inventory.FindItem(itemID)
	if slotIndex == -1 {
		return nil, fmt.Errorf("failed to find added item")
	}

	return inventoryItem, nil
}

// RemoveItem removes an item from inventory
func (s *inventoryService) RemoveItem(ctx context.Context, characterID uuid.UUID, itemID uuid.UUID, quantity int) error {
	inventory, err := s.inventoryRepo.GetByCharacterID(ctx, characterID)
	if err != nil {
		return fmt.Errorf("failed to get inventory: %w", err)
	}

	// Check if item exists and has sufficient quantity
	totalQuantity := inventory.GetItemQuantity(itemID)
	if totalQuantity < quantity {
		return models.NewInsufficientResourcesError("item quantity", quantity, totalQuantity)
	}

	// Remove item from inventory
	err = inventory.RemoveItem(itemID, quantity)
	if err != nil {
		return fmt.Errorf("failed to remove item from inventory: %w", err)
	}

	// Update inventory in database
	err = s.inventoryRepo.Update(ctx, inventory)
	if err != nil {
		return fmt.Errorf("failed to update inventory: %w", err)
	}

	return nil
}

// MoveItem moves an item between slots
func (s *inventoryService) MoveItem(ctx context.Context, characterID uuid.UUID, request *models.MoveItemRequest) error {
	inventory, err := s.inventoryRepo.GetByCharacterID(ctx, characterID)
	if err != nil {
		return fmt.Errorf("failed to get inventory: %w", err)
	}

	// Validate slot indices
	if request.FromSlot < 0 || request.FromSlot >= len(inventory.Slots) {
		return models.NewValidationError("invalid from slot")
	}
	if request.ToSlot < 0 || request.ToSlot >= len(inventory.Slots) {
		return models.NewValidationError("invalid to slot")
	}

	// Check if source slot has an item
	if inventory.Slots[request.FromSlot].Item == nil {
		return models.NewValidationError("source slot is empty")
	}

	// Check if destination slot is empty or can stack
	fromItem := &inventory.Slots[request.FromSlot]
	toItem := &inventory.Slots[request.ToSlot]

	if toItem.Item == nil {
		// Empty destination slot - move item
		*toItem = *fromItem
		toItem.Slot = request.ToSlot
		*fromItem = models.InventoryItem{}
	} else if toItem.Item.ID == fromItem.Item.ID && toItem.Item.Stackable {
		// Same item and stackable - try to stack
		maxStack := toItem.Item.MaxStack
		canAdd := maxStack - toItem.Quantity
		if canAdd > 0 {
			moveAmount := minValue(fromItem.Quantity, canAdd)
			toItem.Quantity += moveAmount
			fromItem.Quantity -= moveAmount

			if fromItem.Quantity <= 0 {
				*fromItem = models.InventoryItem{}
			}
		} else {
			return models.NewValidationError("destination slot is full")
		}
	} else {
		return models.NewValidationError("destination slot is occupied by different item")
	}

	// Update inventory in database
	err = s.inventoryRepo.Update(ctx, inventory)
	if err != nil {
		return fmt.Errorf("failed to update inventory: %w", err)
	}

	return nil
}

// StackItems stacks items from one slot to another
func (s *inventoryService) StackItems(ctx context.Context, characterID uuid.UUID, fromSlot, toSlot int) error {
	inventory, err := s.inventoryRepo.GetByCharacterID(ctx, characterID)
	if err != nil {
		return fmt.Errorf("failed to get inventory: %w", err)
	}

	// Validate slot indices
	if fromSlot < 0 || fromSlot >= len(inventory.Slots) {
		return models.NewValidationError("invalid from slot")
	}
	if toSlot < 0 || toSlot >= len(inventory.Slots) {
		return models.NewValidationError("invalid to slot")
	}

	fromItem := &inventory.Slots[fromSlot]
	toItem := &inventory.Slots[toSlot]

	// Check if both slots have items
	if fromItem.Item == nil {
		return models.NewValidationError("source slot is empty")
	}
	if toItem.Item == nil {
		return models.NewValidationError("destination slot is empty")
	}

	// Check if items are the same and stackable
	if fromItem.Item.ID != toItem.Item.ID {
		return models.NewValidationError("items are different")
	}
	if !fromItem.Item.Stackable {
		return models.NewValidationError("items are not stackable")
	}

	// Stack items
	maxStack := toItem.Item.MaxStack
	canAdd := maxStack - toItem.Quantity
	if canAdd > 0 {
		moveAmount := minValue(fromItem.Quantity, canAdd)
		toItem.Quantity += moveAmount
		fromItem.Quantity -= moveAmount

		if fromItem.Quantity <= 0 {
			*fromItem = models.InventoryItem{}
		}
	} else {
		return models.NewValidationError("destination slot is full")
	}

	// Update inventory in database
	err = s.inventoryRepo.Update(ctx, inventory)
	if err != nil {
		return fmt.Errorf("failed to update inventory: %w", err)
	}

	return nil
}

// SplitStack splits a stack of items
func (s *inventoryService) SplitStack(ctx context.Context, characterID uuid.UUID,
	request *models.SplitStackRequest) (*models.SplitStackResponse, error) {
	sourceSlot := request.FromSlot
	targetSlot := request.ToSlot
	quantity := request.Quantity

	inventory, err := s.inventoryRepo.GetByCharacterID(ctx, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}

	// Validate slot indices
	if sourceSlot < 0 || sourceSlot >= len(inventory.Slots) {
		return nil, models.NewValidationError("invalid source slot")
	}
	if targetSlot < 0 || targetSlot >= len(inventory.Slots) {
		return nil, models.NewValidationError("invalid target slot")
	}

	sourceItem := &inventory.Slots[sourceSlot]
	targetItem := &inventory.Slots[targetSlot]

	// Check if source slot has an item
	if sourceItem.Item == nil {
		return nil, models.NewValidationError("source slot is empty")
	}

	// Check if target slot is empty
	if targetItem.Item != nil {
		return nil, models.NewValidationError("target slot is not empty")
	}

	// Check if quantity is valid
	if quantity <= 0 || quantity >= sourceItem.Quantity {
		return nil, models.NewValidationError("invalid split quantity")
	}

	// Perform split
	targetItem.ItemID = sourceItem.ItemID
	targetItem.Item = sourceItem.Item
	targetItem.Quantity = quantity
	targetItem.Slot = targetSlot

	sourceItem.Quantity -= quantity

	// Update inventory in database
	err = s.inventoryRepo.Update(ctx, inventory)
	if err != nil {
		return nil, fmt.Errorf("failed to update inventory: %w", err)
	}

	return &models.SplitStackResponse{
		Success:      true,
		OriginalItem: sourceItem,
		NewItem:      targetItem,
		Message:      "Stack split successfully",
	}, nil
}

// GetStats returns inventory statistics
func (s *inventoryService) GetStats(ctx context.Context, characterID uuid.UUID) (*models.InventoryStats, error) {
	inventory, err := s.inventoryRepo.GetByCharacterID(ctx, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}

	return inventory.GetStats(), nil
}

// SortInventory sorts inventory items
func (s *inventoryService) SortInventory(ctx context.Context, characterID uuid.UUID) error {
	inventory, err := s.inventoryRepo.GetByCharacterID(ctx, characterID)
	if err != nil {
		return fmt.Errorf("failed to get inventory: %w", err)
	}

	// TODO: Implement sorting logic
	inventory.SortInventory()

	// Update inventory in database
	err = s.inventoryRepo.Update(ctx, inventory)
	if err != nil {
		return fmt.Errorf("failed to update inventory: %w", err)
	}

	return nil
}

// ExpandInventory expands inventory capacity
func (s *inventoryService) ExpandInventory(ctx context.Context, characterID uuid.UUID, additionalSlots int) (*models.Inventory, error) {
	inventory, err := s.inventoryRepo.GetByCharacterID(ctx, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}

	// Expand inventory
	inventory.MaxSlots += additionalSlots

	// Update inventory in database
	err = s.inventoryRepo.Update(ctx, inventory)
	if err != nil {
		return nil, fmt.Errorf("failed to update inventory: %w", err)
	}

	return inventory, nil
}

// UpdateGold updates character's gold amount
func (s *inventoryService) UpdateGold(ctx context.Context, characterID uuid.UUID, amount int) error {
	inventory, err := s.inventoryRepo.GetByCharacterID(ctx, characterID)
	if err != nil {
		return fmt.Errorf("failed to get inventory: %w", err)
	}

	inventory.Gold += amount
	if inventory.Gold < 0 {
		return models.NewInsufficientResourcesError("gold", -amount, inventory.Gold+amount)
	}

	// Update inventory in database
	err = s.inventoryRepo.Update(ctx, inventory)
	if err != nil {
		return fmt.Errorf("failed to update inventory: %w", err)
	}

	return nil
}

// GetGold returns character's gold amount
func (s *inventoryService) GetGold(ctx context.Context, characterID uuid.UUID) (int, error) {
	inventory, err := s.inventoryRepo.GetByCharacterID(ctx, characterID)
	if err != nil {
		return 0, fmt.Errorf("failed to get inventory: %w", err)
	}

	return inventory.Gold, nil
}

// DeleteInventory deletes a character's inventory
func (s *inventoryService) DeleteInventory(ctx context.Context, characterID uuid.UUID) error {
	err := s.inventoryRepo.Delete(ctx, characterID)
	if err != nil {
		return fmt.Errorf("failed to delete inventory: %w", err)
	}

	return nil
}

// UpdateItem updates an inventory item (simplified implementation)
func (s *inventoryService) UpdateItem(ctx context.Context, characterID uuid.UUID,
	itemID uuid.UUID, updates *models.UpdateItemRequest) (*models.InventoryItem, error) {
	inventory, err := s.inventoryRepo.GetByCharacterID(ctx, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}

	_, item := inventory.FindItem(itemID)
	if item == nil {
		return nil, models.NewValidationError("item not found in inventory")
	}

	// Update quantity if provided
	if updates.Quantity != nil {
		if *updates.Quantity <= 0 {
			// Remove item
			err = inventory.RemoveItem(itemID, item.Quantity)
			if err != nil {
				return nil, fmt.Errorf("failed to remove item: %w", err)
			}
		} else {
			// Update quantity
			item.Quantity = *updates.Quantity
		}
	}

	// Update inventory in database
	err = s.inventoryRepo.Update(ctx, inventory)
	if err != nil {
		return nil, fmt.Errorf("failed to update inventory: %w", err)
	}

	return item, nil
}

// GetItem retrieves a specific item from inventory
func (s *inventoryService) GetItem(ctx context.Context, characterID uuid.UUID, itemID uuid.UUID) (*models.InventoryItem, error) {
	inventory, err := s.inventoryRepo.GetByCharacterID(ctx, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}

	_, item := inventory.FindItem(itemID)
	if item == nil {
		return nil, models.NewValidationError("item not found in inventory")
	}

	return item, nil
}

// AddItems adds multiple items to inventory (bulk operation)
func (s *inventoryService) AddItems(ctx context.Context, characterID uuid.UUID,
	request *models.BulkAddItemsRequest) (*models.BulkOperationResponse, error) {
	response := &models.BulkOperationResponse{
		Success:    true,
		Results:    []models.BulkOperationResult{},
		TotalItems: len(request.Items),
		Successful: 0,
		Failed:     0,
	}

	for _, itemRequest := range request.Items {
		result := models.BulkOperationResult{
			ItemID:  itemRequest.ItemID,
			Success: true,
		}

		_, err := s.AddItem(ctx, characterID, itemRequest.ItemID, itemRequest.Quantity)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			response.Failed++
		} else {
			response.Successful++
		}

		response.Results = append(response.Results, result)
	}

	if response.Failed > 0 {
		response.Success = false
	}

	return response, nil
}

// RemoveItems removes multiple items from inventory (bulk operation)
func (s *inventoryService) RemoveItems(ctx context.Context, characterID uuid.UUID,
	request *models.BulkRemoveItemsRequest) (*models.BulkOperationResponse, error) {
	response := &models.BulkOperationResponse{
		Success:    true,
		Results:    []models.BulkOperationResult{},
		TotalItems: len(request.Items),
		Successful: 0,
		Failed:     0,
	}

	for _, itemRequest := range request.Items {
		result := models.BulkOperationResult{
			ItemID:  itemRequest.ItemID,
			Success: true,
		}

		err := s.RemoveItem(ctx, characterID, itemRequest.ItemID, itemRequest.Quantity)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			response.Failed++
		} else {
			response.Successful++
		}

		response.Results = append(response.Results, result)
	}

	if response.Failed > 0 {
		response.Success = false
	}

	return response, nil
}

// ListItems lists items in inventory with filtering
func (s *inventoryService) ListItems(ctx context.Context, characterID uuid.UUID,
	filter *models.InventoryFilterRequest) (*models.InventoryListResponse, error) {
	inventory, err := s.inventoryRepo.GetByCharacterID(ctx, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}

	// Simple filtering implementation
	filteredItems := make([]models.InventoryItem, 0, len(inventory.Slots))
	for _, item := range inventory.Slots {
		if item.Item == nil {
			continue
		}

		// Apply filters
		if filter.ItemType != nil && item.Item.ItemType != *filter.ItemType {
			continue
		}
		if filter.Rarity != nil && item.Item.Rarity != *filter.Rarity {
			continue
		}

		filteredItems = append(filteredItems, item)
	}

	stats := inventory.GetStats()
	return &models.InventoryListResponse{
		Items: filteredItems,
		Stats: *stats,
		Meta: &models.Meta{
			Page:       1,
			Limit:      len(filteredItems),
			Total:      len(filteredItems),
			TotalPages: 1,
		},
	}, nil
}

// ExpandStorage expands inventory storage capacity
func (s *inventoryService) ExpandStorage(ctx context.Context, characterID uuid.UUID,
	request *models.ExpandStorageRequest) (*models.StorageExpansionResponse, error) {
	inventory, err := s.ExpandInventory(ctx, characterID, request.AdditionalSlots)
	if err != nil {
		return nil, fmt.Errorf("failed to expand inventory: %w", err)
	}

	return &models.StorageExpansionResponse{
		Success:    true,
		NewSlots:   request.AdditionalSlots,
		TotalSlots: inventory.MaxSlots,
		Cost:       int64(request.AdditionalSlots * SlotExpansionCost), // 1000 gold per slot
		Message:    fmt.Sprintf("Successfully expanded inventory by %d slots", request.AdditionalSlots),
	}, nil
}

// Helper function for min
func minValue(a, b int) int {
	if a < b {
		return a
	}
	return b
}
