package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"world/internal/models"
	"world/internal/service"
)

// PlayerPositionHandler gère les requêtes HTTP pour les positions des joueurs
type PlayerPositionHandler struct {
	positionService *service.PlayerPositionService
}

// NewPlayerPositionHandler crée un nouveau handler de position
func NewPlayerPositionHandler(positionService *service.PlayerPositionService) *PlayerPositionHandler {
	return &PlayerPositionHandler{
		positionService: positionService,
	}
}

// GetCharacterPosition récupère la position d'un personnage
// @Summary Get character position
// @Description Get current position of a specific character
// @Tags positions
// @Security BearerAuth
// @Produce json
// @Param characterId path string true "Character ID"
// @Success 200 {object} models.PlayerPosition
// @Router /positions/character/{characterId} [get]
func (h *PlayerPositionHandler) GetCharacterPosition(c *gin.Context) {
	characterIDStr := c.Param("characterId")
	if characterIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Character ID is required"})
		return
	}

	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid character ID format"})
		return
	}

	// Vérifier que l'utilisateur peut accéder à ce personnage
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	position, err := h.positionService.GetCharacterPosition(characterID)
	if err != nil {
		logrus.WithError(err).WithField("character_id", characterID).Error("Failed to get character position")
		c.JSON(http.StatusNotFound, gin.H{"error": "Character position not found"})
		return
	}

	// Vérifier que le personnage appartient à l'utilisateur
	userUUID, _ := uuid.Parse(userID.(string))
	if position.UserID != userUUID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to this character"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"position": position,
	})
}

// UpdateCharacterPosition met à jour la position d'un personnage
// @Summary Update character position
// @Description Update the position of a specific character
// @Tags positions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param characterId path string true "Character ID"
// @Param position body models.UpdatePositionRequest true "Position data"
// @Success 200 {object} models.PlayerPosition
// @Router /positions/character/{characterId} [put]
func (h *PlayerPositionHandler) UpdateCharacterPosition(c *gin.Context) {
	characterIDStr := c.Param("characterId")
	if characterIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Character ID is required"})
		return
	}

	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid character ID format"})
		return
	}

	// Récupérer l'ID utilisateur depuis le JWT
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Bind de la requête
	var req models.UpdatePositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Mettre à jour la position
	position, err := h.positionService.UpdatePosition(characterID, userUUID, req)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"character_id": characterID,
			"user_id":      userUUID,
			"zone_id":      req.ZoneID,
		}).Error("Failed to update character position")

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Diffuser la position aux joueurs proches (en arrière-plan)
	go func() {
		if err := h.positionService.BroadcastPosition(characterID); err != nil {
			logrus.WithError(err).WithField("character_id", characterID).Warn("Failed to broadcast position")
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"position": position,
		"message":  "Position updated successfully",
	})
}

// GetZonePositions récupère toutes les positions dans une zone
// @Summary Get zone positions
// @Description Get all player positions in a specific zone
// @Tags positions
// @Security BearerAuth
// @Produce json
// @Param zoneId path string true "Zone ID"
// @Param include_self query bool false "Include requesting player in results"
// @Success 200 {array} models.PlayerPosition
// @Router /positions/zone/{zoneId} [get]
func (h *PlayerPositionHandler) GetZonePositions(c *gin.Context) {
	zoneID := c.Param("zoneId")
	if zoneID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Zone ID is required"})
		return
	}

	positions, err := h.positionService.GetZonePositions(zoneID)
	if err != nil {
		logrus.WithError(err).WithField("zone_id", zoneID).Error("Failed to get zone positions")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve positions"})
		return
	}

	// Filtrer le joueur demandeur si nécessaire
	includeSelf := c.Query("include_self") == "true"
	if !includeSelf {
		userID, exists := c.Get("user_id")
		if exists {
			userUUID, err := uuid.Parse(userID.(string))
			if err == nil {
				var filteredPositions []*models.PlayerPosition
				for _, pos := range positions {
					if pos.UserID != userUUID {
						filteredPositions = append(filteredPositions, pos)
					}
				}
				positions = filteredPositions
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"positions": positions,
		"count":     len(positions),
		"zone_id":   zoneID,
	})
}

// GetNearbyPlayers récupère les joueurs proches
// @Summary Get nearby players
// @Description Get players near the requesting character
// @Tags positions
// @Security BearerAuth
// @Produce json
// @Param radius query float64 false "Search radius in meters" default(50)
// @Success 200 {array} models.PlayerPosition
// @Router /positions/nearby [get]
func (h *PlayerPositionHandler) GetNearbyPlayers(c *gin.Context) {
	// Récupérer l'ID du personnage depuis l'en-tête
	characterIDStr := c.GetHeader("X-Character-ID")
	if characterIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Character ID header is required"})
		return
	}

	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid character ID format"})
		return
	}

	// Récupérer le rayon de recherche
	radiusStr := c.DefaultQuery("radius", "50")
	radius, err := strconv.ParseFloat(radiusStr, 64)
	if err != nil || radius <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid radius parameter"})
		return
	}

	// Limiter le rayon maximum
	if radius > 200 {
		radius = 200
	}

	nearbyPlayers, err := h.positionService.GetNearbyPlayers(characterID, radius)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"character_id": characterID,
			"radius":       radius,
		}).Error("Failed to get nearby players")

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve nearby players"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"players": nearbyPlayers,
		"count":   len(nearbyPlayers),
		"radius":  radius,
	})
}

// SetPlayerOnline marque un joueur comme en ligne
// @Summary Set player online
// @Description Mark a player as online
// @Tags positions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param characterId path string true "Character ID"
// @Success 200 {string} string "Player set online"
// @Router /positions/character/{characterId}/online [post]
func (h *PlayerPositionHandler) SetPlayerOnline(c *gin.Context) {
	characterIDStr := c.Param("characterId")
	if characterIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Character ID is required"})
		return
	}

	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid character ID format"})
		return
	}

	// Vérifier l'autorisation
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	// Vérifier que le personnage appartient à l'utilisateur
	position, err := h.positionService.GetCharacterPosition(characterID)
	if err == nil {
		userUUID, _ := uuid.Parse(userID.(string))
		if position.UserID != userUUID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to this character"})
			return
		}
	}

	err = h.positionService.SetPlayerOnline(characterID)
	if err != nil {
		logrus.WithError(err).WithField("character_id", characterID).Error("Failed to set player online")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update player status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Player set online successfully",
	})
}

// SetPlayerOffline marque un joueur comme hors ligne
// @Summary Set player offline
// @Description Mark a player as offline
// @Tags positions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param characterId path string true "Character ID"
// @Success 200 {string} string "Player set offline"
// @Router /positions/character/{characterId}/offline [post]
func (h *PlayerPositionHandler) SetPlayerOffline(c *gin.Context) {
	characterIDStr := c.Param("characterId")
	if characterIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Character ID is required"})
		return
	}

	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid character ID format"})
		return
	}

	// Vérifier l'autorisation
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	// Vérifier que le personnage appartient à l'utilisateur
	position, err := h.positionService.GetCharacterPosition(characterID)
	if err == nil {
		userUUID, _ := uuid.Parse(userID.(string))
		if position.UserID != userUUID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to this character"})
			return
		}
	}

	err = h.positionService.SetPlayerOffline(characterID)
	if err != nil {
		logrus.WithError(err).WithField("character_id", characterID).Error("Failed to set player offline")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update player status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Player set offline successfully",
	})
}

// GetPlayerStatistics récupère les statistiques des joueurs
// @Summary Get player statistics
// @Description Get statistics about online players and zone distribution
// @Tags positions
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /positions/statistics [get]
func (h *PlayerPositionHandler) GetPlayerStatistics(c *gin.Context) {
	stats, err := h.positionService.GetPlayerStatistics()
	if err != nil {
		logrus.WithError(err).Error("Failed to get player statistics")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve statistics"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// TeleportPlayer téléporte un joueur (admin seulement)
// @Summary Teleport player
// @Description Teleport a player to a specific location (admin only)
// @Tags positions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param characterId path string true "Character ID"
// @Success 200 {string} string "Player teleported successfully"
// @Router /admin/positions/character/{characterId}/teleport [post]
func (h *PlayerPositionHandler) TeleportPlayer(c *gin.Context) {
	characterIDStr := c.Param("characterId")
	if characterIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Character ID is required"})
		return
	}

	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid character ID format"})
		return
	}

	// Récupérer les données de téléportation
	var req struct {
		ZoneID string  `json:"zone_id" binding:"required"`
		X      float64 `json:"x" binding:"required"`
		Y      float64 `json:"y" binding:"required"`
		Z      float64 `json:"z" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Récupérer l'ID utilisateur
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Effectuer la téléportation
	err = h.positionService.TeleportPlayer(characterID, userUUID, req.ZoneID, req.X, req.Y, req.Z)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"character_id": characterID,
			"zone_id":      req.ZoneID,
			"position":     fmt.Sprintf("(%.2f, %.2f, %.2f)", req.X, req.Y, req.Z),
		}).Error("Failed to teleport player")

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logrus.WithFields(logrus.Fields{
		"character_id": characterID,
		"zone_id":      req.ZoneID,
		"position":     fmt.Sprintf("(%.2f, %.2f, %.2f)", req.X, req.Y, req.Z),
	}).Info("Player teleported successfully")

	c.JSON(http.StatusOK, gin.H{
		"message": "Player teleported successfully",
		"destination": gin.H{
			"zone_id": req.ZoneID,
			"x":       req.X,
			"y":       req.Y,
			"z":       req.Z,
		},
	})
}

// CheckCollisions vérifie les collisions avec d'autres joueurs
// @Summary Check collisions
// @Description Check for collisions with other players at a specific position
// @Tags positions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param characterId path string true "Character ID"
// @Success 200 {array} models.PlayerPosition
// @Router /positions/character/{characterId}/collisions [post]
func (h *PlayerPositionHandler) CheckCollisions(c *gin.Context) {
	characterIDStr := c.Param("characterId")
	if characterIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Character ID is required"})
		return
	}

	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid character ID format"})
		return
	}

	// Récupérer les données de position pour vérifier les collisions
	var req struct {
		X      float64 `json:"x" binding:"required"`
		Y      float64 `json:"y" binding:"required"`
		Z      float64 `json:"z" binding:"required"`
		Radius float64 `json:"radius" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Limiter le rayon de collision
	if req.Radius > 10.0 {
		req.Radius = 10.0
	}
	if req.Radius <= 0 {
		req.Radius = 1.0
	}

	// Vérifier l'autorisation
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	// Vérifier que le personnage appartient à l'utilisateur
	position, err := h.positionService.GetCharacterPosition(characterID)
	if err == nil {
		userUUID, _ := uuid.Parse(userID.(string))
		if position.UserID != userUUID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to this character"})
			return
		}
	}

	// Vérifier les collisions
	collisions, err := h.positionService.CheckCollisions(characterID, req.X, req.Y, req.Z, req.Radius)
	if err != nil {
		logrus.WithError(err).WithField("character_id", characterID).Error("Failed to check collisions")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check collisions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"collisions": collisions,
		"count":      len(collisions),
		"radius":     req.Radius,
	})
}

// GetMovementHistory récupère l'historique des mouvements
// @Summary Get movement history
// @Description Get movement history for a character
// @Tags positions
// @Security BearerAuth
// @Produce json
// @Param characterId path string true "Character ID"
// @Param limit query int false "Number of records to return" default(10)
// @Success 200 {array} map[string]interface{}
// @Router /positions/character/{characterId}/history [get]
func (h *PlayerPositionHandler) GetMovementHistory(c *gin.Context) {
	characterIDStr := c.Param("characterId")
	if characterIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Character ID is required"})
		return
	}

	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid character ID format"})
		return
	}

	// Récupérer la limite
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	// Vérifier l'autorisation
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	// Vérifier que le personnage appartient à l'utilisateur
	position, err := h.positionService.GetCharacterPosition(characterID)
	if err == nil {
		userUUID, _ := uuid.Parse(userID.(string))
		if position.UserID != userUUID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to this character"})
			return
		}
	}

	// Récupérer l'historique
	history, err := h.positionService.GetMovementHistory(characterID, limit)
	if err != nil {
		logrus.WithError(err).WithField("character_id", characterID).Error("Failed to get movement history")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve movement history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"history": history,
		"count":   len(history),
		"limit":   limit,
	})
}
