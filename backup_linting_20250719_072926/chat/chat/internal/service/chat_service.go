package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"chat/internal/config"
	"chat/internal/models"
	"chat/internal/repository"
)

type chatService struct {
	channelRepo    repository.ChannelRepository
	messageRepo    repository.MessageRepository
	moderationRepo repository.ModerationRepository
	userRepo       repository.UserRepository
	config         *config.Config
}

// NewChatService crée une nouvelle instance du service de chat
func NewChatService(
	channelRepo repository.ChannelRepository,
	messageRepo repository.MessageRepository,
	moderationRepo repository.ModerationRepository,
	userRepo repository.UserRepository,
	cfg *config.Config,
) ChatService {
	return &chatService{
		channelRepo:    channelRepo,
		messageRepo:    messageRepo,
		moderationRepo: moderationRepo,
		userRepo:       userRepo,
		config:         cfg,
	}
}

// CreateChannel crée un nouveau channel
func (s *chatService) CreateChannel(ctx context.Context, req *models.CreateChannelRequest, ownerID uuid.UUID) (*models.Channel, error) {
	logrus.WithFields(logrus.Fields{
		"owner_id": ownerID,
		"name":     req.Name,
		"type":     req.Type,
	}).Info("Creating new channel")

	// Validation
	if req.Name == "" {
		return nil, fmt.Errorf("channel name is required")
	}

	if req.MaxMembers <= 0 {
		req.MaxMembers = 100 // Valeur par défaut
	}

	// Créer le channel
	channel := &models.Channel{
		ID:          uuid.New(),
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		OwnerID:     &ownerID,
		IsModerated: true, // Par défaut modéré
		IsPrivate:   req.IsPrivate,
		MaxMembers:  req.MaxMembers,
		ZoneID:      req.ZoneID,
		GuildID:     req.GuildID,
		PartyID:     req.PartyID,
		Settings:    req.Settings,
		IsActive:    true,
	}

	// Valeurs par défaut pour les settings
	if channel.Settings.AllowEmotes == false && channel.Settings.AllowLinks == false {
		channel.Settings = models.ChannelSettings{
			AllowEmotes:    true,
			AllowLinks:     true,
			AllowImages:    true,
			SlowMode:       0,
			RequiredLevel:  1,
			BannedWords:    []string{},
			AutoModEnabled: true,
			LogMessages:    true,
		}
	}

	if err := s.channelRepo.Create(ctx, channel); err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	// Ajouter le créateur comme admin du channel
	member := &models.ChannelMember{
		ID:          uuid.New(),
		ChannelID:   channel.ID,
		UserID:      ownerID,
		DisplayName: "Owner", // À récupérer du service Player
		Role:        "admin",
		CanModerate: true,
		CanInvite:   true,
		IsOnline:    true,
		IsMuted:     false,
		JoinedAt:    time.Now(),
		LastSeenAt:  time.Now(),
	}

	if err := s.channelRepo.AddMember(ctx, member); err != nil {
		logrus.WithError(err).Warn("Failed to add owner as member")
	}

	logrus.WithField("channel_id", channel.ID).Info("Channel created successfully")
	return channel, nil
}

// GetChannel récupère un channel par son ID
func (s *chatService) GetChannel(ctx context.Context, channelID uuid.UUID) (*models.Channel, error) {
	channel, err := s.channelRepo.GetByID(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}
	if channel == nil {
		return nil, fmt.Errorf("channel not found")
	}
	return channel, nil
}

// SendMessage envoie un message dans un channel
func (s *chatService) SendMessage(ctx context.Context, channelID, userID uuid.UUID, req *models.SendMessageRequest) (*models.Message, error) {
	logrus.WithFields(logrus.Fields{
		"channel_id": channelID,
		"user_id":    userID,
		"content":    req.Content[:min(50, len(req.Content))], // Log première partie seulement
	}).Info("Sending message")

	// Vérifier que l'utilisateur est membre du channel
	isMember, err := s.channelRepo.IsMember(ctx, channelID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if !isMember {
		return nil, fmt.Errorf("user is not a member of this channel")
	}

	// Vérifier si l'utilisateur est mute
	isMuted, _, err := s.moderationRepo.IsUserMuted(ctx, channelID, userID)
	if err != nil {
		logrus.WithError(err).Warn("Failed to check if user is muted")
	}
	if isMuted {
		return nil, fmt.Errorf("user is muted in this channel")
	}

	// Validation du contenu
	if len(req.Content) > s.config.Chat.MaxMessageLength {
		return nil, fmt.Errorf("message too long (max %d characters)", s.config.Chat.MaxMessageLength)
	}

	// TODO: Validation anti-spam
	// TODO: Filtrage de contenu

	// Créer le message
	messageType := req.Type
	if messageType == "" {
		messageType = models.MessageTypeText
	}

	message := &models.Message{
		ID:          uuid.New(),
		ChannelID:   channelID,
		UserID:      userID,
		Content:     req.Content,
		Type:        messageType,
		Mentions:    req.Mentions,
		Attachments: req.Attachments,
		Reactions:   []models.Reaction{},
		ReplyToID:   req.ReplyToID,
		IsEdited:    false,
		IsDeleted:   false,
		IsModerated: false,
	}

	if err := s.messageRepo.Create(ctx, message); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	logrus.WithField("message_id", message.ID).Info("Message sent successfully")
	return message, nil
}

// GetMessages récupère les messages d'un channel
func (s *chatService) GetMessages(ctx context.Context, channelID uuid.UUID, req *models.GetMessagesRequest, userID uuid.UUID) ([]*models.Message, error) {
	// Vérifier que l'utilisateur est membre du channel
	isMember, err := s.channelRepo.IsMember(ctx, channelID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if !isMember {
		return nil, fmt.Errorf("user is not a member of this channel")
	}

	// Valeurs par défaut
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 50
	}

	messages, err := s.messageRepo.GetChannelMessages(ctx, channelID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	return messages, nil
}

// JoinChannel fait rejoindre un utilisateur à un channel
func (s *chatService) JoinChannel(ctx context.Context, channelID, userID uuid.UUID, req *models.JoinChannelRequest) error {
	logrus.WithFields(logrus.Fields{
		"channel_id": channelID,
		"user_id":    userID,
	}).Info("User joining channel")

	// Vérifier que le channel existe
	channel, err := s.channelRepo.GetByID(ctx, channelID)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}
	if channel == nil {
		return fmt.Errorf("channel not found")
	}

	// Vérifier si l'utilisateur est déjà membre
	isMember, err := s.channelRepo.IsMember(ctx, channelID, userID)
	if err != nil {
		return fmt.Errorf("failed to check membership: %w", err)
	}
	if isMember {
		return fmt.Errorf("user is already a member of this channel")
	}

	// Vérifier les limites de membres
	memberCount, err := s.channelRepo.GetMemberCount(ctx, channelID)
	if err != nil {
		return fmt.Errorf("failed to get member count: %w", err)
	}
	if memberCount >= channel.MaxMembers {
		return fmt.Errorf("channel is full")
	}

	// TODO: Vérifier le mot de passe pour les channels privés
	// TODO: Vérifier le niveau requis

	// Récupérer les infos utilisateur
	userInfo, err := s.userRepo.GetUserInfo(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user info: %w", err)
	}

	// Ajouter l'utilisateur comme membre
	member := &models.ChannelMember{
		ID:          uuid.New(),
		ChannelID:   channelID,
		UserID:      userID,
		DisplayName: userInfo.DisplayName,
		Role:        "member",
		CanModerate: false,
		CanInvite:   false,
		IsOnline:    true,
		IsMuted:     false,
		JoinedAt:    time.Now(),
		LastSeenAt:  time.Now(),
	}

	if err := s.channelRepo.AddMember(ctx, member); err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"channel_id": channelID,
		"user_id":    userID,
	}).Info("User joined channel successfully")

	return nil
}

// LeaveChannel fait quitter un utilisateur d'un channel
func (s *chatService) LeaveChannel(ctx context.Context, channelID, userID uuid.UUID) error {
	logrus.WithFields(logrus.Fields{
		"channel_id": channelID,
		"user_id":    userID,
	}).Info("User leaving channel")

	// Vérifier que l'utilisateur est membre
	isMember, err := s.channelRepo.IsMember(ctx, channelID, userID)
	if err != nil {
		return fmt.Errorf("failed to check membership: %w", err)
	}
	if !isMember {
		return fmt.Errorf("user is not a member of this channel")
	}

	// Retirer l'utilisateur
	if err := s.channelRepo.RemoveMember(ctx, channelID, userID); err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"channel_id": channelID,
		"user_id":    userID,
	}).Info("User left channel successfully")

	return nil
}

// Implémentations simplifiées pour les autres méthodes de l'interface
func (s *chatService) UpdateChannel(ctx context.Context, channelID uuid.UUID, req *models.UpdateChannelRequest, userID uuid.UUID) (*models.Channel, error) {
	// TODO: Implémenter
	return nil, fmt.Errorf("not implemented")
}

func (s *chatService) DeleteChannel(ctx context.Context, channelID uuid.UUID, userID uuid.UUID) error {
	// TODO: Implémenter
	return fmt.Errorf("not implemented")
}

func (s *chatService) GetChannels(ctx context.Context, req *models.GetChannelsRequest, userID uuid.UUID) ([]*models.Channel, error) {
	// TODO: Implémenter
	return []*models.Channel{}, nil
}

func (s *chatService) SearchChannels(ctx context.Context, query string, limit, offset int, userID uuid.UUID) ([]*models.Channel, error) {
	// TODO: Implémenter
	return []*models.Channel{}, nil
}

func (s *chatService) InviteUser(ctx context.Context, channelID, inviterID uuid.UUID, req *models.InviteUserRequest) error {
	// TODO: Implémenter
	return fmt.Errorf("not implemented")
}

func (s *chatService) KickUser(ctx context.Context, channelID, kickerID, targetID uuid.UUID, reason string) error {
	// TODO: Implémenter
	return fmt.Errorf("not implemented")
}

func (s *chatService) GetChannelMembers(ctx context.Context, channelID uuid.UUID, limit, offset int, requesterID uuid.UUID) ([]*models.ChannelMember, error) {
	return s.channelRepo.GetMembers(ctx, channelID, limit, offset)
}

func (s *chatService) UpdateMember(ctx context.Context, channelID, userID uuid.UUID, req *models.UpdateMemberRequest, requesterID uuid.UUID) error {
	// TODO: Implémenter
	return fmt.Errorf("not implemented")
}

func (s *chatService) EditMessage(ctx context.Context, messageID, userID uuid.UUID, req *models.EditMessageRequest) (*models.Message, error) {
	// TODO: Implémenter
	return nil, fmt.Errorf("not implemented")
}

func (s *chatService) DeleteMessage(ctx context.Context, messageID, userID uuid.UUID) error {
	// TODO: Implémenter
	return fmt.Errorf("not implemented")
}

func (s *chatService) SearchMessages(ctx context.Context, req *models.SearchMessagesRequest, userID uuid.UUID) ([]*models.Message, error) {
	// TODO: Implémenter
	return []*models.Message{}, nil
}

func (s *chatService) AddReaction(ctx context.Context, messageID, userID uuid.UUID, req *models.ReactToMessageRequest) error {
	// TODO: Implémenter
	return fmt.Errorf("not implemented")
}

func (s *chatService) RemoveReaction(ctx context.Context, messageID, userID uuid.UUID, emoji string) error {
	// TODO: Implémenter
	return fmt.Errorf("not implemented")
}

func (s *chatService) ModerateUser(ctx context.Context, channelID, moderatorID uuid.UUID, req *models.ModerationRequest) error {
	// TODO: Implémenter
	return fmt.Errorf("not implemented")
}

func (s *chatService) ModerateMessage(ctx context.Context, messageID, moderatorID uuid.UUID, req *models.ModerationRequest) error {
	// TODO: Implémenter
	return fmt.Errorf("not implemented")
}

func (s *chatService) GetModerationLogs(ctx context.Context, channelID uuid.UUID, limit, offset int, requesterID uuid.UUID) ([]*models.ModerationLog, error) {
	// TODO: Implémenter
	return []*models.ModerationLog{}, nil
}

// Fonction utilitaire
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
