package models

import (
	"time"

	"github.com/google/uuid"
)

// Character représente un personnage de jeu
type Character struct {
	ID       uuid.UUID `json:"id" db:"id"`
	PlayerID uuid.UUID `json:"player_id" db:"player_id"`
	Name     string    `json:"name" db:"name"`
	Class    string    `json:"class" db:"class"`
	Race     string    `json:"race" db:"race"`
	Gender   string    `json:"gender" db:"gender"`

	// Apparence
	Appearance CharacterAppearance `json:"appearance" db:"appearance"`

	// Progression
	Level      int   `json:"level" db:"level"`
	Experience int64 `json:"experience" db:"experience"`

	// Position dans le monde
	ZoneID    string  `json:"zone_id" db:"zone_id"`
	PositionX float64 `json:"position_x" db:"position_x"`
	PositionY float64 `json:"position_y" db:"position_y"`
	PositionZ float64 `json:"position_z" db:"position_z"`

	// État
	Status     string    `json:"status" db:"status"` // active, inactive, deleted
	LastPlayed time.Time `json:"last_played" db:"last_played"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`

	// Relations (chargées séparément)
	Stats     *CharacterStats `json:"stats,omitempty" db:"-"`
	Equipment []Equipment     `json:"equipment,omitempty" db:"-"`
	Inventory []InventoryItem `json:"inventory,omitempty" db:"-"`
}

// CharacterAppearance représente l'apparence d'un personnage
type CharacterAppearance struct {
	SkinColor   string   `json:"skin_color"`
	HairColor   string   `json:"hair_color"`
	HairStyle   string   `json:"hair_style"`
	EyeColor    string   `json:"eye_color"`
	Height      int      `json:"height"`    // en cm
	BodyType    string   `json:"body_type"` // slim, normal, muscular
	FaceType    string   `json:"face_type"`
	Accessories []string `json:"accessories"` // scars, tattoos, etc.
}

// Equipment représente un équipement porté
type Equipment struct {
	ID       uuid.UUID `json:"id" db:"id"`
	Slot     string    `json:"slot" db:"slot"`       // helmet, armor, weapon, etc.
	ItemID   string    `json:"item_id" db:"item_id"` // Référence vers le service Inventory
	ItemName string    `json:"item_name" db:"item_name"`
	Equipped time.Time `json:"equipped" db:"equipped"`
}

// InventoryItem représente un objet dans l'inventaire (aperçu)
type InventoryItem struct {
	ID       uuid.UUID `json:"id"`
	ItemID   string    `json:"item_id"`
	ItemName string    `json:"item_name"`
	Quantity int       `json:"quantity"`
	Slot     int       `json:"slot"`
}

// CreateCharacterRequest représente une demande de création de personnage
type CreateCharacterRequest struct {
	Name       string              `json:"name" binding:"required,min=3,max=20"`
	Class      string              `json:"class" binding:"required"`
	Race       string              `json:"race" binding:"required"`
	Gender     string              `json:"gender" binding:"required"`
	Appearance CharacterAppearance `json:"appearance" binding:"required"`
}

// UpdateCharacterRequest représente une demande de mise à jour
type UpdateCharacterRequest struct {
	Name       string              `json:"name" binding:"omitempty,min=3,max=20"`
	Appearance CharacterAppearance `json:"appearance"`
}

// CharacterSummary représente un résumé de personnage pour les listes
type CharacterSummary struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Class      string    `json:"class"`
	Race       string    `json:"race"`
	Level      int       `json:"level"`
	ZoneID     string    `json:"zone_id"`
	LastPlayed time.Time `json:"last_played"`
	Status     string    `json:"status"`
}

// CharacterClasses définit les classes disponibles
var CharacterClasses = map[string]ClassInfo{
	"warrior": {
		Name:         "Warrior",
		Description:  "Master of melee combat and defense",
		PrimaryStats: []string{"strength", "vitality"},
		StartingBonus: map[string]int{
			"strength": 5,
			"vitality": 3,
		},
	},
	"mage": {
		Name:         "Mage",
		Description:  "Wielder of arcane magic and spells",
		PrimaryStats: []string{"intelligence", "mana"},
		StartingBonus: map[string]int{
			"intelligence": 5,
			"mana":         20,
		},
	},
	"archer": {
		Name:         "Archer",
		Description:  "Expert marksman and ranger",
		PrimaryStats: []string{"agility", "strength"},
		StartingBonus: map[string]int{
			"agility":  5,
			"strength": 2,
		},
	},
	"rogue": {
		Name:         "Rogue",
		Description:  "Master of stealth and critical strikes",
		PrimaryStats: []string{"agility", "intelligence"},
		StartingBonus: map[string]int{
			"agility":      4,
			"intelligence": 3,
		},
	},
}

// CharacterRaces définit les races disponibles
var CharacterRaces = map[string]RaceInfo{
	"human": {
		Name:        "Human",
		Description: "Versatile and adaptable",
		StatBonus: map[string]int{
			"strength":     1,
			"agility":      1,
			"intelligence": 1,
			"vitality":     1,
		},
	},
	"elf": {
		Name:        "Elf",
		Description: "Graceful and magically gifted",
		StatBonus: map[string]int{
			"agility":      3,
			"intelligence": 2,
		},
	},
	"dwarf": {
		Name:        "Dwarf",
		Description: "Hardy and strong",
		StatBonus: map[string]int{
			"strength": 3,
			"vitality": 3,
		},
	},
	"orc": {
		Name:        "Orc",
		Description: "Powerful and fierce",
		StatBonus: map[string]int{
			"strength": 4,
			"vitality": 2,
		},
	},
}

// ClassInfo informations sur une classe
type ClassInfo struct {
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	PrimaryStats  []string       `json:"primary_stats"`
	StartingBonus map[string]int `json:"starting_bonus"`
}

// RaceInfo informations sur une race
type RaceInfo struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	StatBonus   map[string]int `json:"stat_bonus"`
}

// IsValidClass vérifie si une classe est valide
func IsValidClass(class string) bool {
	_, exists := CharacterClasses[class]
	return exists
}

// IsValidRace vérifie si une race est valide
func IsValidRace(race string) bool {
	_, exists := CharacterRaces[race]
	return exists
}

// IsValidGender vérifie si le genre est valide
func IsValidGender(gender string) bool {
	validGenders := []string{"male", "female", "other"}
	for _, g := range validGenders {
		if gender == g {
			return true
		}
	}
	return false
}

// GetExperienceToNextLevel calcule l'expérience nécessaire pour le prochain niveau
func (c *Character) GetExperienceToNextLevel() int64 {
	// Formule exponentielle simple: niveau^2 * 100
	nextLevel := c.Level + 1
	return int64(nextLevel * nextLevel * 100)
}

// GetExperienceProgress calcule le pourcentage de progression vers le niveau suivant
func (c *Character) GetExperienceProgress() float64 {
	if c.Level >= 100 { // Max level
		return 100.0
	}

	currentLevelXP := int64(c.Level * c.Level * 100)
	nextLevelXP := c.GetExperienceToNextLevel()
	progressXP := c.Experience - currentLevelXP

	if progressXP <= 0 {
		return 0.0
	}

	return (float64(progressXP) / float64(nextLevelXP-currentLevelXP)) * 100.0
}

// CanLevelUp vérifie si le personnage peut monter de niveau
func (c *Character) CanLevelUp() bool {
	if c.Level >= 100 { // Max level
		return false
	}

	requiredXP := c.GetExperienceToNextLevel()
	return c.Experience >= requiredXP
}

// LevelUp fait monter le personnage d'un niveau
func (c *Character) LevelUp() bool {
	if !c.CanLevelUp() {
		return false
	}

	c.Level++
	c.UpdatedAt = time.Now()
	return true
}

// IsActive vérifie si le personnage est actif
func (c *Character) IsActive() bool {
	return c.Status == "active"
}

// GetDefaultAppearance retourne l'apparence par défaut selon la race
func GetDefaultAppearance(race string) CharacterAppearance {
	defaults := map[string]CharacterAppearance{
		"human": {
			SkinColor: "medium",
			HairColor: "brown",
			HairStyle: "short",
			EyeColor:  "brown",
			Height:    175,
			BodyType:  "normal",
			FaceType:  "oval",
		},
		"elf": {
			SkinColor: "pale",
			HairColor: "blonde",
			HairStyle: "long",
			EyeColor:  "blue",
			Height:    180,
			BodyType:  "slim",
			FaceType:  "angular",
		},
		"dwarf": {
			SkinColor: "tan",
			HairColor: "red",
			HairStyle: "braided",
			EyeColor:  "green",
			Height:    150,
			BodyType:  "muscular",
			FaceType:  "square",
		},
		"orc": {
			SkinColor: "green",
			HairColor: "black",
			HairStyle: "mohawk",
			EyeColor:  "red",
			Height:    190,
			BodyType:  "muscular",
			FaceType:  "angular",
		},
	}

	if appearance, exists := defaults[race]; exists {
		return appearance
	}

	return defaults["human"] // fallback
}
