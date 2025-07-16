package service

import (
	"context"
	"time"

	"github.com/dan-2/github_mmorpg/services/guild/internal/models"
	"github.com/dan-2/github_mmorpg/services/guild/internal/repository"
	"github.com/google/uuid"
)

// guildService implémente GuildService
type guildService struct {
	guildRepo   repository.GuildRepository
	memberRepo  repository.GuildMemberRepository
	logRepo     repository.GuildLogRepository
	permService GuildPermissionService
}

// NewGuildService crée une nouvelle instance de GuildService
func NewGuildService(
	guildRepo repository.GuildRepository,
	memberRepo repository.GuildMemberRepository,
	logRepo repository.GuildLogRepository,
	permService GuildPermissionService,
) GuildService {
	return &guildService{
		guildRepo:   guildRepo,
		memberRepo:  memberRepo,
		logRepo:     logRepo,
		permService: permService,
	}
}

// CreateGuild crée une nouvelle guilde
func (s *guildService) CreateGuild(ctx context.Context, req *models.CreateGuildRequest, creatorID uuid.UUID) (*models.GuildResponse, error) {
	// Vérifier si le nom existe déjà
	existingGuild, _ := s.guildRepo.GetByName(ctx, req.Name)
	if existingGuild != nil {
		return nil, models.ErrGuildAlreadyExists
	}

	// Vérifier si le tag existe déjà
	existingTag, _ := s.guildRepo.GetByTag(ctx, req.Tag)
	if existingTag != nil {
		return nil, models.ErrGuildAlreadyExists
	}

	// Vérifier si le créateur est déjà dans une guilde
	existingMember, _ := s.memberRepo.GetByPlayer(ctx, creatorID)
	if existingMember != nil {
		return nil, models.ErrAlreadyInGuild
	}

	// Créer la guilde
	guild := &models.Guild{
		Name:        req.Name,
		Description: req.Description,
		Tag:         req.Tag,
		Level:       1,
		Experience:  0,
		MaxMembers:  50, // Valeur par défaut
	}

	err := s.guildRepo.Create(ctx, guild)
	if err != nil {
		return nil, err
	}

	// Ajouter le créateur comme leader
	member := &models.GuildMember{
		GuildID:  guild.ID,
		PlayerID: creatorID,
		Role:     "leader",
	}

	err = s.memberRepo.Create(ctx, member)
	if err != nil {
		return nil, err
	}

	// Ajouter un log
	s.logRepo.Create(ctx, &models.GuildLog{
		GuildID:   guild.ID,
		PlayerID:  creatorID,
		Action:    "guild_created",
		Details:   "Guilde créée",
		CreatedAt: time.Now(),
	})

	// Retourner la réponse
	return &models.GuildResponse{
		ID:          guild.ID.String(),
		Name:        guild.Name,
		Description: guild.Description,
		Tag:         guild.Tag,
		Level:       guild.Level,
		Experience:  guild.Experience,
		MaxMembers:  guild.MaxMembers,
		MemberCount: 1,
		CreatedAt:   guild.CreatedAt,
		UpdatedAt:   guild.UpdatedAt,
	}, nil
}

// GetGuild récupère une guilde
func (s *guildService) GetGuild(ctx context.Context, guildID uuid.UUID) (*models.GuildResponse, error) {
	guild, err := s.guildRepo.GetByID(ctx, guildID)
	if err != nil {
		return nil, err
	}

	// Compter les membres
	memberCount, err := s.memberRepo.GetMemberCount(ctx, guildID)
	if err != nil {
		return nil, err
	}

	return &models.GuildResponse{
		ID:          guild.ID.String(),
		Name:        guild.Name,
		Description: guild.Description,
		Tag:         guild.Tag,
		Level:       guild.Level,
		Experience:  guild.Experience,
		MaxMembers:  guild.MaxMembers,
		MemberCount: memberCount,
		CreatedAt:   guild.CreatedAt,
		UpdatedAt:   guild.UpdatedAt,
	}, nil
}

// UpdateGuild met à jour une guilde
func (s *guildService) UpdateGuild(ctx context.Context, guildID uuid.UUID, req *models.UpdateGuildRequest, playerID uuid.UUID) (*models.GuildResponse, error) {
	// Vérifier les permissions
	canEdit, err := s.permService.HasPermission(ctx, guildID, playerID, "edit_guild_info")
	if err != nil || !canEdit {
		return nil, models.ErrInsufficientPermissions
	}

	guild, err := s.guildRepo.GetByID(ctx, guildID)
	if err != nil {
		return nil, err
	}

	// Mettre à jour les champs fournis
	if req.Name != nil {
		// Vérifier si le nom existe déjà
		if *req.Name != guild.Name {
			existingGuild, _ := s.guildRepo.GetByName(ctx, *req.Name)
			if existingGuild != nil {
				return nil, models.ErrGuildAlreadyExists
			}
		}
		guild.Name = *req.Name
	}

	if req.Description != nil {
		guild.Description = *req.Description
	}

	if req.Tag != nil {
		// Vérifier si le tag existe déjà
		if *req.Tag != guild.Tag {
			existingTag, _ := s.guildRepo.GetByTag(ctx, *req.Tag)
			if existingTag != nil {
				return nil, models.ErrGuildAlreadyExists
			}
		}
		guild.Tag = *req.Tag
	}

	err = s.guildRepo.Update(ctx, guild)
	if err != nil {
		return nil, err
	}

	// Ajouter un log
	s.logRepo.Create(ctx, &models.GuildLog{
		GuildID:   guildID,
		PlayerID:  playerID,
		Action:    "guild_updated",
		Details:   "Informations de guilde mises à jour",
		CreatedAt: time.Now(),
	})

	return s.GetGuild(ctx, guildID)
}

// DeleteGuild supprime une guilde
func (s *guildService) DeleteGuild(ctx context.Context, guildID uuid.UUID, playerID uuid.UUID) error {
	// Vérifier les permissions
	canDelete, err := s.permService.HasPermission(ctx, guildID, playerID, "delete_guild")
	if err != nil || !canDelete {
		return models.ErrInsufficientPermissions
	}

	// Ajouter un log avant suppression
	s.logRepo.Create(ctx, &models.GuildLog{
		GuildID:   guildID,
		PlayerID:  playerID,
		Action:    "guild_deleted",
		Details:   "Guilde supprimée",
		CreatedAt: time.Now(),
	})

	return s.guildRepo.Delete(ctx, guildID)
}

// SearchGuilds recherche des guildes
func (s *guildService) SearchGuilds(ctx context.Context, req *models.GuildSearchRequest) (*models.GuildSearchResponse, error) {
	guilds, total, err := s.guildRepo.Search(ctx, req.Name, req.Tag, req.MinLevel, req.MaxLevel, req.Page, req.Limit)
	if err != nil {
		return nil, err
	}

	var responses []models.GuildResponse
	for _, guild := range guilds {
		memberCount, _ := s.memberRepo.GetMemberCount(ctx, guild.ID)
		responses = append(responses, models.GuildResponse{
			ID:          guild.ID.String(),
			Name:        guild.Name,
			Description: guild.Description,
			Tag:         guild.Tag,
			Level:       guild.Level,
			Experience:  guild.Experience,
			MaxMembers:  guild.MaxMembers,
			MemberCount: memberCount,
			CreatedAt:   guild.CreatedAt,
			UpdatedAt:   guild.UpdatedAt,
		})
	}

	return &models.GuildSearchResponse{
		Guilds: responses,
		Total:  total,
		Page:   req.Page,
		Limit:  req.Limit,
	}, nil
}

// ListGuilds liste les guildes
func (s *guildService) ListGuilds(ctx context.Context, page, limit int) (*models.GuildSearchResponse, error) {
	guilds, total, err := s.guildRepo.List(ctx, page, limit)
	if err != nil {
		return nil, err
	}

	var responses []models.GuildResponse
	for _, guild := range guilds {
		memberCount, _ := s.memberRepo.GetMemberCount(ctx, guild.ID)
		responses = append(responses, models.GuildResponse{
			ID:          guild.ID.String(),
			Name:        guild.Name,
			Description: guild.Description,
			Tag:         guild.Tag,
			Level:       guild.Level,
			Experience:  guild.Experience,
			MaxMembers:  guild.MaxMembers,
			MemberCount: memberCount,
			CreatedAt:   guild.CreatedAt,
			UpdatedAt:   guild.UpdatedAt,
		})
	}

	return &models.GuildSearchResponse{
		Guilds: responses,
		Total:  total,
		Page:   page,
		Limit:  limit,
	}, nil
}

// GetGuildStats récupère les statistiques d'une guilde
func (s *guildService) GetGuildStats(ctx context.Context, guildID uuid.UUID) (*models.GuildStatsResponse, error) {
	guild, err := s.guildRepo.GetByID(ctx, guildID)
	if err != nil {
		return nil, err
	}

	memberCount, err := s.memberRepo.GetMemberCount(ctx, guildID)
	if err != nil {
		return nil, err
	}

	// TODO: Implémenter les autres statistiques (membres en ligne, niveau moyen, etc.)
	// Pour l'instant, on retourne des valeurs par défaut

	return &models.GuildStatsResponse{
		GuildID:             guildID.String(),
		TotalMembers:        memberCount,
		OnlineMembers:       0, // TODO: Implémenter
		AverageLevel:        0, // TODO: Implémenter
		TotalExperience:     guild.Experience,
		BankGold:            0, // TODO: Implémenter
		ActiveWars:          0, // TODO: Implémenter
		ActiveAlliances:     0, // TODO: Implémenter
		PendingInvitations:  0, // TODO: Implémenter
		PendingApplications: 0, // TODO: Implémenter
	}, nil
}
