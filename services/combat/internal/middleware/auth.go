package middleware

import (
	"combat/internal/config"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// JWTClaims représente les claims du JWT
type JWTClaims struct {
	UserID      string   `json:"user_id"`
	CharacterID string   `json:"character_id"`
	Username    string   `json:"username"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// AuthMiddleware crée le middleware d'authentification JWT
func AuthMiddleware(config *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Récupérer le token depuis l'en-tête Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":      "Authorization header required",
				"request_id": c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		// Vérifier le format "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":      "Invalid authorization header format",
				"request_id": c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Valider et parser le JWT
		claims, err := validateJWT(tokenString, config.JWT.Secret)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error":      err.Error(),
				"ip":         c.ClientIP(),
				"user_agent": c.Request.UserAgent(),
				"path":       c.Request.URL.Path,
				"request_id": c.GetHeader("X-Request-ID"),
			}).Warn("JWT validation failed")

			c.JSON(http.StatusUnauthorized, gin.H{
				"error":      "Invalid or expired token",
				"request_id": c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		// Vérifier que le token n'est pas expiré
		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":      "Token expired",
				"request_id": c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		// Stocker les informations de l'utilisateur dans le contexte
		c.Set("user_id", claims.UserID)
		c.Set("character_id", claims.CharacterID)
		c.Set("username", claims.Username)
		c.Set("user_role", claims.Role)
		c.Set("permissions", claims.Permissions)
		c.Set("jwt_claims", claims)

		// Log de l'authentification réussie
		logrus.WithFields(logrus.Fields{
			"user_id":      claims.UserID,
			"character_id": claims.CharacterID,
			"username":     claims.Username,
			"role":         claims.Role,
			"path":         c.Request.URL.Path,
			"method":       c.Request.Method,
			"request_id":   c.GetHeader("X-Request-ID"),
		}).Debug("User authenticated successfully")

		c.Next()
	}
}

// OptionalAuthMiddleware permet l'authentification optionnelle
func OptionalAuthMiddleware(config *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// Pas d'en-tête, continuer sans authentification
			c.Next()
			return
		}

		// Tenter l'authentification
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			tokenString := parts[1]
			if claims, err := validateJWT(tokenString, config.JWT.Secret); err == nil {
				// Authentification réussie
				c.Set("user_id", claims.UserID)
				c.Set("character_id", claims.CharacterID)
				c.Set("username", claims.Username)
				c.Set("user_role", claims.Role)
				c.Set("permissions", claims.Permissions)
				c.Set("authenticated", true)
			}
		}

		c.Next()
	}
}

// RequireRole middleware pour vérifier les rôles utilisateur
func RequireRole(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := c.GetString("user_role")
		if userRole == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":      "Authentication required",
				"request_id": c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		// Vérifier si l'utilisateur a l'un des rôles requis
		hasRole := false
		for _, requiredRole := range requiredRoles {
			if userRole == requiredRole {
				hasRole = true
				break
			}
		}

		// Les admins ont accès à tout
		if userRole == "admin" || userRole == "superuser" {
			hasRole = true
		}

		if !hasRole {
			logrus.WithFields(logrus.Fields{
				"user_id":        c.GetString("user_id"),
				"user_role":      userRole,
				"required_roles": requiredRoles,
				"path":           c.Request.URL.Path,
				"request_id":     c.GetHeader("X-Request-ID"),
			}).Warn("Access denied: insufficient permissions")

			c.JSON(http.StatusForbidden, gin.H{
				"error":      "Insufficient permissions",
				"required":   requiredRoles,
				"user_role":  userRole,
				"request_id": c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePermission middleware pour vérifier les permissions spécifiques
func RequirePermission(requiredPermissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permissions, exists := c.Get("permissions")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":      "Authentication required",
				"request_id": c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		userPermissions, ok := permissions.([]string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":      "Invalid permissions format",
				"request_id": c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		// Vérifier si l'utilisateur a au moins une des permissions requises
		hasPermission := false
		for _, required := range requiredPermissions {
			for _, userPerm := range userPermissions {
				if userPerm == required || userPerm == "*" { // * = toutes les permissions
					hasPermission = true
					break
				}
			}
			if hasPermission {
				break
			}
		}

		if !hasPermission {
			logrus.WithFields(logrus.Fields{
				"user_id":              c.GetString("user_id"),
				"user_permissions":     userPermissions,
				"required_permissions": requiredPermissions,
				"path":                 c.Request.URL.Path,
				"request_id":           c.GetHeader("X-Request-ID"),
			}).Warn("Access denied: insufficient permissions")

			c.JSON(http.StatusForbidden, gin.H{
				"error":      "Insufficient permissions",
				"required":   requiredPermissions,
				"request_id": c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireCharacterOwnership vérifie que l'utilisateur possède le personnage
func RequireCharacterOwnership() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestedCharacterID := c.Param("characterId")
		if requestedCharacterID == "" {
			requestedCharacterID = c.Param("id") // Fallback
		}

		userCharacterID := c.GetString("character_id")
		userRole := c.GetString("user_role")

		// Les admins peuvent accéder à tous les personnages
		if userRole == "admin" || userRole == "moderator" {
			c.Next()
			return
		}

		// Vérifier que l'utilisateur possède le personnage
		if requestedCharacterID != "" && requestedCharacterID != userCharacterID {
			logrus.WithFields(logrus.Fields{
				"user_id":                c.GetString("user_id"),
				"user_character_id":      userCharacterID,
				"requested_character_id": requestedCharacterID,
				"path":                   c.Request.URL.Path,
				"request_id":             c.GetHeader("X-Request-ID"),
			}).Warn("Access denied: character ownership mismatch")

			c.JSON(http.StatusForbidden, gin.H{
				"error":      "You can only access your own character",
				"request_id": c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitByUser applique un rate limiting par utilisateur
func RateLimitByUser(requestsPerMinute int) gin.HandlerFunc {
	// Simple in-memory rate limiter par utilisateur
	userLimits := make(map[string][]time.Time)

	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			userID = c.ClientIP() // Fallback sur l'IP si pas d'utilisateur
		}

		now := time.Now()

		// Nettoyer les anciennes requêtes
		if requests, exists := userLimits[userID]; exists {
			validRequests := []time.Time{}
			for _, reqTime := range requests {
				if now.Sub(reqTime) < time.Minute {
					validRequests = append(validRequests, reqTime)
				}
			}
			userLimits[userID] = validRequests
		}

		// Vérifier la limite
		if len(userLimits[userID]) >= requestsPerMinute {
			logrus.WithFields(logrus.Fields{
				"user_id":    userID,
				"limit":      requestsPerMinute,
				"requests":   len(userLimits[userID]),
				"path":       c.Request.URL.Path,
				"request_id": c.GetHeader("X-Request-ID"),
			}).Warn("User rate limit exceeded")

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"limit":       requestsPerMinute,
				"window":      "1 minute",
				"retry_after": config.DefaultRetryAfterSeconds,
				"request_id":  c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		// Ajouter la requête actuelle
		userLimits[userID] = append(userLimits[userID], now)
		c.Next()
	}
}

// validateJWT valide et parse un token JWT
func validateJWT(tokenString, secret string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Vérifier la méthode de signature
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Validation supplémentaire des claims
	if claims.UserID == "" {
		return nil, fmt.Errorf("missing user_id in token")
	}

	if _, err := uuid.Parse(claims.UserID); err != nil {
		return nil, fmt.Errorf("invalid user_id format")
	}

	if claims.CharacterID != "" {
		if _, err := uuid.Parse(claims.CharacterID); err != nil {
			return nil, fmt.Errorf("invalid character_id format")
		}
	}

	return claims, nil
}

// ServiceAuthentication middleware pour l'authentification entre services
func ServiceAuthentication(allowedServices []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceToken := c.GetHeader("X-Service-Token")
		serviceName := c.GetHeader("X-Service-Name")

		if serviceToken == "" || serviceName == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":      "Service authentication required",
				"request_id": c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		// Vérifier que le service est autorisé
		allowed := false
		for _, allowedService := range allowedServices {
			if serviceName == allowedService {
				allowed = true
				break
			}
		}

		if !allowed {
			logrus.WithFields(logrus.Fields{
				"service_name": serviceName,
				"allowed":      allowedServices,
				"path":         c.Request.URL.Path,
				"request_id":   c.GetHeader("X-Request-ID"),
			}).Warn("Unauthorized service access attempt")

			c.JSON(http.StatusForbidden, gin.H{
				"error":      "Service not authorized",
				"request_id": c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		// TODO: Valider le token du service (par example avec un secret partagé)
		// Pour l'instant, on fait confiance si le nom du service est correct

		c.Set("service_name", serviceName)
		c.Set("service_authenticated", true)
		c.Next()
	}
}
