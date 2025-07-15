package external

import (
	"github.com/google/uuid"
	"time"
)

// PlayerInfo représente les informations d'un joueur depuis le service Player
type PlayerInfo struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	DisplayName string    `json:"display_name"`
	Level       int       `json:"level"`
	Experience  int       `json:"experience"`
	IsOnline    bool      `json:"is_online"`
	LastSeen    time.Time `json:"last_seen"`
}

// CharacterInfo représente les informations d'un personnage
type CharacterInfo struct {
	ID       uuid.UUID `json:"id"`
	PlayerID uuid.UUID `json:"player_id"`
	Name     string    `json:"name"`
	Class    string    `json:"class"`
	Level    int       `json:"level"`
	Health   int       `json:"health"`
	Mana     int       `json:"mana"`
	Stats    map[string]interface{} `json:"stats"`
}

// NOTE: ZoneInfo est déjà défini dans world_client.go