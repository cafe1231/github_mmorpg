package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"

	"player/internal/config"
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

// Logger middleware personnalisé pour le service Player
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logrus.WithFields(logrus.Fields{
			"timestamp":    param.TimeStamp.Format(time.RFC3339),
			"client_ip":    param.ClientIP,
			"method":       param.Method,
			"path":         param.Path,
			"status_code":  param.StatusCode,
			"latency_ms":   param.Latency.Milliseconds(),
			"user_agent":   param.Request.UserAgent(),
			"request_id":   param.Request.Header.Get("X-Request-ID"),
			"service":      "player",
		}).Info("HTTP Request")

		return ""
	})
}

// Recovery middleware avec logging amélioré
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logrus.WithFields(logrus.Fields{
			"error":      recovered,
			"path":       c.Request.URL.Path,
			"method":     c.Request.Method,
			"client_ip":  c.ClientIP(),
			"request_id": c.GetHeader("X-Request-ID"),
			"service":    "player",
		}).Error("Panic recovered in player service")

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Internal server error",
			"request_id": c.GetHeader("X-Request-ID"),
		})
	})
}

// CORS middleware configuré pour le service Player
func CORS() gin.HandlerFunc {
	config := cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",
			"http://localhost:8080",
			"https://*.yourgame.com",
		},
		AllowMethods: []string{
			"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Authorization",
			"X-Requested-With",
			"X-Request-ID",
			"X-Client-Version",
			"X-User-ID",
			"X-Username",
			"X-User-Role",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"X-Request-ID",
			"X-Rate-Limit-Remaining",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	return cors.New(config)
}

// RequestID ajoute un ID unique à chaque requête
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}

// RateLimiter pour le service Player
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mutex    sync.RWMutex
	rpm      int
	burst    int
}

// NewRateLimiter crée un nouveau rate limiter
func NewRateLimiter(rpm, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rpm:      rpm,
		burst:    burst,
	}
}

// GetLimiter récupère ou crée un limiteur pour un client
func (rl *RateLimiter) GetLimiter(clientID string) *rate.Limiter {
	rl.mutex.RLock()
	limiter, exists := rl.limiters[clientID]
	rl.mutex.RUnlock()

	if !exists {
		rl.mutex.Lock()
		limiter = rate.NewLimiter(rate.Every(time.Minute/time.Duration(rl.rpm)), rl.burst)
		rl.limiters[clientID] = limiter
		rl.mutex.Unlock()
	}

	return limiter
}

// RateLimit middleware de rate limiting pour le service Player
func RateLimit(cfg config.RateLimitConfig) gin.HandlerFunc {
	limiter := NewRateLimiter(cfg.RequestsPerMinute, cfg.BurstSize)

	// Nettoyage périodique
	go func() {
		ticker := time.NewTicker(cfg.CleanupInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				limiter.mutex.Lock()
				limiter.limiters = make(map[string]*rate.Limiter)
				limiter.mutex.Unlock()
			}
		}
	}()

	return func(c *gin.Context) {
		// Identifier le client (IP + User ID si authentifié)
		clientID := c.ClientIP()
		if userID := c.GetHeader("X-User-ID"); userID != "" {
			clientID = fmt.Sprintf("%s:%s", clientID, userID)
		}

		clientLimiter := limiter.GetLimiter(clientID)

		if !clientLimiter.Allow() {
			logrus.WithFields(logrus.Fields{
				"client_id":  clientID,
				"client_ip":  c.ClientIP(),
				"path":       c.Request.URL.Path,
				"request_id": c.GetHeader("X-Request-ID"),
			}).Warn("Rate limit exceeded")

			c.Header("X-Rate-Limit-Remaining", "0")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":      "Rate limit exceeded",
				"message":    "Too many requests, please slow down",
				"retry_after": 60,
				"request_id": c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		remaining := int(clientLimiter.Tokens())
		c.Header("X-Rate-Limit-Remaining", fmt.Sprintf("%d", remaining))

		c.Next()
	}
}

// JWTAuth middleware d'authentification JWT (compatible avec le service Auth)
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
			// Ajouter les informations utilisateur au contexte
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("email", claims.Email)
			c.Set("user_role", claims.Role)
			c.Set("session_id", claims.SessionID)
			c.Set("permissions", claims.Permissions)

			// Ajouter des en-têtes pour les logs
			c.Header("X-User-ID", claims.UserID.String())
			c.Header("X-Username", claims.Username)
			c.Header("X-User-Role", claims.Role)

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

// RequireRole middleware qui vérifie les rôles utilisateur
func RequireRole(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := c.GetString("user_role")
		if userRole == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":      "User role not found",
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

		// Admin et superuser ont tous les droits
		if userRole == "admin" || userRole == "superuser" {
			hasRole = true
		}

		if !hasRole {
			logrus.WithFields(logrus.Fields{
				"user_role":      userRole,
				"required_roles": requiredRoles,
				"path":           c.Request.URL.Path,
				"request_id":     c.GetHeader("X-Request-ID"),
			}).Warn("Access denied: insufficient permissions")

			c.JSON(http.StatusForbidden, gin.H{
				"error":         "Insufficient permissions",
				"required_roles": requiredRoles,
				"user_role":     userRole,
				"request_id":    c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidateContentType vérifie le type de contenu
func ValidateContentType() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			contentType := c.GetHeader("Content-Type")
			if !strings.Contains(contentType, "application/json") {
				c.JSON(http.StatusUnsupportedMediaType, gin.H{
					"error":      "Content-Type must be application/json",
					"request_id": c.GetHeader("X-Request-ID"),
				})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

// SecurityHeaders ajoute les en-têtes de sécurité
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'none'; connect-src 'self'")
		
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		c.Next()
	}
}

// PlayerOwnership middleware qui vérifie que l'utilisateur accède à ses propres données
func PlayerOwnership() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ce middleware peut être utilisé pour les routes qui nécessitent 
		// une vérification de propriété des données joueur
		// La logique de vérification sera implémentée dans les services
		c.Next()
	}
}

// GetUserIDFromContext récupère l'ID utilisateur depuis le contexte
func GetUserIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, false
	}

	userID, ok := userIDInterface.(uuid.UUID)
	return userID, ok
}

// GetUsernameFromContext récupère le nom d'utilisateur depuis le contexte
func GetUsernameFromContext(c *gin.Context) (string, bool) {
	username := c.GetString("username")
	return username, username != ""
}

// GetUserRoleFromContext récupère le rôle utilisateur depuis le contexte
func GetUserRoleFromContext(c *gin.Context) (string, bool) {
	userRole := c.GetString("user_role")
	return userRole, userRole != ""
}