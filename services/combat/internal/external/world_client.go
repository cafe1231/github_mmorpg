// internal/external/world_client.go
package external

import (
"github.com/google/uuid"
"github.com/sirupsen/logrus"
"combat/internal/config"
)

// WorldClientInterface définit les méthodes pour communiquer avec le service World
type WorldClientInterface interface {
CanCreateCombat(userID uuid.UUID, zoneID string) (bool, error)
ValidatePosition(characterID uuid.UUID, zoneID string, x, y, z float64) (bool, error)
GetZoneInfo(zoneID string) (*ZoneInfo, error)
NotifyCombatStarted(sessionID uuid.UUID, zoneID string) error
NotifyCombatEnded(sessionID uuid.UUID, zoneID string) error
}

// ZoneInfo représente les informations d'une zone
type ZoneInfo struct {
ID          string  `json:"id"`
Name        string  `json:"name"`
Type        string  `json:"type"`
IsPvP       bool    `json:"is_pvp"`
IsSafeZone  bool    `json:"is_safe_zone"`
Level       int     `json:"level"`
MaxPlayers  int     `json:"max_players"`
}

// WorldClient implémente l'interface WorldClientInterface
type WorldClient struct {
baseURL string
config  *config.Config
}

// NewWorldClient crée une nouvelle instance du client World
func NewWorldClient(cfg *config.Config) WorldClientInterface {
return &WorldClient{
baseURL: cfg.Services.World.URL,
config:  cfg,
}
}

// CanCreateCombat vérifie si un utilisateur peut créer un combat dans une zone
func (c *WorldClient) CanCreateCombat(userID uuid.UUID, zoneID string) (bool, error) {
return true, nil
}

// ValidatePosition valide qu'une position est valide dans une zone
func (c *WorldClient) ValidatePosition(characterID uuid.UUID, zoneID string, x, y, z float64) (bool, error) {
return true, nil
}

// GetZoneInfo récupère les informations d'une zone
func (c *WorldClient) GetZoneInfo(zoneID string) (*ZoneInfo, error) {
return &ZoneInfo{
ID:         zoneID,
Name:       "Test Zone",
Type:       "dungeon",
IsPvP:      true,
IsSafeZone: false,
Level:      1,
MaxPlayers: 20,
}, nil
}

// NotifyCombatStarted notifie le service World qu'un combat a commencé
func (c *WorldClient) NotifyCombatStarted(sessionID uuid.UUID, zoneID string) error {
logrus.WithFields(logrus.Fields{
"session_id": sessionID,
"zone_id":    zoneID,
}).Info("Combat started notification sent to World service")
return nil
}

// NotifyCombatEnded notifie le service World qu'un combat s'est terminé
func (c *WorldClient) NotifyCombatEnded(sessionID uuid.UUID, zoneID string) error {
logrus.WithFields(logrus.Fields{
"session_id": sessionID,
"zone_id":    zoneID,
}).Info("Combat ended notification sent to World service")
return nil
}
