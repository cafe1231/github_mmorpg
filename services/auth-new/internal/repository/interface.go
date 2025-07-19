package repository

import (
	"auth/internal/models"

	"github.com/google/uuid"
)

// UserRepositoryInterface définit les méthodes pour la gestion des utilisateurs
type UserRepositoryInterface interface {
	// CRUD basique
	Create(user *models.User) error
	GetByID(id uuid.UUID) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	Update(user *models.User) error
	Delete(id uuid.UUID) error

	// Méthodes de recherche
	GetAll(limit, offset int) ([]*models.User, error)
	Search(query string, limit, offset int) ([]*models.User, error)

	// Statistiques
	Count() (int64, error)
	CountByStatus(status string) (int64, error)
}

// SessionRepositoryInterface définit les méthodes pour la gestion des sessions
type SessionRepositoryInterface interface {
	// CRUD basique
	Create(session *models.UserSession) error
	GetByID(id uuid.UUID) (*models.UserSession, error)
	Update(session *models.UserSession) error
	Delete(id uuid.UUID) error

	// Méthodes spécifiques aux sessions
	GetUserSessions(userID uuid.UUID) ([]*models.UserSession, error)
	GetActiveUserSessions(userID uuid.UUID) ([]*models.UserSession, error)
	RevokeSession(sessionID uuid.UUID) error
	RevokeAllUserSessions(userID uuid.UUID) error
	CleanExpiredSessions() error

	// Méthodes de recherche
	FindByAccessTokenHash(tokenHash string) (*models.UserSession, error)
	FindByRefreshTokenHash(tokenHash string) (*models.UserSession, error)

	// Statistiques
	CountActiveSessions() (int64, error)
	CountUserSessions(userID uuid.UUID) (int64, error)
}
