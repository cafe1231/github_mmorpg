package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Constants for calculations
const (
	PercentageMultiplier = 100
)

// Inventory represents a character's inventory
type Inventory struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	CharacterID uuid.UUID       `json:"character_id" db:"character_id"`
	Slots       []InventoryItem `json:"slots" db:"slots"`
	MaxSlots    int             `json:"max_slots" db:"max_slots"`
	Gold        int             `json:"gold" db:"gold"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
}

// InventorySlot represents a single inventory slot
type InventorySlot struct {
	Index    int   `json:"index"`
	Item     *Item `json:"item,omitempty"`
	Quantity int   `json:"quantity"`
	IsLocked bool  `json:"is_locked"`
}

// InventoryStats represents inventory statistics
type InventoryStats struct {
	UsedSlots     int              `json:"used_slots"`
	TotalSlots    int              `json:"total_slots"`
	UsagePercent  float64          `json:"usage_percent"`
	TotalValue    int              `json:"total_value"`
	ItemCount     int              `json:"item_count"`
	TypeBreakdown map[ItemType]int `json:"type_breakdown"`
}

// Validate validates inventory data
func (inv *Inventory) Validate() error {
	if inv.CharacterID == uuid.Nil {
		return fmt.Errorf("character ID is required")
	}
	if inv.MaxSlots <= 0 {
		return fmt.Errorf("max slots must be positive")
	}
	if inv.Gold < 0 {
		return fmt.Errorf("gold cannot be negative")
	}
	if len(inv.Slots) > inv.MaxSlots {
		return fmt.Errorf("inventory exceeds maximum slots")
	}
	return nil
}

// HasSpace checks if there's space to add an item
func (inv *Inventory) HasSpace(item *Item, quantity int) bool {
	if item.Stackable {
		// Check if there are existing stacks with space
		for _, slot := range inv.Slots {
			if slot.Item != nil && slot.Item.ID == item.ID {
				remainingSpace := item.MaxStack - slot.Quantity
				if remainingSpace >= quantity {
					return true
				}
				quantity -= remainingSpace
			}
		}
	}

	// Count available empty slots
	emptySlots := inv.GetEmptySlotCount()
	requiredSlots := (quantity + item.MaxStack - 1) / item.MaxStack // Division with ceiling

	return emptySlots >= requiredSlots
}

// GetEmptySlotCount returns the number of empty slots
func (inv *Inventory) GetEmptySlotCount() int {
	usedSlots := 0
	for _, slot := range inv.Slots {
		if slot.Item != nil && slot.Quantity > 0 {
			usedSlots++
		}
	}
	return inv.MaxSlots - usedSlots
}

// FindItem searches for an item in inventory
func (inv *Inventory) FindItem(itemID uuid.UUID) (int, *InventoryItem) {
	for i, slot := range inv.Slots {
		if slot.Item != nil && slot.Item.ID == itemID {
			return i, &slot
		}
	}
	return -1, nil
}

// GetItemQuantity returns the total quantity of an item
func (inv *Inventory) GetItemQuantity(itemID uuid.UUID) int {
	total := 0
	for _, slot := range inv.Slots {
		if slot.Item != nil && slot.Item.ID == itemID {
			total += slot.Quantity
		}
	}
	return total
}

// AddItem adds an item to inventory
func (inv *Inventory) AddItem(item *Item, quantity int) error {
	if !inv.HasSpace(item, quantity) {
		return fmt.Errorf("not enough space in inventory")
	}

	remaining := quantity

	// If item is stackable, try to fill existing stacks
	if item.Stackable {
		for i := range inv.Slots {
			slot := &inv.Slots[i]
			if slot.Item != nil && slot.Item.ID == item.ID {
				canAdd := item.MaxStack - slot.Quantity
				if canAdd > 0 {
					addAmount := minValue(canAdd, remaining)
					slot.Quantity += addAmount
					remaining -= addAmount
					if remaining <= 0 {
						break
					}
				}
			}
		}
	}

	// Create new slots for remaining items
	for remaining > 0 {
		// Find first empty slot
		emptyIndex := -1
		for i := range inv.Slots {
			if inv.Slots[i].Item == nil || inv.Slots[i].Quantity == 0 {
				emptyIndex = i
				break
			}
		}

		// If no empty slot, create a new one
		if emptyIndex == -1 {
			if len(inv.Slots) >= inv.MaxSlots {
				return fmt.Errorf("inventory full")
			}
			inv.Slots = append(inv.Slots, InventoryItem{})
			emptyIndex = len(inv.Slots) - 1
		}

		// Add item to slot
		addAmount := minValue(remaining, item.MaxStack)
		inv.Slots[emptyIndex] = InventoryItem{
			ItemID:   item.ID,
			Quantity: addAmount,
			Slot:     emptyIndex,
			Item:     item,
		}
		remaining -= addAmount
	}

	return nil
}

// RemoveItem removes an item from inventory
func (inv *Inventory) RemoveItem(itemID uuid.UUID, quantity int) error {
	totalAvailable := inv.GetItemQuantity(itemID)
	if totalAvailable < quantity {
		return fmt.Errorf("not enough items to remove")
	}

	remaining := quantity
	for i := range inv.Slots {
		slot := &inv.Slots[i]
		if slot.Item != nil && slot.Item.ID == itemID && remaining > 0 {
			removeAmount := minValue(slot.Quantity, remaining)
			slot.Quantity -= removeAmount
			remaining -= removeAmount

			// If slot is empty, clean it
			if slot.Quantity <= 0 {
				inv.Slots[i] = InventoryItem{}
			}
		}
	}

	return nil
}

// GetStats returns inventory statistics
func (inv *Inventory) GetStats() *InventoryStats {
	stats := &InventoryStats{
		TotalSlots:    inv.MaxSlots,
		TypeBreakdown: make(map[ItemType]int),
	}

	usedSlots := 0
	totalValue := 0
	itemCount := 0

	for _, slot := range inv.Slots {
		if slot.Item != nil && slot.Quantity > 0 {
			usedSlots++
			totalValue += slot.Item.Value * slot.Quantity
			itemCount += slot.Quantity
			stats.TypeBreakdown[slot.Item.ItemType] += slot.Quantity
		}
	}

	stats.UsedSlots = usedSlots
	stats.TotalValue = totalValue
	stats.ItemCount = itemCount

	if inv.MaxSlots > 0 {
		stats.UsagePercent = float64(usedSlots) / float64(inv.MaxSlots) * PercentageMultiplier
	}

	return stats
}

// SortInventory sorts inventory by type and rarity
func (inv *Inventory) SortInventory() {
	// TODO: Implement inventory sorting
	// Sort logic by item type, then by rarity, then by name
}

// Helper function for min
func minValue(a, b int) int {
	if a < b {
		return a
	}
	return b
}
