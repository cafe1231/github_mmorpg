package service

import (
	"auth/internal/models"

	"github.com/google/uuid"
)

// AuthServiceInterface définit toutes les méthodes requises pour le service d'authentification
type AuthServiceInterface interface {
	// Authentification de base
	Register(req *models.RegisterRequest) (*models.User, error)
	Login(req models.LoginRequest, ipAddress, userAgent string) (*models.LoginResponse, error)
	RefreshToken(req models.RefreshTokenRequest) (*models.LoginResponse, error)
	Logout(userID uuid.UUID, sessionID uuid.UUID) error
	LogoutAllDevices(userID uuid.UUID) error

	// Gestion du profil utilisateur
	GetProfile(userID uuid.UUID) (*models.User, error)
	UpdateProfile(userID uuid.UUID, req models.UpdateProfileRequest) (*models.User, error)
	ChangePassword(userID uuid.UUID, currentPassword, newPassword string) error

	// Gestion des utilisateurs (Admin)
	ListUsers(limit, offset int, filters map[string]interface{}) ([]*models.User, int64, error)
	GetUser(userID uuid.UUID) (*models.User, error)
	UpdateUser(userID uuid.UUID, req *models.UpdateUserRequest) (*models.User, error)
	SuspendUser(userID uuid.UUID, reason string) error
	ActivateUser(userID uuid.UUID) error

	// Two-Factor Authentication
	EnableTwoFactor(userID uuid.UUID) (*models.TwoFactorSetup, error)
	DisableTwoFactor(userID uuid.UUID, code string) error
	GetTwoFactorQR(userID uuid.UUID) (string, error)

	// Gestion des sessions
	GetSessions(userID uuid.UUID) ([]*models.UserSession, error)
	RevokeSession(userID, sessionID uuid.UUID) error

	// Audit et logs
	GetLoginAttempts(limit, offset int, filters map[string]interface{}) ([]*models.LoginAttempt, int64, error)
	GetAuditLog(limit, offset int, filters map[string]interface{}) ([]*models.AuditLog, int64, error)

	// OAuth
	OAuthRedirect(provider string) (string, error)
	OAuthCallback(provider, code, state string) (*models.LoginResponse, error)

	// Récupération de mot de passe
	ForgotPassword(email string) error
	ResetPassword(token, newPassword string) error

	// Vérification email
	VerifyEmail(token string) error

	// Utilitaires
	GetUserInfo(userID uuid.UUID) (*models.User, error)
	ValidateToken(token string) (*models.JWTClaims, error)
	Close() error
}
