package repository

import (
	"context"
	"time"

	"chat/internal/models"

	"github.com/google/uuid"
)

// ChannelRepository interface pour la gestion des channels
type ChannelRepository interface {
	// CRUD de base
	Create(ctx context.Context, channel *models.Channel) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Channel, error)
	Update(ctx context.Context, channel *models.Channel) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Recherche et filtrage
	GetByType(ctx context.Context, channelType models.ChannelType, limit, offset int) ([]*models.Channel, error)
	GetByZone(ctx context.Context, zoneID string, limit, offset int) ([]*models.Channel, error)
	GetByGuild(ctx context.Context, guildID uuid.UUID, limit, offset int) ([]*models.Channel, error)
	GetUserChannels(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Channel, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*models.Channel, error)

	// Gestion des membres
	AddMember(ctx context.Context, member *models.ChannelMember) error
	RemoveMember(ctx context.Context, channelID, userID uuid.UUID) error
	GetMembers(ctx context.Context, channelID uuid.UUID, limit, offset int) ([]*models.ChannelMember, error)
	GetMember(ctx context.Context, channelID, userID uuid.UUID) (*models.ChannelMember, error)
	UpdateMember(ctx context.Context, member *models.ChannelMember) error
	IsMember(ctx context.Context, channelID, userID uuid.UUID) (bool, error)
	GetMemberCount(ctx context.Context, channelID uuid.UUID) (int, error)

	// Statistiques
	GetChannelStats(ctx context.Context, channelID uuid.UUID) (map[string]interface{}, error)
}

// MessageRepository interface pour la gestion des messages
type MessageRepository interface {
	// CRUD de base
	Create(ctx context.Context, message *models.Message) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Message, error)
	Update(ctx context.Context, message *models.Message) error
	Delete(ctx context.Context, id uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID, reason string) error

	// Récupération des messages
	GetChannelMessages(ctx context.Context, channelID uuid.UUID, req *models.GetMessagesRequest) ([]*models.Message, error)
	GetUserMessages(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Message, error)
	GetReplies(ctx context.Context, messageID uuid.UUID, limit, offset int) ([]*models.Message, error)

	// Recherche
	Search(ctx context.Context, req *models.SearchMessagesRequest) ([]*models.Message, error)
	SearchInChannel(ctx context.Context, channelID uuid.UUID, query string, limit, offset int) ([]*models.Message, error)

	// Réactions
	AddReaction(ctx context.Context, messageID uuid.UUID, userID uuid.UUID, emoji string) error
	RemoveReaction(ctx context.Context, messageID uuid.UUID, userID uuid.UUID, emoji string) error
	GetReactions(ctx context.Context, messageID uuid.UUID) ([]models.Reaction, error)

	// Mentions
	GetMentions(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Message, error)
	MarkMentionAsRead(ctx context.Context, messageID, userID uuid.UUID) error

	// Nettoyage et maintenance
	DeleteOldMessages(ctx context.Context, days int) (int, error)
	GetMessageCount(ctx context.Context, channelID uuid.UUID) (int, error)
}

// ModerationRepository interface pour la gestion de la modération
type ModerationRepository interface {
	// Logs de modération
	CreateLog(ctx context.Context, log *models.ModerationLog) error
	GetLogs(ctx context.Context, channelID uuid.UUID, limit, offset int) ([]*models.ModerationLog, error)
	GetUserLogs(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.ModerationLog, error)
	GetModeratorLogs(ctx context.Context, moderatorID uuid.UUID, limit, offset int) ([]*models.ModerationLog, error)

	// État de modération des utilisateurs
	IsUserMuted(ctx context.Context, channelID, userID uuid.UUID) (bool, *time.Time, error)
	IsUserBanned(ctx context.Context, channelID, userID uuid.UUID) (bool, *time.Time, error)
	MuteUser(ctx context.Context, channelID, userID, moderatorID uuid.UUID, duration time.Duration, reason string) error
	UnmuteUser(ctx context.Context, channelID, userID, moderatorID uuid.UUID) error
	BanUser(ctx context.Context, channelID, userID, moderatorID uuid.UUID, reason string) error
	UnbanUser(ctx context.Context, channelID, userID, moderatorID uuid.UUID) error

	// Nettoyage automatique
	CleanExpiredModerations(ctx context.Context) (int, error)
	GetActiveModerations(ctx context.Context, userID uuid.UUID) ([]*models.ModerationLog, error)
}

// UserRepository interface pour les informations d'utilisateurs dans le contexte du chat
type UserRepository interface {
	// Informations utilisateur
	GetUserInfo(ctx context.Context, userID uuid.UUID) (*models.UserInfo, error)
	GetUsersInfo(ctx context.Context, userIDs []uuid.UUID) ([]*models.UserInfo, error)
	UpdateUserOnlineStatus(ctx context.Context, userID uuid.UUID, isOnline bool) error

	// Recherche d'utilisateurs
	SearchUsers(ctx context.Context, query string, limit, offset int) ([]*models.UserInfo, error)
	GetOnlineUsers(ctx context.Context, channelID uuid.UUID) ([]*models.UserInfo, error)

	// Blocage et amitié (pour messages privés)
	IsUserBlocked(ctx context.Context, userID, blockedUserID uuid.UUID) (bool, error)
	BlockUser(ctx context.Context, userID, blockedUserID uuid.UUID) error
	UnblockUser(ctx context.Context, userID, blockedUserID uuid.UUID) error
	GetBlockedUsers(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}
