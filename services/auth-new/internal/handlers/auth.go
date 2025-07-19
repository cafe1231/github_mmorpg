package handlers

import (
	"auth/internal/config"
	"auth/internal/models"
	"auth/internal/service"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// AuthHandler gère les requêtes HTTP d'authentification
type AuthHandler struct {
	authService service.AuthServiceInterface
	config      *config.Config
}

// NewAuthHandler crée une nouvelle instance du handler d'authentification
func NewAuthHandler(authService service.AuthServiceInterface, config *config.Config) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		config:      config,
	}
}

// Register inscrit un nouvel utilisateur
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	user, err := h.authService.Register(&req)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "username already exists" || err.Error() == "email already exists" {
			status = http.StatusConflict
		}
		h.respondError(c, status, "Registration failed", err.Error())
		return
	}

	logrus.WithFields(logrus.Fields{
		"user_id":    user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"ip_address": c.ClientIP(),
		"request_id": c.GetHeader("X-Request-ID"),
	}).Info("User registered successfully")

	h.respondSuccess(c, http.StatusCreated, "User registered successfully", gin.H{
		"user_id": user.ID,
		"message": "Please check your email to verify your account",
	})
}

// Login connecte un utilisateur
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	// Extraire les informations de la requête
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	loginResp, err := h.authService.Login(req, ipAddress, userAgent)
	if err != nil {
		status := http.StatusUnauthorized
		message := "Authentication failed"

		// Déterminer le type d'erreur pour le status code approprié
		switch err.Error() {
		case "account locked":
			status = http.StatusLocked
			message = "Account temporarily locked due to multiple failed attempts"
		case "account suspended", "account banned":
			status = http.StatusForbidden
			message = "Account access restricted"
		case "email not verified":
			status = http.StatusUnauthorized
			message = "Please verify your email before logging in"
		case "two_factor_required":
			status = http.StatusUnauthorized
			message = "Two-factor authentication required"
		}

		h.respondError(c, status, message, err.Error())
		return
	}

	logrus.WithFields(logrus.Fields{
		"user_id":    loginResp.User.ID,
		"username":   loginResp.User.Username,
		"ip_address": ipAddress,
		"request_id": c.GetHeader("X-Request-ID"),
	}).Info("User logged in successfully")

	c.JSON(http.StatusOK, loginResp)
}

// RefreshToken renouvelle un token d'accès
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	loginResp, err := h.authService.RefreshToken(req)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "Token refresh failed", err.Error())
		return
	}

	logrus.WithFields(logrus.Fields{
		"user_id":    loginResp.User.ID,
		"username":   loginResp.User.Username,
		"request_id": c.GetHeader("X-Request-ID"),
	}).Debug("Token refreshed successfully")

	c.JSON(http.StatusOK, loginResp)
}

// ForgotPassword initie le processus de récupération de mot de passe
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req models.ForgotPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	err := h.authService.ForgotPassword(req.Email)
	if err != nil {
		// Ne pas révéler si l'email existe ou non pour la sécurité
		logrus.WithError(err).WithField("email", req.Email).Warn("Password reset request failed")
	}

	// Toujours retourner succès pour ne pas révéler si l'email existe
	h.respondSuccess(c, http.StatusOK, "Password reset instructions sent", gin.H{
		"message": "If the email exists, you will receive password reset instructions",
	})
}

// ResetPassword réinitialize le mot de passe avec un token
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	err := h.authService.ResetPassword(req.Token, req.NewPassword)
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "Password reset failed", err.Error())
		return
	}

	h.respondSuccess(c, http.StatusOK, "Password reset successfully", nil)
}

// VerifyEmail vérifie l'adresse email avec un token
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		h.respondError(c, http.StatusBadRequest, "Verification token required", "")
		return
	}

	err := h.authService.VerifyEmail(token)
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "Email verification failed", err.Error())
		return
	}

	h.respondSuccess(c, http.StatusOK, "Email verified successfully", nil)
}

// ResendVerification renvoie l'email de vérification (stub)
func (h *AuthHandler) ResendVerification(c *gin.Context) {
	h.respondError(c, http.StatusNotImplemented, "Email verification resend not implemented yet", "")
}

// GetProfile récupère le profil de l'utilisateur connecté
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "User not authenticated", err.Error())
		return
	}

	user, err := h.authService.GetProfile(userID)
	if err != nil {
		h.respondError(c, http.StatusNotFound, "User not found", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// UpdateProfile met à jour le profil utilisateur
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	var req models.UpdateProfileRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "User not authenticated", err.Error())
		return
	}

	user, err := h.authService.UpdateProfile(userID, req)
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "Profile update failed", err.Error())
		return
	}

	logrus.WithFields(logrus.Fields{
		"user_id":    userID,
		"request_id": c.GetHeader("X-Request-ID"),
	}).Info("Profile updated successfully")

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user":    user,
	})
}

// ChangePassword change le mot de passe de l'utilisateur
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req models.ChangePasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	if req.NewPassword != req.NewPasswordConfirm {
		h.respondError(c, http.StatusBadRequest, "Passwords do not match", "")
		return
	}

	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "User not authenticated", err.Error())
		return
	}

	err = h.authService.ChangePassword(userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "Password change failed", err.Error())
		return
	}

	logrus.WithFields(logrus.Fields{
		"user_id":    userID,
		"request_id": c.GetHeader("X-Request-ID"),
	}).Info("Password changed successfully")

	h.respondSuccess(c, http.StatusOK, "Password changed successfully", nil)
}

// Logout déconnecte l'utilisateur
func (h *AuthHandler) Logout(c *gin.Context) {
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "User not authenticated", err.Error())
		return
	}

	sessionID, err := h.getSessionIDFromContext(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "Session not found", err.Error())
		return
	}

	err = h.authService.Logout(userID, sessionID)
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, "Logout failed", err.Error())
		return
	}

	h.respondSuccess(c, http.StatusOK, "Logged out successfully", nil)
}

// LogoutAllDevices déconnecte l'utilisateur de tous ses appareils
func (h *AuthHandler) LogoutAllDevices(c *gin.Context) {
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "User not authenticated", err.Error())
		return
	}

	err = h.authService.LogoutAllDevices(userID)
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, "Logout failed", err.Error())
		return
	}

	h.respondSuccess(c, http.StatusOK, "Logged out from all devices", nil)
}

// GetSessions récupère les sessions actives de l'utilisateur
func (h *AuthHandler) GetSessions(c *gin.Context) {
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "User not authenticated", err.Error())
		return
	}

	sessions, err := h.authService.GetSessions(userID)
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, "Failed to get sessions", err.Error())
		return
	}

	// Convertir en format de réponse
	sessionResponses := make([]models.SessionResponse, len(sessions))
	for i, session := range sessions {
		sessionResponses[i] = models.SessionResponse{
			ID:           session.ID,
			DeviceInfo:   session.DeviceInfo,
			IPAddress:    session.IPAddress,
			CreatedAt:    session.CreatedAt,
			ExpiresAt:    session.ExpiresAt,
			LastActivity: session.LastActivity,
			IsActive:     session.IsActive,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"sessions": sessionResponses,
	})
}

// RevokeSession révoque une session spécifique
func (h *AuthHandler) RevokeSession(c *gin.Context) {
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "User not authenticated", err.Error())
		return
	}

	sessionIDStr := c.Param("id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "Invalid session ID", err.Error())
		return
	}

	err = h.authService.RevokeSession(userID, sessionID)
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "Failed to revoke session", err.Error())
		return
	}

	h.respondSuccess(c, http.StatusOK, "Session revoked successfully", nil)
}

// ValidateToken valide un token (pour les autres services)
func (h *AuthHandler) ValidateToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		h.respondError(c, http.StatusBadRequest, "Authorization header required", "")
		return
	}

	// Extraire le token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		h.respondError(c, http.StatusBadRequest, "Invalid authorization header format", "")
		return
	}

	token := parts[1]
	claims, err := h.authService.ValidateToken(token)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "Invalid token", err.Error())
		return
	}

	c.JSON(http.StatusOK, models.TokenValidationResponse{
		Valid:       true,
		UserID:      claims.UserID,
		Username:    claims.Username,
		Email:       claims.Email,
		Role:        claims.Role,
		Permissions: claims.Permissions,
		ExpiresAt:   claims.ExpiresAt.Time,
	})
}

// GetUserInfo récupère les informations d'un utilisateur (pour les autres services)
func (h *AuthHandler) GetUserInfo(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	user, err := h.authService.GetUserInfo(userID)
	if err != nil {
		h.respondError(c, http.StatusNotFound, "User not found", err.Error())
		return
	}

	c.JSON(http.StatusOK, models.UserInfoResponse{
		ID:               user.ID,
		Username:         user.Username,
		Email:            user.Email,
		FirstName:        user.FirstName,
		LastName:         user.LastName,
		Avatar:           user.Avatar,
		Role:             user.Role,
		Status:           user.Status,
		EmailVerified:    user.EmailVerified,
		EmailVerifiedAt:  user.EmailVerifiedAt,
		CreatedAt:        user.CreatedAt,
		LastLoginAt:      user.LastLoginAt,
		TwoFactorEnabled: user.TwoFactorEnabled,
	})
}

// Two-Factor Authentication handlers
func (h *AuthHandler) EnableTwoFactor(c *gin.Context) {
	h.respondError(c, http.StatusNotImplemented, "Two-factor authentication not implemented yet", "")
}

func (h *AuthHandler) DisableTwoFactor(c *gin.Context) {
	h.respondError(c, http.StatusNotImplemented, "Two-factor authentication not implemented yet", "")
}

func (h *AuthHandler) GetTwoFactorQR(c *gin.Context) {
	h.respondError(c, http.StatusNotImplemented, "Two-factor authentication not implemented yet", "")
}

// VerifyTwoFactor vérifie un code 2FA (stub)
func (h *AuthHandler) VerifyTwoFactor(c *gin.Context) {
	h.respondError(c, http.StatusNotImplemented, "Two-factor authentication not implemented yet", "")
}

// GetBackupCodes récupère les codes de sauvegarde 2FA (stub)
func (h *AuthHandler) GetBackupCodes(c *gin.Context) {
	h.respondError(c, http.StatusNotImplemented, "Two-factor authentication not implemented yet", "")
}

// RegenerateBackupCodes régénère les codes de sauvegarde 2FA (stub)
func (h *AuthHandler) RegenerateBackupCodes(c *gin.Context) {
	h.respondError(c, http.StatusNotImplemented, "Two-factor authentication not implemented yet", "")
}

// OAuth handlers (stubs)
func (h *AuthHandler) OAuthRedirect(c *gin.Context) {
	h.respondError(c, http.StatusNotImplemented, "OAuth not implemented yet", "")
}

func (h *AuthHandler) OAuthCallback(c *gin.Context) {
	h.respondError(c, http.StatusNotImplemented, "OAuth not implemented yet", "")
}

// Admin handlers
func (h *AuthHandler) ListUsers(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	users, total, err := h.authService.ListUsers(limit, offset, nil)
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, "Failed to list users", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users":  users,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *AuthHandler) GetUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	user, err := h.authService.GetUser(userID)
	if err != nil {
		h.respondError(c, http.StatusNotFound, "User not found", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

func (h *AuthHandler) UpdateUser(c *gin.Context) {
	h.respondError(c, http.StatusNotImplemented, "User administration not implemented yet", "")
}

// CreateUser crée un nouvel utilisateur (admin) (stub)
func (h *AuthHandler) CreateUser(c *gin.Context) {
	h.respondError(c, http.StatusNotImplemented, "User creation not implemented yet", "")
}

// DeleteUser supprime un utilisateur (admin) (stub)
func (h *AuthHandler) DeleteUser(c *gin.Context) {
	h.respondError(c, http.StatusNotImplemented, "User deletion not implemented yet", "")
}

// BanUser bannit un utilisateur (admin) (stub)
func (h *AuthHandler) BanUser(c *gin.Context) {
	h.respondError(c, http.StatusNotImplemented, "User banning not implemented yet", "")
}

// UnbanUser débannit un utilisateur (admin) (stub)
func (h *AuthHandler) UnbanUser(c *gin.Context) {
	h.respondError(c, http.StatusNotImplemented, "User unbanning not implemented yet", "")
}

func (h *AuthHandler) SuspendUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	reason := c.PostForm("reason")
	if reason == "" {
		reason = "Administrative action"
	}

	err = h.authService.SuspendUser(userID, reason)
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "Failed to suspend user", err.Error())
		return
	}

	h.respondSuccess(c, http.StatusOK, "User suspended successfully", nil)
}

func (h *AuthHandler) ActivateUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	err = h.authService.ActivateUser(userID)
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "Failed to activate user", err.Error())
		return
	}

	h.respondSuccess(c, http.StatusOK, "User activated successfully", nil)
}

func (h *AuthHandler) GetLoginAttempts(c *gin.Context) {
	h.respondError(c, http.StatusNotImplemented, "Login attempts audit not implemented yet", "")
}

func (h *AuthHandler) GetAuditLog(c *gin.Context) {
	h.respondError(c, http.StatusNotImplemented, "Audit log not implemented yet", "")
}

// GetStatistics récupère les statistiques d'administration (stub)
func (h *AuthHandler) GetStatistics(c *gin.Context) {
	h.respondError(c, http.StatusNotImplemented, "Statistics not implemented yet", "")
}

// GetAllSessions récupère toutes les sessions (admin) (stub)
func (h *AuthHandler) GetAllSessions(c *gin.Context) {
	h.respondError(c, http.StatusNotImplemented, "All sessions admin view not implemented yet", "")
}

// AdminRevokeSession révoque une session (admin) (stub)
func (h *AuthHandler) AdminRevokeSession(c *gin.Context) {
	h.respondError(c, http.StatusNotImplemented, "Admin session revocation not implemented yet", "")
}

// AdminLogoutUser déconnecte un utilisateur (admin) (stub)
func (h *AuthHandler) AdminLogoutUser(c *gin.Context) {
	h.respondError(c, http.StatusNotImplemented, "Admin user logout not implemented yet", "")
}

// SearchUsers recherche des utilisateurs (admin) (stub)
func (h *AuthHandler) SearchUsers(c *gin.Context) {
	h.respondError(c, http.StatusNotImplemented, "User search not implemented yet", "")
}

// GetUserSessionsAdmin récupère les sessions d'un utilisateur (admin) (stub)
func (h *AuthHandler) GetUserSessionsAdmin(c *gin.Context) {
	h.respondError(c, http.StatusNotImplemented, "Admin user sessions view not implemented yet", "")
}

// AddUserWarning ajoute un avertissement à un utilisateur (admin) (stub)
func (h *AuthHandler) AddUserWarning(c *gin.Context) {
	h.respondError(c, http.StatusNotImplemented, "User warnings not implemented yet", "")
}

// GetUserWarnings récupère les avertissements d'un utilisateur (admin) (stub)
func (h *AuthHandler) GetUserWarnings(c *gin.Context) {
	h.respondError(c, http.StatusNotImplemented, "User warnings not implemented yet", "")
}

// UpdateUserActivity met à jour l'activité d'un utilisateur (service) (stub)
func (h *AuthHandler) UpdateUserActivity(c *gin.Context) {
	h.respondError(c, http.StatusNotImplemented, "User activity update not implemented yet", "")
}

// GetUserPermissions récupère les permissions d'un utilisateur (service) (stub)
func (h *AuthHandler) GetUserPermissions(c *gin.Context) {
	h.respondError(c, http.StatusNotImplemented, "User permissions not implemented yet", "")
}

// Méthodes utilitaires

// getUserIDFromContext extrait l'ID utilisateur du contexte JWT
func (h *AuthHandler) getUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, fmt.Errorf("user ID not found in context")
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid user ID format")
	}

	return userID, nil
}

// getSessionIDFromContext extrait l'ID de session du contexte JWT
func (h *AuthHandler) getSessionIDFromContext(c *gin.Context) (uuid.UUID, error) {
	sessionIDInterface, exists := c.Get("session_id")
	if !exists {
		return uuid.Nil, fmt.Errorf("session ID not found in context")
	}

	sessionID, ok := sessionIDInterface.(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid session ID format")
	}

	return sessionID, nil
}

// respondSuccess envoie une réponse de succès standardisée
func (h *AuthHandler) respondSuccess(c *gin.Context, status int, message string, data interface{}) {
	response := models.SuccessResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		RequestID: c.GetHeader("X-Request-ID"),
	}
	c.JSON(status, response)
}

// respondError envoie une réponse d'erreur standardisée
func (h *AuthHandler) respondError(c *gin.Context, status int, message, details string) {
	response := models.ErrorResponse{
		Success:   false,
		Error:     message,
		Details:   details,
		RequestID: c.GetHeader("X-Request-ID"),
	}
	c.JSON(status, response)
}
