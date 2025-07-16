package models

import (
	"time"
)

// GuildResponse représente une réponse de guilde
type GuildResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Tag         string    `json:"tag"`
	Level       int       `json:"level"`
	Experience  int64     `json:"experience"`
	MaxMembers  int       `json:"max_members"`
	MemberCount int       `json:"member_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GuildMemberResponse représente une réponse de membre de guilde
type GuildMemberResponse struct {
	ID           string    `json:"id"`
	PlayerID     string    `json:"player_id"`
	PlayerName   string    `json:"player_name"`
	Role         string    `json:"role"`
	JoinedAt     time.Time `json:"joined_at"`
	LastSeen     time.Time `json:"last_seen"`
	Contribution int64     `json:"contribution"`
	IsOnline     bool      `json:"is_online"`
}

// GuildInvitationResponse représente une réponse d'invitation
type GuildInvitationResponse struct {
	ID          string    `json:"id"`
	GuildID     string    `json:"guild_id"`
	GuildName   string    `json:"guild_name"`
	InviterID   string    `json:"inviter_id"`
	InviterName string    `json:"inviter_name"`
	Message     string    `json:"message"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// GuildApplicationResponse représente une réponse de candidature
type GuildApplicationResponse struct {
	ID         string     `json:"id"`
	GuildID    string     `json:"guild_id"`
	GuildName  string     `json:"guild_name"`
	PlayerID   string     `json:"player_id"`
	PlayerName string     `json:"player_name"`
	Message    string     `json:"message"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	ReviewedAt *time.Time `json:"reviewed_at,omitempty"`
	ReviewedBy *string    `json:"reviewed_by,omitempty"`
}

// GuildLogResponse représente une réponse de log de guilde
type GuildLogResponse struct {
	ID         string    `json:"id"`
	PlayerID   string    `json:"player_id"`
	PlayerName string    `json:"player_name"`
	Action     string    `json:"action"`
	Details    string    `json:"details"`
	CreatedAt  time.Time `json:"created_at"`
}

// GuildBankResponse représente une réponse de banque de guilde
type GuildBankResponse struct {
	ID        string    `json:"id"`
	Gold      int64     `json:"gold"`
	Items     string    `json:"items"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GuildBankTransactionResponse représente une réponse de transaction bancaire
type GuildBankTransactionResponse struct {
	ID          string    `json:"id"`
	PlayerID    string    `json:"player_id"`
	PlayerName  string    `json:"player_name"`
	Type        string    `json:"type"`
	Amount      int64     `json:"amount"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// GuildWarResponse représente une réponse de guerre de guilde
type GuildWarResponse struct {
	ID         string     `json:"id"`
	Guild1ID   string     `json:"guild1_id"`
	Guild1Name string     `json:"guild1_name"`
	Guild2ID   string     `json:"guild2_id"`
	Guild2Name string     `json:"guild2_name"`
	Status     string     `json:"status"`
	StartTime  time.Time  `json:"start_time"`
	EndTime    *time.Time `json:"end_time,omitempty"`
	Score1     int        `json:"score1"`
	Score2     int        `json:"score2"`
	WinnerID   *string    `json:"winner_id,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// GuildAllianceResponse représente une réponse d'alliance
type GuildAllianceResponse struct {
	ID         string     `json:"id"`
	Guild1ID   string     `json:"guild1_id"`
	Guild1Name string     `json:"guild1_name"`
	Guild2ID   string     `json:"guild2_id"`
	Guild2Name string     `json:"guild2_name"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	EndedAt    *time.Time `json:"ended_at,omitempty"`
}

// GuildSearchResponse représente une réponse de recherche de guilde
type GuildSearchResponse struct {
	Guilds []GuildResponse `json:"guilds"`
	Total  int             `json:"total"`
	Page   int             `json:"page"`
	Limit  int             `json:"limit"`
}

// GuildStatsResponse représente une réponse de statistiques de guilde
type GuildStatsResponse struct {
	GuildID             string  `json:"guild_id"`
	TotalMembers        int     `json:"total_members"`
	OnlineMembers       int     `json:"online_members"`
	AverageLevel        float64 `json:"average_level"`
	TotalExperience     int64   `json:"total_experience"`
	BankGold            int64   `json:"bank_gold"`
	ActiveWars          int     `json:"active_wars"`
	ActiveAlliances     int     `json:"active_alliances"`
	PendingInvitations  int     `json:"pending_invitations"`
	PendingApplications int     `json:"pending_applications"`
}

// GuildPermissionResponse représente une réponse de permissions
type GuildPermissionResponse struct {
	CanInvitePlayers      bool `json:"can_invite_players"`
	CanKickMembers        bool `json:"can_kick_members"`
	CanPromoteMembers     bool `json:"can_promote_members"`
	CanDemoteMembers      bool `json:"can_demote_members"`
	CanManageBank         bool `json:"can_manage_bank"`
	CanDeclareWar         bool `json:"can_declare_war"`
	CanCreateAlliance     bool `json:"can_create_alliance"`
	CanManageApplications bool `json:"can_manage_applications"`
	CanViewLogs           bool `json:"can_view_logs"`
	CanEditGuildInfo      bool `json:"can_edit_guild_info"`
}
