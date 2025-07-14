package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"world/internal/config"
	"world/internal/models"
	"world/internal/repository"
)

// ZoneService gère la logique métier des zones
type ZoneService struct {
	zoneRepo     repository.ZoneRepositoryInterface
	positionRepo repository.PlayerPositionRepositoryInterface
	config       *config.Config
}

// NewZoneService crée un nouveau service de zone
func NewZoneService(
	zoneRepo repository.ZoneRepositoryInterface,
	positionRepo repository.PlayerPositionRepositoryInterface,
	config *config.Config,
) *ZoneService {
	return &ZoneService{
		zoneRepo:     zoneRepo,
		positionRepo: positionRepo,
		config:       config,
	}
}

// CreateZone crée une nouvelle zone
func (s *ZoneService) CreateZone(req models.CreateZoneRequest) (*models.Zone, error) {
	// Validation des données
	if err := s.validateCreateZoneRequest(req); err != nil {
		return nil, err
	}

	// Vérifier que la zone n'existe pas déjà
	existingZone, _ := s.zoneRepo.GetByID(req.ID)
	if existingZone != nil {
		return nil, fmt.Errorf("zone with ID '%s' already exists", req.ID)
	}

	// Créer la nouvelle zone
	zone := &models.Zone{
		ID:          req.ID,
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Type:        req.Type,
		Level:       req.Level,
		MinX:        req.MinX,
		MinY:        req.MinY,
		MinZ:        req.MinZ,
		MaxX:        req.MaxX,
		MaxY:        req.MaxY,
		MaxZ:        req.MaxZ,
		SpawnX:      req.SpawnX,
		SpawnY:      req.SpawnY,
		SpawnZ:      req.SpawnZ,
		MaxPlayers:  req.MaxPlayers,
		IsPvP:       req.IsPvP,
		IsSafeZone:  req.IsSafeZone,
		Settings:    req.Settings,
		Status:      models.StatusActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Validation supplémentaire
	if err := s.validateZoneGeometry(zone); err != nil {
		return nil, err
	}

	// Sauvegarder en base
	if err := s.zoneRepo.Create(zone); err != nil {
		return nil, fmt.Errorf("failed to create zone: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"zone_id":   zone.ID,
		"zone_name": zone.Name,
		"zone_type": zone.Type,
	}).Info("Zone created successfully")

	return zone, nil
}

// GetZone récupère une zone par son ID
func (s *ZoneService) GetZone(zoneID string) (*models.Zone, error) {
	zone, err := s.zoneRepo.GetByID(zoneID)
	if err != nil {
		return nil, fmt.Errorf("failed to get zone: %w", err)
	}

	return zone, nil
}

// ListZones récupère toutes les zones
func (s *ZoneService) ListZones() ([]*models.Zone, error) {
	zones, err := s.zoneRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to list zones: %w", err)
	}

	return zones, nil
}

// GetZonesByType récupère les zones par type
func (s *ZoneService) GetZonesByType(zoneType string) ([]*models.Zone, error) {
	zones, err := s.zoneRepo.GetByType(zoneType)
	if err != nil {
		return nil, fmt.Errorf("failed to get zones by type: %w", err)
	}

	return zones, nil
}

// GetZonesByLevel récupère les zones par niveau
func (s *ZoneService) GetZonesByLevel(minLevel, maxLevel int) ([]*models.Zone, error) {
	zones, err := s.zoneRepo.GetByLevel(minLevel, maxLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to get zones by level: %w", err)
	}

	return zones, nil
}

// UpdateZone met à jour une zone
func (s *ZoneService) UpdateZone(zoneID string, req models.UpdateZoneRequest) (*models.Zone, error) {
	// Récupérer la zone existante
	zone, err := s.zoneRepo.GetByID(zoneID)
	if err != nil {
		return nil, fmt.Errorf("zone not found: %w", err)
	}

	// Appliquer les modifications
	if req.Name != nil {
		zone.Name = *req.Name
	}
	if req.DisplayName != nil {
		zone.DisplayName = *req.DisplayName
	}
	if req.Description != nil {
		zone.Description = *req.Description
	}
	if req.Level != nil {
		zone.Level = *req.Level
	}
	if req.MaxPlayers != nil {
		zone.MaxPlayers = *req.MaxPlayers
	}
	if req.IsPvP != nil {
		zone.IsPvP = *req.IsPvP
	}
	if req.IsSafeZone != nil {
		zone.IsSafeZone = *req.IsSafeZone
	}
	if req.Settings != nil {
		zone.Settings = *req.Settings
	}
	if req.Status != nil {
		zone.Status = *req.Status
	}

	zone.UpdatedAt = time.Now()

	// Validation
	if err := s.validateZoneGeometry(zone); err != nil {
		return nil, err
	}

	// Sauvegarder
	if err := s.zoneRepo.Update(zone); err != nil {
		return nil, fmt.Errorf("failed to update zone: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"zone_id":   zone.ID,
		"zone_name": zone.Name,
	}).Info("Zone updated successfully")

	return zone, nil
}

// DeleteZone supprime une zone
func (s *ZoneService) DeleteZone(zoneID string) error {
	// Vérifier que la zone existe
	zone, err := s.zoneRepo.GetByID(zoneID)
	if err != nil {
		return fmt.Errorf("zone not found: %w", err)
	}

	// Vérifier qu'il n'y a pas de joueurs dans la zone
	playerCount, err := s.zoneRepo.GetPlayerCount(zoneID)
	if err != nil {
		return fmt.Errorf("failed to check player count: %w", err)
	}

	if playerCount > 0 {
		return fmt.Errorf("cannot delete zone with %d players inside", playerCount)
	}

	// Supprimer la zone
	if err := s.zoneRepo.Delete(zoneID); err != nil {
		return fmt.Errorf("failed to delete zone: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"zone_id":   zone.ID,
		"zone_name": zone.Name,
	}).Info("Zone deleted successfully")

	return nil
}

// EnterZone fait entrer un joueur dans une zone
func (s *ZoneService) EnterZone(characterID uuid.UUID, userID uuid.UUID, zoneID string) (*models.Zone, error) {
	// Vérifier que la zone existe et est active
	zone, err := s.zoneRepo.GetByID(zoneID)
	if err != nil {
		return nil, fmt.Errorf("zone not found: %w", err)
	}

	if zone.Status != models.StatusActive {
		return nil, fmt.Errorf("zone is not active (status: %s)", zone.Status)
	}

	// Vérifier la limite de joueurs
	if zone.MaxPlayers > 0 && zone.PlayerCount >= zone.MaxPlayers {
		return nil, fmt.Errorf("zone is full (%d/%d players)", zone.PlayerCount, zone.MaxPlayers)
	}

	// Récupérer la position actuelle du joueur (si elle existe)
	currentPosition, err := s.positionRepo.GetByCharacterID(characterID)
	if err != nil {
		// Si pas de position, créer une nouvelle avec le spawn de la zone
		spawnPoint := zone.GetSpawnPoint()
		newPosition := &models.PlayerPosition{
			CharacterID: characterID,
			UserID:      userID,
			ZoneID:      zoneID,
			X:           spawnPoint.X,
			Y:           spawnPoint.Y,
			Z:           spawnPoint.Z,
			Rotation:    0,
			VelocityX:   0,
			VelocityY:   0,
			VelocityZ:   0,
			IsMoving:    false,
			IsOnline:    true,
			LastUpdate:  time.Now(),
		}

		if err := s.positionRepo.Upsert(newPosition); err != nil {
			return nil, fmt.Errorf("failed to set player position: %w", err)
		}
	} else {
		// Mettre à jour la zone si différente
		if currentPosition.ZoneID != zoneID {
			spawnPoint := zone.GetSpawnPoint()
			if err := s.positionRepo.UpdateZone(characterID, zoneID, spawnPoint.X, spawnPoint.Y, spawnPoint.Z); err != nil {
				return nil, fmt.Errorf("failed to update player zone: %w", err)
			}
		} else {
			// Juste marquer comme en ligne
			if err := s.positionRepo.SetOnline(characterID); err != nil {
				return nil, fmt.Errorf("failed to set player online: %w", err)
			}
		}
	}

	logrus.WithFields(logrus.Fields{
		"character_id": characterID,
		"user_id":      userID,
		"zone_id":      zoneID,
		"zone_name":    zone.Name,
	}).Info("Player entered zone")

	return zone, nil
}

// LeaveZone fait sortir un joueur d'une zone
func (s *ZoneService) LeaveZone(characterID uuid.UUID, zoneID string) error {
	// Vérifier que le joueur est dans cette zone
	position, err := s.positionRepo.GetByCharacterID(characterID)
	if err != nil {
		return fmt.Errorf("player position not found: %w", err)
	}

	if position.ZoneID != zoneID {
		return fmt.Errorf("player is not in zone %s (currently in %s)", zoneID, position.ZoneID)
	}

	// Marquer comme hors ligne
	if err := s.positionRepo.SetOffline(characterID); err != nil {
		return fmt.Errorf("failed to set player offline: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"character_id": characterID,
		"zone_id":      zoneID,
	}).Info("Player left zone")

	return nil
}

// GetPlayersInZone récupère tous les joueurs dans une zone
func (s *ZoneService) GetPlayersInZone(zoneID string) ([]*models.PlayerPosition, error) {
	// Vérifier que la zone existe
	_, err := s.zoneRepo.GetByID(zoneID)
	if err != nil {
		return nil, fmt.Errorf("zone not found: %w", err)
	}

	players, err := s.positionRepo.GetByZoneID(zoneID)
	if err != nil {
		return nil, fmt.Errorf("failed to get players in zone: %w", err)
	}

	return players, nil
}

// CheckZoneTransitions vérifie si un joueur peut faire une transition vers une autre zone
func (s *ZoneService) CheckZoneTransitions(characterID uuid.UUID, characterLevel int) ([]*models.ZoneTransition, error) {
	// Récupérer la position actuelle du joueur
	position, err := s.positionRepo.GetByCharacterID(characterID)
	if err != nil {
		return nil, fmt.Errorf("player position not found: %w", err)
	}

	// Récupérer les transitions depuis cette zone
	transitions, err := s.zoneRepo.GetTransitions(position.ZoneID)
	if err != nil {
		return nil, fmt.Errorf("failed to get zone transitions: %w", err)
	}

	var availableTransitions []*models.ZoneTransition

	// Vérifier chaque transition
	for _, transition := range transitions {
		// Vérifier la distance du trigger
		dx := position.X - transition.TriggerX
		dy := position.Y - transition.TriggerY
		dz := position.Z - transition.TriggerZ
		distance := sqrt(dx*dx + dy*dy + dz*dz)

		if distance <= transition.TriggerRadius {
			// Vérifier le niveau requis
			if characterLevel >= transition.RequiredLevel {
				// TODO: Vérifier la quête requise si nécessaire
				availableTransitions = append(availableTransitions, transition)
			}
		}
	}

	return availableTransitions, nil
}

// ProcessZoneTransition traite une transition de zone
func (s *ZoneService) ProcessZoneTransition(characterID uuid.UUID, userID uuid.UUID, transitionID uuid.UUID) (*models.Zone, error) {
	// Récupérer la position actuelle du joueur
	position, err := s.positionRepo.GetByCharacterID(characterID)
	if err != nil {
		return nil, fmt.Errorf("player position not found: %w", err)
	}

	// Récupérer les transitions depuis cette zone
	transitions, err := s.zoneRepo.GetTransitions(position.ZoneID)
	if err != nil {
		return nil, fmt.Errorf("failed to get zone transitions: %w", err)
	}

	// Trouver la transition demandée
	var targetTransition *models.ZoneTransition
	for _, transition := range transitions {
		if transition.ID == transitionID {
			targetTransition = transition
			break
		}
	}

	if targetTransition == nil {
		return nil, fmt.Errorf("transition not found")
	}

	// Vérifier que le joueur est dans la zone de trigger
	dx := position.X - targetTransition.TriggerX
	dy := position.Y - targetTransition.TriggerY
	dz := position.Z - targetTransition.TriggerZ
	distance := sqrt(dx*dx + dy*dy + dz*dz)

	if distance > targetTransition.TriggerRadius {
		return nil, fmt.Errorf("player is not in transition trigger zone")
	}

	// Effectuer la transition
	err = s.positionRepo.UpdateZone(
		characterID,
		targetTransition.ToZoneID,
		targetTransition.DestinationX,
		targetTransition.DestinationY,
		targetTransition.DestinationZ,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to process zone transition: %w", err)
	}

	// Récupérer la nouvelle zone
	newZone, err := s.zoneRepo.GetByID(targetTransition.ToZoneID)
	if err != nil {
		return nil, fmt.Errorf("failed to get destination zone: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"character_id": characterID,
		"from_zone":    targetTransition.FromZoneID,
		"to_zone":      targetTransition.ToZoneID,
		"transition_id": transitionID,
	}).Info("Player zone transition completed")

	return newZone, nil
}

// ValidatePosition vérifie si une position est valide dans une zone
func (s *ZoneService) ValidatePosition(zoneID string, x, y, z float64) error {
	zone, err := s.zoneRepo.GetByID(zoneID)
	if err != nil {
		return fmt.Errorf("zone not found: %w", err)
	}

	if !zone.IsInZone(x, y, z) {
		return fmt.Errorf("position (%.2f, %.2f, %.2f) is outside zone bounds", x, y, z)
	}

	return nil
}

// GetZoneStatistics récupère les statistiques d'une zone
func (s *ZoneService) GetZoneStatistics(zoneID string) (map[string]interface{}, error) {
	zone, err := s.zoneRepo.GetByID(zoneID)
	if err != nil {
		return nil, fmt.Errorf("zone not found: %w", err)
	}

	playerCount, err := s.zoneRepo.GetPlayerCount(zoneID)
	if err != nil {
		return nil, fmt.Errorf("failed to get player count: %w", err)
	}

	stats := map[string]interface{}{
		"zone_id":           zone.ID,
		"zone_name":         zone.Name,
		"zone_type":         zone.Type,
		"current_players":   playerCount,
		"max_players":       zone.MaxPlayers,
		"occupancy_rate":    float64(playerCount) / float64(zone.MaxPlayers) * 100,
		"is_pvp":           zone.IsPvP,
		"is_safe_zone":     zone.IsSafeZone,
		"recommended_level": zone.Level,
		"status":           zone.Status,
	}

	return stats, nil
}

// Méthodes privées de validation

func (s *ZoneService) validateCreateZoneRequest(req models.CreateZoneRequest) error {
	if strings.TrimSpace(req.ID) == "" {
		return fmt.Errorf("zone ID is required")
	}

	if len(req.ID) < 3 || len(req.ID) > 50 {
		return fmt.Errorf("zone ID must be between 3 and 50 characters")
	}

	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("zone name is required")
	}

	if strings.TrimSpace(req.DisplayName) == "" {
		return fmt.Errorf("zone display name is required")
	}

	validTypes := []string{models.ZoneTypeCity, models.ZoneTypeDungeon, models.ZoneTypeWilderness, models.ZoneTypePvP, models.ZoneTypeSafe}
	if !contains(validTypes, req.Type) {
		return fmt.Errorf("invalid zone type: %s", req.Type)
	}

	if req.Level < 1 || req.Level > 100 {
		return fmt.Errorf("zone level must be between 1 and 100")
	}

	if req.MaxPlayers < 1 || req.MaxPlayers > 1000 {
		return fmt.Errorf("max players must be between 1 and 1000")
	}

	return nil
}

func (s *ZoneService) validateZoneGeometry(zone *models.Zone) error {
	if zone.MinX >= zone.MaxX {
		return fmt.Errorf("zone MinX must be less than MaxX")
	}

	if zone.MinY >= zone.MaxY {
		return fmt.Errorf("zone MinY must be less than MaxY")
	}

	if zone.MinZ >= zone.MaxZ {
		return fmt.Errorf("zone MinZ must be less than MaxZ")
	}

	// Vérifier que le point de spawn est dans la zone
	if !zone.IsInZone(zone.SpawnX, zone.SpawnY, zone.SpawnZ) {
		return fmt.Errorf("spawn point (%.2f, %.2f, %.2f) is outside zone bounds", 
			zone.SpawnX, zone.SpawnY, zone.SpawnZ)
	}

	return nil
}

// Fonctions utilitaires

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func sqrt(x float64) float64 {
	if x == 0 {
		return 0
	}
	
	// Implémentation simple de la racine carrée par méthode de Newton
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}