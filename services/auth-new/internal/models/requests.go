package models

import (
	"github.com/google/uuid"
	"time"
)

// Requêtes d'authentification

// RegisterRequest représente une demande d'inscription
type RegisterRequest struct {
	Username        string `json:"username" binding:"required,min=3,max=30"`
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=8"`
	PasswordConfirm string `json:"password_confirm" binding:"required"`
	FirstName       string `json:"first_name" binding:"required,min=1,max=50"`
	LastName        string `json:"last_name" binding:"required,min=1,max=50"`
	Avatar          string `json:"avatar"`
	AcceptTerms     bool   `json:"accept_terms" binding:"required"`
}

// LoginRequest représente une demande de connexion
type LoginRequest struct {
	Username      string `json:"username" binding:"required"`
	Password      string `json:"password" binding:"required"`
	RememberMe    bool   `json:"remember_me"`
	TwoFactorCode string `json:"two_factor_code"`
}

// RefreshTokenRequest représente une demande de renouvellement de token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ChangePasswordRequest pour le changement de mot de passe
type ChangePasswordRequest struct {
	CurrentPassword    string `json:"current_password" binding:"required"`
	NewPassword        string `json:"new_password" binding:"required"`
	NewPasswordConfirm string `json:"new_password_confirm" binding:"required"`
}

// UpdateProfileRequest pour la mise à jour du profil
type UpdateProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Avatar    string `json:"avatar"`
}

// UpdateUserRequest pour la mise à jour admin d'un utilisateur
type UpdateUserRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	Status    string `json:"status"`
}

// ForgotPasswordRequest représente une demande de mot de passe oublié
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest représente une demande de réinitialisation de mot de passe
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// OAuthLoginRequest représente une demande de connexion OAuth
type OAuthLoginRequest struct {
	Provider string `json:"provider" binding:"required"`
	Code     string `json:"code" binding:"required"`
	State    string `json:"state"`
}

// TwoFactorSetup contient les informations de configuration 2FA
type TwoFactorSetup struct {
	Secret string `json:"secret"`
	QRCode string `json:"qr_code"`
}

// LoginAttempt représente une tentative de connexion
type LoginAttempt struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	UserID    *uuid.UUID `json:"user_id" db:"user_id"`
	Username  string     `json:"username" db:"username"`
	IPAddress string     `json:"ip_address" db:"ip_address"`
	UserAgent string     `json:"user_agent" db:"user_agent"`
	Success   bool       `json:"success" db:"success"`
	Reason    string     `json:"reason" db:"reason"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

// AuditLog représente une entrée du journal d'audit
type AuditLog struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	UserID    *uuid.UUID `json:"user_id" db:"user_id"`
	Action    string     `json:"action" db:"action"`
	Resource  string     `json:"resource" db:"resource"`
	Details   string     `json:"details" db:"details"`
	IPAddress string     `json:"ip_address" db:"ip_address"`
	UserAgent string     `json:"user_agent" db:"user_agent"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

// Réponses

// LoginResponse représente la réponse d'une connexion réussie
type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         *User     `json:"user"`
	Permissions  []string  `json:"permissions"`
}

// TokenValidationResponse représente la réponse de validation de token
type TokenValidationResponse struct {
	Valid       bool      `json:"valid"`
	UserID      uuid.UUID `json:"user_id,omitempty"`
	Username    string    `json:"username,omitempty"`
	Email       string    `json:"email,omitempty"`
	Role        string    `json:"role,omitempty"`
	Permissions []string  `json:"permissions,omitempty"`
	ExpiresAt   time.Time `json:"expires_at,omitempty"`
}

// UserInfoResponse représente les informations d'un utilisateur
type UserInfoResponse struct {
	ID               uuid.UUID  `json:"id"`
	Username         string     `json:"username"`
	Email            string     `json:"email"`
	FirstName        string     `json:"first_name"`
	LastName         string     `json:"last_name"`
	Avatar           string     `json:"avatar"`
	Role             string     `json:"role"`
	Status           string     `json:"status"`
	EmailVerified    bool       `json:"email_verified"`
	EmailVerifiedAt  *time.Time `json:"email_verified_at"`
	CreatedAt        time.Time  `json:"created_at"`
	LastLoginAt      *time.Time `json:"last_login_at"`
	TwoFactorEnabled bool       `json:"two_factor_enabled"`
}

// SessionResponse représente les informations d'une session
type SessionResponse struct {
	ID           uuid.UUID `json:"id"`
	DeviceInfo   string    `json:"device_info"`
	IPAddress    string    `json:"ip_address"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	LastActivity time.Time `json:"last_activity"`
	IsActive     bool      `json:"is_active"`
}

// SuccessResponse représente une réponse de succès générique
type SuccessResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

// ErrorResponse représente une réponse d'erreur
type ErrorResponse struct {
	Success   bool   `json:"success"`
	Error     string `json:"error"`
	Message   string `json:"message,omitempty"`
	Details   string `json:"details,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}
