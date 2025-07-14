// internal/handler/combat.go
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"combat/internal/models"
	"combat/internal/service"
)

// CombatHandler gère les endpoints de combat
type CombatHandler struct {
	combatService service.CombatServiceInterface
}

// NewCombatHandler crée une nouvelle instance du handler combat
func NewCombatHandler(combatService service.CombatServiceInterface) *CombatHandler {
	return &CombatHandler{
		combatService: combatService,
	}
}

// StartCombat démarre un nouveau combat
func (h *CombatHandler) StartCombat(c *gin.Context) {
	var req models.StartCombatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid request body",
			"details":    err.Error(),
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	// Récupérer l'ID utilisateur depuis le middleware JWT
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":      "User ID not found in token",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	// Convertir userID en uuid.UUID
	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid user ID format",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	session, err := h.combatService.CreateCombatSession(userUUID, req)
	if err != nil {
		logrus.WithError(err).Error("Failed to create combat session")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to create combat session",
			"details":    err.Error(),
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"session":    session,
		"message":    "Combat session created successfully",
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

// JoinCombat permet de rejoindre un combat
func (h *CombatHandler) JoinCombat(c *gin.Context) {
	sessionIDStr := c.Param("sessionId")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid session ID",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	var req models.JoinCombatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid request body",
			"details":    err.Error(),
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	if err := h.combatService.JoinCombat(sessionID, req); err != nil {
		logrus.WithError(err).Error("Failed to join combat")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Failed to join combat",
			"details":    err.Error(),
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Successfully joined combat",
		"session_id": sessionID,
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

// LeaveCombat permet de quitter un combat
func (h *CombatHandler) LeaveCombat(c *gin.Context) {
	sessionIDStr := c.Param("sessionId")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid session ID",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	characterIDStr := c.Query("characterId")
	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid character ID",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	if err := h.combatService.LeaveCombat(sessionID, characterID); err != nil {
		logrus.WithError(err).Error("Failed to leave combat")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Failed to leave combat",
			"details":    err.Error(),
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Successfully left combat",
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

// PerformAction exécute une action de combat
func (h *CombatHandler) PerformAction(c *gin.Context) {
	var req models.PerformActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid request body",
			"details":    err.Error(),
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	action, err := h.combatService.PerformAction(req)
	if err != nil {
		logrus.WithError(err).Error("Failed to perform combat action")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Failed to perform action",
			"details":    err.Error(),
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"action":     action,
		"message":    "Action performed successfully",
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

// GetCombatStatus récupère le statut d'un combat
func (h *CombatHandler) GetCombatStatus(c *gin.Context) {
	sessionIDStr := c.Param("sessionId")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid session ID",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	status, err := h.combatService.GetCombatSession(sessionID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get combat status")
		c.JSON(http.StatusNotFound, gin.H{
			"error":      "Combat session not found",
			"details":    err.Error(),
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

// GetActiveCombats récupère tous les combats actifs
func (h *CombatHandler) GetActiveCombats(c *gin.Context) {
	combats, err := h.combatService.GetActiveCombats()
	if err != nil {
		logrus.WithError(err).Error("Failed to get active combats")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to retrieve active combats",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"combats":    combats,
		"count":      len(combats),
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

// GetCombatLogs récupère les logs d'un personnage
func (h *CombatHandler) GetCombatLogs(c *gin.Context) {
	characterIDStr := c.Param("characterId")
	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid character ID",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 1000 {
		limit = 50
	}

	// Pour l'instant, retourner une réponse vide
	// L'implémentation complète nécessiterait le CombatLogService
	c.JSON(http.StatusOK, gin.H{
		"logs":       []interface{}{},
		"character_id": characterID,
		"limit":      limit,
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

// GetSessionLogs récupère les logs d'une session
func (h *CombatHandler) GetSessionLogs(c *gin.Context) {
	sessionIDStr := c.Param("sessionId")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid session ID",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 1000 {
		limit = 100
	}

	// Pour l'instant, retourner une réponse vide
	// L'implémentation complète nécessiterait une méthode dans le service
	c.JSON(http.StatusOK, gin.H{
		"logs":       []interface{}{},
		"session_id": sessionID,
		"limit":      limit,
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

// Endpoints admin et statistiques

// GetAllCombatSessions récupère toutes les sessions (admin)
func (h *CombatHandler) GetAllCombatSessions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"sessions":   []interface{}{},
		"total":      0,
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

// ForceEndCombat force la fin d'un combat (admin)
func (h *CombatHandler) ForceEndCombat(c *gin.Context) {
	sessionIDStr := c.Param("sessionId")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid session ID",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	if err := h.combatService.EndCombat(sessionID, 0); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to end combat",
			"details":    err.Error(),
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Combat ended successfully",
		"session_id": sessionID,
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

// GetCombatStatistics récupère les statistiques de combat
func (h *CombatHandler) GetCombatStatistics(c *gin.Context) {
	stats, err := h.combatService.GetCombatStatistics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to get statistics",
			"details":    err.Error(),
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statistics": stats,
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

// SimulateCombat simule un combat (debug)
func (h *CombatHandler) SimulateCombat(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message":    "Combat simulation completed",
		"result":     "This is a debug endpoint",
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

// EndCombat termine un combat (alias pour ForceEndCombat)
func (h *CombatHandler) EndCombat(c *gin.Context) {
	h.ForceEndCombat(c)
}

// Méthodes PvP

// ChallengePvP crée un défi PvP
func (h *CombatHandler) ChallengePvP(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message":    "PvP challenge - TODO: Implement",
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

// AcceptPvPChallenge accepte un défi PvP
func (h *CombatHandler) AcceptPvPChallenge(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message":    "PvP challenge accepted - TODO: Implement",
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

// DeclinePvPChallenge refuse un défi PvP
func (h *CombatHandler) DeclinePvPChallenge(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message":    "PvP challenge declined - TODO: Implement",
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

// GetPvPRankings récupère les classements PvP
func (h *CombatHandler) GetPvPRankings(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"rankings":   []interface{}{},
		"message":    "PvP rankings - TODO: Implement",
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

// Méthodes d'effets

// GetActiveEffects récupère les effets actifs d'un personnage
func (h *CombatHandler) GetActiveEffects(c *gin.Context) {
	characterIDStr := c.Param("characterId")
	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid character ID",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	// Pour l'instant, retourner une liste vide
	c.JSON(http.StatusOK, gin.H{
		"character_id": characterID,
		"effects":      []interface{}{},
		"request_id":   c.GetHeader("X-Request-ID"),
	})
}

// RemoveEffect supprime un effet
func (h *CombatHandler) RemoveEffect(c *gin.Context) {
	effectIDStr := c.Param("effectId")
	effectID, err := uuid.Parse(effectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid effect ID",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	// Pour l'instant, simuler la suppression
	c.JSON(http.StatusOK, gin.H{
		"message":    "Effect removed successfully",
		"effect_id":  effectID,
		"request_id": c.GetHeader("X-Request-ID"),
	})
}