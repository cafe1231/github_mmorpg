package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"chat/internal/models"
	"chat/internal/service"
)

type ChatHandler struct {
	chatService service.ChatService
}

// NewChatHandler crée une nouvelle instance du handler de chat
func NewChatHandler(chatService service.ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

// CreateChannel crée un nouveau channel
// POST /api/v1/channels
func (h *ChatHandler) CreateChannel(c *gin.Context) {
	var req models.CreateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Récupérer l'ID utilisateur depuis le JWT (middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	ownerID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	channel, err := h.chatService.CreateChannel(c.Request.Context(), &req, ownerID)
	if err != nil {
		logrus.WithError(err).Error("Failed to create channel")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, channel)
}

// GetChannel récupère un channel par ID
// GET /api/v1/channels/:id
func (h *ChatHandler) GetChannel(c *gin.Context) {
	channelIDStr := c.Param("id")
	channelID, err := uuid.Parse(channelIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	channel, err := h.chatService.GetChannel(c.Request.Context(), channelID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get channel")
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, channel)
}

// JoinChannel fait rejoindre l'utilisateur à un channel
// POST /api/v1/channels/:id/join
func (h *ChatHandler) JoinChannel(c *gin.Context) {
	channelIDStr := c.Param("id")
	channelID, err := uuid.Parse(channelIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	var req models.JoinChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.chatService.JoinChannel(c.Request.Context(), channelID, userUUID, &req); err != nil {
		logrus.WithError(err).Error("Failed to join channel")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully joined channel"})
}

// LeaveChannel fait quitter l'utilisateur d'un channel
// POST /api/v1/channels/:id/leave
func (h *ChatHandler) LeaveChannel(c *gin.Context) {
	channelIDStr := c.Param("id")
	channelID, err := uuid.Parse(channelIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.chatService.LeaveChannel(c.Request.Context(), channelID, userUUID); err != nil {
		logrus.WithError(err).Error("Failed to leave channel")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully left channel"})
}

// SendMessage envoie un message dans un channel
// POST /api/v1/channels/:id/messages
func (h *ChatHandler) SendMessage(c *gin.Context) {
	channelIDStr := c.Param("id")
	channelID, err := uuid.Parse(channelIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	var req models.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	message, err := h.chatService.SendMessage(c.Request.Context(), channelID, userUUID, &req)
	if err != nil {
		logrus.WithError(err).Error("Failed to send message")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, message)
}

// GetMessages récupère les messages d'un channel
// GET /api/v1/channels/:id/messages
func (h *ChatHandler) GetMessages(c *gin.Context) {
	channelIDStr := c.Param("id")
	channelID, err := uuid.Parse(channelIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	// Parse query parameters
	var req models.GetMessagesRequest
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			req.Limit = limit
		}
	}
	if req.Limit <= 0 {
		req.Limit = 50
	}

	if beforeStr := c.Query("before"); beforeStr != "" {
		if beforeID, err := uuid.Parse(beforeStr); err == nil {
			req.Before = &beforeID
		}
	}

	if afterStr := c.Query("after"); afterStr != "" {
		if afterID, err := uuid.Parse(afterStr); err == nil {
			req.After = &afterID
		}
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	messages, err := h.chatService.GetMessages(c.Request.Context(), channelID, &req, userUUID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get messages")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"total":    len(messages),
		"limit":    req.Limit,
	})
}

// GetChannelMembers récupère les membres d'un channel
// GET /api/v1/channels/:id/members
func (h *ChatHandler) GetChannelMembers(c *gin.Context) {
	channelIDStr := c.Param("id")
	channelID, err := uuid.Parse(channelIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	// Parse query parameters
	limit := 50
	offset := 0
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	members, err := h.chatService.GetChannelMembers(c.Request.Context(), channelID, limit, offset, userUUID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get channel members")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"members": members,
		"total":   len(members),
		"limit":   limit,
		"offset":  offset,
	})
}
