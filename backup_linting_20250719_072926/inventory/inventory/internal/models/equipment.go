package models

import (
	"time"

	"github.com/google/uuid"
)

type EquipmentSlot string

const (
	SlotHead     EquipmentSlot = "head"
	SlotChest    EquipmentSlot = "chest"
	SlotLegs     EquipmentSlot = "legs"
	SlotFeet     EquipmentSlot = "feet"
	SlotHands    EquipmentSlot = "hands"
	SlotMainHand EquipmentSlot = "main_hand"
	SlotOffHand  EquipmentSlot = "off_hand"
	SlotRing1    EquipmentSlot = "ring1"
	SlotRing2    EquipmentSlot = "ring2"
	SlotNecklace EquipmentSlot = "necklace"
	SlotEarring1 EquipmentSlot = "earring1"
	SlotEarring2 EquipmentSlot = "earring2"
)

type Equipment struct {
	ID          uuid.UUID     `json:"id" db:"id"`
	CharacterID uuid.UUID     `json:"character_id" db:"character_id"`
	Slot        EquipmentSlot `json:"slot" db:"slot"`
	ItemID      *uuid.UUID    `json:"item_id,omitempty" db:"item_id"`
	Item        *Item         `json:"item,omitempty"`
	EquippedAt  time.Time     `json:"equipped_at" db:"equipped_at"`
	CreatedAt   time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at" db:"updated_at"`
}

type EquipmentSet struct {
	CharacterID uuid.UUID                    `json:"character_id"`
	Equipment   map[EquipmentSlot]*Equipment `json:"equipment"`
	TotalStats  map[string]int               `json:"total_stats"`
	SetBonuses  []SetBonus                   `json:"set_bonuses"`
}

type SetBonus struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	RequiredSet string         `json:"required_set"`
	PiecesCount int            `json:"pieces_count"`
	Stats       map[string]int `json:"stats"`
	Active      bool           `json:"active"`
}

// ValidateEquipment validates if an item can be equipped in a slot
func (e *Equipment) ValidateEquipment(item *Item) error {
	if item.ItemType == ItemTypeWeapon {
		if e.Slot != SlotMainHand && e.Slot != SlotOffHand {
			return NewValidationError("weapon can only be equipped in main or off hand")
		}
	} else if item.ItemType == ItemTypeArmor {
		armorSlot := item.Metadata["slot"]
		if armorSlot != string(e.Slot) {
			return NewValidationError("armor piece doesn't match equipment slot")
		}
	}

	return nil
}

// CalculateSetBonuses calculates active set bonuses for equipped items
func (es *EquipmentSet) CalculateSetBonuses() {
	setBonuses := make(map[string]int)

	// Count items per set
	for _, equipment := range es.Equipment {
		if equipment.Item != nil {
			setName, exists := equipment.Item.Metadata["set"]
			if exists {
				if setNameStr, ok := setName.(string); ok {
					setBonuses[setNameStr]++
				}
			}
		}
	}

	// Apply set bonuses based on piece count
	es.SetBonuses = []SetBonus{}
	for setName, count := range setBonuses {
		if count >= 2 {
			bonus := SetBonus{
				Name:        setName + " Set",
				RequiredSet: setName,
				PiecesCount: count,
				Active:      true,
			}

			// Define set bonuses (simplified example)
			switch setName {
			case "warrior":
				if count >= 2 {
					bonus.Stats = map[string]int{"strength": 10, "defense": 5}
				}
				if count >= 4 {
					bonus.Stats["strength"] += 15
					bonus.Stats["defense"] += 10
				}
			case "mage":
				if count >= 2 {
					bonus.Stats = map[string]int{"intelligence": 10, "mana": 50}
				}
				if count >= 4 {
					bonus.Stats["intelligence"] += 15
					bonus.Stats["mana"] += 100
				}
			}

			es.SetBonuses = append(es.SetBonuses, bonus)
		}
	}
}

// CalculateTotalStats calculates total stats from all equipped items including set bonuses
func (es *EquipmentSet) CalculateTotalStats() {
	es.TotalStats = make(map[string]int)

	// Add stats from individual items
	for _, equipment := range es.Equipment {
		if equipment.Item != nil && equipment.Item.Stats != nil {
			stats := equipment.Item.Stats
			es.TotalStats["attack"] += stats.Attack
			es.TotalStats["defense"] += stats.Defense
			es.TotalStats["magic_attack"] += stats.MagicAttack
			es.TotalStats["magic_defense"] += stats.MagicDefense
			es.TotalStats["health"] += stats.Health
			es.TotalStats["mana"] += stats.Mana
			es.TotalStats["speed"] += stats.Speed
		}
	}

	// Add set bonus stats
	for _, bonus := range es.SetBonuses {
		if bonus.Active {
			for stat, value := range bonus.Stats {
				es.TotalStats[stat] += value
			}
		}
	}
}

// GetEquippedItem returns the item equipped in a specific slot
func (es *EquipmentSet) GetEquippedItem(slot EquipmentSlot) *Item {
	if equipment, exists := es.Equipment[slot]; exists && equipment.Item != nil {
		return equipment.Item
	}
	return nil
}

// IsSlotEmpty checks if a equipment slot is empty
func (es *EquipmentSet) IsSlotEmpty(slot EquipmentSlot) bool {
	equipment, exists := es.Equipment[slot]
	return !exists || equipment.ItemID == nil
}
