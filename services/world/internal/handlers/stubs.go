package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"world/internal/models"
	"world/internal/service"
)

// NPCHandler gère les requêtes HTTP pour les NPCs
type NPCHandler struct {
	npcService  *service.NPCService
	zoneService *service.ZoneService
}

func NewNPCHandler(npcService *service.NPCService, zoneService *service.ZoneService) *NPCHandler {
	return &NPCHandler{
		npcService:  npcService,
		zoneService: zoneService,
	}
}

func (h *NPCHandler) ListNPCs(c *gin.Context) {
	// Paramètres de pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	npcType := c.Query("type")
	zoneID := c.Query("zone_id")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	// Si une zone spécifique est demandée
	if zoneID != "" {
		npcs, err := h.npcService.GetNPCsByZone(c.Request.Context(), zoneID, npcType)
		if err != nil {
			logrus.WithError(err).Error("Failed to get NPCs by zone")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve NPCs"})
			return
		}

		// Pagination manuelle
		offset := (page - 1) * limit
		end := offset + limit
		total := len(npcs)

		if offset >= total {
			c.JSON(http.StatusOK, gin.H{
				"npcs":    []models.NPC{},
				"count":   0,
				"total":   total,
				"page":    page,
				"limit":   limit,
				"zone_id": zoneID,
			})
			return
		}

		if end > total {
			end = total
		}

		c.JSON(http.StatusOK, gin.H{
			"npcs":    npcs[offset:end],
			"count":   end - offset,
			"total":   total,
			"page":    page,
			"limit":   limit,
			"zone_id": zoneID,
		})
		return
	}

	// TODO: Implémenter GetAllNPCs avec pagination dans le service
	c.JSON(http.StatusOK, gin.H{
		"npcs":    []models.NPC{},
		"count":   0,
		"total":   0,
		"page":    page,
		"limit":   limit,
		"message": "Use zone_id parameter to filter by zone",
	})
}

func (h *NPCHandler) GetNPC(c *gin.Context) {
	npcID := c.Param("id")
	if npcID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "NPC ID is required"})
		return
	}

	// TODO: Implémenter GetNPCByID dans le service
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Get single NPC not implemented yet"})
}

func (h *NPCHandler) InteractWithNPC(c *gin.Context) {
	npcID := c.Param("id")
	if npcID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "NPC ID is required"})
		return
	}

	// Récupérer l'ID du joueur depuis le token JWT
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	logrus.WithFields(logrus.Fields{
		"npc_id":  npcID,
		"user_id": userID,
	}).Info("NPC interaction requested")

	// TODO: Implémenter la logique d'interaction NPC
	c.JSON(http.StatusOK, gin.H{
		"message":            "NPC interaction successful",
		"npc_id":             npcID,
		"user_id":            userID,
		"interaction_result": "Default interaction response",
	})
}

func (h *NPCHandler) GetNPCsByZone(c *gin.Context) {
	zoneID := c.Param("zoneId")
	if zoneID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Zone ID is required"})
		return
	}

	npcType := c.Query("type")
	x, _ := strconv.ParseFloat(c.Query("x"), 64)
	y, _ := strconv.ParseFloat(c.Query("y"), 64)
	z, _ := strconv.ParseFloat(c.Query("z"), 64)
	radius, _ := strconv.ParseFloat(c.Query("radius"), 64)

	var npcs []*models.NPC
	var err error

	// Si des coordonnées et un rayon sont fournis, chercher les NPCs proches
	if radius > 0 && (x != 0 || y != 0 || z != 0) {
		npcs, err = h.npcService.GetNearbyNPCs(c.Request.Context(), zoneID, x, y, z, radius)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"zone_id": zoneID,
				"x":       x, "y": y, "z": z,
				"radius": radius,
			}).Error("Failed to get nearby NPCs")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve nearby NPCs"})
			return
		}
	} else {
		// Sinon, récupérer tous les NPCs de la zone
		npcs, err = h.npcService.GetNPCsByZone(c.Request.Context(), zoneID, npcType)
		if err != nil {
			logrus.WithError(err).WithField("zone_id", zoneID).Error("Failed to get NPCs by zone")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve NPCs"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"npcs":    npcs,
		"count":   len(npcs),
		"zone_id": zoneID,
	})
}

func (h *NPCHandler) CreateNPC(c *gin.Context) {
	var npc models.NPC
	if err := c.ShouldBindJSON(&npc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.npcService.CreateNPC(c.Request.Context(), &npc); err != nil {
		logrus.WithError(err).Error("Failed to create NPC")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create NPC"})
		return
	}

	logrus.WithField("npc_id", npc.ID).Info("NPC created successfully")
	c.JSON(http.StatusCreated, gin.H{
		"message": "NPC created successfully",
		"npc":     npc,
	})
}

func (h *NPCHandler) UpdateNPC(c *gin.Context) {
	npcIDStr := c.Param("id")
	if npcIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "NPC ID is required"})
		return
	}

	npcID, err := uuid.Parse(npcIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid NPC ID format"})
		return
	}

	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implémenter UpdateNPC dans le service avec des mises à jour partielles
	// Pour l'instant, on gère quelques cas spécifiques
	if health, exists := updateData["health"]; exists {
		if healthInt, ok := health.(float64); ok {
			err := h.npcService.UpdateNPCHealth(c.Request.Context(), npcID, int(healthInt))
			if err != nil {
				logrus.WithError(err).WithField("npc_id", npcID).Error("Failed to update NPC health")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update NPC health"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "NPC health updated successfully"})
			return
		}
	}

	if position, exists := updateData["position"]; exists {
		if posMap, ok := position.(map[string]interface{}); ok {
			x, _ := posMap["x"].(float64)
			y, _ := posMap["y"].(float64)
			z, _ := posMap["z"].(float64)
			rotation, _ := posMap["rotation"].(float64)

			err := h.npcService.UpdateNPCPosition(c.Request.Context(), npcID, x, y, z, rotation)
			if err != nil {
				logrus.WithError(err).WithField("npc_id", npcID).Error("Failed to update NPC position")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update NPC position"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "NPC position updated successfully"})
			return
		}
	}

	c.JSON(http.StatusBadRequest, gin.H{"error": "No valid update fields provided"})
}

func (h *NPCHandler) DeleteNPC(c *gin.Context) {
	npcIDStr := c.Param("id")
	if npcIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "NPC ID is required"})
		return
	}

	// TODO: Implémenter DeleteNPC dans le service
	logrus.WithField("npc_id", npcIDStr).Info("NPC deletion requested")
	c.JSON(http.StatusOK, gin.H{
		"message": "NPC deletion functionality will be implemented",
		"npc_id":  npcIDStr,
	})
}

// WorldEventHandler gère les requêtes HTTP pour les événements du monde
type WorldEventHandler struct {
	eventService *service.WorldEventService
}

func NewWorldEventHandler(eventService *service.WorldEventService) *WorldEventHandler {
	return &WorldEventHandler{
		eventService: eventService,
	}
}

func (h *WorldEventHandler) ListEvents(c *gin.Context) {
	status := c.Query("status")
	zoneID := c.Query("zone_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if limit < 1 || limit > 100 {
		limit = 20
	}

	var events []*models.WorldEvent
	var err error

	switch status {
	case "active":
		events, err = h.eventService.GetActiveEvents(c.Request.Context(), zoneID)
	case "upcoming":
		events, err = h.eventService.GetUpcomingEvents(c.Request.Context(), zoneID, limit)
	default:
		// Combiner les événements actifs et à venir
		activeEvents, activeErr := h.eventService.GetActiveEvents(c.Request.Context(), zoneID)
		upcomingEvents, upcomingErr := h.eventService.GetUpcomingEvents(c.Request.Context(), zoneID, limit)

		if activeErr != nil {
			err = activeErr
		} else if upcomingErr != nil {
			err = upcomingErr
		} else {
			events = append(activeEvents, upcomingEvents...)
		}
	}

	if err != nil {
		logrus.WithError(err).Error("Failed to get events")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events": events,
		"count":  len(events),
		"filter": gin.H{
			"status":  status,
			"zone_id": zoneID,
		},
	})
}

func (h *WorldEventHandler) GetActiveEvents(c *gin.Context) {
	zoneID := c.Query("zone_id")

	events, err := h.eventService.GetActiveEvents(c.Request.Context(), zoneID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get active events")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve active events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events": events,
		"count":  len(events),
	})
}

func (h *WorldEventHandler) GetZoneEvents(c *gin.Context) {
	zoneID := c.Param("zoneId")
	if zoneID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Zone ID is required"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Récupérer les événements actifs et à venir pour cette zone
	activeEvents, err := h.eventService.GetActiveEvents(c.Request.Context(), zoneID)
	if err != nil {
		logrus.WithError(err).WithField("zone_id", zoneID).Error("Failed to get active events for zone")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve zone events"})
		return
	}

	upcomingEvents, err := h.eventService.GetUpcomingEvents(c.Request.Context(), zoneID, limit)
	if err != nil {
		logrus.WithError(err).WithField("zone_id", zoneID).Error("Failed to get upcoming events for zone")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve zone events"})
		return
	}

	allEvents := append(activeEvents, upcomingEvents...)

	c.JSON(http.StatusOK, gin.H{
		"events":  allEvents,
		"count":   len(allEvents),
		"zone_id": zoneID,
		"breakdown": gin.H{
			"active":   len(activeEvents),
			"upcoming": len(upcomingEvents),
		},
	})
}

func (h *WorldEventHandler) ParticipateInEvent(c *gin.Context) {
	eventIDStr := c.Param("id")
	if eventIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Event ID is required"})
		return
	}

	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID format"})
		return
	}

	// Récupérer l'ID du joueur depuis le token JWT
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	logrus.WithFields(logrus.Fields{
		"event_id": eventID,
		"user_id":  userID,
	}).Info("Event participation requested")

	// TODO: Implémenter la logique de participation aux événements
	c.JSON(http.StatusOK, gin.H{
		"message":  "Event participation successful",
		"event_id": eventID,
		"user_id":  userID,
	})
}

func (h *WorldEventHandler) CreateEvent(c *gin.Context) {
	var event models.WorldEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.eventService.CreateEvent(c.Request.Context(), &event); err != nil {
		logrus.WithError(err).Error("Failed to create event")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event"})
		return
	}

	logrus.WithField("event_id", event.ID).Info("Event created successfully")
	c.JSON(http.StatusCreated, gin.H{
		"message": "Event created successfully",
		"event":   event,
	})
}

func (h *WorldEventHandler) UpdateEvent(c *gin.Context) {
	eventIDStr := c.Param("id")
	if eventIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Event ID is required"})
		return
	}

	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID format"})
		return
	}

	var action struct {
		Action string `json:"action"`
	}
	if err := c.ShouldBindJSON(&action); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	switch action.Action {
	case "start":
		err = h.eventService.StartEvent(c.Request.Context(), eventID)
	case "end":
		err = h.eventService.EndEvent(c.Request.Context(), eventID)
	case "cancel":
		err = h.eventService.CancelEvent(c.Request.Context(), eventID)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action. Use: start, end, or cancel"})
		return
	}

	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"event_id": eventID,
			"action":   action.Action,
		}).Error("Failed to update event")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update event"})
		return
	}

	logrus.WithFields(logrus.Fields{
		"event_id": eventID,
		"action":   action.Action,
	}).Info("Event updated successfully")

	c.JSON(http.StatusOK, gin.H{
		"message":  "Event updated successfully",
		"event_id": eventID,
		"action":   action.Action,
	})
}

func (h *WorldEventHandler) DeleteEvent(c *gin.Context) {
	eventIDStr := c.Param("id")
	if eventIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Event ID is required"})
		return
	}

	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID format"})
		return
	}

	if err := h.eventService.CancelEvent(c.Request.Context(), eventID); err != nil {
		logrus.WithError(err).WithField("event_id", eventID).Error("Failed to delete event")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete event"})
		return
	}

	logrus.WithField("event_id", eventID).Info("Event deleted successfully")
	c.JSON(http.StatusOK, gin.H{
		"message":  "Event deleted successfully",
		"event_id": eventID,
	})
}

// WeatherHandler gère les requêtes HTTP pour la météo
type WeatherHandler struct {
	weatherService *service.WeatherService
}

func NewWeatherHandler(weatherService *service.WeatherService) *WeatherHandler {
	return &WeatherHandler{
		weatherService: weatherService,
	}
}

func (h *WeatherHandler) GetZoneWeather(c *gin.Context) {
	zoneID := c.Param("zoneId")
	if zoneID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Zone ID is required"})
		return
	}

	weather, err := h.weatherService.GetCurrentWeather(c.Request.Context(), zoneID)
	if err != nil {
		logrus.WithError(err).WithField("zone_id", zoneID).Error("Failed to get zone weather")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve weather"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"weather": weather,
		"zone_id": zoneID,
	})
}

func (h *WeatherHandler) GetWeatherForecast(c *gin.Context) {
	zoneID := c.Param("zoneId")
	if zoneID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Zone ID is required"})
		return
	}

	// Pour l'instant, on retourne la météo actuelle comme "prévision"
	// TODO: Implémenter un vrai système de prévisions
	currentWeather, err := h.weatherService.GetCurrentWeather(c.Request.Context(), zoneID)
	if err != nil {
		logrus.WithError(err).WithField("zone_id", zoneID).Error("Failed to get weather forecast")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve weather forecast"})
		return
	}

	// Créer une prévision basique (météo actuelle + variations)
	forecast := []*models.Weather{currentWeather}

	c.JSON(http.StatusOK, gin.H{
		"forecast": forecast,
		"zone_id":  zoneID,
		"note":     "Enhanced forecast system coming soon",
	})
}

func (h *WeatherHandler) SetWeather(c *gin.Context) {
	zoneID := c.Param("zoneId")
	if zoneID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Zone ID is required"})
		return
	}

	var weatherRequest struct {
		Type          string  `json:"type" binding:"required"`
		Intensity     float64 `json:"intensity"`
		Temperature   float64 `json:"temperature"`
		WindSpeed     float64 `json:"wind_speed"`
		WindDirection float64 `json:"wind_direction"`
		Visibility    float64 `json:"visibility"`
		Duration      int     `json:"duration"` // en minutes
	}

	if err := c.ShouldBindJSON(&weatherRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Créer l'objet météo
	weather := &models.Weather{
		ZoneID:        zoneID,
		Type:          weatherRequest.Type,
		Intensity:     weatherRequest.Intensity,
		Temperature:   weatherRequest.Temperature,
		WindSpeed:     weatherRequest.WindSpeed,
		WindDirection: weatherRequest.WindDirection,
		Visibility:    weatherRequest.Visibility,
		StartTime:     time.Now(),
	}

	// Définir la durée
	if weatherRequest.Duration > 0 {
		weather.EndTime = time.Now().Add(time.Duration(weatherRequest.Duration) * time.Minute)
	}

	if err := h.weatherService.SetWeather(c.Request.Context(), weather); err != nil {
		logrus.WithError(err).WithField("zone_id", zoneID).Error("Failed to set weather")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set weather"})
		return
	}

	logrus.WithFields(logrus.Fields{
		"zone_id": zoneID,
		"type":    weather.Type,
	}).Info("Weather set successfully")

	c.JSON(http.StatusOK, gin.H{
		"message": "Weather set successfully",
		"weather": weather,
	})
}
