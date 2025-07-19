package models

import (
	"github.com/google/uuid"
)

// CreateGuildRequest représente une requête de création de guilde
type CreateGuildRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=50"`
	Description string `json:"description" binding:"max=500"`
	Tag         string `json:"tag" binding:"required,min=2,max=10"`
	MaxMembers  int    `json:"max_members" binding:"min=5,max=100"`
}

// UpdateGuildRequest représente une requête de mise à jour de guilde
type UpdateGuildRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,min=3,max=50"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=500"`
	Tag         *string `json:"tag,omitempty" binding:"omitempty,min=2,max=10"`
	MaxMembers  *int    `json:"max_members,omitempty" binding:"omitempty,min=5,max=100"`
}

// GuildSearchRequest représente une requête de recherche de guilde
type GuildSearchRequest struct {
	Name     *string `json:"name,omitempty"`
	Tag      *string `json:"tag,omitempty"`
	MinLevel *int    `json:"min_level,omitempty"`
	MaxLevel *int    `json:"max_level,omitempty"`
	Page     int     `json:"page"`
	Limit    int     `json:"limit"`
}

// UpdateMemberRoleRequest représente une requête de mise à jour de rôle
type UpdateMemberRoleRequest struct {
	PlayerID uuid.UUID `json:"player_id" binding:"required"`
	Role     string    `json:"role" binding:"required,oneof=leader officer member"`
}

// InvitePlayerRequest représente une requête d'invitation
type InvitePlayerRequest struct {
	PlayerID uuid.UUID `json:"player_id" binding:"required"`
	Message  string    `json:"message" binding:"max=200"`
}

// ApplyToGuildRequest représente une requête de candidature
type ApplyToGuildRequest struct {
	GuildID uuid.UUID `json:"guild_id" binding:"required"`
	Message string    `json:"message" binding:"max=200"`
}

// CreateGuildWarRequest représente une requête de déclaration de guerre
type CreateGuildWarRequest struct {
	TargetGuildID uuid.UUID `json:"target_guild_id" binding:"required"`
	Reason        string    `json:"reason" binding:"max=200"`
	Duration      int       `json:"duration" binding:"min=1,max=7"` // jours
}

// CreateAllianceRequest représente une requête de création d'alliance
type CreateAllianceRequest struct {
	TargetGuildID uuid.UUID `json:"target_guild_id" binding:"required"`
	Message       string    `json:"message" binding:"max=200"`
}

// GuildBankTransactionRequest représente une requête de transaction bancaire
type GuildBankTransactionRequest struct {
	Amount      int64  `json:"amount" binding:"required,min=1"`
	Description string `json:"description" binding:"max=200"`
}
