package models

import (
	"time"

	"github.com/google/uuid"
)

// ChannelType définit les types de channels de chat
type ChannelType string

const (
	ChannelTypeGlobal  ChannelType = "global"
	ChannelTypeZone    ChannelType = "zone"
	ChannelTypeGuild   ChannelType = "guild"
	ChannelTypeParty   ChannelType = "party"
	ChannelTypePrivate ChannelType = "private"
	ChannelTypeSystem  ChannelType = "system"
	ChannelTypeTrade   ChannelType = "trade"
	ChannelTypeGeneral ChannelType = "general"
)

// MessageType définit les types de messages
type MessageType string

const (
	MessageTypeText    MessageType = "text"
	MessageTypeSystem  MessageType = "system"
	MessageTypeJoin    MessageType = "join"
	MessageTypeLeave   MessageType = "leave"
	MessageTypeCommand MessageType = "command"
	MessageTypeEmote   MessageType = "emote"
)

// ModerationAction définit les actions de modération possibles
type ModerationAction string

const (
	ModerationActionMute   ModerationAction = "mute"
	ModerationActionBan    ModerationAction = "ban"
	ModerationActionKick   ModerationAction = "kick"
	ModerationActionWarn   ModerationAction = "warn"
	ModerationActionUnmute ModerationAction = "unmute"
	ModerationActionUnban  ModerationAction = "unban"
)

// Channel représente un channel de chat
type Channel struct {
	ID          uuid.UUID   `json:"id" db:"id"`
	Name        string      `json:"name" db:"name"`
	Type        ChannelType `json:"type" db:"type"`
	Description string      `json:"description" db:"description"`

	// Propriétaire et modération
	OwnerID     *uuid.UUID `json:"owner_id" db:"owner_id"`
	IsModerated bool       `json:"is_moderated" db:"is_moderated"`
	IsPrivate   bool       `json:"is_private" db:"is_private"`
	MaxMembers  int        `json:"max_members" db:"max_members"`

	// Données contextuelles
	ZoneID  *string    `json:"zone_id" db:"zone_id"`   // Pour les channels de zone
	GuildID *uuid.UUID `json:"guild_id" db:"guild_id"` // Pour les channels de guild
	PartyID *uuid.UUID `json:"party_id" db:"party_id"` // Pour les channels de party

	// Configuration
	Settings ChannelSettings `json:"settings" db:"settings"`

	// État
	IsActive    bool `json:"is_active" db:"is_active"`
	MemberCount int  `json:"member_count" db:"-"` // Calculé dynamiquement

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// Relations (chargées séparément)
	Members     []*ChannelMember `json:"members,omitempty" db:"-"`
	LastMessage *Message         `json:"last_message,omitempty" db:"-"`
}

// ChannelSettings représente les paramètres d'un channel
type ChannelSettings struct {
	AllowEmotes    bool     `json:"allow_emotes"`
	AllowLinks     bool     `json:"allow_links"`
	AllowImages    bool     `json:"allow_images"`
	SlowMode       int      `json:"slow_mode"`      // secondes entre messages
	RequiredLevel  int      `json:"required_level"` // niveau requis pour rejoindre
	BannedWords    []string `json:"banned_words"`
	AutoModEnabled bool     `json:"auto_mod_enabled"`
	LogMessages    bool     `json:"log_messages"`
}

// ChannelMember représente l'appartenance d'un utilisateur à un channel
type ChannelMember struct {
	ID          uuid.UUID `json:"id" db:"id"`
	ChannelID   uuid.UUID `json:"channel_id" db:"channel_id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	DisplayName string    `json:"display_name" db:"display_name"`

	// Permissions
	Role        string `json:"role" db:"role"` // member, moderator, admin
	CanModerate bool   `json:"can_moderate" db:"can_moderate"`
	CanInvite   bool   `json:"can_invite" db:"can_invite"`

	// État
	IsOnline   bool       `json:"is_online" db:"is_online"`
	IsMuted    bool       `json:"is_muted" db:"is_muted"`
	MutedUntil *time.Time `json:"muted_until" db:"muted_until"`

	// Timestamps
	JoinedAt   time.Time `json:"joined_at" db:"joined_at"`
	LastSeenAt time.Time `json:"last_seen_at" db:"last_seen_at"`

	// Relations
	User *UserInfo `json:"user,omitempty" db:"-"`
}

// Message représente un message de chat
type Message struct {
	ID        uuid.UUID   `json:"id" db:"id"`
	ChannelID uuid.UUID   `json:"channel_id" db:"channel_id"`
	UserID    uuid.UUID   `json:"user_id" db:"user_id"`
	Content   string      `json:"content" db:"content"`
	Type      MessageType `json:"type" db:"type"`

	// Contenu enrichi
	Mentions    []uuid.UUID `json:"mentions" db:"mentions"`       // IDs des utilisateurs mentionnés
	Attachments []string    `json:"attachments" db:"attachments"` // URLs des fichiers attachés
	Reactions   []Reaction  `json:"reactions" db:"reactions"`

	// Métadonnées
	IsEdited  bool       `json:"is_edited" db:"is_edited"`
	EditedAt  *time.Time `json:"edited_at" db:"edited_at"`
	IsDeleted bool       `json:"is_deleted" db:"is_deleted"`
	DeletedAt *time.Time `json:"deleted_at" db:"deleted_at"`

	// Modération
	IsModerated bool   `json:"is_moderated" db:"is_moderated"`
	ModReason   string `json:"mod_reason" db:"mod_reason"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// Relations (chargées séparément)
	User      *UserInfo  `json:"user,omitempty" db:"-"`
	Channel   *Channel   `json:"channel,omitempty" db:"-"`
	ReplyTo   *Message   `json:"reply_to,omitempty" db:"-"`
	ReplyToID *uuid.UUID `json:"reply_to_id" db:"reply_to_id"`
}

// Reaction représente une réaction emoji sur un message
type Reaction struct {
	Emoji   string      `json:"emoji"`
	Count   int         `json:"count"`
	UserIDs []uuid.UUID `json:"user_ids"`
}

// UserInfo représente les informations d'un utilisateur dans le contexte du chat
type UserInfo struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	Avatar      string    `json:"avatar"`
	Level       int       `json:"level"`
	Title       string    `json:"title"`
	IsOnline    bool      `json:"is_online"`
	LastSeen    time.Time `json:"last_seen"`
}

// ModerationLog représente un log d'action de modération
type ModerationLog struct {
	ID           uuid.UUID        `json:"id" db:"id"`
	ChannelID    uuid.UUID        `json:"channel_id" db:"channel_id"`
	ModeratorID  uuid.UUID        `json:"moderator_id" db:"moderator_id"`
	TargetUserID uuid.UUID        `json:"target_user_id" db:"target_user_id"`
	Action       ModerationAction `json:"action" db:"action"`
	Reason       string           `json:"reason" db:"reason"`
	Duration     *time.Duration   `json:"duration" db:"duration"` // Pour mute temporaire
	ExpiresAt    *time.Time       `json:"expires_at" db:"expires_at"`
	MessageID    *uuid.UUID       `json:"message_id" db:"message_id"` // Message concerné si applicable

	// Métadonnées
	Metadata map[string]interface{} `json:"metadata" db:"metadata"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// Relations
	Moderator  *UserInfo `json:"moderator,omitempty" db:"-"`
	TargetUser *UserInfo `json:"target_user,omitempty" db:"-"`
	Channel    *Channel  `json:"channel,omitempty" db:"-"`
}

// ChatEvent représente un événement de chat en temps réel
type ChatEvent struct {
	Type      string      `json:"type"`
	ChannelID uuid.UUID   `json:"channel_id"`
	UserID    uuid.UUID   `json:"user_id"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// WebSocketMessage représente un message WebSocket
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Event   string      `json:"event"`
	Channel string      `json:"channel,omitempty"`
	Data    interface{} `json:"data"`
}

// ConnectionInfo représente les informations d'une connection WebSocket
type ConnectionInfo struct {
	UserID      uuid.UUID `json:"user_id"`
	Username    string    `json:"username"`
	ConnectedAt time.Time `json:"connected_at"`
	LastPing    time.Time `json:"last_ping"`
	Channels    []string  `json:"channels"`
}
