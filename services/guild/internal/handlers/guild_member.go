package handlers

import (
	"net/http"
	"strconv"

	"github.com/dan-2/github_mmorpg/services/guild/internal/models"
	"github.com/dan-2/github_mmorpg/services/guild/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GuildMemberHandler gère les requêtes liées aux membres de guilde
type GuildMemberHandler struct {
	memberService service.GuildMemberService
}

// NewGuildMemberHandler crée une nouvelle instance de GuildMemberHandler
func NewGuildMemberHandler(memberService service.GuildMemberService) *GuildMemberHandler {
	return &GuildMemberHandler{
		memberService: memberService,
	}
}

// JoinGuild permet à un joueur de rejoindre une guilde
func (h *GuildMemberHandler) JoinGuild(c *gin.Context) {
	guildIDStr := c.Param("guild_id")
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

	err = h.memberService.JoinGuild(c.Request.Context(), guildID, playerID)
	if err != nil {
		switch err {
		case models.ErrGuildNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case models.ErrAlreadyInGuild:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur interne du serveur"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Joueur ajouté à la guilde avec succès"})
}

// LeaveGuild permet à un joueur de quitter une guilde
func (h *GuildMemberHandler) LeaveGuild(c *gin.Context) {
	guildIDStr := c.Param("guild_id")
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

	err = h.memberService.LeaveGuild(c.Request.Context(), guildID, playerID)
	if err != nil {
		switch err {
		case models.ErrGuildNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case models.ErrNotInGuild:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur interne du serveur"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Joueur retiré de la guilde avec succès"})
}

// KickMember permet d'expulser un membre de la guilde
func (h *GuildMemberHandler) KickMember(c *gin.Context) {
	guildIDStr := c.Param("guild_id")
	guildID, err := uuid.Parse(guildIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de guilde invalide"})
		return
	}

	targetPlayerIDStr := c.Param("player_id")
	targetPlayerID, err := uuid.Parse(targetPlayerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de joueur cible invalide"})
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

	err = h.memberService.KickMember(c.Request.Context(), guildID, targetPlayerID, playerID)
	if err != nil {
		switch err {
		case models.ErrGuildNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case models.ErrInsufficientPermissions:
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case models.ErrNotInGuild:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur interne du serveur"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Membre expulsé avec succès"})
}

// UpdateMemberRole met à jour le rôle d'un membre
func (h *GuildMemberHandler) UpdateMemberRole(c *gin.Context) {
	guildIDStr := c.Param("guild_id")
	guildID, err := uuid.Parse(guildIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de guilde invalide"})
		return
	}

	var req models.UpdateMemberRoleRequest
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

	err = h.memberService.UpdateMemberRole(c.Request.Context(), guildID, &req, playerID)
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

	c.JSON(http.StatusOK, gin.H{"message": "Rôle mis à jour avec succès"})
}

// GetMembers récupère la liste des membres d'une guilde
func (h *GuildMemberHandler) GetMembers(c *gin.Context) {
	guildIDStr := c.Param("guild_id")
	guildID, err := uuid.Parse(guildIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de guilde invalide"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	members, total, err := h.memberService.GetMembers(c.Request.Context(), guildID, page, limit)
	if err != nil {
		if err == models.ErrGuildNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur interne du serveur"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"members": members,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// GetMember récupère les informations d'un membre spécifique
func (h *GuildMemberHandler) GetMember(c *gin.Context) {
	guildIDStr := c.Param("guild_id")
	guildID, err := uuid.Parse(guildIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de guilde invalide"})
		return
	}

	playerIDStr := c.Param("player_id")
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de joueur invalide"})
		return
	}

	member, err := h.memberService.GetMember(c.Request.Context(), guildID, playerID)
	if err != nil {
		if err == models.ErrGuildNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur interne du serveur"})
		}
		return
	}

	c.JSON(http.StatusOK, member)
}
