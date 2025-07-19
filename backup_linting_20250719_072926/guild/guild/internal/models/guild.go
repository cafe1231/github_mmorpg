package models

import (
	"time"

	"github.com/google/uuid"
)

// Guild représente une guilde
type Guild struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Tag         string    `json:"tag" db:"tag"`
	Level       int       `json:"level" db:"level"`
	Experience  int64     `json:"experience" db:"experience"`
	MaxMembers  int       `json:"max_members" db:"max_members"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// GuildMember représente un membre de guilde
type GuildMember struct {
	ID           uuid.UUID `json:"id" db:"id"`
	GuildID      uuid.UUID `json:"guild_id" db:"guild_id"`
	PlayerID     uuid.UUID `json:"player_id" db:"player_id"`
	Role         string    `json:"role" db:"role"` // "leader", "officer", "member"
	JoinedAt     time.Time `json:"joined_at" db:"joined_at"`
	LastSeen     time.Time `json:"last_seen" db:"last_seen"`
	Contribution int64     `json:"contribution" db:"contribution"`
}

// GuildInvitation représente une invitation à rejoindre une guilde
type GuildInvitation struct {
	ID        uuid.UUID `json:"id" db:"id"`
	GuildID   uuid.UUID `json:"guild_id" db:"guild_id"`
	InviterID uuid.UUID `json:"inviter_id" db:"inviter_id"`
	InviteeID uuid.UUID `json:"invitee_id" db:"invitee_id"`
	Message   string    `json:"message" db:"message"`
	Status    string    `json:"status" db:"status"` // "pending", "accepted", "declined", "expired"
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
}

// GuildApplication représente une candidature pour rejoindre une guilde
type GuildApplication struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	GuildID    uuid.UUID  `json:"guild_id" db:"guild_id"`
	PlayerID   uuid.UUID  `json:"player_id" db:"player_id"`
	Message    string     `json:"message" db:"message"`
	Status     string     `json:"status" db:"status"` // "pending", "accepted", "declined"
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	ReviewedAt *time.Time `json:"reviewed_at" db:"reviewed_at"`
	ReviewedBy *uuid.UUID `json:"reviewed_by" db:"reviewed_by"`
}

// GuildLog représente un log d'activité de guilde
type GuildLog struct {
	ID        uuid.UUID `json:"id" db:"id"`
	GuildID   uuid.UUID `json:"guild_id" db:"guild_id"`
	PlayerID  uuid.UUID `json:"player_id" db:"player_id"`
	Action    string    `json:"action" db:"action"`
	Details   string    `json:"details" db:"details"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// GuildRank représente un rang de guilde
type GuildRank struct {
	ID          uuid.UUID `json:"id" db:"id"`
	GuildID     uuid.UUID `json:"guild_id" db:"guild_id"`
	Name        string    `json:"name" db:"name"`
	Level       int       `json:"level" db:"level"`
	Permissions []string  `json:"permissions" db:"permissions"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// GuildBank représente la banque de guilde
type GuildBank struct {
	ID        uuid.UUID `json:"id" db:"id"`
	GuildID   uuid.UUID `json:"guild_id" db:"guild_id"`
	Gold      int64     `json:"gold" db:"gold"`
	Items     string    `json:"items" db:"items"` // JSON array of items
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// GuildBankTransaction représente une transaction de la banque de guilde
type GuildBankTransaction struct {
	ID          uuid.UUID `json:"id" db:"id"`
	GuildID     uuid.UUID `json:"guild_id" db:"guild_id"`
	PlayerID    uuid.UUID `json:"player_id" db:"player_id"`
	Type        string    `json:"type" db:"type"` // "deposit", "withdraw", "transfer"
	Amount      int64     `json:"amount" db:"amount"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// GuildWar représente une guerre de guilde
type GuildWar struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	Guild1ID  uuid.UUID  `json:"guild1_id" db:"guild1_id"`
	Guild2ID  uuid.UUID  `json:"guild2_id" db:"guild2_id"`
	Status    string     `json:"status" db:"status"` // "pending", "active", "ended"
	StartTime time.Time  `json:"start_time" db:"start_time"`
	EndTime   *time.Time `json:"end_time" db:"end_time"`
	Score1    int        `json:"score1" db:"score1"`
	Score2    int        `json:"score2" db:"score2"`
	WinnerID  *uuid.UUID `json:"winner_id" db:"winner_id"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

// GuildAlliance représente une alliance entre guildes
type GuildAlliance struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	Guild1ID  uuid.UUID  `json:"guild1_id" db:"guild1_id"`
	Guild2ID  uuid.UUID  `json:"guild2_id" db:"guild2_id"`
	Status    string     `json:"status" db:"status"` // "pending", "active", "ended"
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	EndedAt   *time.Time `json:"ended_at" db:"ended_at"`
}
