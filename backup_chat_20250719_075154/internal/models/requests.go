package models

import (
	"github.com/google/uuid"
)

// CreateChannelRequest représente une requête de création de channel
type CreateChannelRequest struct {
	Name        string          `json:"name" binding:"required,min=1,max=50"`
	Type        ChannelType     `json:"type" binding:"required"`
	Description string          `json:"description" binding:"max=500"`
	IsPrivate   bool            `json:"is_private"`
	MaxMembers  int             `json:"max_members" binding:"min=2,max=1000"`
	ZoneID      *string         `json:"zone_id,omitempty"`
	GuildID     *uuid.UUID      `json:"guild_id,omitempty"`
	PartyID     *uuid.UUID      `json:"party_id,omitempty"`
	Settings    ChannelSettings `json:"settings"`
}

// UpdateChannelRequest représente une requête de mise à jour de channel
type UpdateChannelRequest struct {
	Name        *string          `json:"name,omitempty" binding:"omitempty,min=1,max=50"`
	Description *string          `json:"description,omitempty" binding:"omitempty,max=500"`
	IsPrivate   *bool            `json:"is_private,omitempty"`
	MaxMembers  *int             `json:"max_members,omitempty" binding:"omitempty,min=2,max=1000"`
	Settings    *ChannelSettings `json:"settings,omitempty"`
}

// SendMessageRequest représente une requête d'envoi de message
type SendMessageRequest struct {
	Content     string      `json:"content" binding:"required,min=1,max=2000"`
	Type        MessageType `json:"type,omitempty"`
	Mentions    []uuid.UUID `json:"mentions,omitempty"`
	ReplyToID   *uuid.UUID  `json:"reply_to_id,omitempty"`
	Attachments []string    `json:"attachments,omitempty"`
}

// EditMessageRequest représente une requête de modification de message
type EditMessageRequest struct {
	Content string `json:"content" binding:"required,min=1,max=2000"`
}

// JoinChannelRequest représente une requête pour rejoindre un channel
type JoinChannelRequest struct {
	Password string `json:"password,omitempty"`
}

// InviteUserRequest représente une requête d'invitation d'utilisateur
type InviteUserRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
	Role   string    `json:"role,omitempty"` // member, moderator
}

// UpdateMemberRequest représente une requête de mise à jour de membre
type UpdateMemberRequest struct {
	Role        *string `json:"role,omitempty"`
	CanModerate *bool   `json:"can_moderate,omitempty"`
	CanInvite   *bool   `json:"can_invite,omitempty"`
}

// ModerationRequest représente une requête d'action de modération
type ModerationRequest struct {
	Action    ModerationAction `json:"action" binding:"required"`
	TargetID  uuid.UUID        `json:"target_id" binding:"required"` // UserID ou MessageID
	Reason    string           `json:"reason" binding:"required,min=1,max=500"`
	Duration  *int             `json:"duration,omitempty"` // en minutes pour mute temporaire
	MessageID *uuid.UUID       `json:"message_id,omitempty"`
}

// ReactToMessageRequest représente une requête de réaction à un message
type ReactToMessageRequest struct {
	Emoji string `json:"emoji" binding:"required,min=1,max=10"`
}

// SearchMessagesRequest représente une requête de recherche de messages
type SearchMessagesRequest struct {
	Query     string     `json:"query" form:"query" binding:"required,min=1"`
	ChannelID *uuid.UUID `json:"channel_id" form:"channel_id"`
	UserID    *uuid.UUID `json:"user_id" form:"user_id"`
	Type      *string    `json:"type" form:"type"`
	Limit     int        `json:"limit" form:"limit" binding:"min=1,max=100"`
	Offset    int        `json:"offset" form:"offset" binding:"min=0"`
}

// GetMessagesRequest représente une requête de récupération de messages
type GetMessagesRequest struct {
	Limit  int        `json:"limit" form:"limit" binding:"min=1,max=100"`
	Before *uuid.UUID `json:"before" form:"before"`
	After  *uuid.UUID `json:"after" form:"after"`
	UserID *uuid.UUID `json:"user_id" form:"user_id"`
	Type   *string    `json:"type" form:"type"`
}

// GetChannelsRequest représente une requête de récupération de channels
type GetChannelsRequest struct {
	Type    *ChannelType `json:"type" form:"type"`
	ZoneID  *string      `json:"zone_id" form:"zone_id"`
	GuildID *uuid.UUID   `json:"guild_id" form:"guild_id"`
	Limit   int          `json:"limit" form:"limit" binding:"min=1,max=100"`
	Offset  int          `json:"offset" form:"offset" binding:"min=0"`
}

// WebSocketAuthRequest représente une requête d'authentification WebSocket
type WebSocketAuthRequest struct {
	Token string `json:"token" binding:"required"`
}

// WebSocketJoinChannelRequest représente une requête de join WebSocket
type WebSocketJoinChannelRequest struct {
	ChannelID string `json:"channel_id" binding:"required"`
}

// WebSocketLeaveChannelRequest représente une requête de leave WebSocket
type WebSocketLeaveChannelRequest struct {
	ChannelID string `json:"channel_id" binding:"required"`
}

// WebSocketSendMessageRequest représente une requête d'envoi de message via WebSocket
type WebSocketSendMessageRequest struct {
	ChannelID string      `json:"channel_id" binding:"required"`
	Content   string      `json:"content" binding:"required,min=1,max=2000"`
	Type      MessageType `json:"type,omitempty"`
	Mentions  []uuid.UUID `json:"mentions,omitempty"`
	ReplyToID *uuid.UUID  `json:"reply_to_id,omitempty"`
}

// TypingIndicatorRequest représente une requête d'indicateur de frappe
type TypingIndicatorRequest struct {
	ChannelID string `json:"channel_id" binding:"required"`
	IsTyping  bool   `json:"is_typing"`
}
