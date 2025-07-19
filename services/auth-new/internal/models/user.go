package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// User représente un utilisateur du système d'authentification
type User struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	Username         string     `json:"username" db:"username"`
	Email            string     `json:"email" db:"email"`
	PasswordHash     string     `json:"-" db:"password_hash"` // Jamais exposé en JSON
	FirstName        string     `json:"first_name" db:"first_name"`
	LastName         string     `json:"last_name" db:"last_name"`
	Avatar           string     `json:"avatar" db:"avatar"`
	Role             string     `json:"role" db:"role"`
	Status           string     `json:"status" db:"status"` // active, suspended, banned
	EmailVerified    bool       `json:"email_verified" db:"email_verified"`
	EmailVerifiedAt  *time.Time `json:"email_verified_at" db:"email_verified_at"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
	LastLoginAt      *time.Time `json:"last_login_at" db:"last_login_at"`
	LastLoginIP      string     `json:"last_login_ip" db:"last_login_ip"`
	LoginAttempts    int        `json:"-" db:"login_attempts"`
	LockedUntil      *time.Time `json:"-" db:"locked_until"`
	TwoFactorSecret  string     `json:"-" db:"two_factor_secret"`
	TwoFactorEnabled bool       `json:"two_factor_enabled" db:"two_factor_enabled"`
}

// UserSession représente une session utilisateur active
type UserSession struct {
	ID           uuid.UUID `json:"id" db:"id"`
	UserID       uuid.UUID `json:"user_id" db:"user_id"`
	AccessToken  string    `json:"-" db:"access_token_hash"`
	RefreshToken string    `json:"-" db:"refresh_token_hash"`
	DeviceInfo   string    `json:"device_info" db:"device_info"`
	IPAddress    string    `json:"ip_address" db:"ip_address"`
	UserAgent    string    `json:"user_agent" db:"user_agent"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	LastActivity time.Time `json:"last_activity" db:"last_activity"`
	IsActive     bool      `json:"is_active" db:"is_active"`
}

// OAuthAccount représente un compte OAuth lié
type OAuthAccount struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	UserID           uuid.UUID  `json:"user_id" db:"user_id"`
	Provider         string     `json:"provider" db:"provider"`
	ProviderUserID   string     `json:"provider_user_id" db:"provider_user_id"`
	ProviderUsername string     `json:"provider_username" db:"provider_username"`
	ProviderEmail    string     `json:"provider_email" db:"provider_email"`
	ProviderAvatar   string     `json:"provider_avatar" db:"provider_avatar"`
	AccessToken      string     `json:"-" db:"access_token"`
	RefreshToken     string     `json:"-" db:"refresh_token"`
	ExpiresAt        *time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// JWTClaims représente les claims du JWT
type JWTClaims struct {
	UserID      uuid.UUID `json:"user_id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	Permissions []string  `json:"permissions"`
	SessionID   uuid.UUID `json:"session_id"`
	TokenType   string    `json:"token_type"` // access, refresh
	jwt.RegisteredClaims
}

// Implémentation de l'interface jwt.Claims pour jwt/v5
func (c *JWTClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	return c.RegisteredClaims.ExpiresAt, nil
}

func (c *JWTClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	return c.RegisteredClaims.IssuedAt, nil
}

func (c *JWTClaims) GetNotBefore() (*jwt.NumericDate, error) {
	return c.RegisteredClaims.NotBefore, nil
}

func (c *JWTClaims) GetIssuer() (string, error) {
	return c.RegisteredClaims.Issuer, nil
}

func (c *JWTClaims) GetSubject() (string, error) {
	return c.RegisteredClaims.Subject, nil
}

func (c *JWTClaims) GetAudience() (jwt.ClaimStrings, error) {
	return c.RegisteredClaims.Audience, nil
}

// Constants pour les rôles et statutes
const (
	// Rôles utilisateur
	RolePlayer    = "player"
	RoleModerator = "moderator"
	RoleAdmin     = "admin"
	RoleSuperUser = "superuser"

	// Statutes utilisateur
	StatusActive    = "active"
	StatusSuspended = "suspended"
	StatusBanned    = "banned"
	StatusPending   = "pending"

	// Permissions
	PermissionLogin        = "auth.login"
	PermissionManageUsers  = "auth.manage_users"
	PermissionViewAuditLog = "auth.view_audit_log"
	PermissionAdminPanel   = "auth.admin_panel"
)

// Méthodes utilitaires pour User

// CanLogin vérifie si l'utilisateur peut se connecter
func (u *User) CanLogin() bool {
	if u.Status != StatusActive {
		return false
	}

	if u.LockedUntil != nil && time.Now().Before(*u.LockedUntil) {
		return false
	}

	return true
}

// IsAccountLocked vérifie si le compte est verrouillé
func (u *User) IsAccountLocked() bool {
	return u.LockedUntil != nil && time.Now().Before(*u.LockedUntil)
}

// IsActive vérifie si l'utilisateur est actif
func (u *User) IsActive() bool {
	return u.Status == StatusActive
}

// IsSuspended vérifie si l'utilisateur est suspendu
func (u *User) IsSuspended() bool {
	return u.Status == StatusSuspended
}

// IsBanned vérifie si l'utilisateur est banni
func (u *User) IsBanned() bool {
	return u.Status == StatusBanned
}

// HasRole vérifie si l'utilisateur a un rôle spécifique
func (u *User) HasRole(role string) bool {
	return u.Role == role
}

// IsAdmin vérifie si l'utilisateur est admin ou superuser
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin || u.Role == RoleSuperUser
}

// GetDisplayName retourne le nom d'affichage de l'utilisateur
func (u *User) GetDisplayName() string {
	if u.FirstName != "" && u.LastName != "" {
		return u.FirstName + " " + u.LastName
	}
	return u.Username
}

// IsExpired vérifie si la session est expirée
func (s *UserSession) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// GetUserPermissions retourne les permissions selon le rôle
func GetUserPermissions(role string) []string {
	switch role {
	case RolePlayer:
		return []string{PermissionLogin}
	case RoleModerator:
		return []string{PermissionLogin, PermissionViewAuditLog}
	case RoleAdmin:
		return []string{PermissionLogin, PermissionManageUsers, PermissionViewAuditLog}
	case RoleSuperUser:
		return []string{PermissionLogin, PermissionManageUsers, PermissionViewAuditLog, PermissionAdminPanel}
	default:
		return []string{}
	}
}
