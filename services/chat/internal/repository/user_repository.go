package repository

import (
	"chat/internal/models"
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type userRepository struct {
	db *sql.DB
}

// NewUserRepository crée une nouvelle instance du repository des utilisateurs
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

// GetUserInfo récupère les informations d'un utilisateur
func (r *userRepository) GetUserInfo(ctx context.Context, userID uuid.UUID) (*models.UserInfo, error) {
	// Implementation simplifiée - dans un vrai système, on ferait appel au service Player
	user := &models.UserInfo{
		ID:          userID,
		Username:    "user_" + userID.String()[:8],
		DisplayName: "User",
		Avatar:      "",
		Level:       1,
		Title:       "",
		IsOnline:    true,
	}
	return user, nil
}

// GetUsersInfo récupère les informations de plusieurs utilisateurs
func (r *userRepository) GetUsersInfo(ctx context.Context, userIDs []uuid.UUID) ([]*models.UserInfo, error) {
	var users []*models.UserInfo
	for _, userID := range userIDs {
		user, err := r.GetUserInfo(ctx, userID)
		if err == nil {
			users = append(users, user)
		}
	}
	return users, nil
}

// UpdateUserOnlineStatus met à jour le statut en ligne d'un utilisateur
func (r *userRepository) UpdateUserOnlineStatus(ctx context.Context, userID uuid.UUID, isOnline bool) error {
	// Implementation simplifiée - à compléter plus tard
	return nil
}

// SearchUsers recherche des utilisateurs
func (r *userRepository) SearchUsers(ctx context.Context, query string, limit, offset int) ([]*models.UserInfo, error) {
	// Implementation simplifiée - à compléter plus tard
	return []*models.UserInfo{}, nil
}

// GetOnlineUsers récupère les utilisateurs en ligne d'un channel
func (r *userRepository) GetOnlineUsers(ctx context.Context, channelID uuid.UUID) ([]*models.UserInfo, error) {
	// Implementation simplifiée - à compléter plus tard
	return []*models.UserInfo{}, nil
}

// IsUserBlocked vérifie si un utilisateur est bloqué
func (r *userRepository) IsUserBlocked(ctx context.Context, userID, blockedUserID uuid.UUID) (bool, error) {
	// Implementation simplifiée - à compléter plus tard
	return false, nil
}

// BlockUser bloque un utilisateur
func (r *userRepository) BlockUser(ctx context.Context, userID, blockedUserID uuid.UUID) error {
	// Implementation simplifiée - à compléter plus tard
	return nil
}

// UnblockUser débloque un utilisateur
func (r *userRepository) UnblockUser(ctx context.Context, userID, blockedUserID uuid.UUID) error {
	// Implementation simplifiée - à compléter plus tard
	return nil
}

// GetBlockedUsers récupère la liste des utilisateurs bloqués
func (r *userRepository) GetBlockedUsers(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	// Implementation simplifiée - à compléter plus tard
	return []uuid.UUID{}, nil
}
