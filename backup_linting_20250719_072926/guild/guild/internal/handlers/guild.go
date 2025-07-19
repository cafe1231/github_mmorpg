package handlers

import (
	"net/http"
	"strconv"

	"github.com/dan-2/github_mmorpg/services/guild/internal/models"
	"github.com/dan-2/github_mmorpg/services/guild/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GuildHandler gère les requêtes liées aux guildes
type GuildHandler struct {
	guildService service.GuildService
}

// NewGuildHandler crée une nouvelle instance de GuildHandler
func NewGuildHandler(guildService service.GuildService) *GuildHandler {
	return &GuildHandler{
		guildService: guildService,
	}
}

// CreateGuild crée une nouvelle guilde
func (h *GuildHandler) CreateGuild(c *gin.Context) {
	var req models.CreateGuildRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Récupérer l'ID du joueur depuis le contexte (supposé être défini par le middleware d'auth)
	playerIDStr := c.GetString("player_id")
	if playerIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Joueur non authentifié"})
		return
	}

	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de joueur invalide"})
		return
	}

	guild, err := h.guildService.CreateGuild(c.Request.Context(), &req, playerID)
	if err != nil {
		switch err {
		case models.ErrGuildAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case models.ErrAlreadyInGuild:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur interne du serveur"})
		}
		return
	}

	c.JSON(http.StatusCreated, guild)
}

// GetGuild récupère une guilde
func (h *GuildHandler) GetGuild(c *gin.Context) {
	guildIDStr := c.Param("id")
	guildID, err := uuid.Parse(guildIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de guilde invalide"})
		return
	}

	guild, err := h.guildService.GetGuild(c.Request.Context(), guildID)
	if err != nil {
		if err == models.ErrGuildNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur interne du serveur"})
		}
		return
	}

	c.JSON(http.StatusOK, guild)
}

// UpdateGuild met à jour une guilde
func (h *GuildHandler) UpdateGuild(c *gin.Context) {
	guildIDStr := c.Param("id")
	guildID, err := uuid.Parse(guildIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de guilde invalide"})
		return
	}

	var req models.UpdateGuildRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	playerIDStr := c.GetString("player_id")
	if playerIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Joueur non authentifié"})
		return
	}

	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de joueur invalide"})
		return
	}

	guild, err := h.guildService.UpdateGuild(c.Request.Context(), guildID, &req, playerID)
	if err != nil {
		switch err {
		case models.ErrGuildNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case models.ErrInsufficientPermissions:
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case models.ErrGuildAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur interne du serveur"})
		}
		return
	}

	c.JSON(http.StatusOK, guild)
}

// DeleteGuild supprime une guilde
func (h *GuildHandler) DeleteGuild(c *gin.Context) {
	guildIDStr := c.Param("id")
	guildID, err := uuid.Parse(guildIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de guilde invalide"})
		return
	}

	playerIDStr := c.GetString("player_id")
	if playerIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Joueur non authentifié"})
		return
	}

	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de joueur invalide"})
		return
	}

	err = h.guildService.DeleteGuild(c.Request.Context(), guildID, playerID)
	if err != nil {
		switch err {
		case models.ErrGuildNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case models.ErrInsufficientPermissions:
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur interne du serveur"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Guilde supprimée avec succès"})
}

// SearchGuilds recherche des guildes
func (h *GuildHandler) SearchGuilds(c *gin.Context) {
	var req models.GuildSearchRequest

	// Récupérer les paramètres de requête
	if name := c.Query("name"); name != "" {
		req.Name = &name
	}
	if tag := c.Query("tag"); tag != "" {
		req.Tag = &tag
	}
	if minLevelStr := c.Query("min_level"); minLevelStr != "" {
		if minLevel, err := strconv.Atoi(minLevelStr); err == nil {
			req.MinLevel = &minLevel
		}
	}
	if maxLevelStr := c.Query("max_level"); maxLevelStr != "" {
		if maxLevel, err := strconv.Atoi(maxLevelStr); err == nil {
			req.MaxLevel = &maxLevel
		}
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	req.Page = page
	req.Limit = limit

	result, err := h.guildService.SearchGuilds(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur interne du serveur"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ListGuilds liste les guildes
func (h *GuildHandler) ListGuilds(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	result, err := h.guildService.ListGuilds(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur interne du serveur"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetGuildStats récupère les statistiques d'une guilde
func (h *GuildHandler) GetGuildStats(c *gin.Context) {
	guildIDStr := c.Param("id")
	guildID, err := uuid.Parse(guildIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de guilde invalide"})
		return
	}

	stats, err := h.guildService.GetGuildStats(c.Request.Context(), guildID)
	if err != nil {
		if err == models.ErrGuildNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur interne du serveur"})
		}
		return
	}

	c.JSON(http.StatusOK, stats)
}
