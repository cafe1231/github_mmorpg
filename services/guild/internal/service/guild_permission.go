package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dan-2/github_mmorpg/services/guild/internal/models"
	"github.com/google/uuid"
)

// guildPermissionService implémente GuildPermissionService
type guildPermissionService struct {
	db *sql.DB
}

// NewGuildPermissionService crée une nouvelle instance de GuildPermissionService
func NewGuildPermissionService(db *sql.DB) GuildPermissionService {
	return &guildPermissionService{db: db}
}

// GetPermissions récupère toutes les permissions d'un joueur dans une guilde
func (s *guildPermissionService) GetPermissions(ctx context.Context, guildID, playerID uuid.UUID) (*models.GuildPermissionResponse, error) {
	query := `
		SELECT role, permissions
		FROM guild_members 
		WHERE guild_id = $1 AND player_id = $2
	`

	var role string
	var permissionsJSON string
	err := s.db.QueryRowContext(ctx, query, guildID, playerID).Scan(&role, &permissionsJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrGuildNotFound
		}
		return nil, fmt.Errorf("erreur lors de la récupération des permissions: %w", err)
	}

	// Déterminer les permissions basées sur le rôle
	permissions := &models.GuildPermissionResponse{
		CanInvitePlayers:      role == "leader" || role == "officer",
		CanKickMembers:        role == "leader" || role == "officer",
		CanPromoteMembers:     role == "leader",
		CanDemoteMembers:      role == "leader",
		CanManageBank:         role == "leader" || role == "officer",
		CanDeclareWar:         role == "leader",
		CanCreateAlliance:     role == "leader",
		CanManageApplications: role == "leader" || role == "officer",
		CanViewLogs:           role == "leader" || role == "officer",
		CanEditGuildInfo:      role == "leader",
	}

	return permissions, nil
}

// HasPermission vérifie si un joueur a une permission spécifique
func (s *guildPermissionService) HasPermission(ctx context.Context, guildID, playerID uuid.UUID, permission string) (bool, error) {
	permissions, err := s.GetPermissions(ctx, guildID, playerID)
	if err != nil {
		return false, err
	}

	switch permission {
	case "invite_players":
		return permissions.CanInvitePlayers, nil
	case "kick_members":
		return permissions.CanKickMembers, nil
	case "promote_members":
		return permissions.CanPromoteMembers, nil
	case "demote_members":
		return permissions.CanDemoteMembers, nil
	case "manage_bank":
		return permissions.CanManageBank, nil
	case "declare_war":
		return permissions.CanDeclareWar, nil
	case "create_alliance":
		return permissions.CanCreateAlliance, nil
	case "manage_applications":
		return permissions.CanManageApplications, nil
	case "view_logs":
		return permissions.CanViewLogs, nil
	case "edit_guild_info":
		return permissions.CanEditGuildInfo, nil
	default:
		return false, fmt.Errorf("permission inconnue: %s", permission)
	}
}
