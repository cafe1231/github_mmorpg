package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"combat/internal/config"
	"combat/internal/models"
	"combat/internal/service"
)

// PvPHandler gère les requêtes HTTP pour le PvP
type PvPHandler struct {
	pvpService service.PvPServiceInterface
	config     *config.Config
}

// NewPvPHandler crée un nouveau handler PvP
func NewPvPHandler(pvpService service.PvPServiceInterface, config *config.Config) *PvPHandler {
	return &PvPHandler{
		pvpService: pvpService,
		config:     config,
	}
}

// CreateChallenge crée un défi PvP
// @Summary Créer un défi PvP
// @Description Crée un nouveau défi PvP entre joueurs
// @Tags pvp
// @Accept json
// @Produce json
// @Param request body models.CreateChallengeRequest true "Paramètres du défi"
// @Success 201 {object} models.PvPChallenge
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/pvp/challenge [post]
func (h *PvPHandler) CreateChallenge(c *gin.Context) {
	var req models.CreateChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid request body",
			"details":    err.Error(),
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	// Récupérer l'ID du challenger depuis le contexte
	challengerID, err := uuid.Parse(c.GetString("character_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid challenger ID",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	challenge, err := h.pvpService.CreateChallenge(challengerID, &req)
	if err != nil {
		logrus.WithError(err).Error("Failed to create PvP challenge")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      err.Error(),
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	c.JSON(http.StatusCreated, challenge)
}

// GetChallenges récupère les défis d'un joueur
// @Summary Liste des défis
// @Description Récupère les défis envoyés et reçus par un joueur
// @Tags pvp
// @Produce json
// @Param status query string false "Statut des défis (pending, accepted, declined, expired)"
// @Param type query string false "Type de défis (sent, received, all)"
// @Success 200 {array} models.PvPChallenge
// @Failure 400 {object} models.ErrorResponse
// @Router /api/v1/pvp/challenges [get]
func (h *PvPHandler) GetChallenges(c *gin.Context) {
	playerID, err := uuid.Parse(c.GetString("character_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid player ID",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	req := models.GetChallengesRequest{
		PlayerID: playerID,
		Status:   c.Query("status"),
		Type:     c.Query("type"),
	}

	challenges, err := h.pvpService.GetChallenges(&req)
	if err != nil {
		logrus.WithError(err).Error("Failed to get challenges")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to retrieve challenges",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	c.JSON(http.StatusOK, challenges)
}

// GetChallenge récupère un défi spécifique
// @Summary Détails d'un défi
// @Description Récupère les détails d'un défi PvP
// @Tags pvp
// @Produce json
// @Param id path string true "ID du défi"
// @Success 200 {object} models.PvPChallenge
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/pvp/challenges/{id} [get]
func (h *PvPHandler) GetChallenge(c *gin.Context) {
	idStr := c.Param("id")
	challengeID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid challenge ID",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	challenge, err := h.pvpService.GetChallenge(challengeID)
	if err != nil {
		if err.Error() == "challenge not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":      "Challenge not found",
				"request_id": c.GetHeader("X-Request-ID"),
			})
			return
		}

		logrus.WithError(err).Error("Failed to get challenge")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to retrieve challenge",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	c.JSON(http.StatusOK, challenge)
}

// RespondToChallenge répond à un défi PvP
// @Summary Répondre à un défi
// @Description Accepte ou refuse un défi PvP
// @Tags pvp
// @Accept json
// @Produce json
// @Param id path string true "ID du défi"
// @Param request body models.RespondToChallengeRequest true "Réponse au défi"
// @Success 200 {object} models.ChallengeResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/pvp/challenges/{id}/respond [post]
func (h *PvPHandler) RespondToChallenge(c *gin.Context) {
	idStr := c.Param("id")
	challengeID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid challenge ID",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	var req models.RespondToChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid request body",
			"details":    err.Error(),
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	// Récupérer l'ID du joueur depuis le contexte
	playerID, err := uuid.Parse(c.GetString("character_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid player ID",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}
	req.PlayerID = playerID

	response, err := h.pvpService.RespondToChallenge(challengeID, &req)
	if err != nil {
		if err.Error() == "challenge not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":      "Challenge not found",
				"request_id": c.GetHeader("X-Request-ID"),
			})
			return
		}

		logrus.WithError(err).Error("Failed to respond to challenge")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      err.Error(),
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// CancelChallenge annule un défi PvP
// @Summary Annuler un défi
// @Description Annule un défi PvP en attente
// @Tags pvp
// @Produce json
// @Param id path string true "ID du défi"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/pvp/challenges/{id} [delete]
func (h *PvPHandler) CancelChallenge(c *gin.Context) {
	idStr := c.Param("id")
	challengeID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid challenge ID",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	playerID, err := uuid.Parse(c.GetString("character_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid player ID",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	if err := h.pvpService.CancelChallenge(challengeID, playerID); err != nil {
		if err.Error() == "challenge not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":      "Challenge not found",
				"request_id": c.GetHeader("X-Request-ID"),
			})
			return
		}

		logrus.WithError(err).Error("Failed to cancel challenge")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      err.Error(),
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Challenge canceled successfully",
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

// GetRankings récupère les classements PvP
// @Summary Classements PvP
// @Description Récupère les classements PvP par saison
// @Tags pvp
// @Produce json
// @Param season query string false "Saison (défaut: current)"
// @Param limit query int false "Nombre de résultats (défaut: 50, max: 100)"
// @Param offset query int false "Décalage pour la pagination (défaut: 0)"
// @Success 200 {object} models.RankingsResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /api/v1/pvp/rankings [get]
func (h *PvPHandler) GetRankings(c *gin.Context) {
	req := models.GetRankingsRequest{
		Season: c.DefaultQuery("season", "current"),
		Limit:  50,
		Offset: 0,
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			req.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			req.Offset = offset
		}
	}

	rankings, err := h.pvpService.GetRankings(&req)
	if err != nil {
		logrus.WithError(err).Error("Failed to get rankings")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to retrieve rankings",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	c.JSON(http.StatusOK, rankings)
}

// GetPvPStatistics récupère les statistiques PvP d'un joueur
// @Summary Statistiques PvP
// @Description Récupère les statistiques PvP d'un joueur
// @Tags pvp
// @Produce json
// @Param characterId path string true "ID du personnage"
// @Param season query string false "Saison (défaut: current)"
// @Success 200 {object} models.PvPStatistics
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/pvp/statistics/{characterId} [get]
func (h *PvPHandler) GetPvPStatistics(c *gin.Context) {
	characterIDStr := c.Param("characterId")
	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid character ID",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	season := c.DefaultQuery("season", "current")

	stats, err := h.pvpService.GetPlayerStatistics(characterID, season)
	if err != nil {
		if err.Error() == "statistics not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":      "Statistics not found",
				"request_id": c.GetHeader("X-Request-ID"),
			})
			return
		}

		logrus.WithError(err).Error("Failed to get PvP statistics")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to retrieve statistics",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetSeasonInfo récupère les informations de la saison actuelle
// @Summary Informations de saison
// @Description Récupère les informations de la saison PvP actuelle
// @Tags pvp
// @Produce json
// @Success 200 {object} models.SeasonInfo
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/pvp/season [get]
func (h *PvPHandler) GetSeasonInfo(c *gin.Context) {
	seasonInfo, err := h.pvpService.GetCurrentSeasonInfo()
	if err != nil {
		logrus.WithError(err).Error("Failed to get season info")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to retrieve season information",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	c.JSON(http.StatusOK, seasonInfo)
}

// JoinQueue rejoint la file d'attente PvP
// @Summary Rejoindre la file d'attente
// @Description Rejoint la file d'attente pour le matchmaking PvP
// @Tags pvp
// @Accept json
// @Produce json
// @Param request body models.JoinQueueRequest true "Paramètres de la file d'attente"
// @Success 200 {object} models.QueueResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /api/v1/pvp/queue [post]
func (h *PvPHandler) JoinQueue(c *gin.Context) {
	var req models.JoinQueueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid request body",
			"details":    err.Error(),
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	playerID, err := uuid.Parse(c.GetString("character_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid player ID",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}
	req.PlayerID = playerID

	response, err := h.pvpService.JoinQueue(&req)
	if err != nil {
		logrus.WithError(err).Error("Failed to join queue")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      err.Error(),
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// LeaveQueue quitte la file d'attente PvP
// @Summary Quitter la file d'attente
// @Description Quitte la file d'attente PvP
// @Tags pvp
// @Produce json
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /api/v1/pvp/queue [delete]
func (h *PvPHandler) LeaveQueue(c *gin.Context) {
	playerID, err := uuid.Parse(c.GetString("character_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid player ID",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	if err := h.pvpService.LeaveQueue(playerID); err != nil {
		logrus.WithError(err).Error("Failed to leave queue")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      err.Error(),
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Successfully left queue",
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

// GetQueueStatus récupère le statut de la file d'attente
// @Summary Statut de la file d'attente
// @Description Récupère le statut actuel de la file d'attente du joueur
// @Tags pvp
// @Produce json
// @Success 200 {object} models.QueueStatus
// @Failure 400 {object} models.ErrorResponse
// @Router /api/v1/pvp/queue/status [get]
func (h *PvPHandler) GetQueueStatus(c *gin.Context) {
	playerID, err := uuid.Parse(c.GetString("character_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid player ID",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	status, err := h.pvpService.GetQueueStatus(playerID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get queue status")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to retrieve queue status",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	c.JSON(http.StatusOK, status)
}
