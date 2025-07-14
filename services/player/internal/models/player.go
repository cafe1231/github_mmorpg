package models

import (
	"time"

	"github.com/google/uuid"
)

// Player représente un joueur (lié au service Auth)
type Player struct {
	ID           uuid.UUID `json:"id" db:"id"`
	UserID       uuid.UUID `json:"user_id" db:"user_id"` // ID du service Auth
	DisplayName  string    `json:"display_name" db:"display_name"`
	Avatar       string    `json:"avatar" db:"avatar"`
	Title        string    `json:"title" db:"title"`
	GuildID      *uuid.UUID `json:"guild_id" db:"guild_id"`
	
	// Statistiques globales du joueur
	TotalPlayTime    int       `json:"total_play_time" db:"total_play_time"` // en minutes
	LastSeen         time.Time `json:"last_seen" db:"last_seen"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
	
	// Préférences
	Preferences      PlayerPreferences `json:"preferences" db:"preferences"`
	
	// Relations (chargées séparément)
	Characters       []*Character `json:"characters,omitempty" db:"-"`
	CharacterCount   int         `json:"character_count" db:"-"`
}

// PlayerPreferences représente les préférences du joueur
type PlayerPreferences struct {
	Language        string            `json:"language"`
	Theme           string            `json:"theme"`
	SoundEnabled    bool              `json:"sound_enabled"`
	MusicEnabled    bool              `json:"music_enabled"`
	ChatSettings    ChatSettings      `json:"chat_settings"`
	NotificationSettings NotificationSettings `json:"notification_settings"`
}

// ChatSettings paramètres de chat
type ChatSettings struct {
	GlobalChatEnabled bool `json:"global_chat_enabled"`
	GuildChatEnabled  bool `json:"guild_chat_enabled"`
	PrivateMessages   bool `json:"private_messages"`
	TradeMessages     bool `json:"trade_messages"`
}

// NotificationSettings paramètres de notifications
type NotificationSettings struct {
	FriendRequests    bool `json:"friend_requests"`
	GuildInvitations  bool `json:"guild_invitations"`
	TradeRequests     bool `json:"trade_requests"`
	SystemMessages    bool `json:"system_messages"`
}

// CreatePlayerRequest représente une demande de création de profil joueur
type CreatePlayerRequest struct {
	DisplayName string `json:"display_name" binding:"required,min=3,max=20"`
	Avatar      string `json:"avatar"`
}

// UpdatePlayerRequest représente une demande de mise à jour de profil
type UpdatePlayerRequest struct {
	DisplayName string            `json:"display_name" binding:"omitempty,min=3,max=20"`
	Avatar      string            `json:"avatar"`
	Title       string            `json:"title"`
	Preferences PlayerPreferences `json:"preferences"`
}

// PlayerStats représente les statistiques globales d'un joueur
type PlayerStats struct {
	PlayerID         uuid.UUID `json:"player_id" db:"player_id"`
	TotalCharacters  int       `json:"total_characters" db:"total_characters"`
	HighestLevel     int       `json:"highest_level" db:"highest_level"`
	TotalExperience  int64     `json:"total_experience" db:"total_experience"`
	AchievementCount int       `json:"achievement_count" db:"achievement_count"`
	
	// Statistiques de jeu
	TotalKills       int       `json:"total_kills" db:"total_kills"`
	TotalDeaths      int       `json:"total_deaths" db:"total_deaths"`
	TotalQuests      int       `json:"total_quests" db:"total_quests"`
	TotalGold        int64     `json:"total_gold" db:"total_gold"`
	
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// PlayerResponse représente la réponse complète d'un profil joueur
type PlayerResponse struct {
	Player     *Player      `json:"player"`
	Stats      *PlayerStats `json:"stats,omitempty"`
	Characters []*Character `json:"characters,omitempty"`
}

// GetDefaultPreferences retourne les préférences par défaut
func GetDefaultPreferences() PlayerPreferences {
	return PlayerPreferences{
		Language:     "en",
		Theme:        "dark",
		SoundEnabled: true,
		MusicEnabled: true,
		ChatSettings: ChatSettings{
			GlobalChatEnabled: true,
			GuildChatEnabled:  true,
			PrivateMessages:   true,
			TradeMessages:     true,
		},
		NotificationSettings: NotificationSettings{
			FriendRequests:   true,
			GuildInvitations: true,
			TradeRequests:    true,
			SystemMessages:   true,
		},
	}
}

// IsOnline vérifie si le joueur est considéré comme en ligne
func (p *Player) IsOnline() bool {
	// Considéré en ligne si vu dans les 5 dernières minutes
	return time.Since(p.LastSeen) < 5*time.Minute
}

// CanCreateCharacter vérifie si le joueur peut créer un nouveau personnage
func (p *Player) CanCreateCharacter(maxCharacters int) bool {
	return p.CharacterCount < maxCharacters
}

// UpdatePlayTime met à jour le temps de jeu total
func (p *Player) UpdatePlayTime(minutes int) {
	p.TotalPlayTime += minutes
	p.LastSeen = time.Now()
	p.UpdatedAt = time.Now()
}