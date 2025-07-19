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

// RateLimiter interface pour différents types de limiteurs
type RateLimiter interface {
	Allow(key string) bool
	AllowN(key string, n int) bool
	GetInfo(key string) RateLimitInfo
}

// RateLimitInfo informations sur l'état du rate limiting
type RateLimitInfo struct {
	Limit      int           `json:"limit"`
	Remaining  int           `json:"remaining"`
	ResetTime  time.Time     `json:"reset_time"`
	RetryAfter time.Duration `json:"retry_after"`
}

// MemoryRateLimiter implémentation en mémoire du rate limiter
type MemoryRateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
	cleanup  time.Duration
}

// NewMemoryRateLimiter crée un nouveau rate limiter en mémoire
func NewMemoryRateLimiter(requestsPerMinute int, burst int) *MemoryRateLimiter {
	rl := &MemoryRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rate.Limit(requestsPerMinute) / 60, // Convertir en requêtes par seconde
		burst:    burst,
		cleanup:  5 * time.Minute,
	}

	// Démarrer le nettoyage périodique
	go rl.cleanupRoutine()

	return rl
}

// Allow vérifie si une requête est autorisée
func (rl *MemoryRateLimiter) Allow(key string) bool {
	return rl.AllowN(key, 1)
}

// AllowN vérifie si N requêtes sont autorisées
func (rl *MemoryRateLimiter) AllowN(key string, n int) bool {
	limiter := rl.getLimiter(key)
	return limiter.AllowN(time.Now(), n)
}

// GetInfo retourne les informations sur le rate limiting
func (rl *MemoryRateLimiter) GetInfo(key string) RateLimitInfo {
	limiter := rl.getLimiter(key)

	// Calculer les tokens restants (approximation)
	tokens := int(limiter.Tokens())
	if tokens > rl.burst {
		tokens = rl.burst
	}

	return RateLimitInfo{
		Limit:      rl.burst,
		Remaining:  tokens,
		ResetTime:  time.Now().Add(time.Duration(float64(rl.burst-tokens) / float64(rl.rate))),
		RetryAfter: time.Second,
	}
}

// getLimiter récupère ou crée un limiter pour une clé
func (rl *MemoryRateLimiter) getLimiter(key string) *rate.Limiter {
	rl.mu.RLock()
	limiter, exists := rl.limiters[key]
	rl.mu.RUnlock()

	if !exists {
		rl.mu.Lock()
		// Double-check après acquisition du lock exclusif
		if limiter, exists = rl.limiters[key]; !exists {
			limiter = rate.NewLimiter(rl.rate, rl.burst)
			rl.limiters[key] = limiter
		}
		rl.mu.Unlock()
	}

	return limiter
}

// cleanupRoutine nettoie périodiquement les limiters inactifs
func (rl *MemoryRateLimiter) cleanupRoutine() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		// Supprimer les limiters qui ont tous leurs tokens (inactifs)
		for key, limiter := range rl.limiters {
			if limiter.Tokens() >= float64(rl.burst) {
				delete(rl.limiters, key)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimit middleware de rate limiting global
func RateLimit(config config.RateLimitConfig) gin.HandlerFunc {
	limiter := NewMemoryRateLimiter(config.RequestsPerMinute, config.BurstSize)

	return func(c *gin.Context) {
		key := getClientKey(c)

		if !limiter.Allow(key) {
			info := limiter.GetInfo(key)

			logrus.WithFields(logrus.Fields{
				"client_key": key,
				"client_ip":  c.ClientIP(),
				"path":       c.Request.URL.Path,
				"method":     c.Request.Method,
				"user_agent": c.Request.UserAgent(),
				"limit":      info.Limit,
				"remaining":  info.Remaining,
				"request_id": c.GetHeader("X-Request-ID"),
			}).Warn("Rate limit exceeded")

			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", info.Limit))
			c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", info.Remaining))
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", info.ResetTime.Unix()))
			c.Header("Retry-After", fmt.Sprintf("%.0f", info.RetryAfter.Seconds()))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"limit":       info.Limit,
				"remaining":   info.Remaining,
				"reset_time":  info.ResetTime.Unix(),
				"retry_after": int(info.RetryAfter.Seconds()),
				"request_id":  c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		// Ajouter les headers de rate limiting
		info := limiter.GetInfo(key)
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", info.Limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", info.Remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", info.ResetTime.Unix()))

		c.Next()
	}
}

// CombatRateLimit rate limiting spécifique aux actions de combat
func CombatRateLimit(actionsPerMinute int) gin.HandlerFunc {
	limiter := NewMemoryRateLimiter(actionsPerMinute, actionsPerMinute/4) // Burst = 25% du taux

	return func(c *gin.Context) {
		// Utiliser l'ID du personnage comme clé pour les actions de combat
		characterID := c.GetString("character_id")
		if characterID == "" {
			characterID = getClientKey(c) // Fallback sur IP
		}

		key := fmt.Sprintf("combat:%s", characterID)

		if !limiter.Allow(key) {
			info := limiter.GetInfo(key)

			logrus.WithFields(logrus.Fields{
				"character_id": characterID,
				"client_ip":    c.ClientIP(),
				"action":       c.Request.URL.Path,
				"limit":        info.Limit,
				"remaining":    info.Remaining,
				"request_id":   c.GetHeader("X-Request-ID"),
			}).Warn("Combat action rate limit exceeded")

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many combat actions",
				"limit":       info.Limit,
				"remaining":   info.Remaining,
				"retry_after": int(info.RetryAfter.Seconds()),
				"message":     "Please wait before performing another action",
				"request_id":  c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AuthRateLimit rate limiting pour les tentatives d'authentification
func AuthRateLimit(attemptsPerMinute int) gin.HandlerFunc {
	limiter := NewMemoryRateLimiter(attemptsPerMinute, 3) // Burst très bas pour la sécurité

	return func(c *gin.Context) {
		key := fmt.Sprintf("auth:%s", c.ClientIP())

		if !limiter.Allow(key) {
			info := limiter.GetInfo(key)

			logrus.WithFields(logrus.Fields{
				"client_ip":  c.ClientIP(),
				"user_agent": c.Request.UserAgent(),
				"path":       c.Request.URL.Path,
				"limit":      info.Limit,
				"remaining":  info.Remaining,
				"request_id": c.GetHeader("X-Request-ID"),
			}).Warn("Authentication rate limit exceeded")

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many authentication attempts",
				"limit":       info.Limit,
				"retry_after": int(info.RetryAfter.Seconds()),
				"message":     "Please wait before trying again",
				"request_id":  c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// PvPRateLimit rate limiting pour les actions PvP
func PvPRateLimit(actionsPerMinute int) gin.HandlerFunc {
	limiter := NewMemoryRateLimiter(actionsPerMinute, actionsPerMinute/6) // Burst = ~17% du taux

	return func(c *gin.Context) {
		characterID := c.GetString("character_id")
		if characterID == "" {
			characterID = getClientKey(c)
		}

		key := fmt.Sprintf("pvp:%s", characterID)

		if !limiter.Allow(key) {
			info := limiter.GetInfo(key)

			logrus.WithFields(logrus.Fields{
				"character_id": characterID,
				"client_ip":    c.ClientIP(),
				"action":       c.Request.URL.Path,
				"limit":        info.Limit,
				"remaining":    info.Remaining,
				"request_id":   c.GetHeader("X-Request-ID"),
			}).Warn("PvP action rate limit exceeded")

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many PvP actions",
				"limit":       info.Limit,
				"remaining":   info.Remaining,
				"retry_after": int(info.RetryAfter.Seconds()),
				"message":     "Please wait before performing another PvP action",
				"request_id":  c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdaptiveRateLimit rate limiting adaptatif basé sur la charge
type AdaptiveRateLimit struct {
	baseLimiter *MemoryRateLimiter
	currentLoad float64
	maxLoad     float64
	mu          sync.RWMutex
}

// NewAdaptiveRateLimit crée un rate limiter adaptatif
func NewAdaptiveRateLimit(baseRate int, burst int, maxLoad float64) *AdaptiveRateLimit {
	return &AdaptiveRateLimit{
		baseLimiter: NewMemoryRateLimiter(baseRate, burst),
		maxLoad:     maxLoad,
	}
}

// UpdateLoad met à jour la charge actuelle du système
func (arl *AdaptiveRateLimit) UpdateLoad(load float64) {
	arl.mu.Lock()
	defer arl.mu.Unlock()
	arl.currentLoad = load
}

// Middleware retourne le middleware de rate limiting adaptatif
func (arl *AdaptiveRateLimit) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		arl.mu.RLock()
		load := arl.currentLoad
		arl.mu.RUnlock()

		// Ajuster le rate limiting basé sur la charge
		penalty := 1.0
		if load > arl.maxLoad*0.8 { // Commencer à pénaliser à 80% de la charge max
			penalty = 1.0 + ((load - arl.maxLoad*0.8) / (arl.maxLoad * 0.2))
		}

		key := getClientKey(c)

		// Appliquer la pénalité en demandant plus de tokens
		tokensNeeded := int(penalty)
		if !arl.baseLimiter.AllowN(key, tokensNeeded) {
			info := arl.baseLimiter.GetInfo(key)

			logrus.WithFields(logrus.Fields{
				"client_key":    key,
				"system_load":   load,
				"penalty":       penalty,
				"tokens_needed": tokensNeeded,
				"request_id":    c.GetHeader("X-Request-ID"),
			}).Warn("Adaptive rate limit exceeded due to high system load")

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded due to high system load",
				"system_load": load,
				"retry_after": int(info.RetryAfter.Seconds()),
				"request_id":  c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// getClientKey génère une clé unique pour identifier le client
func getClientKey(c *gin.Context) string {
	// Priorité : User ID > Character ID > IP Address
	if userID := c.GetString("user_id"); userID != "" {
		return fmt.Sprintf("user:%s", userID)
	}

	if characterID := c.GetString("character_id"); characterID != "" {
		return fmt.Sprintf("char:%s", characterID)
	}

	return fmt.Sprintf("ip:%s", c.ClientIP())
}

// WhitelistRateLimit rate limiter avec liste blanche
func WhitelistRateLimit(config config.RateLimitConfig, whitelist []string) gin.HandlerFunc {
	limiter := NewMemoryRateLimiter(config.RequestsPerMinute, config.BurstSize)

	// Convertir la liste blanche en map pour une recherche plus rapide
	whitelistMap := make(map[string]bool)
	for _, ip := range whitelist {
		whitelistMap[ip] = true
	}

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		// Vérifier si l'IP est dans la liste blanche
		if whitelistMap[clientIP] {
			c.Next()
			return
		}

		// Appliquer le rate limiting normal
		key := getClientKey(c)
		if !limiter.Allow(key) {
			info := limiter.GetInfo(key)

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"retry_after": int(info.RetryAfter.Seconds()),
				"request_id":  c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// BurstProtection protection contre les burst attacks
func BurstProtection(maxBurstSize int, timeWindow time.Duration) gin.HandlerFunc {
	type clientBurst struct {
		requests  []time.Time
		lastReset time.Time
	}

	clients := make(map[string]*clientBurst)
	mu := sync.RWMutex{}

	return func(c *gin.Context) {
		key := getClientKey(c)
		now := time.Now()

		mu.Lock()
		defer mu.Unlock()

		client, exists := clients[key]
		if !exists {
			client = &clientBurst{
				requests:  []time.Time{},
				lastReset: now,
			}
			clients[key] = client
		}

		// Nettoyer les anciennes requêtes
		validRequests := []time.Time{}
		for _, reqTime := range client.requests {
			if now.Sub(reqTime) < timeWindow {
				validRequests = append(validRequests, reqTime)
			}
		}
		client.requests = validRequests

		// Vérifier si le burst est dépassé
		if len(client.requests) >= maxBurstSize {
			logrus.WithFields(logrus.Fields{
				"client_key":  key,
				"burst_size":  len(client.requests),
				"max_burst":   maxBurstSize,
				"time_window": timeWindow,
				"request_id":  c.GetHeader("X-Request-ID"),
			}).Warn("Burst protection triggered")

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Burst limit exceeded",
				"max_burst":   maxBurstSize,
				"time_window": timeWindow.String(),
				"retry_after": int(timeWindow.Seconds()),
				"request_id":  c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		// Ajouter la requête actuelle
		client.requests = append(client.requests, now)

		c.Next()
	}
}
