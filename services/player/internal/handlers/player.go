package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"player/internal/config"
	"player/internal/middleware"
	"player/internal/models"
	"player/internal/service"
)

// PlayerHandler gère les routes des joueurs
type PlayerHandler struct {
	playerService *service.PlayerService
	config        *config.Config
}

// NewPlayerHandler crée un nouveau handler de joueur
func NewPlayerHandler(playerService *service.PlayerService, config *config.Config) *PlayerHandler {
	return &PlayerHandler{
		playerService: playerService,
		config:        config,
	}
}

// CreatePlayer godoc
// @Summary      Création d'un profil joueur
// @Description  Crée un nouveau profil joueur pour l'utilisateur connecté
// @Tags         player
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body models.CreatePlayerRequest true "Données du profil joueur"
// @Success      201  {object}  models.Player
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      409  {object}  map[string]interface{}
// @Router       /api/v1/player/profile [post]
func (h *PlayerHandler) CreatePlayer(c *gin.Context) {
	var req models.CreatePlayerRequest

	// Validation des données JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Invalid player creation data")
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

	// Créer le profil joueur
	player, err := h.playerService.CreatePlayer(userID, req)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Player creation failed")

		statusCode := http.StatusBadRequest
		if err.Error() == "player profile already exists for this user" {
			statusCode = http.StatusConflict
		} else if err.Error() == "display name already taken" {
			statusCode = http.StatusConflict
		}

		c.JSON(statusCode, gin.H{
			"error":      "Player creation failed",
			"message":    err.Error(),
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	logrus.WithFields(logrus.Fields{
		"player_id":    player.ID,
		"user_id":      userID,
		"display_name": player.DisplayName,
	}).Info("Player profile created successfully")

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Player profile created successfully",
		"player":     player,
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

// GetPlayer godoc
// @Summary      Profil du joueur connecté
// @Description  Récupère le profil complet du joueur connecté avec ses personnages
// @Tags         player
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  models.PlayerResponse
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /api/v1/player/profile [get]
func (h *PlayerHandler) GetPlayer(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":      "User ID not found",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	playerResponse, err := h.playerService.GetPlayer(userID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to get player")

		statusCode := http.StatusInternalServerError
		if err.Error() == "player not found" {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"error":      "Player not found",
			"message":    "Please create a player profile first",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"player":     playerResponse.Player,
		"stats":      playerResponse.Stats,
		"characters": playerResponse.Characters,
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

// UpdatePlayer godoc
// @Summary      Mise à jour du profil joueur
// @Description  Met à jour les informations du profil joueur
// @Tags         player
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body models.UpdatePlayerRequest true "Nouvelles données du profil"
// @Success      200  {object}  models.Player
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /api/v1/player/profile [put]
func (h *PlayerHandler) UpdatePlayer(c *gin.Context) {
	var req models.UpdatePlayerRequest

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

	player, err := h.playerService.UpdatePlayer(userID, req)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Player update failed")

		statusCode := http.StatusBadRequest
		if err.Error() == "player not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "display name already taken" {
			statusCode = http.StatusConflict
		}

		c.JSON(statusCode, gin.H{
			"error":      "Player update failed",
			"message":    err.Error(),
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	logrus.WithFields(logrus.Fields{
		"player_id":    player.ID,
		"user_id":      userID,
		"display_name": player.DisplayName,
	}).Info("Player profile updated successfully")

	c.JSON(http.StatusOK, gin.H{
		"message":    "Player profile updated successfully",
		"player":     player,
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

// GetPlayerStats godoc
// @Summary      Statistiques du joueur
// @Description  Récupère les statistiques complètes du joueur
// @Tags         player
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  models.PlayerStats
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /api/v1/player/stats [get]
func (h *PlayerHandler) GetPlayerStats(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":      "User ID not found",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	stats, err := h.playerService.GetPlayerStats(userID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to get player stats")

		statusCode := http.StatusInternalServerError
		if err.Error() == "player not found" {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"error":      "Failed to get player stats",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats":      stats,
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

// UpdatePlayTime godoc
// @Summary      Mise à jour du temps de jeu
// @Description  Met à jour le temps de jeu total du joueur
// @Tags         player
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body object true "Minutes de jeu à ajouter"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /api/v1/player/playtime [post]
func (h *PlayerHandler) UpdatePlayTime(c *gin.Context) {
	var req struct {
		Minutes int `json:"minutes" binding:"required,min=1"`
	}

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

	err := h.playerService.UpdatePlayTime(userID, req.Minutes)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to update play time")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to update play time",
			"message":    err.Error(),
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Play time updated successfully",
		"minutes":    req.Minutes,
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

// UpdateLastSeen godoc
// @Summary      Mise à jour de la dernière connection
// @Description  Met à jour l'horodatage de la dernière activité du joueur
// @Tags         player
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /api/v1/player/ping [post]
func (h *PlayerHandler) UpdateLastSeen(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":      "User ID not found",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	err := h.playerService.UpdateLastSeen(userID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to update last seen")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to update last seen",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Last seen updated successfully",
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

// CanCreateCharacter godoc
// @Summary      Vérification de création de personnage
// @Description  Vérifie si le joueur peut créer un nouveau personnage
// @Tags         player
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /api/v1/player/can-create-character [get]
func (h *PlayerHandler) CanCreateCharacter(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":      "User ID not found",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	canCreate, err := h.playerService.CanCreateCharacter(userID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to check character creation")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to check character creation availability",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"can_create":  canCreate,
		"max_allowed": h.config.Game.MaxCharactersPerPlayer,
		"request_id":  c.GetHeader("X-Request-ID"),
	})
}

// SearchPlayers godoc
// @Summary      Recherche de joueurs
// @Description  Recherche des joueurs par nom d'affichage
// @Tags         player
// @Security     BearerAuth
// @Produce      json
// @Param        q    query     string  true  "Terme de recherche"
// @Success      200  {array}   models.Player
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /api/v1/player/search [get]
func (h *PlayerHandler) SearchPlayers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Search query is required",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	if len(query) < 3 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Search query must be at least 3 characters long",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	players, err := h.playerService.SearchPlayersByDisplayName(query)
	if err != nil {
		logrus.WithError(err).WithField("query", query).Error("Player search failed")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Search failed",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"players":    players,
		"total":      len(players),
		"query":      query,
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

// Admin endpoints

// ListPlayers godoc
// @Summary      Liste des joueurs (admin)
// @Description  Récupère une liste de tous les joueurs avec pagination
// @Tags         admin
// @Security     BearerAuth
// @Produce      json
// @Param        limit   query     int  false  "Nombre de résultats par page (max 100)"
// @Param        offset  query     int  false  "Nombre de résultats à ignorer"
// @Success      200     {array}   models.Player
// @Failure      401     {object}  map[string]interface{}
// @Failure      403     {object}  map[string]interface{}
// @Router       /api/v1/admin/players [get]
func (h *PlayerHandler) ListPlayers(c *gin.Context) {
	// Parse query parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	players, err := h.playerService.ListPlayers(limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to list players")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to retrieve players",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"players":    players,
		"total":      len(players),
		"limit":      limit,
		"offset":     offset,
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

// GetPlayerByID godoc
// @Summary      Détails d'un joueur (admin)
// @Description  Récupère les détails d'un joueur par son ID
// @Tags         admin
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "Player ID"
// @Success      200  {object}  models.Player
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /api/v1/admin/players/{id} [get]
func (h *PlayerHandler) GetPlayerByID(c *gin.Context) {
	playerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid player ID",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	player, err := h.playerService.GetPlayerByID(playerID)
	if err != nil {
		logrus.WithError(err).WithField("player_id", playerID).Error("Failed to get player by ID")

		statusCode := http.StatusInternalServerError
		if err.Error() == "player not found" {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"error":      "Player not found",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"player":     player,
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

// DeletePlayerProfile godoc
// @Summary      Suppression de profil joueur (admin)
// @Description  Supprime complètement le profil d'un joueur et ses personnages
// @Tags         admin
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "Player User ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /api/v1/admin/players/{id} [delete]
func (h *PlayerHandler) DeletePlayerProfile(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid user ID",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	err = h.playerService.DeletePlayer(userID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Player deletion failed")

		statusCode := http.StatusInternalServerError
		if err.Error() == "player not found" {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"error":      "Player deletion failed",
			"message":    err.Error(),
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	logrus.WithField("user_id", userID).Info("Player profile deleted by admin")

	c.JSON(http.StatusOK, gin.H{
		"message":    "Player profile deleted successfully",
		"user_id":    userID,
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

// Debug endpoints

// ShowPlayerDebug affiche les informations de debug d'un joueur
func (h *PlayerHandler) ShowPlayerDebug(c *gin.Context) {
	if h.config.Server.Environment == "production" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Debug endpoints disabled in production"})
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

	username, _ := middleware.GetUsernameFromContext(c)
	userRole, _ := middleware.GetUserRoleFromContext(c)

	debugInfo := gin.H{
		"user_context": gin.H{
			"user_id":  userID,
			"username": username,
			"role":     userRole,
		},
		"game_config": gin.H{
			"max_characters_per_player": h.config.Game.MaxCharactersPerPlayer,
			"max_level":                 h.config.Game.MaxLevel,
			"starting_level":            h.config.Game.StartingLevel,
			"available_classes":         h.config.Game.AvailableClasses,
			"available_races":           h.config.Game.AvailableRaces,
			"starting_stats":            h.config.Game.StartingStats,
		},
		"service_info": gin.H{
			"environment": h.config.Server.Environment,
			"debug":       h.config.Server.Debug,
			"auth_url":    h.config.Auth.ServiceURL,
		},
	}

	c.JSON(http.StatusOK, debugInfo)
}

// GetPlayerSummary endpoint pour récupérer un résumé du joueur (utilisé par d'autres services)
func (h *PlayerHandler) GetPlayerSummary(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid user ID",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	playerResponse, err := h.playerService.GetPlayer(userID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to get player summary")

		statusCode := http.StatusInternalServerError
		if err.Error() == "player not found" {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"error":      "Player not found",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	// Retourner un résumé simplifié pour les autres services
	summary := gin.H{
		"player_id":       playerResponse.Player.ID,
		"user_id":         playerResponse.Player.UserID,
		"display_name":    playerResponse.Player.DisplayName,
		"avatar":          playerResponse.Player.Avatar,
		"title":           playerResponse.Player.Title,
		"guild_id":        playerResponse.Player.GuildID,
		"is_online":       playerResponse.Player.IsOnline(),
		"character_count": len(playerResponse.Characters),
		"last_seen":       playerResponse.Player.LastSeen,
	}

	if playerResponse.Stats != nil {
		summary["highest_level"] = playerResponse.Stats.HighestLevel
		summary["total_experience"] = playerResponse.Stats.TotalExperience
	}

	c.JSON(http.StatusOK, gin.H{
		"player_summary": summary,
		"request_id":     c.GetHeader("X-Request-ID"),
	})
}

// GetPlayerCharactersSummary endpoint pour récupérer un résumé des personnages d'un joueur
func (h *PlayerHandler) GetPlayerCharactersSummary(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid user ID",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	playerResponse, err := h.playerService.GetPlayer(userID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to get player characters")

		statusCode := http.StatusInternalServerError
		if err.Error() == "player not found" {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"error":      "Player not found",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	// Créer un résumé des personnages
	characters := make([]gin.H, 0, len(playerResponse.Characters))
	for _, char := range playerResponse.Characters {
		characters = append(characters, gin.H{
			"id":          char.ID,
			"name":        char.Name,
			"class":       char.Class,
			"race":        char.Race,
			"level":       char.Level,
			"zone_id":     char.ZoneID,
			"status":      char.Status,
			"last_played": char.LastPlayed,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"characters": characters,
		"total":      len(characters),
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

// ValidateDisplayName endpoint pour valider un nom d'affichage
func (h *PlayerHandler) ValidateDisplayName(c *gin.Context) {
	var req struct {
		DisplayName string `json:"display_name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid request data",
			"details":    err.Error(),
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	// Validation de base du nom
	isValid := true
	message := "Display name is valid"

	if len(req.DisplayName) < 3 {
		isValid = false
		message = "Display name must be at least 3 characters long"
	} else if len(req.DisplayName) > 20 {
		isValid = false
		message = "Display name must be less than 20 characters long"
	}

	// Vérifier l'unicité si le nom est valide
	if isValid {
		_, err := h.playerService.SearchPlayersByDisplayName(req.DisplayName)
		if err == nil {
			// Si on trouve des résultats, vérifier l'exactitude
			players, _ := h.playerService.ListPlayers(100, 0)
			for _, player := range players {
				if player.DisplayName == req.DisplayName {
					isValid = false
					message = "Display name is already taken"
					break
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":        isValid,
		"message":      message,
		"display_name": req.DisplayName,
		"request_id":   c.GetHeader("X-Request-ID"),
	})
}

// GetOnlinePlayers endpoint pour récupérer les joueurs en ligne
func (h *PlayerHandler) GetOnlinePlayers(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	players, err := h.playerService.ListPlayers(limit, 0)
	if err != nil {
		logrus.WithError(err).Error("Failed to get players for online check")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to retrieve players",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	// Filtrer les joueurs en ligne
	onlinePlayers := make([]gin.H, 0)
	for _, player := range players {
		if player.IsOnline() {
			onlinePlayers = append(onlinePlayers, gin.H{
				"id":           player.ID,
				"display_name": player.DisplayName,
				"avatar":       player.Avatar,
				"title":        player.Title,
				"last_seen":    player.LastSeen,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"online_players": onlinePlayers,
		"total_online":   len(onlinePlayers),
		"request_id":     c.GetHeader("X-Request-ID"),
	})
}

// GetPlayerPreferences endpoint pour récupérer les préférences d'un joueur
func (h *PlayerHandler) GetPlayerPreferences(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":      "User ID not found",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	playerResponse, err := h.playerService.GetPlayer(userID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to get player preferences")
		c.JSON(http.StatusNotFound, gin.H{
			"error":      "Player not found",
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"preferences": playerResponse.Player.Preferences,
		"request_id":  c.GetHeader("X-Request-ID"),
	})
}

// UpdatePlayerPreferences endpoint pour mettre à jour les préférences
func (h *PlayerHandler) UpdatePlayerPreferences(c *gin.Context) {
	var req struct {
		Preferences models.PlayerPreferences `json:"preferences" binding:"required"`
	}

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

	updateReq := models.UpdatePlayerRequest{
		Preferences: req.Preferences,
	}

	_, err := h.playerService.UpdatePlayer(userID, updateReq)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to update player preferences")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to update preferences",
			"message":    err.Error(),
			"request_id": c.GetHeader("X-Request-ID"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Preferences updated successfully",
		"preferences": req.Preferences,
		"request_id":  c.GetHeader("X-Request-ID"),
	})
}

