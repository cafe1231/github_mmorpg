package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"world/internal/models"
)

// Stubs pour les autres handlers en attendant leur implémentation complète

// NPCHandler gère les requêtes HTTP pour les NPCs
type NPCHandler struct{}

func NewNPCHandler() *NPCHandler {
	return &NPCHandler{}
}

func (h *NPCHandler) ListNPCs(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"npcs":  []models.NPC{},
		"count": 0,
	})
}

func (h *NPCHandler) GetNPC(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"error": "NPC not found"})
}

func (h *NPCHandler) InteractWithNPC(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "NPC interaction not implemented yet"})
}

func (h *NPCHandler) GetNPCsByZone(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"npcs":    []models.NPC{},
		"count":   0,
		"zone_id": c.Param("zoneId"),
	})
}

func (h *NPCHandler) CreateNPC(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "NPC creation not implemented yet"})
}

func (h *NPCHandler) UpdateNPC(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "NPC update not implemented yet"})
}

func (h *NPCHandler) DeleteNPC(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "NPC deletion not implemented yet"})
}

// WorldEventHandler gère les requêtes HTTP pour les événements du monde
type WorldEventHandler struct{}

func NewWorldEventHandler() *WorldEventHandler {
	return &WorldEventHandler{}
}

func (h *WorldEventHandler) ListEvents(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"events": []models.WorldEvent{},
		"count":  0,
	})
}

func (h *WorldEventHandler) GetActiveEvents(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"events": []models.WorldEvent{},
		"count":  0,
	})
}

func (h *WorldEventHandler) GetZoneEvents(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"events":  []models.WorldEvent{},
		"count":   0,
		"zone_id": c.Param("zoneId"),
	})
}

func (h *WorldEventHandler) ParticipateInEvent(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Event participation not implemented yet"})
}

func (h *WorldEventHandler) CreateEvent(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Event creation not implemented yet"})
}

func (h *WorldEventHandler) UpdateEvent(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Event update not implemented yet"})
}

func (h *WorldEventHandler) DeleteEvent(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Event deletion not implemented yet"})
}

// WeatherHandler gère les requêtes HTTP pour la météo
type WeatherHandler struct{}

func NewWeatherHandler() *WeatherHandler {
	return &WeatherHandler{}
}

func (h *WeatherHandler) GetZoneWeather(c *gin.Context) {
	zoneID := c.Param("zoneId")
	c.JSON(http.StatusOK, gin.H{
		"weather": models.Weather{
			ZoneID:      zoneID,
			Type:        "clear",
			Intensity:   0.3,
			Temperature: 22.0,
			WindSpeed:   5.0,
			WindDirection: 180.0,
			Visibility:  1000.0,
			IsActive:    true,
		},
	})
}

func (h *WeatherHandler) GetWeatherForecast(c *gin.Context) {
	zoneID := c.Param("zoneId")
	c.JSON(http.StatusOK, gin.H{
		"forecast": []models.Weather{
			{
				ZoneID:      zoneID,
				Type:        "clear",
				Intensity:   0.3,
				Temperature: 22.0,
				WindSpeed:   5.0,
				Visibility:  1000.0,
				IsActive:    false,
			},
		},
		"zone_id": zoneID,
	})
}

func (h *WeatherHandler) SetWeather(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Weather setting not implemented yet"})
}