package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/dan-2/github_mmorpg/services/analytics/internal/models"
	"github.com/dan-2/github_mmorpg/services/analytics/internal/service"
)

type AnalyticsHandler struct {
	analyticsService service.AnalyticsService
	metricsService   service.MetricsService
	loggingService   service.LoggingService
}

func NewAnalyticsHandler(
	analyticsService service.AnalyticsService,
	metricsService service.MetricsService,
	loggingService service.LoggingService,
) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
		metricsService:   metricsService,
		loggingService:   loggingService,
	}
}

// TrackEvent enregistre un événement
func (h *AnalyticsHandler) TrackEvent(c *gin.Context) {
	var req models.CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.analyticsService.TrackEvent(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de l'enregistrement de l'événement"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Événement enregistré avec succès"})
}

// GetEvents récupère les événements avec filtres
func (h *AnalyticsHandler) GetEvents(c *gin.Context) {
	eventType := c.Query("type")
	fromStr := c.Query("from")
	toStr := c.Query("to")
	playerIDStr := c.Query("player_id")
	guildIDStr := c.Query("guild_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	var from, to *time.Time
	var playerID, guildID *uuid.UUID

	if fromStr != "" {
		if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
			from = &t
		}
	}

	if toStr != "" {
		if t, err := time.Parse(time.RFC3339, toStr); err == nil {
			to = &t
		}
	}

	if playerIDStr != "" {
		if id, err := uuid.Parse(playerIDStr); err == nil {
			playerID = &id
		}
	}

	if guildIDStr != "" {
		if id, err := uuid.Parse(guildIDStr); err == nil {
			guildID = &id
		}
	}

	var eventTypePtr *string
	if eventType != "" {
		eventTypePtr = &eventType
	}

	events, total, err := h.analyticsService.GetEvents(c.Request.Context(), eventTypePtr, from, to, playerID, guildID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des événements"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events": events,
		"total":  total,
		"page":   page,
		"limit":  limit,
	})
}

// GetEvent récupère un événement spécifique
func (h *AnalyticsHandler) GetEvent(c *gin.Context) {
	eventIDStr := c.Param("id")
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID d'événement invalide"})
		return
	}

	event, err := h.analyticsService.GetEvent(c.Request.Context(), eventID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Événement non trouvé"})
		return
	}

	c.JSON(http.StatusOK, event)
}

// RecordMetric enregistre une métrique
func (h *AnalyticsHandler) RecordMetric(c *gin.Context) {
	var req struct {
		Name  string            `json:"name" binding:"required"`
		Value float64           `json:"value" binding:"required"`
		Tags  map[string]string `json:"tags,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.metricsService.RecordMetric(c.Request.Context(), req.Name, req.Value, req.Tags)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de l'enregistrement de la métrique"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Métrique enregistrée avec succès"})
}

// GetMetrics récupère les métriques
func (h *AnalyticsHandler) GetMetrics(c *gin.Context) {
	var req models.GetMetricsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	metrics, err := h.metricsService.GetMetrics(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des métriques"})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// Log enregistre un log
func (h *AnalyticsHandler) Log(c *gin.Context) {
	var req struct {
		Level   string                 `json:"level" binding:"required"`
		Message string                 `json:"message" binding:"required"`
		Context map[string]interface{} `json:"context,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.loggingService.Log(c.Request.Context(), req.Level, req.Message, req.Context)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de l'enregistrement du log"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Log enregistré avec succès"})
}

// GetLogs récupère les logs
func (h *AnalyticsHandler) GetLogs(c *gin.Context) {
	var req models.GetLogsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logs, total, err := h.loggingService.GetLogs(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"total": total,
		"page":  req.Page,
		"limit": req.Limit,
	})
}
