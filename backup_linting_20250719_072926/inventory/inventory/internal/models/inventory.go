package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Inventory représente l'inventaire d'un personnage
type Inventory struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	CharacterID uuid.UUID       `json:"character_id" db:"character_id"`
	Slots       []InventoryItem `json:"slots" db:"slots"`
	MaxSlots    int             `json:"max_slots" db:"max_slots"`
	Gold        int             `json:"gold" db:"gold"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
}

// InventorySlot représente un slot d'inventaire avec métadonnées
type InventorySlot struct {
	Index    int   `json:"index"`
	Item     *Item `json:"item,omitempty"`
	Quantity int   `json:"quantity"`
	IsLocked bool  `json:"is_locked"`
}

// InventoryStats représente les statistiques d'un inventaire
type InventoryStats struct {
	UsedSlots     int              `json:"used_slots"`
	TotalSlots    int              `json:"total_slots"`
	UsagePercent  float64          `json:"usage_percent"`
	TotalValue    int              `json:"total_value"`
	ItemCount     int              `json:"item_count"`
	TypeBreakdown map[ItemType]int `json:"type_breakdown"`
}

// Validate vérifie la validité d'un inventaire
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

// HasSpace vérifie s'il y a de l'espace pour ajouter un objet
func (inv *Inventory) HasSpace(item *Item, quantity int) bool {
	if item.Stackable {
		// Vérifier s'il y a des stacks existants avec de l'espace
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

	// Compter les slots vides disponibles
	emptySlots := inv.GetEmptySlotCount()
	requiredSlots := (quantity + item.MaxStack - 1) / item.MaxStack // Division avec arrondi au supérieur

	return emptySlots >= requiredSlots
}

// GetEmptySlotCount retourne le nombre de slots vides
func (inv *Inventory) GetEmptySlotCount() int {
	usedSlots := 0
	for _, slot := range inv.Slots {
		if slot.Item != nil && slot.Quantity > 0 {
			usedSlots++
		}
	}
	return inv.MaxSlots - usedSlots
}

// FindItem cherche un objet dans l'inventaire
func (inv *Inventory) FindItem(itemID uuid.UUID) (int, *InventoryItem) {
	for i, slot := range inv.Slots {
		if slot.Item != nil && slot.Item.ID == itemID {
			return i, &slot
		}
	}
	return -1, nil
}

// GetItemQuantity retourne la quantité totale d'un objet
func (inv *Inventory) GetItemQuantity(itemID uuid.UUID) int {
	total := 0
	for _, slot := range inv.Slots {
		if slot.Item != nil && slot.Item.ID == itemID {
			total += slot.Quantity
		}
	}
	return total
}

// AddItem ajoute un objet à l'inventaire
func (inv *Inventory) AddItem(item *Item, quantity int) error {
	if !inv.HasSpace(item, quantity) {
		return fmt.Errorf("not enough space in inventory")
	}

	remaining := quantity

	// Si l'objet est stackable, essayer de remplir les stacks existants
	if item.Stackable {
		for i := range inv.Slots {
			slot := &inv.Slots[i]
			if slot.Item != nil && slot.Item.ID == item.ID {
				canAdd := item.MaxStack - slot.Quantity
				if canAdd > 0 {
					addAmount := min(canAdd, remaining)
					slot.Quantity += addAmount
					remaining -= addAmount
					if remaining <= 0 {
						break
					}
				}
			}
		}
	}

	// Créer de nouveaux slots pour le restant
	for remaining > 0 {
		// Trouver le premier slot vide
		emptyIndex := -1
		for i := range inv.Slots {
			if inv.Slots[i].Item == nil || inv.Slots[i].Quantity == 0 {
				emptyIndex = i
				break
			}
		}

		// Si aucun slot vide, en créer un nouveau
		if emptyIndex == -1 {
			if len(inv.Slots) >= inv.MaxSlots {
				return fmt.Errorf("inventory full")
			}
			inv.Slots = append(inv.Slots, InventoryItem{})
			emptyIndex = len(inv.Slots) - 1
		}

		// Ajouter l'objet au slot
		addAmount := min(remaining, item.MaxStack)
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

// RemoveItem retire un objet de l'inventaire
func (inv *Inventory) RemoveItem(itemID uuid.UUID, quantity int) error {
	totalAvailable := inv.GetItemQuantity(itemID)
	if totalAvailable < quantity {
		return fmt.Errorf("not enough items to remove")
	}

	remaining := quantity
	for i := range inv.Slots {
		slot := &inv.Slots[i]
		if slot.Item != nil && slot.Item.ID == itemID && remaining > 0 {
			removeAmount := min(slot.Quantity, remaining)
			slot.Quantity -= removeAmount
			remaining -= removeAmount

			// Si le slot est vide, le nettoyer
			if slot.Quantity <= 0 {
				inv.Slots[i] = InventoryItem{}
			}
		}
	}

	return nil
}

// GetStats retourne les statistiques de l'inventaire
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
		stats.UsagePercent = float64(usedSlots) / float64(inv.MaxSlots) * 100
	}

	return stats
}

// SortInventory trie l'inventaire par type et rareté
func (inv *Inventory) SortInventory() {
	// TODO: Implémenter le tri de l'inventaire
	// Logique de tri par type d'objet, puis par rareté, puis par nom
}

// Helper function pour min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
