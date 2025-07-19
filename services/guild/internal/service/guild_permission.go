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
		CanInvitePlayers:      role == RoleLeader || role == RoleOfficer,
		CanKickMembers:        role == RoleLeader || role == RoleOfficer,
		CanPromoteMembers:     role == RoleLeader,
		CanDemoteMembers:      role == RoleLeader,
		CanManageBank:         role == RoleLeader || role == RoleOfficer,
		CanDeclareWar:         role == RoleLeader,
		CanCreateAlliance:     role == RoleLeader,
		CanManageApplications: role == RoleLeader || role == RoleOfficer,
		CanViewLogs:           role == RoleLeader || role == RoleOfficer,
		CanEditGuildInfo:      role == RoleLeader,
	}

	return permissions, nil
}

// HasPermission vérifie si un joueur a une permission spécifique dans une guilde
func (s *guildPermissionService) HasPermission(ctx context.Context, guildID, playerID uuid.UUID, permission string) (bool, error) {
	// Récupérer le rôle du joueur dans la guilde
	var role string
	query := `SELECT role FROM guild_members WHERE guild_id = $1 AND player_id = $2`
	err := s.db.QueryRowContext(ctx, query, guildID, playerID).Scan(&role)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // Joueur pas dans la guilde
		}
		return false, err
	}

	// Vérifier les permissions selon le rôle
	switch permission {
	case "edit_guild_info":
		return role == RoleLeader || role == RoleOfficer, nil
	case "delete_guild":
		return role == RoleLeader, nil
	case "kick_member":
		return role == RoleLeader || role == RoleOfficer, nil
	case "promote_member":
		return role == RoleLeader, nil
	case "invite_member":
		return role == RoleLeader || role == RoleOfficer, nil
	case "manage_roles":
		return role == RoleLeader, nil
	default:
		return false, nil
	}
}

// GetRole récupère le rôle d'un joueur dans une guilde
func (s *guildPermissionService) GetRole(ctx context.Context, guildID, playerID uuid.UUID) (string, error) {
	var role string
	query := `SELECT role FROM guild_members WHERE guild_id = $1 AND player_id = $2`
	err := s.db.QueryRowContext(ctx, query, guildID, playerID).Scan(&role)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", models.ErrMemberNotFound
		}
		return "", err
	}
	return role, nil
}

// IsLeader vérifie si un joueur est le leader d'une guilde
func (s *guildPermissionService) IsLeader(ctx context.Context, guildID, playerID uuid.UUID) (bool, error) {
	role, err := s.GetRole(ctx, guildID, playerID)
	if err != nil {
		return false, err
	}
	return role == RoleLeader, nil
}

// IsOfficer vérifie si un joueur est officier d'une guilde
func (s *guildPermissionService) IsOfficer(ctx context.Context, guildID, playerID uuid.UUID) (bool, error) {
	role, err := s.GetRole(ctx, guildID, playerID)
	if err != nil {
		return false, err
	}
	return role == RoleLeader || role == RoleOfficer, nil
}
