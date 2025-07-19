package repository

import (
	"auth/internal/models"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// UserRepository implémente l'interface UserRepositoryInterface
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository crée une nouvelle instance du repository utilisateur
func NewUserRepository(db *sqlx.DB) UserRepositoryInterface {
	return &UserRepository{db: db}
}

// Create crée un nouvel utilisateur
func (r *UserRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (
			id, username, email, password_hash, first_name, last_name, 
			avatar, role, status, email_verified, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)`

	_, err := r.db.Exec(query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.Avatar,
		user.Role,
		user.Status,
		user.EmailVerified,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID récupère un utilisateur par son ID
func (r *UserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, username, email, password_hash, first_name, last_name,
		       avatar, role, status, email_verified, email_verified_at,
		       created_at, updated_at, last_login_at, last_login_ip,
		       login_attempts, locked_until, two_factor_secret, two_factor_enabled
		FROM users 
		WHERE id = $1`

	err := r.db.Get(&user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

// GetByUsername récupère un utilisateur par son nom d'utilisateur
func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, username, email, password_hash, first_name, last_name,
		       avatar, role, status, email_verified, email_verified_at,
		       created_at, updated_at, last_login_at, last_login_ip,
		       login_attempts, locked_until, two_factor_secret, two_factor_enabled
		FROM users 
		WHERE username = $1`

	err := r.db.Get(&user, query, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}

// GetByEmail récupère un utilisateur par son email
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, username, email, password_hash, first_name, last_name,
		       avatar, role, status, email_verified, email_verified_at,
		       created_at, updated_at, last_login_at, last_login_ip,
		       login_attempts, locked_until, two_factor_secret, two_factor_enabled
		FROM users 
		WHERE email = $1`

	err := r.db.Get(&user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// Update met à jour un utilisateur
func (r *UserRepository) Update(user *models.User) error {
	query := `
		UPDATE users SET
			username = $2,
			email = $3,
			password_hash = $4,
			first_name = $5,
			last_name = $6,
			avatar = $7,
			role = $8,
			status = $9,
			email_verified = $10,
			email_verified_at = $11,
			updated_at = $12,
			last_login_at = $13,
			last_login_ip = $14,
			login_attempts = $15,
			locked_until = $16,
			two_factor_secret = $17,
			two_factor_enabled = $18
		WHERE id = $1`

	_, err := r.db.Exec(query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.Avatar,
		user.Role,
		user.Status,
		user.EmailVerified,
		user.EmailVerifiedAt,
		user.UpdatedAt,
		user.LastLoginAt,
		user.LastLoginIP,
		user.LoginAttempts,
		user.LockedUntil,
		user.TwoFactorSecret,
		user.TwoFactorEnabled,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// Delete supprime un utilisateur
func (r *UserRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// GetAll récupère tous les utilisateurs avec pagination
func (r *UserRepository) GetAll(limit, offset int) ([]*models.User, error) {
	var users []*models.User
	query := `
		SELECT id, username, email, password_hash, first_name, last_name, 
		       avatar, role, status, email_verified, email_verified_at,
		       created_at, updated_at, last_login_at, last_login_ip,
		       login_attempts, locked_until, two_factor_secret, two_factor_enabled
		FROM users 
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	err := r.db.Select(&users, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}

	return users, nil
}

// Search recherche des utilisateurs par nom d'utilisateur ou email
func (r *UserRepository) Search(searchQuery string, limit, offset int) ([]*models.User, error) {
	var users []*models.User
	query := `
		SELECT id, username, email, password_hash, first_name, last_name, 
		       avatar, role, status, email_verified, email_verified_at,
		       created_at, updated_at, last_login_at, last_login_ip,
		       login_attempts, locked_until, two_factor_secret, two_factor_enabled
		FROM users 
		WHERE username ILIKE $1 OR email ILIKE $1 OR first_name ILIKE $1 OR last_name ILIKE $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	searchPattern := "%" + searchQuery + "%"
	err := r.db.Select(&users, query, searchPattern, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	return users, nil
}

// Count compte le nombre total d'utilisateurs
func (r *UserRepository) Count() (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM users`

	err := r.db.Get(&count, query)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

// CountByStatus compte les utilisateurs par statut
func (r *UserRepository) CountByStatus(status string) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM users WHERE status = $1`

	err := r.db.Get(&count, query, status)
	if err != nil {
		return 0, fmt.Errorf("failed to count users by status: %w", err)
	}

	return count, nil
}
