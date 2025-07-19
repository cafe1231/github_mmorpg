package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"

	"auth/internal/models"
)

// SessionRepository implémente SessionRepositoryInterface
type SessionRepository struct {
	db *sqlx.DB
}

// NewSessionRepository crée un nouveau repository de sessions
func NewSessionRepository(db *sqlx.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create crée une nouvelle session
func (r *SessionRepository) Create(session *models.UserSession) error {
	query := `
		INSERT INTO user_sessions (
			id, user_id, access_token_hash, refresh_token_hash,
			device_info, ip_address, user_agent, created_at,
			expires_at, last_activity, is_active
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.Exec(query,
		session.ID,
		session.UserID,
		session.AccessToken,
		session.RefreshToken,
		session.DeviceInfo,
		session.IPAddress,
		session.UserAgent,
		session.CreatedAt,
		session.ExpiresAt,
		session.LastActivity,
		session.IsActive,
	)

	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// GetByID récupère une session par son ID
func (r *SessionRepository) GetByID(id uuid.UUID) (*models.UserSession, error) {
	query := `
		SELECT id, user_id, access_token_hash, refresh_token_hash,
		       device_info, ip_address, user_agent, created_at,
		       expires_at, last_activity, is_active
		FROM user_sessions 
		WHERE id = $1
	`

	var session models.UserSession
	err := r.db.Get(&session, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

// Update met à jour une session
func (r *SessionRepository) Update(session *models.UserSession) error {
	query := `
		UPDATE user_sessions SET 
			access_token_hash = $2,
			refresh_token_hash = $3,
			device_info = $4,
			ip_address = $5,
			user_agent = $6,
			expires_at = $7,
			last_activity = $8,
			is_active = $9
		WHERE id = $1
	`

	_, err := r.db.Exec(query,
		session.ID,
		session.AccessToken,
		session.RefreshToken,
		session.DeviceInfo,
		session.IPAddress,
		session.UserAgent,
		session.ExpiresAt,
		session.LastActivity,
		session.IsActive,
	)

	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

// Delete supprime une session
func (r *SessionRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM user_sessions WHERE id = $1`

	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// GetUserSessions récupère toutes les sessions d'un utilisateur
func (r *SessionRepository) GetUserSessions(userID uuid.UUID) ([]*models.UserSession, error) {
	query := `
		SELECT id, user_id, access_token_hash, refresh_token_hash,
		       device_info, ip_address, user_agent, created_at,
		       expires_at, last_activity, is_active
		FROM user_sessions 
		WHERE user_id = $1
		ORDER BY last_activity DESC
	`

	var sessions []*models.UserSession
	err := r.db.Select(&sessions, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}

	return sessions, nil
}

// GetActiveUserSessions récupère les sessions actives d'un utilisateur
func (r *SessionRepository) GetActiveUserSessions(userID uuid.UUID) ([]*models.UserSession, error) {
	query := `
		SELECT id, user_id, access_token_hash, refresh_token_hash,
		       device_info, ip_address, user_agent, created_at,
		       expires_at, last_activity, is_active
		FROM user_sessions 
		WHERE user_id = $1 AND is_active = true AND expires_at > NOW()
		ORDER BY last_activity DESC
	`

	var sessions []*models.UserSession
	err := r.db.Select(&sessions, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active user sessions: %w", err)
	}

	return sessions, nil
}

// RevokeSession révoque une session spécifique
func (r *SessionRepository) RevokeSession(sessionID uuid.UUID) error {
	query := `UPDATE user_sessions SET is_active = false WHERE id = $1`

	_, err := r.db.Exec(query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	return nil
}

// RevokeAllUserSessions révoque toutes les sessions d'un utilisateur
func (r *SessionRepository) RevokeAllUserSessions(userID uuid.UUID) error {
	query := `UPDATE user_sessions SET is_active = false WHERE user_id = $1`

	_, err := r.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke all user sessions: %w", err)
	}

	return nil
}

// CleanExpiredSessions supprime les sessions expirées
func (r *SessionRepository) CleanExpiredSessions() error {
	query := `DELETE FROM user_sessions WHERE expires_at < NOW()`

	result, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to clean expired sessions: %w", err)
	}

	affected, _ := result.RowsAffected()
	logrus.WithField("cleaned_sessions", affected).Info("Cleaned expired sessions")

	return nil
}

// FindByAccessTokenHash trouve une session par hash du token d'accès
func (r *SessionRepository) FindByAccessTokenHash(tokenHash string) (*models.UserSession, error) {
	query := `
		SELECT id, user_id, access_token_hash, refresh_token_hash,
		       device_info, ip_address, user_agent, created_at,
		       expires_at, last_activity, is_active
		FROM user_sessions 
		WHERE access_token_hash = $1 AND is_active = true AND expires_at > NOW()
	`

	var session models.UserSession
	err := r.db.Get(&session, query, tokenHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to find session by access token: %w", err)
	}

	return &session, nil
}

// FindByRefreshTokenHash trouve une session par hash du refresh token
func (r *SessionRepository) FindByRefreshTokenHash(tokenHash string) (*models.UserSession, error) {
	query := `
		SELECT id, user_id, access_token_hash, refresh_token_hash,
		       device_info, ip_address, user_agent, created_at,
		       expires_at, last_activity, is_active
		FROM user_sessions 
		WHERE refresh_token_hash = $1 AND is_active = true
	`

	var session models.UserSession
	err := r.db.Get(&session, query, tokenHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to find session by refresh token: %w", err)
	}

	return &session, nil
}

// CountActiveSessions compte le nombre de sessions actives
func (r *SessionRepository) CountActiveSessions() (int64, error) {
	query := `SELECT COUNT(*) FROM user_sessions WHERE is_active = true AND expires_at > NOW()`

	var count int64
	err := r.db.Get(&count, query)
	if err != nil {
		return 0, fmt.Errorf("failed to count active sessions: %w", err)
	}

	return count, nil
}

// CountUserSessions compte le nombre de sessions d'un utilisateur
func (r *SessionRepository) CountUserSessions(userID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM user_sessions WHERE user_id = $1 AND is_active = true AND expires_at > NOW()`

	var count int64
	err := r.db.Get(&count, query, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to count user sessions: %w", err)
	}

	return count, nil
}

// ScheduleCleanup démarre le nettoyage périodique des sessions
func (r *SessionRepository) ScheduleCleanup(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := r.CleanExpiredSessions(); err != nil {
					logrus.WithError(err).Error("Failed to clean expired sessions")
				}
			}
		}
	}()

	logrus.WithField("interval", interval).Info("Session cleanup scheduler started")
}
