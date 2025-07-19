package service

import (
	"auth/internal/config"
	"auth/internal/models"
	"auth/internal/repository"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// AuthService gère l'authentification des utilisateurs
type AuthService struct {
	userRepo    repository.UserRepositoryInterface
	sessionRepo repository.SessionRepositoryInterface
	config      *config.Config
}

// NewAuthService crée un nouveau service d'authentification
func NewAuthService(
	userRepo repository.UserRepositoryInterface,
	sessionRepo repository.SessionRepositoryInterface,
	config *config.Config,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		config:      config,
	}
}

// Register inscrit un nouvel utilisateur
func (s *AuthService) Register(req *models.RegisterRequest) (*models.User, error) {
	// Validation des données
	if err := s.validateRegistration(req); err != nil {
		return nil, err
	}

	// Vérifier si l'utilisateur existe déjà
	existingUser, _ := s.userRepo.GetByUsername(req.Username)
	if existingUser != nil {
		return nil, fmt.Errorf("username already exists")
	}

	existingUser, _ = s.userRepo.GetByEmail(req.Email)
	if existingUser != nil {
		return nil, fmt.Errorf("email already exists")
	}

	// Hacher le mot de passe
	hashedPassword, err := s.hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Créer l'utilisateur
	user := &models.User{
		ID:           uuid.New(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Role:         models.RolePlayer,
		Status:       models.StatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Sauvegarder en base
	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Envoyer email de vérification (en arrière-plan)
	go s.sendEmailVerification(user)

	// Log de l'inscription
	logrus.WithFields(logrus.Fields{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
	}).Info("User registered successfully")

	// Ne pas retourner le hash du mot de passe
	user.PasswordHash = ""
	return user, nil
}

// Login connecte un utilisateur
func (s *AuthService) Login(req models.LoginRequest, ipAddress, userAgent string) (*models.LoginResponse, error) {
	// Récupérer l'utilisateur
	user, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		// Log de la tentative échouée
		s.logLoginAttempt(nil, req.Username, ipAddress, userAgent, false, "user_not_found")
		return nil, fmt.Errorf("invalid credentials")
	}

	// Vérifier si le compte est verrouillé
	if user.IsAccountLocked() {
		s.logLoginAttempt(&user.ID, req.Username, ipAddress, userAgent, false, "account_locked")
		return nil, fmt.Errorf("account is temporarily locked")
	}

	// Vérifier si le compte peut se connecter
	if !user.CanLogin() {
		s.logLoginAttempt(&user.ID, req.Username, ipAddress, userAgent, false, "account_disabled")
		return nil, fmt.Errorf("account is disabled")
	}

	// Vérifier le mot de passe
	if err := s.verifyPassword(req.Password, user.PasswordHash); err != nil {
		// Incrémenter les tentatives échouées
		s.incrementFailedAttempts(user)
		s.logLoginAttempt(&user.ID, req.Username, ipAddress, userAgent, false, "wrong_password")
		return nil, fmt.Errorf("invalid credentials")
	}

	// Vérifier 2FA si activé
	if user.TwoFactorEnabled && req.TwoFactorCode == "" {
		return nil, fmt.Errorf("two_factor_required")
	}

	if user.TwoFactorEnabled && req.TwoFactorCode != "" {
		if !s.verifyTwoFactorCode(user.TwoFactorSecret, req.TwoFactorCode) {
			s.logLoginAttempt(&user.ID, req.Username, ipAddress, userAgent, false, "invalid_2fa")
			return nil, fmt.Errorf("invalid two-factor code")
		}
	}

	// Générer les tokens
	accessToken, refreshToken, expiresAt, err := s.generateTokens(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Créer la session
	session := &models.UserSession{
		ID:           uuid.New(),
		UserID:       user.ID,
		AccessToken:  s.hashToken(accessToken),
		RefreshToken: s.hashToken(refreshToken),
		DeviceInfo:   s.extractDeviceInfo(userAgent),
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		CreatedAt:    time.Now(),
		ExpiresAt:    expiresAt,
		LastActivity: time.Now(),
		IsActive:     true,
	}

	if err := s.sessionRepo.Create(session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Réinitialiser les tentatives échouées
	s.resetFailedAttempts(user)

	// Mettre à jour la dernière connection
	user.LastLoginAt = &session.CreatedAt
	user.LastLoginIP = ipAddress
	if err := s.userRepo.Update(user); err != nil {
		logrus.WithError(err).Error("Failed to update user last login info")
	}

	// Log de la connection réussie
	s.logLoginAttempt(&user.ID, req.Username, ipAddress, userAgent, true, "success")

	logrus.WithFields(logrus.Fields{
		"user_id":    user.ID,
		"username":   user.Username,
		"ip_address": ipAddress,
		"session_id": session.ID,
	}).Info("User logged in successfully")

	// Préparer la réponse
	user.PasswordHash = ""
	user.TwoFactorSecret = ""

	return &models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.config.JWT.AccessTokenExpiration.Seconds()),
		ExpiresAt:    expiresAt,
		User:         user,
		Permissions:  models.GetUserPermissions(user.Role),
	}, nil
}

// RefreshToken renouvelle un token d'accès
func (s *AuthService) RefreshToken(req models.RefreshTokenRequest) (*models.LoginResponse, error) {
	// Valider le refresh token
	claims, err := s.validateToken(req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if claims.TokenType != "refresh" {
		return nil, fmt.Errorf("token is not a refresh token")
	}

	// Récupérer l'utilisateur
	user, err := s.userRepo.GetByID(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	if !user.CanLogin() {
		return nil, fmt.Errorf("user cannot login")
	}

	// Générer de nouveaux tokens
	newAccessToken, newRefreshToken, expiresAt, err := s.generateTokens(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new tokens: %w", err)
	}

	// Préparer la réponse
	user.PasswordHash = ""
	user.TwoFactorSecret = ""

	return &models.LoginResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.config.JWT.AccessTokenExpiration.Seconds()),
		ExpiresAt:    expiresAt,
		User:         user,
		Permissions:  models.GetUserPermissions(user.Role),
	}, nil
}

// ValidateToken valide un token JWT
func (s *AuthService) ValidateToken(tokenString string) (*models.JWTClaims, error) {
	return s.validateToken(tokenString)
}

// Logout déconnecte un utilisateur
func (s *AuthService) Logout(userID uuid.UUID, sessionID uuid.UUID) error {
	return s.sessionRepo.RevokeSession(sessionID)
}

// LogoutAllDevices déconnecte un utilisateur de tous ses appareils
func (s *AuthService) LogoutAllDevices(userID uuid.UUID) error {
	return s.sessionRepo.RevokeAllUserSessions(userID)
}

// ChangePassword change le mot de passe d'un utilisateur
func (s *AuthService) ChangePassword(userID uuid.UUID, currentPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Vérifier le mot de passe actuel
	if err := s.verifyPassword(currentPassword, user.PasswordHash); err != nil {
		return fmt.Errorf("current password is incorrect")
	}

	// Valider le nouveau mot de passe
	if err := s.validatePassword(newPassword); err != nil {
		return err
	}

	// Hacher le nouveau mot de passe
	hashedPassword, err := s.hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Mettre à jour le mot de passe
	user.PasswordHash = hashedPassword
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"user_id":  user.ID,
		"username": user.Username,
	}).Info("Password changed successfully")

	return nil
}

// GetProfile récupère le profil d'un utilisateur
func (s *AuthService) GetProfile(userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Ne pas exposer le hash du mot de passe
	user.PasswordHash = ""
	return user, nil
}

// UpdateProfile met à jour le profil d'un utilisateur
func (s *AuthService) UpdateProfile(userID uuid.UUID, req models.UpdateProfileRequest) (*models.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Mettre à jour les champs modifiables
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}

	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	// Ne pas exposer le hash du mot de passe
	user.PasswordHash = ""
	return user, nil
}

// ListUsers liste les utilisateurs (admin uniquement)
func (s *AuthService) ListUsers(limit, offset int, filters map[string]interface{}) ([]*models.User, int64, error) {
	users, err := s.userRepo.GetAll(limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	// Ne pas exposer les mots de passe
	for _, user := range users {
		user.PasswordHash = ""
	}

	// Compter le total
	total, err := s.userRepo.Count()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	return users, total, nil
}

// GetUser récupère un utilisateur par son ID (admin uniquement)
func (s *AuthService) GetUser(userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Ne pas exposer le hash du mot de passe
	user.PasswordHash = ""
	return user, nil
}

// UpdateUser met à jour un utilisateur (admin uniquement)
func (s *AuthService) UpdateUser(userID uuid.UUID, req *models.UpdateUserRequest) (*models.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Mettre à jour les champs autorisés
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.Email != "" {
		// Vérifier que l'email n'est pas déjà utilisé
		existingUser, _ := s.userRepo.GetByEmail(req.Email)
		if existingUser != nil && existingUser.ID != userID {
			return nil, fmt.Errorf("email already in use")
		}
		user.Email = req.Email
	}
	if req.Role != "" {
		user.Role = req.Role
	}
	if req.Status != "" {
		user.Status = req.Status
	}

	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Ne pas exposer le hash du mot de passe
	user.PasswordHash = ""
	return user, nil
}

// SuspendUser suspend un utilisateur
func (s *AuthService) SuspendUser(userID uuid.UUID, reason string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	user.Status = models.StatusSuspended
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("failed to suspend user: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"user_id":  userID,
		"username": user.Username,
		"reason":   reason,
	}).Info("User suspended")

	return nil
}

// ActivateUser active un utilisateur
func (s *AuthService) ActivateUser(userID uuid.UUID) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	user.Status = models.StatusActive
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("failed to activate user: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"user_id":  userID,
		"username": user.Username,
	}).Info("User activated")

	return nil
}

// GetSessions récupère les sessions actives d'un utilisateur
func (s *AuthService) GetSessions(userID uuid.UUID) ([]*models.UserSession, error) {
	return s.sessionRepo.GetUserSessions(userID)
}

// RevokeSession révoque une session spécifique
func (s *AuthService) RevokeSession(userID, sessionID uuid.UUID) error {
	// Vérifier que la session appartient à l'utilisateur
	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	if session.UserID != userID {
		return fmt.Errorf("session does not belong to user")
	}

	return s.sessionRepo.RevokeSession(sessionID)
}

// EnableTwoFactor active l'authentification à deux facteurs
func (s *AuthService) EnableTwoFactor(userID uuid.UUID) (*models.TwoFactorSetup, error) {
	return nil, fmt.Errorf("two-factor authentication not implemented yet")
}

// DisableTwoFactor désactive l'authentification à deux facteurs
func (s *AuthService) DisableTwoFactor(userID uuid.UUID, code string) error {
	return fmt.Errorf("two-factor authentication not implemented yet")
}

// GetTwoFactorQR génère un QR code pour l'authentification à deux facteurs
func (s *AuthService) GetTwoFactorQR(userID uuid.UUID) (string, error) {
	return "", fmt.Errorf("two-factor authentication not implemented yet")
}

// GetLoginAttempts récupère les tentatives de connection
func (s *AuthService) GetLoginAttempts(limit, offset int, filters map[string]interface{}) ([]*models.LoginAttempt, int64, error) {
	return []*models.LoginAttempt{}, 0, fmt.Errorf("login attempts audit not implemented yet")
}

// GetAuditLog récupère le journal d'audit
func (s *AuthService) GetAuditLog(limit, offset int, filters map[string]interface{}) ([]*models.AuditLog, int64, error) {
	return []*models.AuditLog{}, 0, fmt.Errorf("audit log not implemented yet")
}

// OAuthRedirect génère une URL de redirection OAuth
func (s *AuthService) OAuthRedirect(provider string) (string, error) {
	return "", fmt.Errorf("OAuth not implemented yet")
}

// OAuthCallback traite le callback OAuth
func (s *AuthService) OAuthCallback(provider, code, state string) (*models.LoginResponse, error) {
	return nil, fmt.Errorf("OAuth not implemented yet")
}

// ForgotPassword initie le processus de récupération de mot de passe
func (s *AuthService) ForgotPassword(email string) error {
	return fmt.Errorf("password recovery not implemented yet")
}

// ResetPassword réinitialize le mot de passe avec un token
func (s *AuthService) ResetPassword(token, newPassword string) error {
	return fmt.Errorf("password reset not implemented yet")
}

// VerifyEmail vérifie l'adresse email avec un token
func (s *AuthService) VerifyEmail(token string) error {
	return fmt.Errorf("email verification not implemented yet")
}

// GetUserInfo récupère les informations utilisateur pour les autres services
func (s *AuthService) GetUserInfo(userID uuid.UUID) (*models.User, error) {
	return s.GetProfile(userID)
}

// Close ferme le service d'authentification
func (s *AuthService) Close() error {
	logrus.Info("Auth service closed")
	return nil
}

// Méthodes privées

// validateRegistration valide les données d'inscription
func (s *AuthService) validateRegistration(req *models.RegisterRequest) error {
	// Validation du nom d'utilisateur
	if len(req.Username) < 3 || len(req.Username) > 30 {
		return fmt.Errorf("username must be between 3 and 30 characters")
	}

	// Caractères autorisés pour le nom d'utilisateur
	usernameRegex := regexp.MustCompile("^[a-zA-Z0-9_-]+$")
	if !usernameRegex.MatchString(req.Username) {
		return fmt.Errorf("username can only contain letters, numbers, underscores and hyphens")
	}

	// Validation de l'email
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(req.Email) {
		return fmt.Errorf("invalid email format")
	}

	// Validation du mot de passe
	if err := s.validatePassword(req.Password); err != nil {
		return err
	}

	// Vérification de la confirmation du mot de passe
	if req.Password != req.PasswordConfirm {
		return fmt.Errorf("passwords do not match")
	}

	// Validation des noms
	if strings.TrimSpace(req.FirstName) == "" {
		return fmt.Errorf("first name is required")
	}
	if strings.TrimSpace(req.LastName) == "" {
		return fmt.Errorf("last name is required")
	}

	// Acceptation des conditions
	if !req.AcceptTerms {
		return fmt.Errorf("you must accept the terms and conditions")
	}

	return nil
}

// validatePassword valide un mot de passe selon les règles de sécurité
func (s *AuthService) validatePassword(password string) error {
	if len(password) < s.config.Security.PasswordMinLength {
		return fmt.Errorf("password must be at least %d characters long", s.config.Security.PasswordMinLength)
	}

	if s.config.Security.PasswordRequireUpper {
		hasUpper := regexp.MustCompile("[A-Z]").MatchString(password)
		if !hasUpper {
			return fmt.Errorf("password must contain at least one uppercase letter")
		}
	}

	if s.config.Security.PasswordRequireLower {
		hasLower := regexp.MustCompile("[a-z]").MatchString(password)
		if !hasLower {
			return fmt.Errorf("password must contain at least one lowercase letter")
		}
	}

	if s.config.Security.PasswordRequireDigit {
		hasDigit := regexp.MustCompile(`\d`).MatchString(password)
		if !hasDigit {
			return fmt.Errorf("password must contain at least one digit")
		}
	}

	if s.config.Security.PasswordRequireSpecial {
		hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>/?]`).MatchString(password)
		if !hasSpecial {
			return fmt.Errorf("password must contain at least one special character")
		}
	}

	return nil
}

// hashPassword hache un mot de passe avec bcrypt
func (s *AuthService) hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.config.Security.BCryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// verifyPassword vérifie un mot de passe contre son hash
func (s *AuthService) verifyPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// generateTokens génère les tokens JWT (access et refresh)
func (s *AuthService) generateTokens(user *models.User) (accessTokenStr, refreshTokenStr string, expiresAt time.Time, err error) {
	now := time.Now()
	accessExpiry := now.Add(s.config.JWT.AccessTokenExpiration)
	refreshExpiry := now.Add(s.config.JWT.RefreshTokenExpiration)

	// Claims pour le token d'accès
	accessClaims := &models.JWTClaims{
		UserID:      user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Role:        user.Role,
		Permissions: models.GetUserPermissions(user.Role),
		SessionID:   uuid.New(),
		TokenType:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.config.JWT.Issuer,
			Subject:   user.ID.String(),
		},
	}

	// Générer le token d'accès
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.config.JWT.Secret))
	if err != nil {
		return "", "", time.Time{}, err
	}

	// Claims pour le refresh token
	refreshClaims := &models.JWTClaims{
		UserID:    user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		SessionID: accessClaims.SessionID,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.config.JWT.Issuer,
			Subject:   user.ID.String(),
		},
	}

	// Générer le refresh token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.config.JWT.Secret))
	if err != nil {
		return "", "", time.Time{}, err
	}

	return accessTokenString, refreshTokenString, accessExpiry, nil
}

// validateToken valide un token JWT
func (s *AuthService) validateToken(tokenString string) (*models.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWT.Secret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*models.JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// hashToken hache un token pour le stockage sécurisé
func (s *AuthService) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// extractDeviceInfo extrait les informations de l'appareil depuis le User-Agent
func (s *AuthService) extractDeviceInfo(userAgent string) string {
	// Simplification pour l'exemple - en production, utiliser une librairie dédiée
	if strings.Contains(userAgent, "Mobile") {
		return "Mobile Device"
	}
	if strings.Contains(userAgent, "Windows") {
		return "Windows PC"
	}
	if strings.Contains(userAgent, "Mac") {
		return "Mac"
	}
	if strings.Contains(userAgent, "Linux") {
		return "Linux"
	}
	return "Unknown Device"
}

// incrementFailedAttempts incrémente le compteur d'échecs de connection
func (s *AuthService) incrementFailedAttempts(user *models.User) {
	user.LoginAttempts++

	// Verrouiller le compte si trop d'échecs
	if user.LoginAttempts >= s.config.Security.MaxLoginAttempts {
		lockUntil := time.Now().Add(s.config.Security.LockoutDuration)
		user.LockedUntil = &lockUntil

		logrus.WithFields(logrus.Fields{
			"user_id":      user.ID,
			"username":     user.Username,
			"attempts":     user.LoginAttempts,
			"locked_until": lockUntil,
		}).Warn("User account locked due to failed login attempts")
	}

	if err := s.userRepo.Update(user); err != nil {
		logrus.WithError(err).Error("Failed to update user failed attempts")
	}
}

// resetFailedAttempts remet à zéro le compteur d'échecs
func (s *AuthService) resetFailedAttempts(user *models.User) {
	if user.LoginAttempts > 0 || user.LockedUntil != nil {
		user.LoginAttempts = 0
		user.LockedUntil = nil
		if err := s.userRepo.Update(user); err != nil {
			logrus.WithError(err).Error("Failed to reset user failed attempts")
		}
	}
}

// logLoginAttempt enregistre une tentative de connection
func (s *AuthService) logLoginAttempt(userID *uuid.UUID, username, ipAddress, userAgent string, success bool, reason string) {
	// En production, utiliser une table d'audit dédiée
	logrus.WithFields(logrus.Fields{
		"user_id":    userID,
		"username":   username,
		"ip_address": ipAddress,
		"user_agent": userAgent,
		"success":    success,
		"reason":     reason,
	}).Info("Login attempt logged")
}

// sendEmailVerification envoie un email de vérification (stub)
func (s *AuthService) sendEmailVerification(user *models.User) {
	// TODO: Implémenter l'envoi d'email
	logrus.WithFields(logrus.Fields{
		"user_id": user.ID,
		"email":   user.Email,
	}).Info("Email verification would be sent")
}

// verifyTwoFactorCode vérifie un code 2FA (stub)
func (s *AuthService) verifyTwoFactorCode(secret, code string) bool {
	// TODO: Implémenter la vérification TOTP
	return false
}
