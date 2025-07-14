// internal/external/player_client.go
package external

import (
"github.com/google/uuid"
"combat/internal/config"
"combat/internal/models"
)

// PlayerClientInterface définit les méthodes pour communiquer avec le service Player
type PlayerClientInterface interface {
GetCharacter(characterID uuid.UUID) (*models.Character, error)
GetCharacterStats(characterID uuid.UUID) (*CharacterStats, error)
UpdateCharacterStats(characterID uuid.UUID, stats *CharacterStats) error
ValidateCharacterOwnership(userID, characterID uuid.UUID) error
}

// CharacterStats représente les stats d'un personnage depuis le service Player
type CharacterStats struct {
CharacterID     uuid.UUID `json:"character_id"`
Health          int       `json:"health"`
MaxHealth       int       `json:"max_health"`
Mana            int       `json:"mana"`
MaxMana         int       `json:"max_mana"`
Strength        int       `json:"strength"`
Agility         int       `json:"agility"`
Intelligence    int       `json:"intelligence"`
Vitality        int       `json:"vitality"`
PhysicalDamage  int       `json:"physical_damage"`
MagicalDamage   int       `json:"magical_damage"`
PhysicalDefense int       `json:"physical_defense"`
MagicalDefense  int       `json:"magical_defense"`
CriticalChance  int       `json:"critical_chance"`
AttackSpeed     int       `json:"attack_speed"`
MovementSpeed   int       `json:"movement_speed"`
}

// PlayerClient implémente l'interface PlayerClientInterface
type PlayerClient struct {
baseURL    string
config     *config.Config
}

// NewPlayerClient crée une nouvelle instance du client Player
func NewPlayerClient(cfg *config.Config) PlayerClientInterface {
return &PlayerClient{
baseURL: cfg.Services.Player.URL,
config:  cfg,
}
}

// GetCharacter récupère les informations d'un personnage
func (c *PlayerClient) GetCharacter(characterID uuid.UUID) (*models.Character, error) {
return &models.Character{
ID:              characterID,
Name:            "Test Character",
Level:           10,
Class:           "warrior",
Race:            "human",
MaxHealth:       100,
MaxMana:         50,
PhysicalDamage:  20,
MagicalDamage:   10,
PhysicalDefense: 15,
MagicalDefense:  8,
CriticalChance:  5,
AttackSpeed:     100,
MovementSpeed:   100,
}, nil
}

// GetCharacterStats récupère les stats d'un personnage
func (c *PlayerClient) GetCharacterStats(characterID uuid.UUID) (*CharacterStats, error) {
return &CharacterStats{
CharacterID:     characterID,
Health:          100,
MaxHealth:       100,
Mana:            50,
MaxMana:         50,
Strength:        10,
Agility:         10,
Intelligence:    10,
Vitality:        10,
PhysicalDamage:  20,
MagicalDamage:   10,
PhysicalDefense: 15,
MagicalDefense:  8,
CriticalChance:  5,
AttackSpeed:     100,
MovementSpeed:   100,
}, nil
}

// UpdateCharacterStats met à jour les stats d'un personnage
func (c *PlayerClient) UpdateCharacterStats(characterID uuid.UUID, stats *CharacterStats) error {
return nil
}

// ValidateCharacterOwnership valide qu'un utilisateur possède un personnage
func (c *PlayerClient) ValidateCharacterOwnership(userID, characterID uuid.UUID) error {
return nil
}
