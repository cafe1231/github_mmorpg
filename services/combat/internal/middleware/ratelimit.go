// internal/middleware/ratelimit.go
package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"

	"combat/internal/config"
)

// RateLimiter gère les limiteurs de taux par client
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	config   config.RateLimitConfig
}

// NewRateLimiter crée un nouveau limiteur de taux
func NewRateLimiter(config config.RateLimitConfig) *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		config:   config,
	}

	// Nettoyage périodique des limiteurs inactifs
	go rl.cleanupLimiters()

	return rl
}

// GetLimiter récupère ou crée un limiteur pour un client
func (rl *RateLimiter) GetLimiter(clientID string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[clientID]
	if !exists {
		// Créer un nouveau limiteur : actions par minute avec burst
		limiter = rate.NewLimiter(
			rate.Every(time.Minute/time.Duration(rl.config.CombatActionsPerMinute)),
			rl.config.BurstSize,
		)
		rl.limiters[clientID] = limiter
	}

	return limiter
}

// cleanupLimiters nettoie les limiteurs inactifs
func (rl *RateLimiter) cleanupLimiters() {
	ticker := time.NewTicker(rl.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for clientID, limiter := range rl.limiters {
			// Supprimer les limiteurs qui n'ont pas été utilisés récemment
			if limiter.TokensAt(time.Now()) == float64(rl.config.BurstSize) {
				delete(rl.limiters, clientID)
			}
		}
		rl.mu.Unlock()
	}
}

// Instance globale du rate limiter
var globalRateLimiter *RateLimiter

// RateLimit middleware de limitation de taux
func RateLimit(config config.RateLimitConfig) gin.HandlerFunc {
	if globalRateLimiter == nil {
		globalRateLimiter = NewRateLimiter(config)
	}

	return func(c *gin.Context) {
		// Identifier le client (IP + user_id si authentifié)
		clientID := c.ClientIP()
		if userID, exists := c.Get("user_id"); exists {
			clientID = fmt.Sprintf("%s_%s", clientID, userID)
		}

		// Récupérer le limiteur pour ce client
		limiter := globalRateLimiter.GetLimiter(clientID)

		// Vérifier la limite
		if !limiter.Allow() {
			logrus.WithFields(logrus.Fields{
				"client_id":  clientID,
				"path":       c.Request.URL.Path,
				"method":     c.Request.Method,
				"request_id": c.GetHeader("X-Request-ID"),
			}).Warn("Combat rate limit exceeded")

			c.Header("X-Rate-Limit-Remaining", "0")
			c.Header("Retry-After", "60")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"message":     "Too many combat actions, please slow down",
				"retry_after": 60,
				"request_id":  c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		// Ajouter les headers de rate limiting
		remaining := int(limiter.TokensAt(time.Now()))
		c.Header("X-Rate-Limit-Remaining", fmt.Sprintf("%d", remaining))

		c.Next()
	}
}

// RequestID middleware pour générer un ID unique par requête
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("combat-%d", time.Now().UnixNano())
		}
		
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}