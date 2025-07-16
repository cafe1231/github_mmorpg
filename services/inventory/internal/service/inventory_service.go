package service

import (
	"context"
	"fmt"
	"time"

	"inventory/internal/models"
	"inventory/internal/repository"

	"github.com/google/uuid"
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
func (s *inventoryService) CreateInventory(ctx context.Context, characterID uuid.UUID, slots int, maxWeight float64) (*models.Inventory, error) {
	now := time.Now()
	inventory := &models.Inventory{
		ID:          uuid.New(),
		CharacterID: characterID,
		MaxSlots:    slots,
		Slots:       []models.InventoryItem{},
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

// GetInventory retrieves inventory by character ID
func (s *inventoryService) GetInventory(ctx context.Context, characterID uuid.UUID) (*models.Inventory, error) {
	inventory, err := s.inventoryRepo.GetByCharacterID(ctx, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}

	return inventory, nil
}

// AddItem adds an item to inventory
func (s *inventoryService) AddItem(ctx context.Context, characterID uuid.UUID, itemID uuid.UUID, quantity int) (*models.InventoryItem, error) {
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
		return nil, models.NewValidationError("not enough space in inventory")
	}

	// Add item to inventory using the model's method
	err = inventory.AddItem(item, quantity)
	if err != nil {
		return nil, fmt.Errorf("failed to add item to inventory: %w", err)
	}

	// Update inventory in database
	err = s.inventoryRepo.Update(ctx, inventory)
	if err != nil {
		return nil, fmt.Errorf("failed to update inventory: %w", err)
	}

	// Find the added item to return
	_, addedItem := inventory.FindItem(itemID)
	if addedItem == nil {
		return nil, fmt.Errorf("item was not added correctly")
	}

	return addedItem, nil
}

// RemoveItem removes an item from inventory
func (s *inventoryService) RemoveItem(ctx context.Context, characterID uuid.UUID, itemID uuid.UUID, quantity int) error {
	// Get inventory
	inventory, err := s.inventoryRepo.GetByCharacterID(ctx, characterID)
	if err != nil {
		return fmt.Errorf("failed to get inventory: %w", err)
	}

	// Remove item using model method
	err = inventory.RemoveItem(itemID, quantity)
	if err != nil {
		return fmt.Errorf("failed to remove item: %w", err)
	}

	// Update inventory in database
	err = s.inventoryRepo.Update(ctx, inventory)
	if err != nil {
		return fmt.Errorf("failed to update inventory: %w", err)
	}

	return nil
}

// MoveItem moves an item to a different slot
func (s *inventoryService) MoveItem(ctx context.Context, characterID uuid.UUID, request *models.MoveItemRequest) error {
	fromSlot := request.FromSlot
	toSlot := request.ToSlot
	// Get inventory
	inventory, err := s.inventoryRepo.GetByCharacterID(ctx, characterID)
	if err != nil {
		return fmt.Errorf("failed to get inventory: %w", err)
	}

	// Validate slots
	if fromSlot < 0 || fromSlot >= len(inventory.Slots) || toSlot < 0 || toSlot >= inventory.MaxSlots {
		return models.NewValidationError("invalid slot index")
	}

	// Check if from slot has an item
	if fromSlot >= len(inventory.Slots) || inventory.Slots[fromSlot].Item == nil {
		return models.NewValidationError("no item in source slot")
	}

	// Simple swap logic
	if toSlot < len(inventory.Slots) {
		// Swap items
		inventory.Slots[fromSlot], inventory.Slots[toSlot] = inventory.Slots[toSlot], inventory.Slots[fromSlot]
	} else {
		// Extend slots if needed and move item
		for len(inventory.Slots) <= toSlot {
			inventory.Slots = append(inventory.Slots, models.InventoryItem{})
		}
		inventory.Slots[toSlot] = inventory.Slots[fromSlot]
		inventory.Slots[fromSlot] = models.InventoryItem{}
	}

	// Update inventory in database
	err = s.inventoryRepo.Update(ctx, inventory)
	if err != nil {
		return fmt.Errorf("failed to update inventory: %w", err)
	}

	return nil
}

// StackItems stacks items together
func (s *inventoryService) StackItems(ctx context.Context, characterID uuid.UUID, fromSlot, toSlot int) error {
	// Get inventory
	inventory, err := s.inventoryRepo.GetByCharacterID(ctx, characterID)
	if err != nil {
		return fmt.Errorf("failed to get inventory: %w", err)
	}

	// Validate slots
	if fromSlot < 0 || fromSlot >= len(inventory.Slots) || toSlot < 0 || toSlot >= len(inventory.Slots) {
		return models.NewValidationError("invalid slot index")
	}

	fromItem := &inventory.Slots[fromSlot]
	toItem := &inventory.Slots[toSlot]

	// Check if both slots have items
	if fromItem.Item == nil || toItem.Item == nil {
		return models.NewValidationError("both slots must have items")
	}

	// Check if items can be stacked
	if !fromItem.Item.CanStackWith(toItem.Item) {
		return models.NewValidationError("items cannot be stacked together")
	}

	// Calculate new quantities
	totalQuantity := fromItem.Quantity + toItem.Quantity
	if totalQuantity <= toItem.Item.MaxStack {
		// All items fit in target stack
		toItem.Quantity = totalQuantity
		inventory.Slots[fromSlot] = models.InventoryItem{}
	} else {
		// Some items remain in source stack
		toItem.Quantity = toItem.Item.MaxStack
		fromItem.Quantity = totalQuantity - toItem.Item.MaxStack
	}

	// Update inventory in database
	err = s.inventoryRepo.Update(ctx, inventory)
	if err != nil {
		return fmt.Errorf("failed to update inventory: %w", err)
	}

	return nil
}

// SplitStack splits a stack of items
func (s *inventoryService) SplitStack(ctx context.Context, characterID uuid.UUID, request *models.SplitStackRequest) (*models.SplitStackResponse, error) {
	sourceSlot := request.FromSlot
	targetSlot := request.ToSlot
	quantity := request.Quantity
	// Get inventory
	inventory, err := s.inventoryRepo.GetByCharacterID(ctx, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}

	// Validate slots and quantity
	if sourceSlot < 0 || sourceSlot >= len(inventory.Slots) || targetSlot < 0 || targetSlot >= inventory.MaxSlots {
		return nil, models.NewValidationError("invalid slot index")
	}

	sourceItem := &inventory.Slots[sourceSlot]
	if sourceItem.Item == nil || sourceItem.Quantity <= quantity {
		return nil, models.NewValidationError("invalid split operation")
	}

	// Check if target slot is empty
	if targetSlot < len(inventory.Slots) && inventory.Slots[targetSlot].Item != nil {
		return nil, models.NewValidationError("target slot is not empty")
	}

	// Extend slots if needed
	for len(inventory.Slots) <= targetSlot {
		inventory.Slots = append(inventory.Slots, models.InventoryItem{})
	}

	// Split the stack
	newItem := models.InventoryItem{
		ItemID:   sourceItem.ItemID,
		Quantity: quantity,
		Slot:     targetSlot,
		Item:     sourceItem.Item,
	}
	inventory.Slots[targetSlot] = newItem
	sourceItem.Quantity -= quantity

	// Update inventory in database
	err = s.inventoryRepo.Update(ctx, inventory)
	if err != nil {
		return nil, fmt.Errorf("failed to update inventory: %w", err)
	}

	return &models.SplitStackResponse{
		Success:      true,
		OriginalItem: sourceItem,
		NewItem:      &newItem,
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

// SortInventory sorts the inventory
func (s *inventoryService) SortInventory(ctx context.Context, characterID uuid.UUID) error {
	inventory, err := s.inventoryRepo.GetByCharacterID(ctx, characterID)
	if err != nil {
		return fmt.Errorf("failed to get inventory: %w", err)
	}

	inventory.SortInventory()

	err = s.inventoryRepo.Update(ctx, inventory)
	if err != nil {
		return fmt.Errorf("failed to update inventory: %w", err)
	}

	return nil
}

// ExpandInventory expands inventory slots
func (s *inventoryService) ExpandInventory(ctx context.Context, characterID uuid.UUID, additionalSlots int) (*models.Inventory, error) {
	inventory, err := s.inventoryRepo.GetByCharacterID(ctx, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}

	inventory.MaxSlots += additionalSlots
	inventory.UpdatedAt = time.Now()

	err = s.inventoryRepo.Update(ctx, inventory)
	if err != nil {
		return nil, fmt.Errorf("failed to update inventory: %w", err)
	}

	return inventory, nil
}

// UpdateGold updates character's gold
func (s *inventoryService) UpdateGold(ctx context.Context, characterID uuid.UUID, amount int) error {
	inventory, err := s.inventoryRepo.GetByCharacterID(ctx, characterID)
	if err != nil {
		return fmt.Errorf("failed to get inventory: %w", err)
	}

	newGold := inventory.Gold + amount
	if newGold < 0 {
		return models.NewValidationError("insufficient gold")
	}

	inventory.Gold = newGold
	inventory.UpdatedAt = time.Now()

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
func (s *inventoryService) UpdateItem(ctx context.Context, characterID uuid.UUID, itemID uuid.UUID, updates *models.UpdateItemRequest) (*models.InventoryItem, error) {
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
func (s *inventoryService) AddItems(ctx context.Context, characterID uuid.UUID, request *models.BulkAddItemsRequest) (*models.BulkOperationResponse, error) {
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
func (s *inventoryService) RemoveItems(ctx context.Context, characterID uuid.UUID, request *models.BulkRemoveItemsRequest) (*models.BulkOperationResponse, error) {
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
func (s *inventoryService) ListItems(ctx context.Context, characterID uuid.UUID, filter *models.InventoryFilterRequest) (*models.InventoryListResponse, error) {
	inventory, err := s.inventoryRepo.GetByCharacterID(ctx, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}

	// Simple filtering implementation
	var filteredItems []models.InventoryItem
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
func (s *inventoryService) ExpandStorage(ctx context.Context, characterID uuid.UUID, request *models.ExpandStorageRequest) (*models.StorageExpansionResponse, error) {
	inventory, err := s.ExpandInventory(ctx, characterID, request.AdditionalSlots)
	if err != nil {
		return nil, fmt.Errorf("failed to expand inventory: %w", err)
	}

	return &models.StorageExpansionResponse{
		Success:    true,
		NewSlots:   request.AdditionalSlots,
		TotalSlots: inventory.MaxSlots,
		Cost:       int64(request.AdditionalSlots * 1000), // 1000 gold per slot
		Message:    fmt.Sprintf("Successfully expanded inventory by %d slots", request.AdditionalSlots),
	}, nil
}
