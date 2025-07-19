package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ItemType représente le type d'un objet
type ItemType string

const (
	ItemTypeWeapon     ItemType = "weapon"
	ItemTypeArmor      ItemType = "armor"
	ItemTypeConsumable ItemType = "consumable"
	ItemTypeMaterial   ItemType = "material"
	ItemTypeQuest      ItemType = "quest"
	ItemTypeMisc       ItemType = "misc"
)

// ItemRarity représente la rareté d'un objet
type ItemRarity string

const (
	ItemRarityCommon    ItemRarity = "common"
	ItemRarityUncommon  ItemRarity = "uncommon"
	ItemRarityRare      ItemRarity = "rare"
	ItemRarityEpic      ItemRarity = "epic"
	ItemRarityLegendary ItemRarity = "legendary"
)

// ItemSlot représente l'emplacement d'équipement
type ItemSlot string

const (
	ItemSlotHead     ItemSlot = "head"
	ItemSlotChest    ItemSlot = "chest"
	ItemSlotLegs     ItemSlot = "legs"
	ItemSlotFeet     ItemSlot = "feet"
	ItemSlotMainHand ItemSlot = "main_hand"
	ItemSlotOffHand  ItemSlot = "off_hand"
	ItemSlotRing     ItemSlot = "ring"
	ItemSlotNecklace ItemSlot = "necklace"
)

// ItemStats représente les statistiques d'un objet
type ItemStats struct {
	Attack         int     `json:"attack,omitempty"`
	Defense        int     `json:"defense,omitempty"`
	MagicAttack    int     `json:"magic_attack,omitempty"`
	MagicDefense   int     `json:"magic_defense,omitempty"`
	Health         int     `json:"health,omitempty"`
	Mana           int     `json:"mana,omitempty"`
	CriticalChance float64 `json:"critical_chance,omitempty"`
	Speed          int     `json:"speed,omitempty"`
	Durability     int     `json:"durability,omitempty"`
	MaxDurability  int     `json:"max_durability,omitempty"`
}

// ItemRequirements représente les prérequis pour utiliser un objet
type ItemRequirements struct {
	MinLevel   int            `json:"min_level,omitempty"`
	ClassType  string         `json:"class_type,omitempty"`
	Skills     map[string]int `json:"skills,omitempty"`
	Attributes map[string]int `json:"attributes,omitempty"`
}

// Item représente un objet dans le jeu
type Item struct {
	ID           uuid.UUID              `json:"id" db:"id"`
	Name         string                 `json:"name" db:"name"`
	Description  string                 `json:"description" db:"description"`
	ItemType     ItemType               `json:"item_type" db:"item_type"`
	Rarity       ItemRarity             `json:"rarity" db:"rarity"`
	Stackable    bool                   `json:"stackable" db:"stackable"`
	MaxStack     int                    `json:"max_stack" db:"max_stack"`
	Tradable     bool                   `json:"tradable" db:"tradable"`
	Value        int                    `json:"value" db:"value"`
	Stats        *ItemStats             `json:"stats,omitempty" db:"stats"`
	Requirements *ItemRequirements      `json:"requirements,omitempty" db:"requirements"`
	Metadata     map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
}

// InventoryItem représente un objet dans l'inventaire d'un joueur
type InventoryItem struct {
	ItemID   uuid.UUID `json:"item_id"`
	Quantity int       `json:"quantity"`
	Slot     int       `json:"slot"`
	Item     *Item     `json:"item,omitempty"`
}

// Validate vérifie la validité d'un objet
func (i *Item) Validate() error {
	if i.Name == "" {
		return fmt.Errorf("item name is required")
	}
	if i.ItemType == "" {
		return fmt.Errorf("item type is required")
	}
	if i.MaxStack < 1 {
		i.MaxStack = 1
	}
	if i.Value < 0 {
		i.Value = 0
	}
	return nil
}

// CanStackWith vérifie si cet objet peut être empilé avec un autre
func (i *Item) CanStackWith(other *Item) bool {
	return i.Stackable && other.Stackable && i.ID == other.ID
}

// GetEquipmentSlot retourne l'emplacement d'équipement pour cet objet
func (i *Item) GetEquipmentSlot() ItemSlot {
	if i.Metadata == nil {
		return ""
	}
	if slot, ok := i.Metadata["equipment_slot"].(string); ok {
		return ItemSlot(slot)
	}
	return ""
}

// IsEquippable vérifie si l'objet peut être équipé
func (i *Item) IsEquippable() bool {
	return i.GetEquipmentSlot() != ""
}
