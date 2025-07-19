package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/dan-2/github_mmorpg/services/guild/internal/models"
	"github.com/google/uuid"
)

// guildMemberService implémente GuildMemberService
type guildMemberService struct {
	db *sql.DB
}

// NewGuildMemberService crée une nouvelle instance de GuildMemberService
func NewGuildMemberService(db *sql.DB) GuildMemberService {
	return &guildMemberService{db: db}
}

// JoinGuild permet à un joueur de rejoindre une guilde
func (s *guildMemberService) JoinGuild(ctx context.Context, guildID, playerID uuid.UUID) error {
	// Vérifier que la guilde existe
	var guildExists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM guilds WHERE id = $1)", guildID).Scan(&guildExists)
	if err != nil {
		return fmt.Errorf("erreur lors de la vérification de la guilde: %w", err)
	}
	if !guildExists {
		return models.ErrGuildNotFound
	}

	// Vérifier que le joueur n'est pas déjà dans une guilde
	var alreadyInGuild bool
	err = s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM guild_members WHERE player_id = $1)", playerID).Scan(&alreadyInGuild)
	if err != nil {
		return fmt.Errorf("erreur lors de la vérification du joueur: %w", err)
	}
	if alreadyInGuild {
		return models.ErrAlreadyInGuild
	}

	// Vérifier la capacité de la guilde
	var memberCount, maxMembers int
	err = s.db.QueryRowContext(ctx, `
		SELECT 
			(SELECT COUNT(*) FROM guild_members WHERE guild_id = $1),
			(SELECT max_members FROM guilds WHERE id = $1)
	`, guildID).Scan(&memberCount, &maxMembers)
	if err != nil {
		return fmt.Errorf("erreur lors de la vérification de la capacité: %w", err)
	}
	if memberCount >= maxMembers {
		return models.ErrGuildFull
	}

	// Ajouter le membre
	query := `
		INSERT INTO guild_members (id, guild_id, player_id, role, joined_at, last_seen, contribution)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	memberID := uuid.New()
	now := time.Now()

	_, err = s.db.ExecContext(ctx, query, memberID, guildID, playerID, "member", now, now, 0)
	if err != nil {
		return fmt.Errorf("erreur lors de l'ajout du membre: %w", err)
	}

	return nil
}

// LeaveGuild permet à un joueur de quitter une guilde
func (s *guildMemberService) LeaveGuild(ctx context.Context, guildID, playerID uuid.UUID) error {
	// Vérifier que le joueur est dans la guilde
	var isMember bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM guild_members WHERE guild_id = $1 AND player_id = $2)", guildID, playerID).Scan(&isMember)
	if err != nil {
		return fmt.Errorf("erreur lors de la vérification de l'appartenance: %w", err)
	}
	if !isMember {
		return models.ErrNotInGuild
	}

	// Supprimer le membre
	_, err = s.db.ExecContext(ctx, "DELETE FROM guild_members WHERE guild_id = $1 AND player_id = $2", guildID, playerID)
	if err != nil {
		return fmt.Errorf("erreur lors de la suppression du membre: %w", err)
	}

	return nil
}

// KickMember permet d'expulser un membre de la guilde
func (s *guildMemberService) KickMember(ctx context.Context, guildID, targetPlayerID, playerID uuid.UUID) error {
	// Vérifier les permissions du joueur qui expulse
	hasPermission, err := s.hasKickPermission(ctx, guildID, playerID)
	if err != nil {
		return err
	}
	if !hasPermission {
		return models.ErrInsufficientPermissions
	}

	// Vérifier que le joueur cible est dans la guilde
	var isMember bool
	err = s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM guild_members WHERE guild_id = $1 AND player_id = $2)", guildID, targetPlayerID).Scan(&isMember)
	if err != nil {
		return fmt.Errorf("erreur lors de la vérification de l'appartenance: %w", err)
	}
	if !isMember {
		return models.ErrNotInGuild
	}

	// Expulser le membre
	_, err = s.db.ExecContext(ctx, "DELETE FROM guild_members WHERE guild_id = $1 AND player_id = $2", guildID, targetPlayerID)
	if err != nil {
		return fmt.Errorf("erreur lors de l'expulsion du membre: %w", err)
	}

	return nil
}

// UpdateMemberRole met à jour le rôle d'un membre
func (s *guildMemberService) UpdateMemberRole(ctx context.Context, guildID uuid.UUID, req *models.UpdateMemberRoleRequest, playerID uuid.UUID) error {
	// Vérifier les permissions du joueur qui modifie
	hasPermission, err := s.hasPromotePermission(ctx, guildID, playerID)
	if err != nil {
		return err
	}
	if !hasPermission {
		return models.ErrInsufficientPermissions
	}

	// Vérifier que le joueur cible est dans la guilde
	var isMember bool
	err = s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM guild_members WHERE guild_id = $1 AND player_id = $2)", guildID, req.PlayerID).Scan(&isMember)
	if err != nil {
		return fmt.Errorf("erreur lors de la vérification de l'appartenance: %w", err)
	}
	if !isMember {
		return models.ErrNotInGuild
	}

	// Mettre à jour le rôle
	_, err = s.db.ExecContext(ctx, "UPDATE guild_members SET role = $1 WHERE guild_id = $2 AND player_id = $3", req.Role, guildID, req.PlayerID)
	if err != nil {
		return fmt.Errorf("erreur lors de la mise à jour du rôle: %w", err)
	}

	return nil
}

// GetMembers récupère la liste des membres d'une guilde
func (s *guildMemberService) GetMembers(ctx context.Context, guildID uuid.UUID, page, limit int) ([]*models.GuildMemberResponse, int, error) {
	// Compter le total
	var total int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM guild_members WHERE guild_id = $1", guildID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("erreur lors du comptage des membres: %w", err)
	}

	// Récupérer les membres avec pagination
	query := `
		SELECT id, player_id, role, joined_at, last_seen, contribution
		FROM guild_members 
		WHERE guild_id = $1 
		ORDER BY role DESC, joined_at ASC 
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.QueryContext(ctx, query, guildID, limit, (page-1)*limit)
	if err != nil {
		return nil, 0, fmt.Errorf("erreur lors de la récupération des membres: %w", err)
	}
	defer rows.Close()

	var members []*models.GuildMemberResponse
	for rows.Next() {
		member := &models.GuildMemberResponse{}
		var memberID uuid.UUID
		var playerID uuid.UUID
		err := rows.Scan(&memberID, &playerID, &member.Role, &member.JoinedAt, &member.LastSeen, &member.Contribution)
		if err != nil {
			return nil, 0, fmt.Errorf("erreur lors du scan des membres: %w", err)
		}
		member.ID = memberID.String()
		member.PlayerID = playerID.String()
		// TODO: Récupérer le nom du joueur depuis le service Player
		member.PlayerName = "Unknown"
		member.IsOnline = time.Since(member.LastSeen) < 5*time.Minute
		members = append(members, member)
	}

	return members, total, nil
}

// GetMember récupère les informations d'un membre spécifique
func (s *guildMemberService) GetMember(ctx context.Context, guildID, playerID uuid.UUID) (*models.GuildMemberResponse, error) {
	query := `
		SELECT id, player_id, role, joined_at, last_seen, contribution
		FROM guild_members 
		WHERE guild_id = $1 AND player_id = $2
	`

	member := &models.GuildMemberResponse{}
	var memberID uuid.UUID
	var playerUUID uuid.UUID
	err := s.db.QueryRowContext(ctx, query, guildID, playerID).Scan(&memberID, &playerUUID, &member.Role, &member.JoinedAt, &member.LastSeen, &member.Contribution)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrGuildNotFound
		}
		return nil, fmt.Errorf("erreur lors de la récupération du membre: %w", err)
	}

	member.ID = memberID.String()
	member.PlayerID = playerUUID.String()
	// TODO: Récupérer le nom du joueur depuis le service Player
	member.PlayerName = "Unknown"
	member.IsOnline = time.Since(member.LastSeen) < 5*time.Minute

	return member, nil
}

// UpdateLastSeen met à jour la dernière connexion d'un joueur
func (s *guildMemberService) UpdateLastSeen(ctx context.Context, playerID uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, "UPDATE guild_members SET last_seen = NOW() WHERE player_id = $1", playerID)
	if err != nil {
		return fmt.Errorf("erreur lors de la mise à jour de last_seen: %w", err)
	}
	return nil
}

// Méthodes utilitaires pour les permissions
func (s *guildMemberService) hasKickPermission(ctx context.Context, guildID, playerID uuid.UUID) (bool, error) {
	var role string
	err := s.db.QueryRowContext(ctx, "SELECT role FROM guild_members WHERE guild_id = $1 AND player_id = $2", guildID, playerID).Scan(&role)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, models.ErrNotInGuild
		}
		return false, err
	}
	return role == "leader" || role == "officer", nil
}

func (s *guildMemberService) hasPromotePermission(ctx context.Context, guildID, playerID uuid.UUID) (bool, error) {
	var role string
	err := s.db.QueryRowContext(ctx, "SELECT role FROM guild_members WHERE guild_id = $1 AND player_id = $2", guildID, playerID).Scan(&role)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, models.ErrNotInGuild
		}
		return false, err
	}
	return role == "leader", nil
}
