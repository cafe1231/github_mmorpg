package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"world/internal/config"
	"world/internal/models"
	"world/internal/repository"
)

// PlayerPositionService gère la logique métier des positions des joueurs
type PlayerPositionService struct {
	positionRepo repository.PlayerPositionRepositoryInterface
	zoneRepo     repository.ZoneRepositoryInterface
	config       *config.Config
}

// NewPlayerPositionService crée un nouveau service de position de joueur
func NewPlayerPositionService(
	positionRepo repository.PlayerPositionRepositoryInterface,
	zoneRepo repository.ZoneRepositoryInterface,
	config *config.Config,
) *PlayerPositionService {
	return &PlayerPositionService{
		positionRepo: positionRepo,
		zoneRepo:     zoneRepo,
		config:       config,
	}
}

// UpdatePosition met à jour la position d'un joueur
func (s *PlayerPositionService) UpdatePosition(characterID uuid.UUID, userID uuid.UUID, req models.UpdatePositionRequest) (*models.PlayerPosition, error) {
	// Valider la zone
	zone, err := s.zoneRepo.GetByID(req.ZoneID)
	if err != nil {
		return nil, fmt.Errorf("zone not found: %w", err)
	}

	if zone.Status != models.StatusActive {
		return nil, fmt.Errorf("zone is not active")
	}

	// Valider la position dans les limites de la zone
	if !zone.IsInZone(req.X, req.Y, req.Z) {
		return nil, fmt.Errorf("position is outside zone bounds")
	}

	// Récupérer la position actuelle
	currentPosition, err := s.positionRepo.GetByCharacterID(characterID)
	if err != nil {
		// Si pas de position existante, créer une nouvelle
		newPosition := &models.PlayerPosition{
			CharacterID: characterID,
			UserID:      userID,
			ZoneID:      req.ZoneID,
			X:           req.X,
			Y:           req.Y,
			Z:           req.Z,
			Rotation:    req.Rotation,
			VelocityX:   req.VelocityX,
			VelocityY:   req.VelocityY,
			VelocityZ:   req.VelocityZ,
			IsMoving:    req.IsMoving,
			IsOnline:    true,
			LastUpdate:  time.Now(),
		}

		if err := s.positionRepo.Upsert(newPosition); err != nil {
			return nil, fmt.Errorf("failed to create position: %w", err)
		}

		return newPosition, nil
	}

	// Valider le mouvement (anti-cheat basique)
	if err := s.validateMovement(currentPosition, &req); err != nil {
		return nil, fmt.Errorf("invalid movement: %w", err)
	}

	// Mettre à jour la position
	updatedPosition := &models.PlayerPosition{
		CharacterID: characterID,
		UserID:      userID,
		ZoneID:      req.ZoneID,
		X:           req.X,
		Y:           req.Y,
		Z:           req.Z,
		Rotation:    req.Rotation,
		VelocityX:   req.VelocityX,
		VelocityY:   req.VelocityY,
		VelocityZ:   req.VelocityZ,
		IsMoving:    req.IsMoving,
		IsOnline:    true,
		LastUpdate:  time.Now(),
	}

	if err := s.positionRepo.Upsert(updatedPosition); err != nil {
		return nil, fmt.Errorf("failed to update position: %w", err)
	}

	// Log pour debugging (niveau debug seulement)
	if s.config.Server.Environment == "development" {
		logrus.WithFields(logrus.Fields{
			"character_id": characterID,
			"zone_id":      req.ZoneID,
			"position":     fmt.Sprintf("(%.2f, %.2f, %.2f)", req.X, req.Y, req.Z),
			"is_moving":    req.IsMoving,
		}).Debug("Player position updated")
	}

	return updatedPosition, nil
}

// GetCharacterPosition récupère la position d'un personnage
func (s *PlayerPositionService) GetCharacterPosition(characterID uuid.UUID) (*models.PlayerPosition, error) {
	position, err := s.positionRepo.GetByCharacterID(characterID)
	if err != nil {
		return nil, fmt.Errorf("position not found: %w", err)
	}

	return position, nil
}

// GetZonePositions récupère toutes les positions dans une zone
func (s *PlayerPositionService) GetZonePositions(zoneID string) ([]*models.PlayerPosition, error) {
	// Vérifier que la zone existe
	_, err := s.zoneRepo.GetByID(zoneID)
	if err != nil {
		return nil, fmt.Errorf("zone not found: %w", err)
	}

	positions, err := s.positionRepo.GetByZoneID(zoneID)
	if err != nil {
		return nil, fmt.Errorf("failed to get zone positions: %w", err)
	}

	return positions, nil
}

// GetNearbyPlayers récupère les joueurs proches d'une position
func (s *PlayerPositionService) GetNearbyPlayers(characterID uuid.UUID, radius float64) ([]*models.PlayerPosition, error) {
	// Récupérer la position du joueur
	position, err := s.positionRepo.GetByCharacterID(characterID)
	if err != nil {
		return nil, fmt.Errorf("player position not found: %w", err)
	}

	// Limiter le rayon de recherche
	maxRadius := s.config.Game.MaxRenderDistance
	if radius > maxRadius {
		radius = maxRadius
	}

	// Récupérer les joueurs proches
	nearbyPlayers, err := s.positionRepo.GetNearbyPlayers(position.ZoneID, position.X, position.Y, position.Z, radius)
	if err != nil {
		return nil, fmt.Errorf("failed to get nearby players: %w", err)
	}

	// Filtrer le joueur lui-même
	var filteredPlayers []*models.PlayerPosition
	for _, player := range nearbyPlayers {
		if player.CharacterID != characterID {
			filteredPlayers = append(filteredPlayers, player)
		}
	}

	return filteredPlayers, nil
}

// SetPlayerOffline marque un joueur comme hors ligne
func (s *PlayerPositionService) SetPlayerOffline(characterID uuid.UUID) error {
	if err := s.positionRepo.SetOffline(characterID); err != nil {
		return fmt.Errorf("failed to set player offline: %w", err)
	}

	logrus.WithField("character_id", characterID).Info("Player set offline")
	return nil
}

// SetPlayerOnline marque un joueur comme en ligne
func (s *PlayerPositionService) SetPlayerOnline(characterID uuid.UUID) error {
	if err := s.positionRepo.SetOnline(characterID); err != nil {
		return fmt.Errorf("failed to set player online: %w", err)
	}

	logrus.WithField("character_id", characterID).Info("Player set online")
	return nil
}

// CleanupOfflinePlayers nettoie les joueurs hors ligne
func (s *PlayerPositionService) CleanupOfflinePlayers() error {
	timeout := 5 * time.Minute // Timeout par défaut
	
	if err := s.positionRepo.CleanupOfflinePlayers(timeout); err != nil {
		return fmt.Errorf("failed to cleanup offline players: %w", err)
	}

	return nil
}

// GetPlayerStatistics récupère les statistiques des joueurs
func (s *PlayerPositionService) GetPlayerStatistics() (map[string]interface{}, error) {
	// Nombre total de joueurs en ligne
	onlineCount, err := s.positionRepo.GetOnlinePlayerCount()
	if err != nil {
		return nil, fmt.Errorf("failed to get online player count: %w", err)
	}

	// Répartition par zone
	zoneCounts, err := s.positionRepo.GetZonePlayerCounts()
	if err != nil {
		return nil, fmt.Errorf("failed to get zone player counts: %w", err)
	}

	stats := map[string]interface{}{
		"total_online_players": onlineCount,
		"players_per_zone":     zoneCounts,
		"timestamp":           time.Now().Unix(),
	}

	return stats, nil
}

// TeleportPlayer téléporte un joueur à une position
func (s *PlayerPositionService) TeleportPlayer(characterID uuid.UUID, userID uuid.UUID, zoneID string, x, y, z float64) error {
	// Vérifier que la zone de destination existe
	zone, err := s.zoneRepo.GetByID(zoneID)
	if err != nil {
		return fmt.Errorf("destination zone not found: %w", err)
	}

	if zone.Status != models.StatusActive {
		return fmt.Errorf("destination zone is not active")
	}

	// Vérifier que la position est valide
	if !zone.IsInZone(x, y, z) {
		return fmt.Errorf("teleport position is outside zone bounds")
	}

	// Effectuer la téléportation
	if err := s.positionRepo.UpdateZone(characterID, zoneID, x, y, z); err != nil {
		return fmt.Errorf("failed to teleport player: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"character_id": characterID,
		"zone_id":      zoneID,
		"position":     fmt.Sprintf("(%.2f, %.2f, %.2f)", x, y, z),
	}).Info("Player teleported")

	return nil
}

// CheckCollisions vérifie les collisions avec d'autres joueurs
func (s *PlayerPositionService) CheckCollisions(characterID uuid.UUID, x, y, z, radius float64) ([]*models.PlayerPosition, error) {
	// Récupérer la position actuelle
	position, err := s.positionRepo.GetByCharacterID(characterID)
	if err != nil {
		return nil, fmt.Errorf("player position not found: %w", err)
	}

	// Récupérer les joueurs proches
	nearbyPlayers, err := s.positionRepo.GetNearbyPlayers(position.ZoneID, x, y, z, radius)
	if err != nil {
		return nil, fmt.Errorf("failed to check collisions: %w", err)
	}

	// Filtrer les collisions
	var collisions []*models.PlayerPosition
	for _, player := range nearbyPlayers {
		if player.CharacterID != characterID {
			collisions = append(collisions, player)
		}
	}

	return collisions, nil
}

// ValidateMovement valide un mouvement (anti-cheat basique)
func (s *PlayerPositionService) validateMovement(current *models.PlayerPosition, newReq *models.UpdatePositionRequest) error {
	// Calculer le temps écoulé depuis la dernière mise à jour
	timeDelta := time.Since(current.LastUpdate).Seconds()
	
	// Si trop peu de temps s'est écoulé, accepter le mouvement
	if timeDelta < 0.01 { // 10ms minimum
		return nil
	}

	// Calculer la distance parcourue
	dx := newReq.X - current.X
	dy := newReq.Y - current.Y
	dz := newReq.Z - current.Z
	distance := sqrt(dx*dx + dy*dy + dz*dz)

	// Calculer la vitesse
	speed := distance / timeDelta

	// Vitesse maximale autorisée (ajustable selon la classe, monture, etc.)
	maxSpeed := 20.0 // mètres par seconde

	if speed > maxSpeed {
		return fmt.Errorf("movement too fast: %.2f m/s (max: %.2f m/s)", speed, maxSpeed)
	}

	// Vérifier les changements de zone
	if current.ZoneID != newReq.ZoneID {
		// Pour un changement de zone, s'assurer qu'il y a une transition valide
		// Cette vérification pourrait être plus sophistiquée
		logrus.WithFields(logrus.Fields{
			"character_id": current.CharacterID,
			"from_zone":    current.ZoneID,
			"to_zone":      newReq.ZoneID,
		}).Debug("Zone change detected")
	}

	// Vérifier les téléportations suspectes (distance > seuil en un seul tick)
	if distance > 50.0 && timeDelta < 1.0 {
		return fmt.Errorf("suspicious teleportation: %.2f meters in %.2f seconds", distance, timeDelta)
	}

	return nil
}

// BroadcastPosition diffuse la position d'un joueur aux joueurs proches
func (s *PlayerPositionService) BroadcastPosition(characterID uuid.UUID) error {
	// Récupérer la position du joueur
	position, err := s.positionRepo.GetByCharacterID(characterID)
	if err != nil {
		return fmt.Errorf("player position not found: %w", err)
	}

	// Récupérer les joueurs proches
	nearbyPlayers, err := s.positionRepo.GetNearbyPlayers(
		position.ZoneID, 
		position.X, 
		position.Y, 
		position.Z, 
		s.config.Game.MaxRenderDistance,
	)
	if err != nil {
		return fmt.Errorf("failed to get nearby players: %w", err)
	}

	// TODO: Implémenter la diffusion via WebSocket/NATS
	// Pour l'instant, juste logger
	if len(nearbyPlayers) > 0 {
		logrus.WithFields(logrus.Fields{
			"character_id":    characterID,
			"nearby_players":  len(nearbyPlayers),
			"zone_id":        position.ZoneID,
		}).Debug("Position broadcast to nearby players")
	}

	return nil
}

// GetMovementHistory récupère l'historique des mouvements (si implémenté)
func (s *PlayerPositionService) GetMovementHistory(characterID uuid.UUID, limit int) ([]map[string]interface{}, error) {
	// Pour l'instant, retourner juste la position actuelle
	position, err := s.positionRepo.GetByCharacterID(characterID)
	if err != nil {
		return nil, fmt.Errorf("player position not found: %w", err)
	}

	history := []map[string]interface{}{
		{
			"timestamp": position.LastUpdate.Unix(),
			"zone_id":   position.ZoneID,
			"x":         position.X,
			"y":         position.Y,
			"z":         position.Z,
			"is_moving": position.IsMoving,
		},
	}

	return history, nil
}