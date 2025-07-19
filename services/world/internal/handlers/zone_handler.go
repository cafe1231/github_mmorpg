package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"world/internal/models"
	"world/internal/service"
)

// ZoneHandler gère les requêtes HTTP pour les zones
type ZoneHandler struct {
	zoneService *service.ZoneService
}

// NewZoneHandler crée un nouveau handler de zone
func NewZoneHandler(zoneService *service.ZoneService) *ZoneHandler {
	return &ZoneHandler{
		zoneService: zoneService,
	}
}

// ListZones récupère toutes les zones
// @Summary List all zones
// @Description Get all available zones
// @Tags zones
// @Security BearerAuth
// @Produce json
// @Param type query string false "Zone type filter"
// @Param min_level query int false "Minimum level filter"
// @Param max_level query int false "Maximum level filter"
// @Success 200 {array} models.Zone
// @Router /zones [get]
func (h *ZoneHandler) ListZones(c *gin.Context) {
	// Récupérer les paramètres de filtrage optionnels
	zoneType := c.Query("type")
	minLevelStr := c.Query("min_level")
	maxLevelStr := c.Query("max_level")

	var zones []*models.Zone
	var err error

	// Appliquer les filtres
	if zoneType != "" {
		zones, err = h.zoneService.GetZonesByType(zoneType)
	} else if minLevelStr != "" && maxLevelStr != "" {
		minLevel, err1 := strconv.Atoi(minLevelStr)
		maxLevel, err2 := strconv.Atoi(maxLevelStr)

		if err1 != nil || err2 != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid level parameters"})
			return
		}

		zones, err = h.zoneService.GetZonesByLevel(minLevel, maxLevel)
	} else {
		zones, err = h.zoneService.ListZones()
	}

	if err != nil {
		logrus.WithError(err).Error("Failed to list zones")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve zones"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"zones": zones,
		"count": len(zones),
	})
}

// GetZone récupère une zone par son ID
// @Summary Get zone by ID
// @Description Get detailed information about a specific zone
// @Tags zones
// @Security BearerAuth
// @Produce json
// @Param id path string true "Zone ID"
// @Success 200 {object} models.Zone
// @Router /zones/{id} [get]
func (h *ZoneHandler) GetZone(c *gin.Context) {
	zoneID := c.Param("id")
	if zoneID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Zone ID is required"})
		return
	}

	zone, err := h.zoneService.GetZone(zoneID)
	if err != nil {
		logrus.WithError(err).WithField("zone_id", zoneID).Error("Failed to get zone")
		c.JSON(http.StatusNotFound, gin.H{"error": "Zone not found"})
		return
	}

	// Récupérer les statistiques de la zone
	stats, err := h.zoneService.GetZoneStatistics(zoneID)
	if err != nil {
		logrus.WithError(err).WithField("zone_id", zoneID).Warn("Failed to get zone statistics")
	}

	response := gin.H{
		"zone": zone,
	}

	if stats != nil {
		response["statistics"] = stats
	}

	c.JSON(http.StatusOK, response)
}

// EnterZone fait entrer un joueur dans une zone
// @Summary Enter a zone
// @Description Make a player enter a specific zone
// @Tags zones
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Zone ID"
// @Success 200 {object} models.Zone
// @Router /zones/{id}/enter [post]
func (h *ZoneHandler) EnterZone(c *gin.Context) {
	zoneID := c.Param("id")
	if zoneID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Zone ID is required"})
		return
	}

	// Récupérer l'ID utilisateur depuis le JWT
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	// Récupérer l'ID du personnage depuis le contexte ou les paramètres
	characterIDStr := c.GetHeader("X-Character-ID")
	if characterIDStr == "" {
		// Alternative: récupérer depuis le body ou query param
		var body struct {
			CharacterID string `json:"character_id"`
		}
		if err := c.ShouldBindJSON(&body); err == nil && body.CharacterID != "" {
			characterIDStr = body.CharacterID
		}
	}

	if characterIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Character ID is required"})
		return
	}

	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid character ID format"})
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Faire entrer le joueur dans la zone
	zone, err := h.zoneService.EnterZone(characterID, userUUID, zoneID)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"character_id": characterID,
			"user_id":      userUUID,
			"zone_id":      zoneID,
		}).Error("Failed to enter zone")

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logrus.WithFields(logrus.Fields{
		"character_id": characterID,
		"user_id":      userUUID,
		"zone_id":      zoneID,
	}).Info("Player entered zone successfully")

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully entered zone",
		"zone":    zone,
		"spawn_point": gin.H{
			"x": zone.SpawnX,
			"y": zone.SpawnY,
			"z": zone.SpawnZ,
		},
	})
}

// LeaveZone fait sortir un joueur d'une zone
// @Summary Leave a zone
// @Description Make a player leave a specific zone
// @Tags zones
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Zone ID"
// @Success 200 {string} string "Successfully left zone"
// @Router /zones/{id}/leave [post]
func (h *ZoneHandler) LeaveZone(c *gin.Context) {
	zoneID := c.Param("id")
	if zoneID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Zone ID is required"})
		return
	}

	// Récupérer l'ID du personnage
	characterIDStr := c.GetHeader("X-Character-ID")
	if characterIDStr == "" {
		var body struct {
			CharacterID string `json:"character_id"`
		}
		if err := c.ShouldBindJSON(&body); err == nil && body.CharacterID != "" {
			characterIDStr = body.CharacterID
		}
	}

	if characterIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Character ID is required"})
		return
	}

	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid character ID format"})
		return
	}

	// Faire sortir le joueur de la zone
	err = h.zoneService.LeaveZone(characterID, zoneID)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"character_id": characterID,
			"zone_id":      zoneID,
		}).Error("Failed to leave zone")

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logrus.WithFields(logrus.Fields{
		"character_id": characterID,
		"zone_id":      zoneID,
	}).Info("Player left zone successfully")

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully left zone",
	})
}

// GetPlayersInZone récupère les joueurs dans une zone
// @Summary Get players in zone
// @Description Get list of all players currently in a specific zone
// @Tags zones
// @Security BearerAuth
// @Produce json
// @Param id path string true "Zone ID"
// @Success 200 {array} models.PlayerPosition
// @Router /zones/{id}/players [get]
func (h *ZoneHandler) GetPlayersInZone(c *gin.Context) {
	zoneID := c.Param("id")
	if zoneID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Zone ID is required"})
		return
	}

	players, err := h.zoneService.GetPlayersInZone(zoneID)
	if err != nil {
		logrus.WithError(err).WithField("zone_id", zoneID).Error("Failed to get players in zone")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve players"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"players": players,
		"count":   len(players),
		"zone_id": zoneID,
	})
}

// GetNPCsInZone récupère les NPCs dans une zone
// @Summary Get NPCs in zone
// @Description Get list of all NPCs in a specific zone
// @Tags zones
// @Security BearerAuth
// @Produce json
// @Param id path string true "Zone ID"
// @Success 200 {array} models.NPC
// @Router /zones/{id}/npcs [get]
func (h *ZoneHandler) GetNPCsInZone(c *gin.Context) {
	zoneID := c.Param("id")
	if zoneID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Zone ID is required"})
		return
	}

	// TODO: Implémenter via NPCService
	// Pour l'instant, retourner une réponse vide
	c.JSON(http.StatusOK, gin.H{
		"npcs":    []models.NPC{},
		"count":   0,
		"zone_id": zoneID,
	})
}

// CreateZone crée une nouvelle zone (admin seulement)
// @Summary Create zone
// @Description Create a new zone (admin only)
// @Tags zones
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param zone body models.CreateZoneRequest true "Zone data"
// @Success 201 {object} models.Zone
// @Router /admin/zones [post]
func (h *ZoneHandler) CreateZone(c *gin.Context) {
	var req models.CreateZoneRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	zone, err := h.zoneService.CreateZone(req)
	if err != nil {
		logrus.WithError(err).WithField("zone_id", req.ID).Error("Failed to create zone")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logrus.WithFields(logrus.Fields{
		"zone_id":   zone.ID,
		"zone_name": zone.Name,
		"zone_type": zone.Type,
	}).Info("Zone created successfully")

	c.JSON(http.StatusCreated, gin.H{
		"message": "Zone created successfully",
		"zone":    zone,
	})
}

// UpdateZone met à jour une zone (admin seulement)
// @Summary Update zone
// @Description Update an existing zone (admin only)
// @Tags zones
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Zone ID"
// @Param zone body models.UpdateZoneRequest true "Zone update data"
// @Success 200 {object} models.Zone
// @Router /admin/zones/{id} [put]
func (h *ZoneHandler) UpdateZone(c *gin.Context) {
	zoneID := c.Param("id")
	if zoneID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Zone ID is required"})
		return
	}

	var req models.UpdateZoneRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	zone, err := h.zoneService.UpdateZone(zoneID, req)
	if err != nil {
		logrus.WithError(err).WithField("zone_id", zoneID).Error("Failed to update zone")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logrus.WithFields(logrus.Fields{
		"zone_id":   zone.ID,
		"zone_name": zone.Name,
	}).Info("Zone updated successfully")

	c.JSON(http.StatusOK, gin.H{
		"message": "Zone updated successfully",
		"zone":    zone,
	})
}

// DeleteZone supprime une zone (admin seulement)
// @Summary Delete zone
// @Description Delete a zone (admin only)
// @Tags zones
// @Security BearerAuth
// @Param id path string true "Zone ID"
// @Success 200 {string} string "Zone deleted successfully"
// @Router /admin/zones/{id} [delete]
func (h *ZoneHandler) DeleteZone(c *gin.Context) {
	zoneID := c.Param("id")
	if zoneID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Zone ID is required"})
		return
	}

	err := h.zoneService.DeleteZone(zoneID)
	if err != nil {
		logrus.WithError(err).WithField("zone_id", zoneID).Error("Failed to delete zone")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logrus.WithField("zone_id", zoneID).Info("Zone deleted successfully")

	c.JSON(http.StatusOK, gin.H{
		"message": "Zone deleted successfully",
	})
}
