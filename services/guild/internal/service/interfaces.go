package service

import (
	"context"

	"github.com/dan-2/github_mmorpg/services/guild/internal/models"
	"github.com/google/uuid"
)

// GuildService définit les opérations sur les guildes
type GuildService interface {
	CreateGuild(ctx context.Context, req *models.CreateGuildRequest, creatorID uuid.UUID) (*models.GuildResponse, error)
	GetGuild(ctx context.Context, guildID uuid.UUID) (*models.GuildResponse, error)
	UpdateGuild(ctx context.Context, guildID uuid.UUID, req *models.UpdateGuildRequest, playerID uuid.UUID) (*models.GuildResponse, error)
	DeleteGuild(ctx context.Context, guildID uuid.UUID, playerID uuid.UUID) error
	SearchGuilds(ctx context.Context, req *models.GuildSearchRequest) (*models.GuildSearchResponse, error)
	ListGuilds(ctx context.Context, page, limit int) (*models.GuildSearchResponse, error)
	GetGuildStats(ctx context.Context, guildID uuid.UUID) (*models.GuildStatsResponse, error)
}

// GuildMemberService définit les opérations sur les membres
type GuildMemberService interface {
	JoinGuild(ctx context.Context, guildID, playerID uuid.UUID) error
	LeaveGuild(ctx context.Context, guildID, playerID uuid.UUID) error
	KickMember(ctx context.Context, guildID, targetPlayerID, playerID uuid.UUID) error
	UpdateMemberRole(ctx context.Context, guildID uuid.UUID, req *models.UpdateMemberRoleRequest,
		playerID uuid.UUID) error
	GetMembers(ctx context.Context, guildID uuid.UUID, page, limit int) ([]*models.GuildMemberResponse, int, error)
	GetMember(ctx context.Context, guildID, playerID uuid.UUID) (*models.GuildMemberResponse, error)
	UpdateLastSeen(ctx context.Context, playerID uuid.UUID) error
}

// GuildInvitationService définit les opérations sur les invitations
type GuildInvitationService interface {
	InvitePlayer(ctx context.Context, guildID uuid.UUID, req *models.InvitePlayerRequest, inviterID uuid.UUID) error
	AcceptInvitation(ctx context.Context, invitationID, playerID uuid.UUID) error
	DeclineInvitation(ctx context.Context, invitationID, playerID uuid.UUID) error
	GetInvitations(ctx context.Context, guildID uuid.UUID, status *string, page, limit int) ([]*models.GuildInvitationResponse, int, error)
	GetPlayerInvitations(ctx context.Context, playerID uuid.UUID, status *string, page, limit int) (
		[]*models.GuildInvitationResponse, int, error)
	CancelInvitation(ctx context.Context, invitationID, playerID uuid.UUID) error
}

// GuildApplicationService définit les opérations sur les candidatures
type GuildApplicationService interface {
	ApplyToGuild(ctx context.Context, req *models.ApplyToGuildRequest, playerID uuid.UUID) error
	AcceptApplication(ctx context.Context, applicationID, playerID uuid.UUID) error
	DeclineApplication(ctx context.Context, applicationID, playerID uuid.UUID) error
	GetApplications(ctx context.Context, guildID uuid.UUID, status *string, page, limit int) ([]*models.GuildApplicationResponse, int, error)
	GetPlayerApplications(ctx context.Context, playerID uuid.UUID, status *string, page, limit int) (
		[]*models.GuildApplicationResponse, int, error)
	CancelApplication(ctx context.Context, applicationID, playerID uuid.UUID) error
}

// GuildBankService définit les opérations sur la banque de guilde
type GuildBankService interface {
	GetBank(ctx context.Context, guildID uuid.UUID) (*models.GuildBankResponse, error)
	DepositGold(ctx context.Context, guildID uuid.UUID, req *models.GuildBankTransactionRequest, playerID uuid.UUID) error
	WithdrawGold(ctx context.Context, guildID uuid.UUID, req *models.GuildBankTransactionRequest, playerID uuid.UUID) error
	GetTransactions(ctx context.Context, guildID uuid.UUID, page, limit int) ([]*models.GuildBankTransactionResponse, int, error)
	GetPlayerTransactions(ctx context.Context, playerID uuid.UUID, page, limit int) ([]*models.GuildBankTransactionResponse, int, error)
}

// GuildWarService définit les opérations sur les guerres de guilde
type GuildWarService interface {
	DeclareWar(ctx context.Context, guildID uuid.UUID, req *models.CreateGuildWarRequest, playerID uuid.UUID) error
	GetWars(ctx context.Context, guildID uuid.UUID, status *string, page, limit int) ([]*models.GuildWarResponse, int, error)
	GetActiveWars(ctx context.Context, guildID uuid.UUID) ([]*models.GuildWarResponse, error)
	EndWar(ctx context.Context, warID uuid.UUID, winnerID *uuid.UUID, playerID uuid.UUID) error
	UpdateWarScore(ctx context.Context, warID uuid.UUID, guildID uuid.UUID, score int, playerID uuid.UUID) error
}

// GuildAllianceService définit les opérations sur les alliances
type GuildAllianceService interface {
	CreateAlliance(ctx context.Context, guildID uuid.UUID, req *models.CreateAllianceRequest, playerID uuid.UUID) error
	AcceptAlliance(ctx context.Context, allianceID, playerID uuid.UUID) error
	DeclineAlliance(ctx context.Context, allianceID, playerID uuid.UUID) error
	GetAlliances(ctx context.Context, guildID uuid.UUID, status *string, page, limit int) ([]*models.GuildAllianceResponse, int, error)
	GetActiveAlliances(ctx context.Context, guildID uuid.UUID) ([]*models.GuildAllianceResponse, error)
	EndAlliance(ctx context.Context, allianceID, playerID uuid.UUID) error
}

// GuildLogService définit les opérations sur les logs
type GuildLogService interface {
	AddLog(ctx context.Context, guildID, playerID uuid.UUID, action, details string) error
	GetLogs(ctx context.Context, guildID uuid.UUID, action *string, page, limit int) ([]*models.GuildLogResponse, int, error)
	CleanOldLogs(ctx context.Context, days int) error
}

// GuildPermissionService définit les opérations sur les permissions
type GuildPermissionService interface {
	GetPermissions(ctx context.Context, guildID, playerID uuid.UUID) (*models.GuildPermissionResponse, error)
	HasPermission(ctx context.Context, guildID, playerID uuid.UUID, permission string) (bool, error)
}
