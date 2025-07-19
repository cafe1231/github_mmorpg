package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"player/internal/config"
	"player/internal/middleware"
	"player/internal/models"
	"player/internal/service"
)

// CharacterHandler gère les routes des personnages
type CharacterHandler struct {
	characterService *service.CharacterService
	config           *config.Config
}

// NewCharacterHandler crée un nouveau handler de personnage
func NewCharacterHandler(characterService *service.CharacterService, config *config.Config) *CharacterHandler {
	return &CharacterHandler{
		characterService: characterService,
		config:           config,
	}
}

// CreateCharacter godoc
// @Summary      Création d'un nouveau personnage
// @Description  Crée un nouveau personnage pour le joueur connecté
// @Tags         character
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body models.CreateCharacterRequest true "Données du personnage"
// @Success      201  {object}  models.Character
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      409  {object}  map[string]interface{}
// @Router       /api/v1/characters [post]
func (h *CharacterHandler) CreateCharacter(c *gin.Context) {
	var req models.CreateCharacterRequest

	// Validation des données JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Invalid character creation data")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid request data",
			"details":    err.Error(),
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	// Récupérer l'ID utilisateur du contexte
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":      "User ID not found",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	// Créer le personnage
	character, err := h.characterService.CreateCharacter(userID, req)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Character creation failed")

		statusCode := http.StatusBadRequest
		if err.Error() == "character name already taken" {
			statusCode = http.StatusConflict
		} else if err.Error() == "maximum number of characters reached" {
			statusCode = http.StatusConflict
		}

		c.JSON(statusCode, gin.H{
			"error":      "Character creation failed",
			"message":    err.Error(),
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	logrus.WithFields(logrus.Fields{
		"character_id": character.ID,
		"user_id":      userID,
		"name":         character.Name,
		"class":        character.Class,
		"race":         character.Race,
	}).Info("Character created successfully")

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Character created successfully",
		"character":  character,
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

// GetCharacters godoc
// @Summary      Liste des personnages du joueur
// @Description  Récupère tous les personnages du joueur connecté
// @Tags         character
// @Security     BearerAuth
// @Produce      json
// @Success      200  {array}   models.Character
// @Failure      401  {object}  map[string]interface{}
// @Router       /api/v1/characters [get]
func (h *CharacterHandler) GetCharacters(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":      "User ID not found",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	characters, err := h.characterService.GetCharactersByPlayer(userID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to get characters")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to retrieve characters",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"characters":  characters,
		"total":       len(characters),
		"max_allowed": h.config.Game.MaxCharactersPerPlayer,
		"request_id":  c.GetHeader("X-Request-ID"),
	})
}

// GetCharacter godoc
// @Summary      Détails d'un personnage
// @Description  Récupère les détails complets d'un personnage
// @Tags         character
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "Character ID"
// @Success      200  {object}  models.Character
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /api/v1/characters/{id} [get]
func (h *CharacterHandler) GetCharacter(c *gin.Context) {
	characterID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid character ID",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":      "User ID not found",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	character, err := h.characterService.GetCharacter(characterID, userID)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"character_id": characterID,
			"user_id":      userID,
		}).Error("Failed to get character")

		statusCode := http.StatusInternalServerError
		if err.Error() == "character not found" || err.Error() == "access denied: not the owner of this character" {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"error":      "Character not found",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"character":  character,
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

// UpdateCharacter godoc
// @Summary      Mise à jour d'un personnage
// @Description  Met à jour les informations d'un personnage
// @Tags         character
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id      path      string  true  "Character ID"
// @Param        request body      models.UpdateCharacterRequest true "Nouvelles données"
// @Success      200     {object}  models.Character
// @Failure      400     {object}  map[string]interface{}
// @Failure      401     {object}  map[string]interface{}
// @Failure      404     {object}  map[string]interface{}
// @Router       /api/v1/characters/{id} [put]
func (h *CharacterHandler) UpdateCharacter(c *gin.Context) {
	characterID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid character ID",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	var req models.UpdateCharacterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid request data",
			"details":    err.Error(),
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":      "User ID not found",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	character, err := h.characterService.UpdateCharacter(characterID, userID, req)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"character_id": characterID,
			"user_id":      userID,
		}).Error("Character update failed")

		statusCode := http.StatusBadRequest
		if err.Error() == "character not found" || err.Error() == "access denied: not the owner of this character" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "character name already taken" {
			statusCode = http.StatusConflict
		}

		c.JSON(statusCode, gin.H{
			"error":      "Character update failed",
			"message":    err.Error(),
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	logrus.WithFields(logrus.Fields{
		"character_id": characterID,
		"user_id":      userID,
	}).Info("Character updated successfully")

	c.JSON(http.StatusOK, gin.H{
		"message":    "Character updated successfully",
		"character":  character,
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

// DeleteCharacter godoc
// @Summary      Suppression d'un personnage
// @Description  Supprime un personnage (soft delete)
// @Tags         character
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "Character ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /api/v1/characters/{id} [delete]
func (h *CharacterHandler) DeleteCharacter(c *gin.Context) {
	characterID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid character ID",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":      "User ID not found",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	err = h.characterService.DeleteCharacter(characterID, userID)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"character_id": characterID,
			"user_id":      userID,
		}).Error("Character deletion failed")

		statusCode := http.StatusInternalServerError
		if err.Error() == "character not found" || err.Error() == "access denied: not the owner of this character" {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"error":      "Character deletion failed",
			"message":    err.Error(),
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	logrus.WithFields(logrus.Fields{
		"character_id": characterID,
		"user_id":      userID,
	}).Info("Character deleted successfully")

	c.JSON(http.StatusOK, gin.H{
		"message":    "Character deleted successfully",
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

// GetCharacterStats godoc
// @Summary      Statistiques d'un personnage
// @Description  Récupère toutes les statistiques d'un personnage
// @Tags         character
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "Character ID"
// @Success      200  {object}  models.StatsResponse
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /api/v1/characters/{id}/stats [get]
func (h *CharacterHandler) GetCharacterStats(c *gin.Context) {
	characterID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid character ID",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":      "User ID not found",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	stats, err := h.characterService.GetCharacterStats(characterID, userID)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"character_id": characterID,
			"user_id":      userID,
		}).Error("Failed to get character stats")

		statusCode := http.StatusInternalServerError
		if err.Error() == "character not found" || err.Error() == "access denied: not the owner of this character" {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"error":      "Failed to get character stats",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats":      stats,
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

// UpdateCharacterStats godoc
// @Summary      Mise à jour des statistiques
// @Description  Met à jour les statistiques d'un personnage (dépense de points)
// @Tags         character
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id      path      string  true  "Character ID"
// @Param        request body      models.UpdateStatsRequest true "Points à dépenser"
// @Success      200     {object}  models.CharacterStats
// @Failure      400     {object}  map[string]interface{}
// @Failure      401     {object}  map[string]interface{}
// @Failure      404     {object}  map[string]interface{}
// @Router       /api/v1/characters/{id}/stats [put]
func (h *CharacterHandler) UpdateCharacterStats(c *gin.Context) {
	characterID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid character ID",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	var req models.UpdateStatsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid request data",
			"details":    err.Error(),
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":      "User ID not found",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	stats, err := h.characterService.UpdateCharacterStats(characterID, userID, req)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"character_id": characterID,
			"user_id":      userID,
		}).Error("Character stats update failed")

		statusCode := http.StatusBadRequest
		if err.Error() == "character not found" || err.Error() == "access denied: not the owner of this character" {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"error":      "Stats update failed",
			"message":    err.Error(),
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	logrus.WithFields(logrus.Fields{
		"character_id": characterID,
		"user_id":      userID,
	}).Info("Character stats updated successfully")

	c.JSON(http.StatusOK, gin.H{
		"message":    "Stats updated successfully",
		"stats":      stats,
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

// GetGameInfo godoc
// @Summary      Informations du jeu
// @Description  Récupère les classes, races et configuration du jeu
// @Tags         character
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/characters/game-info [get]
func (h *CharacterHandler) GetGameInfo(c *gin.Context) {
	gameInfo := h.characterService.GetAvailableClassesAndRaces()

	c.JSON(http.StatusOK, gin.H{
		"game_info":  gameInfo,
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

// AddExperience godoc
// @Summary      Ajouter de l'expérience (admin/debug)
// @Description  Ajoute de l'expérience à un personnage
// @Tags         character
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path      string  true   "Character ID"
// @Param        request  body      object  true   "Experience to add"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      401      {object}  map[string]interface{}
// @Router       /api/v1/characters/{id}/experience [post]
func (h *CharacterHandler) AddExperience(c *gin.Context) {
	if h.config.Server.Environment == "production" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Debug endpoints disabled in production",
		})
		return
	}

	characterID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid character ID",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	var req struct {
		Experience int64 `json:"experience" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid request data",
			"details":    err.Error(),
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	err = h.characterService.AddExperience(characterID, req.Experience)
	if err != nil {
		logrus.WithError(err).WithField("character_id", characterID).Error("Failed to add experience")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to add experience",
			"message":    err.Error(),
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Experience added successfully",
		"experience": req.Experience,
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

// ShowCharacterDebug affiche les informations de debug d'un personnage
func (h *CharacterHandler) ShowCharacterDebug(c *gin.Context) {
	if h.config.Server.Environment == "production" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Debug endpoints disabled in production"})
		return
	}

	characterID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"message":      "Character debug info",
		"character_id": characterID,
		"config": gin.H{
			"max_level":                 h.config.Game.MaxLevel,
			"max_characters_per_player": h.config.Game.MaxCharactersPerPlayer,
			"available_classes":         h.config.Game.AvailableClasses,
			"available_races":           h.config.Game.AvailableRaces,
		},
	})
}
