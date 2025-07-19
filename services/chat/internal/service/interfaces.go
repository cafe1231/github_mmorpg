package service

import (
	"chat/internal/models"
	"context"
	"time"

	"github.com/google/uuid"
)

// ChatService interface principale pour les fonctionnalités de chat
type ChatService interface {
	// Gestion des channels
	CreateChannel(ctx context.Context, req *models.CreateChannelRequest, ownerID uuid.UUID) (*models.Channel, error)
	GetChannel(ctx context.Context, channelID uuid.UUID) (*models.Channel, error)
	UpdateChannel(ctx context.Context, channelID uuid.UUID, req *models.UpdateChannelRequest, userID uuid.UUID) (*models.Channel, error)
	DeleteChannel(ctx context.Context, channelID, userID uuid.UUID) error
	GetChannels(ctx context.Context, req *models.GetChannelsRequest, userID uuid.UUID) ([]*models.Channel, error)
	SearchChannels(ctx context.Context, query string, limit, offset int, userID uuid.UUID) ([]*models.Channel, error)

	// Gestion des membres
	JoinChannel(ctx context.Context, channelID, userID uuid.UUID, req *models.JoinChannelRequest) error
	LeaveChannel(ctx context.Context, channelID, userID uuid.UUID) error
	InviteUser(ctx context.Context, channelID, inviterID uuid.UUID, req *models.InviteUserRequest) error
	KickUser(ctx context.Context, channelID, kickerID, targetID uuid.UUID, reason string) error
	GetChannelMembers(ctx context.Context, channelID uuid.UUID, limit, offset int, requesterID uuid.UUID) ([]*models.ChannelMember, error)
	UpdateMember(ctx context.Context, channelID, userID uuid.UUID, req *models.UpdateMemberRequest, requesterID uuid.UUID) error

	// Gestion des messages
	SendMessage(ctx context.Context, channelID, userID uuid.UUID, req *models.SendMessageRequest) (*models.Message, error)
	GetMessages(ctx context.Context, channelID uuid.UUID, req *models.GetMessagesRequest, userID uuid.UUID) ([]*models.Message, error)
	EditMessage(ctx context.Context, messageID, userID uuid.UUID, req *models.EditMessageRequest) (*models.Message, error)
	DeleteMessage(ctx context.Context, messageID, userID uuid.UUID) error
	SearchMessages(ctx context.Context, req *models.SearchMessagesRequest, userID uuid.UUID) ([]*models.Message, error)

	// Réactions
	AddReaction(ctx context.Context, messageID, userID uuid.UUID, req *models.ReactToMessageRequest) error
	RemoveReaction(ctx context.Context, messageID, userID uuid.UUID, emoji string) error

	// Modération
	ModerateUser(ctx context.Context, channelID, moderatorID uuid.UUID, req *models.ModerationRequest) error
	ModerateMessage(ctx context.Context, messageID, moderatorID uuid.UUID, req *models.ModerationRequest) error
	GetModerationLogs(ctx context.Context, channelID uuid.UUID, limit, offset int, requesterID uuid.UUID) ([]*models.ModerationLog, error)
}

// WebSocketService interface pour la gestion des connections WebSocket
type WebSocketService interface {
	// Gestion des connections
	RegisterConnection(userID uuid.UUID, conn WebSocketConnection) error
	UnregisterConnection(userID uuid.UUID) error
	GetConnection(userID uuid.UUID) (WebSocketConnection, bool)
	GetConnectedUsers(channelID uuid.UUID) []uuid.UUID

	// Diffusion de messages
	BroadcastToChannel(channelID uuid.UUID, event *models.ChatEvent) error
	BroadcastToUser(userID uuid.UUID, event *models.ChatEvent) error
	BroadcastToUsers(userIDs []uuid.UUID, event *models.ChatEvent) error

	// Gestion des channels
	JoinChannelWS(userID uuid.UUID, channelID string) error
	LeaveChannelWS(userID uuid.UUID, channelID string) error

	// Événements
	HandleMessage(userID uuid.UUID, message *models.WebSocketMessage) error
	HandleTyping(userID uuid.UUID, req *models.TypingIndicatorRequest) error
	HandlePing(userID uuid.UUID) error
}

// WebSocketConnection interface pour une connection WebSocket
type WebSocketConnection interface {
	Send(message []byte) error
	Close() error
	GetUserID() uuid.UUID
	GetConnectedAt() time.Time
	IsConnected() bool
}

// NotificationService interface pour les notifications
type NotificationService interface {
	// Notifications de messages
	NotifyNewMessage(ctx context.Context, message *models.Message, mentions []uuid.UUID) error
	NotifyMessageEdit(ctx context.Context, message *models.Message) error
	NotifyMessageDelete(ctx context.Context, messageID, channelID uuid.UUID) error

	// Notifications de channels
	NotifyChannelJoin(ctx context.Context, channelID, userID uuid.UUID) error
	NotifyChannelLeave(ctx context.Context, channelID, userID uuid.UUID) error
	NotifyChannelUpdate(ctx context.Context, channel *models.Channel) error

	// Notifications de modération
	NotifyModeration(ctx context.Context, log *models.ModerationLog) error

	// Notifications système
	NotifySystemMessage(ctx context.Context, channelID uuid.UUID, message string) error
}
