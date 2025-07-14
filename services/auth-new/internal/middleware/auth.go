package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"auth/internal/models"
)

// JWTAuth middleware d'authentification JWT
func JWTAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":      "Authorization header required",
				"message":    "Please provide a valid JWT token",
				"request_id": c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":      "Invalid authorization header format",
				"message":    "Use 'Bearer <token>' format",
				"request_id": c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			logrus.WithError(err).WithField("request_id", c.GetHeader("X-Request-ID")).Warn("JWT validation failed")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":      "Invalid token",
				"message":    "Token validation failed",
				"request_id": c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(*models.JWTClaims); ok && token.Valid {
			// Ajouter les informations utilisateur au contexte
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("user_role", claims.Role)
			c.Set("session_id", claims.SessionID)
			c.Set("permissions", claims.Permissions)

			// Ajouter des en-têtes pour les services downstream
			c.Header("X-User-ID", claims.UserID.String())
			c.Header("X-Username", claims.Username)
			c.Header("X-User-Role", claims.Role)

			logrus.WithFields(logrus.Fields{
				"user_id":    claims.UserID,
				"username":   claims.Username,
				"role":       claims.Role,
				"path":       c.Request.URL.Path,
				"method":     c.Request.Method,
				"request_id": c.GetHeader("X-Request-ID"),
			}).Debug("Authenticated request")

			c.Next()
		} else {
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
			// Pas de token, mais on continue quand même
			c.Next()
			return
		}

		// Si il y a un token, on essaie de le valider
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			// Token mal formaté, on continue sans authentification
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

		token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(jwtSecret), nil
		})

		if err == nil {
			if claims, ok := token.Claims.(*models.JWTClaims); ok && token.Valid {
				// Token valide, ajouter les infos au contexte
				c.Set("user_id", claims.UserID)
				c.Set("username", claims.Username)
				c.Set("user_role", claims.Role)
				c.Set("session_id", claims.SessionID)
				c.Set("permissions", claims.Permissions)
				c.Set("authenticated", true)

				logrus.WithFields(logrus.Fields{
					"user_id":    claims.UserID,
					"username":   claims.Username,
					"path":       c.Request.URL.Path,
					"request_id": c.GetHeader("X-Request-ID"),
				}).Debug("Optional auth: user authenticated")
			}
		} else {
			// Token invalide, continuer sans authentification
			logrus.WithError(err).WithFields(logrus.Fields{
				"path":       c.Request.URL.Path,
				"request_id": c.GetHeader("X-Request-ID"),
			}).Debug("Optional auth: invalid token, continuing unauthenticated")
		}

		c.Next()
	}
}

// RequireRole middleware pour vérifier les rôles
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":      "Authentication required",
				"message":    "User role not found in context",
				"request_id": c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		role := userRole.(string)
		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				logrus.WithFields(logrus.Fields{
					"user_role":     role,
					"allowed_roles": allowedRoles,
					"path":          c.Request.URL.Path,
					"request_id":    c.GetHeader("X-Request-ID"),
				}).Debug("Role check passed")
				c.Next()
				return
			}
		}

		// Log de l'accès refusé
		userID, _ := c.Get("user_id")
		username, _ := c.Get("username")
		
		logrus.WithFields(logrus.Fields{
			"user_id":       userID,
			"username":      username,
			"user_role":     role,
			"allowed_roles": allowedRoles,
			"path":          c.Request.URL.Path,
			"method":        c.Request.Method,
			"client_ip":     c.ClientIP(),
			"request_id":    c.GetHeader("X-Request-ID"),
		}).Warn("Access denied: insufficient role")

		c.JSON(http.StatusForbidden, gin.H{
			"error":        "Insufficient permissions",
			"message":      "You don't have permission to access this resource",
			"required_roles": allowedRoles,
			"your_role":    role,
			"request_id":   c.GetHeader("X-Request-ID"),
		})
		c.Abort()
	}
}

// RequirePermission middleware pour vérifier les permissions spécifiques
func RequirePermission(requiredPermissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userPermissions, exists := c.Get("permissions")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":      "Authentication required",
				"message":    "User permissions not found in context",
				"request_id": c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		permissions, ok := userPermissions.([]string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":      "Internal error",
				"message":    "Invalid permissions format",
				"request_id": c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		// Vérifier que l'utilisateur a toutes les permissions requises
		for _, requiredPerm := range requiredPermissions {
			hasPermission := false
			for _, userPerm := range permissions {
				if userPerm == requiredPerm {
					hasPermission = true
					break
				}
			}
			
			if !hasPermission {
				// Log de l'accès refusé
				userID, _ := c.Get("user_id")
				username, _ := c.Get("username")
				
				logrus.WithFields(logrus.Fields{
					"user_id":              userID,
					"username":             username,
					"user_permissions":     permissions,
					"required_permissions": requiredPermissions,
					"missing_permission":   requiredPerm,
					"path":                 c.Request.URL.Path,
					"method":               c.Request.Method,
					"client_ip":            c.ClientIP(),
					"request_id":           c.GetHeader("X-Request-ID"),
				}).Warn("Access denied: missing permission")

				c.JSON(http.StatusForbidden, gin.H{
					"error":                "Insufficient permissions",
					"message":              "You don't have the required permission to access this resource",
					"required_permissions": requiredPermissions,
					"missing_permission":   requiredPerm,
					"request_id":           c.GetHeader("X-Request-ID"),
				})
				c.Abort()
				return
			}
		}

		logrus.WithFields(logrus.Fields{
			"user_permissions":     permissions,
			"required_permissions": requiredPermissions,
			"path":                 c.Request.URL.Path,
			"request_id":           c.GetHeader("X-Request-ID"),
		}).Debug("Permission check passed")

		c.Next()
	}
}

// RequireOwnership middleware pour vérifier que l'utilisateur est propriétaire de la ressource
func RequireOwnership() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":      "Authentication required",
				"message":    "User ID not found in context",
				"request_id": c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		// Récupérer l'ID de la ressource depuis l'URL
		resourceUserID := c.Param("userID")
		if resourceUserID == "" {
			resourceUserID = c.Param("id") // Fallback
		}

		if resourceUserID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":      "Resource ID required",
				"message":    "User ID parameter missing from URL",
				"request_id": c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		// Convertir en UUID
		resourceUUID, err := uuid.Parse(resourceUserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":      "Invalid resource ID",
				"message":    "User ID must be a valid UUID",
				"request_id": c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		currentUserID := userID.(uuid.UUID)

		// Vérifier si l'utilisateur est propriétaire OU a un rôle admin
		userRole, _ := c.Get("user_role")
		role, _ := userRole.(string)

		if currentUserID != resourceUUID && role != models.RoleAdmin && role != models.RoleSuperUser {
			logrus.WithFields(logrus.Fields{
				"current_user_id": currentUserID,
				"resource_user_id": resourceUUID,
				"user_role":       role,
				"path":            c.Request.URL.Path,
				"method":          c.Request.Method,
				"client_ip":       c.ClientIP(),
				"request_id":      c.GetHeader("X-Request-ID"),
			}).Warn("Access denied: not resource owner")

			c.JSON(http.StatusForbidden, gin.H{
				"error":      "Access denied",
				"message":    "You can only access your own resources",
				"request_id": c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ExtractUserInfo helper pour extraire les informations utilisateur du contexte
func ExtractUserInfo(c *gin.Context) (uuid.UUID, string, string, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, "", "", errors.New("user ID not found in context")
	}

	username, exists := c.Get("username")
	if !exists {
		return uuid.Nil, "", "", errors.New("username not found in context")
	}

	role, exists := c.Get("user_role")
	if !exists {
		return uuid.Nil, "", "", errors.New("user role not found in context")
	}

	return userID.(uuid.UUID), username.(string), role.(string), nil
}

// IsAuthenticated helper pour vérifier si l'utilisateur est authentifié
func IsAuthenticated(c *gin.Context) bool {
	_, exists := c.Get("user_id")
	return exists
}

// HasRole helper pour vérifier si l'utilisateur a un rôle spécifique
func HasRole(c *gin.Context, roles ...string) bool {
	userRole, exists := c.Get("user_role")
	if !exists {
		return false
	}

	role := userRole.(string)
	for _, allowedRole := range roles {
		if role == allowedRole {
			return true
		}
	}
	return false
}

// HasPermission helper pour vérifier si l'utilisateur a une permission spécifique
func HasPermission(c *gin.Context, permission string) bool {
	userPermissions, exists := c.Get("permissions")
	if !exists {
		return false
	}

	permissions, ok := userPermissions.([]string)
	if !ok {
		return false
	}

	for _, perm := range permissions {
		if perm == permission {
			return true
		}
	}
	return false
}