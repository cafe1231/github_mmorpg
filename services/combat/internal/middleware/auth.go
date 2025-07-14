// internal/middleware/auth.go
package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// JWTClaims représente les claims du JWT (compatible avec le service Auth)
type JWTClaims struct {
	UserID      uuid.UUID `json:"user_id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	Permissions []string  `json:"permissions"`
	SessionID   uuid.UUID `json:"session_id"`
	TokenType   string    `json:"token_type"`
	jwt.RegisteredClaims
}

// JWTAuth middleware d'authentification JWT
func JWTAuth(jwtSecret string) gin.HandlerFunc {
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

		// Vérifier le format Bearer
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":      "Invalid authorization header format",
				"request_id": c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		// Parser et valider le token
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error":      err.Error(),
				"path":       c.Request.URL.Path,
				"client_ip":  c.ClientIP(),
				"request_id": c.GetHeader("X-Request-ID"),
			}).Warn("JWT validation failed")

			c.JSON(http.StatusUnauthorized, gin.H{
				"error":      "Invalid token",
				"request_id": c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		// Vérifier les claims
		if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
			// Stocker les informations de l'utilisateur dans le contexte
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("role", claims.Role)
			c.Set("permissions", claims.Permissions)
			c.Set("session_id", claims.SessionID)

			logrus.WithFields(logrus.Fields{
				"user_id":    claims.UserID,
				"username":   claims.Username,
				"path":       c.Request.URL.Path,
				"request_id": c.GetHeader("X-Request-ID"),
			}).Debug("JWT authentication successful")

			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":      "Invalid token claims",
				"request_id": c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}
	}
}

// RequireRole middleware pour vérifier le rôle de l'utilisateur
func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":      "No role found in token",
				"request_id": c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		if role.(string) != requiredRole && role.(string) != "admin" && role.(string) != "superuser" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":      "Insufficient permissions",
				"required":   requiredRole,
				"current":    role.(string),
				"request_id": c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}