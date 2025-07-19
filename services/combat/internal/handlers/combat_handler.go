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

// Constantes pour les valeurs de requête
const (
	QueryValueTrue = "true"
)

// CombatHandler gère les requêtes HTTP liées au combat
type CombatHandler struct {
	combatService service.CombatServiceInterface
	config        *config.Config
}

// NewCombatHandler crée un nouveau handler de combat
func NewCombatHandler(combatService service.CombatServiceInterface, config *config.Config) *CombatHandler {
	return &CombatHandler{
		combatService: combatService,
		config:        config,
	}
}

// CreateCombat crée un nouveau combat
func (h *CombatHandler) CreateCombat(c *gin.Context) {
	var req models.CreateCombatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Récupérer l'ID utilisateur depuis le contexte JWT
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	combat, err := h.combatService.CreateCombat(&req)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to create combat")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create combat",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"combat":  combat,
		"message": "Combat created successfully",
	})
}

// GetCombat récupère un combat par son ID
func (h *CombatHandler) GetCombat(c *gin.Context) {
	combatIDStr := c.Param("id")
	combatID, err := uuid.Parse(combatIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid combat ID"})
		return
	}

	combat, err := h.combatService.GetCombat(combatID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Combat not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"combat":  combat,
	})
}

// GetCombatStatus récupère le statut détaillé d'un combat
func (h *CombatHandler) GetCombatStatus(c *gin.Context) {
	combatIDStr := c.Param("id")
	combatID, err := uuid.Parse(combatIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid combat ID"})
		return
	}

	// Parser les paramètres de requête
	req := &models.GetCombatStatusRequest{
		IncludeParticipants: c.Query("include_participants") == QueryValueTrue,
		IncludeActions:      c.Query("include_actions") == QueryValueTrue,
		IncludeEffects:      c.Query("include_effects") == QueryValueTrue,
		IncludeLogs:         c.Query("include_logs") == QueryValueTrue,
	}

	status, err := h.combatService.GetCombatStatus(combatID, req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Failed to get combat status",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"status":  status,
	})
}

// StartCombat démarre un combat
func (h *CombatHandler) StartCombat(c *gin.Context) {
	combatIDStr := c.Param("id")
	combatID, err := uuid.Parse(combatIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid combat ID"})
		return
	}

	if err := h.combatService.StartCombat(combatID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to start combat",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Combat started successfully",
	})
}

// EndCombat termine un combat
func (h *CombatHandler) EndCombat(c *gin.Context) {
	combatIDStr := c.Param("id")
	combatID, err := uuid.Parse(combatIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid combat ID"})
		return
	}

	var req models.EndCombatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	result, err := h.combatService.EndCombat(combatID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to end combat",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"result":  result,
		"message": "Combat ended successfully",
	})
}

// JoinCombat ajoute un participant à un combat
func (h *CombatHandler) JoinCombat(c *gin.Context) {
	combatIDStr := c.Param("id")
	combatID, err := uuid.Parse(combatIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid combat ID"})
		return
	}

	var req models.JoinCombatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	if err := h.combatService.JoinCombat(combatID, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to join combat",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Successfully joined combat",
	})
}

// LeaveCombat retire un participant d'un combat
func (h *CombatHandler) LeaveCombat(c *gin.Context) {
	combatIDStr := c.Param("id")
	combatID, err := uuid.Parse(combatIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid combat ID"})
		return
	}

	characterIDStr := c.Query("character_id")
	if characterIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Character ID required"})
		return
	}

	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid character ID"})
		return
	}

	var req models.LeaveCombatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Debug("Failed to bind optional request body")
	}

	if err := h.combatService.LeaveCombat(combatID, characterID, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to leave combat",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Successfully left combat",
	})
}

// ExecuteAction exécute une action de combat
func (h *CombatHandler) ExecuteAction(c *gin.Context) {
	combatIDStr := c.Param("id")
	combatID, err := uuid.Parse(combatIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid combat ID"})
		return
	}

	var req models.ActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Récupérer l'ID du personnage acteur
	actorIDStr := c.GetHeader("X-Character-ID")
	if actorIDStr == "" {
		actorIDStr = c.Query("actor_id")
	}
	if actorIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Actor ID required"})
		return
	}

	actorID, err := uuid.Parse(actorIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid actor ID"})
		return
	}

	result, err := h.combatService.ExecuteAction(combatID, actorID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to execute action",
			"details": err.Error(),
		})
		return
	}

	// Statut de réponse selon le succès de l'action
	statusCode := http.StatusOK
	if !result.Success {
		statusCode = http.StatusBadRequest
	}

	c.JSON(statusCode, gin.H{
		"success": result.Success,
		"result":  result,
	})
}

// ValidateAction valide une action sans l'exécuter
func (h *CombatHandler) ValidateAction(c *gin.Context) {
	combatIDStr := c.Param("id")
	combatID, err := uuid.Parse(combatIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid combat ID"})
		return
	}

	var req models.ValidateActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Récupérer l'ID du personnage acteur
	actorIDStr := c.GetHeader("X-Character-ID")
	if actorIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Actor ID required"})
		return
	}

	actorID, err := uuid.Parse(actorIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid actor ID"})
		return
	}

	validation, err := h.combatService.ValidateAction(combatID, actorID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to validate action",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"validation": validation,
	})
}

// GetAvailableActions récupère les actions disponibles pour un participant
func (h *CombatHandler) GetAvailableActions(c *gin.Context) {
	combatIDStr := c.Param("id")
	combatID, err := uuid.Parse(combatIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid combat ID"})
		return
	}

	actorIDStr := c.Query("actor_id")
	if actorIDStr == "" {
		actorIDStr = c.GetHeader("X-Character-ID")
	}
	if actorIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Actor ID required"})
		return
	}

	actorID, err := uuid.Parse(actorIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid actor ID"})
		return
	}

	actions, err := h.combatService.GetAvailableActions(combatID, actorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get available actions",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"actions": actions,
	})
}

// ProcessTurn traite un tour de combat
func (h *CombatHandler) ProcessTurn(c *gin.Context) {
	combatIDStr := c.Param("id")
	combatID, err := uuid.Parse(combatIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid combat ID"})
		return
	}

	if err := h.combatService.ProcessTurn(combatID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to process turn",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Turn processed successfully",
	})
}

// AdvanceTurn avance au tour suivant
func (h *CombatHandler) AdvanceTurn(c *gin.Context) {
	combatIDStr := c.Param("id")
	combatID, err := uuid.Parse(combatIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid combat ID"})
		return
	}

	if err := h.combatService.AdvanceTurn(combatID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to advance turn",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Turn advanced successfully",
	})
}

// GetCurrentTurn récupère les informations du tour actuel
func (h *CombatHandler) GetCurrentTurn(c *gin.Context) {
	combatIDStr := c.Param("id")
	combatID, err := uuid.Parse(combatIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid combat ID"})
		return
	}

	turnInfo, err := h.combatService.GetCurrentTurn(combatID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get turn info",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"turn_info": turnInfo,
	})
}

// SearchCombats recherche des combats
func (h *CombatHandler) SearchCombats(c *gin.Context) {
	req := &models.SearchCombatsRequest{}

	// Parser les paramètres de query
	if combatType := c.Query("combat_type"); combatType != "" {
		ct := models.CombatType(combatType)
		req.CombatType = &ct
	}

	if status := c.Query("status"); status != "" {
		cs := models.CombatStatus(status)
		req.Status = &cs
	}

	if zoneID := c.Query("zone_id"); zoneID != "" {
		req.ZoneID = &zoneID
	}

	if participantID := c.Query("participant_id"); participantID != "" {
		pid, err := uuid.Parse(participantID)
		if err == nil {
			req.ParticipantID = &pid
		}
	}

	// Pagination
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			req.Limit = limit
		}
	}
	if req.Limit == 0 {
		req.Limit = 20
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			req.Offset = offset
		}
	}

	req.IncludeFinished = c.Query("include_finished") == QueryValueTrue

	result, err := h.combatService.SearchCombats(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to search combats",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"result":  result,
	})
}

// GetCombatHistory récupère l'historique de combat
func (h *CombatHandler) GetCombatHistory(c *gin.Context) {
	req := &models.GetCombatHistoryRequest{}

	// Parser les paramètres
	if characterID := c.Query("character_id"); characterID != "" {
		cid, err := uuid.Parse(characterID)
		if err == nil {
			req.CharacterID = &cid
		}
	}

	if userID := c.Query("user_id"); userID != "" {
		uid, err := uuid.Parse(userID)
		if err == nil {
			req.UserID = &uid
		}
	}

	if combatType := c.Query("combat_type"); combatType != "" {
		ct := models.CombatType(combatType)
		req.CombatType = &ct
	}

	// Pagination
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			req.Limit = limit
		}
	}
	if req.Limit == 0 {
		req.Limit = 50
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			req.Offset = offset
		}
	}

	req.WinsOnly = c.Query("wins_only") == QueryValueTrue
	req.LossesOnly = c.Query("losses_only") == QueryValueTrue

	result, err := h.combatService.GetCombatHistory(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get combat history",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"result":  result,
	})
}

// GetStatistics récupère les statistiques de combat
func (h *CombatHandler) GetStatistics(c *gin.Context) {
	req := &models.GetStatisticsRequest{}

	// Parser les paramètres
	if characterID := c.Query("character_id"); characterID != "" {
		cid, err := uuid.Parse(characterID)
		if err == nil {
			req.CharacterID = &cid
		}
	}

	if userID := c.Query("user_id"); userID != "" {
		uid, err := uuid.Parse(userID)
		if err == nil {
			req.UserID = &uid
		}
	}

	if combatType := c.Query("combat_type"); combatType != "" {
		ct := models.CombatType(combatType)
		req.CombatType = &ct
	}

	if period := c.Query("period"); period != "" {
		req.Period = period
	}

	req.Detailed = c.Query("detailed") == "true"

	result, err := h.combatService.GetStatistics(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get statistics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"result":  result,
	})
}

// GetParticipants récupère les participants d'un combat
func (h *CombatHandler) GetParticipants(c *gin.Context) {
	combatIDStr := c.Param("id")
	combatID, err := uuid.Parse(combatIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid combat ID"})
		return
	}

	participants, err := h.combatService.GetParticipants(combatID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get participants",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"participants": participants,
	})
}

// Méthodes d'administration et de debug

// GetActiveCombatCount retourne le nombre de combats actifs
func (h *CombatHandler) GetActiveCombatCount(c *gin.Context) {
	count, err := h.combatService.GetActiveCombatCount()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get active combat count",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"active_combats": count,
	})
}

// Méthodes stubs pour les fonctionnalités à implémenter

func (h *CombatHandler) DeleteCombat(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Delete combat not implemented yet",
	})
}

func (h *CombatHandler) UpdateParticipant(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Update participant not implemented yet",
	})
}

func (h *CombatHandler) ApplyEffect(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Apply effect not implemented yet",
	})
}

func (h *CombatHandler) RemoveEffect(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Remove effect not implemented yet",
	})
}

func (h *CombatHandler) GetCombatEffects(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Get combat effects not implemented yet",
	})
}

func (h *CombatHandler) ListAllCombats(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "List all combats not implemented yet",
	})
}

func (h *CombatHandler) ForceEndCombat(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Force end combat not implemented yet",
	})
}

func (h *CombatHandler) AdminAction(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Admin action not implemented yet",
	})
}

func (h *CombatHandler) GetSuspiciousActivities(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Get suspicious activities not implemented yet",
	})
}

func (h *CombatHandler) BanUser(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Ban user not implemented yet",
	})
}

func (h *CombatHandler) GetCombatStatusForService(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Get combat status for service not implemented yet",
	})
}

func (h *CombatHandler) ValidateCharacterStats(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Validate character stats not implemented yet",
	})
}
