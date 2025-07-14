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

// JWTClaims reprÃ©sente les claims du JWT
type JWTClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
	jwt.RegisteredClaims
}

// JWTAuth middleware d'authentification JWT pour le gateway
func JWTAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// RÃ©cupÃ©rer le token depuis l'en-tÃªte Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logrus.WithFields(logrus.Fields{
				"path":       c.Request.URL.Path,
				"client_ip":  c.ClientIP(),
				"request_id": c.GetHeader("X-Request-ID"),
			}).Warn("Missing Authorization header")

			c.JSON(http.StatusUnauthorized, gin.H{
				"error":      "Authorization header required",
				"message":    "Please provide a valid authentication token",
				"request_id": c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		// VÃ©rifier le format Bearer
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			logrus.WithFields(logrus.Fields{
				"auth_header": authHeader,
				"path":        c.Request.URL.Path,
				"client_ip":   c.ClientIP(),
				"request_id":  c.GetHeader("X-Request-ID"),
			}).Warn("Invalid Authorization header format")

			c.JSON(http.StatusUnauthorized, gin.H{
				"error":      "Invalid authorization header format",
				"message":    "Format should be: Bearer <token>",
				"request_id": c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		// Parser et valider le token
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			// VÃ©rifier la mÃ©thode de signature
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

			// DiffÃ©rencier les types d'erreurs JWT
			var message string
			if errors.Is(err, jwt.ErrTokenExpired) {
				message = "Token has expired, please login again"
			} else if errors.Is(err, jwt.ErrTokenMalformed) {
				message = "Invalid token format"
			} else if errors.Is(err, jwt.ErrTokenNotValidYet) {
				message = "Token not valid yet"
			} else {
				message = "Invalid token"
			}

			c.JSON(http.StatusUnauthorized, gin.H{
				"error":      "Token validation failed",
				"message":    message,
				"request_id": c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		// VÃ©rifier les claims
		if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
			// Validation supplÃ©mentaire des claims
			if claims.UserID == uuid.Nil {
				logrus.WithFields(logrus.Fields{
					"claims":     claims,
					"path":       c.Request.URL.Path,
					"client_ip":  c.ClientIP(),
					"request_id": c.GetHeader("X-Request-ID"),
				}).Warn("Invalid user ID in token")

				c.JSON(http.StatusUnauthorized, gin.H{
					"error":      "Invalid token claims",
					"message":    "Token contains invalid user information",
					"request_id": c.GetHeader("X-Request-ID"),
				})
				c.Abort()
				return
			}

			if claims.Username == "" {
				logrus.WithFields(logrus.Fields{
					"user_id":    claims.UserID,
					"path":       c.Request.URL.Path,
					"client_ip":  c.ClientIP(),
					"request_id": c.GetHeader("X-Request-ID"),
				}).Warn("Empty username in token")

				c.JSON(http.StatusUnauthorized, gin.H{
					"error":      "Invalid token claims",
					"message":    "Token contains invalid username",
					"request_id": c.GetHeader("X-Request-ID"),
				})
				c.Abort()
				return
			}

			// Ajouter les informations utilisateur au contexte
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("user_role", claims.Role)

			// Ajouter des en-tÃªtes pour les services downstream
			c.Header("X-User-ID", claims.UserID.String())
			c.Header("X-Username", claims.Username)
			c.Header("X-User-Role", claims.Role)

			// Log de la requÃªte authentifiÃ©e
			logrus.WithFields(logrus.Fields{
				"user_id":    claims.UserID,
				"username":   claims.Username,
				"role":       claims.Role,
				"path":       c.Request.URL.Path,
				"method":     c.Request.Method,
				"client_ip":  c.ClientIP(),
				"request_id": c.GetHeader("X-Request-ID"),
			}).Debug("Authenticated request")

			// Continuer vers le handler suivant
			c.Next()
		} else {
			logrus.WithFields(logrus.Fields{
				"token_valid": token.Valid,
				"path":        c.Request.URL.Path,
				"client_ip":   c.ClientIP(),
				"request_id":  c.GetHeader("X-Request-ID"),
			}).Warn("Invalid token claims structure")

			c.JSON(http.StatusUnauthorized, gin.H{
				"error":      "Invalid token claims",
				"message":    "Token structure is invalid",
				"request_id": c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}
	}
}

// OptionalJWTAuth middleware d'authentification JWT optionnelle
func OptionalJWTAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// Pas de token, mais on continue quand mÃªme
			c.Next()
			return
		}

		// Si il y a un token, on essaie de le valider
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			// Token mal formatÃ©, on continue sans authentification
			logrus.WithFields(logrus.Fields{
				"auth_header": authHeader,
				"path":        c.Request.URL.Path,
				"client_ip":   c.ClientIP(),
				"request_id":  c.GetHeader("X-Request-ID"),
			}).Debug("Malformed auth header in optional auth")
			c.Next()
			return
		}

		tokenString := tokenParts[1]

		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(jwtSecret), nil
		})

		if err == nil {
			if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
				// Token valide, ajouter les infos au contexte
				c.Set("user_id", claims.UserID)
				c.Set("username", claims.Username)
				c.Set("user_role", claims.Role)

				// Headers pour les services downstream
				c.Header("X-User-ID", claims.UserID.String())
				c.Header("X-Username", claims.Username)
				c.Header("X-User-Role", claims.Role)

				logrus.WithFields(logrus.Fields{
					"user_id":    claims.UserID,
					"username":   claims.Username,
					"path":       c.Request.URL.Path,
					"request_id": c.GetHeader("X-Request-ID"),
				}).Debug("Optional auth successful")
			}
		} else {
			logrus.WithFields(logrus.Fields{
				"error":      err.Error(),
				"path":       c.Request.URL.Path,
				"client_ip":  c.ClientIP(),
				"request_id": c.GetHeader("X-Request-ID"),
			}).Debug("Optional auth failed, continuing without auth")
		}

		// Continuer dans tous les cas
		c.Next()
	}
}

// GetUserIDFromContext rÃ©cupÃ¨re l'ID utilisateur depuis le contexte Gin
func GetUserIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, false
	}

	id, ok := userID.(uuid.UUID)
	return id, ok
}

// GetUsernameFromContext rÃ©cupÃ¨re le nom d'utilisateur depuis le contexte Gin
func GetUsernameFromContext(c *gin.Context) (string, bool) {
	username, exists := c.Get("username")
	if !exists {
		return "", false
	}

	name, ok := username.(string)
	return name, ok
}

// GetUserRoleFromContext rÃ©cupÃ¨re le rÃ´le utilisateur depuis le contexte Gin
func GetUserRoleFromContext(c *gin.Context) (string, bool) {
	userRole, exists := c.Get("user_role")
	if !exists {
		return "", false
	}

	role, ok := userRole.(string)
	return role, ok
}

// IsAdmin vÃ©rifie si l'utilisateur connectÃ© est un administrateur
func IsAdmin(c *gin.Context) bool {
	role, exists := GetUserRoleFromContext(c)
	return exists && role == "admin"
}

// IsModerator vÃ©rifie si l'utilisateur connectÃ© est un modÃ©rateur ou admin
func IsModerator(c *gin.Context) bool {
	role, exists := GetUserRoleFromContext(c)
	return exists && (role == "moderator" || role == "admin")
}

// RequireOwnership vÃ©rifie que l'utilisateur est propriÃ©taire d'une ressource
func RequireOwnership(getResourceOwnerID func(*gin.Context) (uuid.UUID, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserIDFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":      "Authentication required",
				"request_id": c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		userRole, _ := GetUserRoleFromContext(c)

		// Admin a tous les droits
		if userRole == "admin" {
			c.Next()
			return
		}

		// VÃ©rifier la propriÃ©tÃ© de la ressource
		resourceOwnerID, err := getResourceOwnerID(c)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error":      err.Error(),
				"user_id":    userID,
				"path":       c.Request.URL.Path,
				"request_id": c.GetHeader("X-Request-ID"),
			}).Error("Failed to get resource owner ID")

			c.JSON(http.StatusInternalServerError, gin.H{
				"error":      "Failed to verify ownership",
				"request_id": c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		if userID != resourceOwnerID {
			logrus.WithFields(logrus.Fields{
				"user_id":           userID,
				"resource_owner_id": resourceOwnerID,
				"path":              c.Request.URL.Path,
				"request_id":        c.GetHeader("X-Request-ID"),
			}).Warn("Access denied: not the owner of this resource")

			c.JSON(http.StatusForbidden, gin.H{
				"error":      "Access denied: not the owner of this resource",
				"request_id": c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
