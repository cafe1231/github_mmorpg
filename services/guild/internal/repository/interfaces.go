package repository

import (
	"context"

	"github.com/dan-2/github_mmorpg/services/guild/internal/models"
	"github.com/google/uuid"
)

// GuildRepository définit les opérations sur les guildes
type GuildRepository interface {
	Create(ctx context.Context, guild *models.Guild) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Guild, error)
	GetByName(ctx context.Context, name string) (*models.Guild, error)
	GetByTag(ctx context.Context, tag string) (*models.Guild, error)
	Update(ctx context.Context, guild *models.Guild) error
	Delete(ctx context.Context, id uuid.UUID) error
	Search(ctx context.Context, name, tag *string, minLevel, maxLevel *int, page, limit int) ([]*models.Guild, int, error)
	List(ctx context.Context, page, limit int) ([]*models.Guild, int, error)
}

// GuildMemberRepository définit les opérations sur les membres de guilde
type GuildMemberRepository interface {
	Create(ctx context.Context, member *models.GuildMember) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.GuildMember, error)
	GetByGuildAndPlayer(ctx context.Context, guildID, playerID uuid.UUID) (*models.GuildMember, error)
	GetByGuild(ctx context.Context, guildID uuid.UUID, page, limit int) ([]*models.GuildMember, int, error)
	GetByPlayer(ctx context.Context, playerID uuid.UUID) (*models.GuildMember, error)
	Update(ctx context.Context, member *models.GuildMember) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetMemberCount(ctx context.Context, guildID uuid.UUID) (int, error)
	UpdateLastSeen(ctx context.Context, memberID uuid.UUID) error
	UpdateContribution(ctx context.Context, memberID uuid.UUID, contribution int64) error
}

// GuildInvitationRepository définit les opérations sur les invitations
type GuildInvitationRepository interface {
	Create(ctx context.Context, invitation *models.GuildInvitation) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.GuildInvitation, error)
	GetByGuild(ctx context.Context, guildID uuid.UUID, status *string, page, limit int) ([]*models.GuildInvitation, int, error)
	GetByPlayer(ctx context.Context, playerID uuid.UUID, status *string, page, limit int) ([]*models.GuildInvitation, int, error)
	Update(ctx context.Context, invitation *models.GuildInvitation) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}

// GuildApplicationRepository définit les opérations sur les candidatures
type GuildApplicationRepository interface {
	Create(ctx context.Context, application *models.GuildApplication) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.GuildApplication, error)
	GetByGuild(ctx context.Context, guildID uuid.UUID, status *string, page, limit int) ([]*models.GuildApplication, int, error)
	GetByPlayer(ctx context.Context, playerID uuid.UUID, status *string, page, limit int) ([]*models.GuildApplication, int, error)
	Update(ctx context.Context, application *models.GuildApplication) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// GuildLogRepository définit les opérations sur les logs de guilde
type GuildLogRepository interface {
	Create(ctx context.Context, log *models.GuildLog) error
	GetByGuild(ctx context.Context, guildID uuid.UUID, action *string, page, limit int) ([]*models.GuildLog, int, error)
	DeleteOldLogs(ctx context.Context, days int) error
}

// GuildBankRepository définit les opérations sur la banque de guilde
type GuildBankRepository interface {
	Create(ctx context.Context, bank *models.GuildBank) error
	GetByGuild(ctx context.Context, guildID uuid.UUID) (*models.GuildBank, error)
	Update(ctx context.Context, bank *models.GuildBank) error
	UpdateGold(ctx context.Context, guildID uuid.UUID, amount int64) error
}

// GuildBankTransactionRepository définit les opérations sur les transactions bancaires
type GuildBankTransactionRepository interface {
	Create(ctx context.Context, transaction *models.GuildBankTransaction) error
	GetByGuild(ctx context.Context, guildID uuid.UUID, page, limit int) ([]*models.GuildBankTransaction, int, error)
	GetByPlayer(ctx context.Context, playerID uuid.UUID, page, limit int) ([]*models.GuildBankTransaction, int, error)
}

// GuildWarRepository définit les opérations sur les guerres de guilde
type GuildWarRepository interface {
	Create(ctx context.Context, war *models.GuildWar) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.GuildWar, error)
	GetByGuild(ctx context.Context, guildID uuid.UUID, status *string, page, limit int) ([]*models.GuildWar, int, error)
	GetActive(ctx context.Context, guildID uuid.UUID) ([]*models.GuildWar, error)
	Update(ctx context.Context, war *models.GuildWar) error
	EndWar(ctx context.Context, warID uuid.UUID, winnerID *uuid.UUID) error
}

// GuildAllianceRepository définit les opérations sur les alliances
type GuildAllianceRepository interface {
	Create(ctx context.Context, alliance *models.GuildAlliance) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.GuildAlliance, error)
	GetByGuild(ctx context.Context, guildID uuid.UUID, status *string, page, limit int) ([]*models.GuildAlliance, int, error)
	GetActive(ctx context.Context, guildID uuid.UUID) ([]*models.GuildAlliance, error)
	Update(ctx context.Context, alliance *models.GuildAlliance) error
	EndAlliance(ctx context.Context, allianceID uuid.UUID) error
}
