package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"

	"gateway/internal/config"
)

// MÃ©triques Prometheus
var (
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_requests_total",
			Help: "Total number of requests processed by the gateway",
		},
		[]string{"method", "endpoint", "status_code", "service"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gateway_request_duration_seconds",
			Help:    "Request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "service"},
	)

	activeConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "gateway_active_connections",
			Help: "Number of active connections",
		},
	)

	rateLimitHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_rate_limit_hits_total",
			Help: "Total number of rate limit hits",
		},
		[]string{"client_ip"},
	)
)

func init() {
	prometheus.MustRegister(requestsTotal)
	prometheus.MustRegister(requestDuration)
	prometheus.MustRegister(activeConnections)
	prometheus.MustRegister(rateLimitHits)
}

// Logger middleware personnalisÃ© pour le gateway
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Log structurÃ© pour le gateway
		logrus.WithFields(logrus.Fields{
			"timestamp":   param.TimeStamp.Format(time.RFC3339),
			"client_ip":   param.ClientIP,
			"method":      param.Method,
			"path":        param.Path,
			"status_code": param.StatusCode,
			"latency_ms":  param.Latency.Milliseconds(),
			"user_agent":  param.Request.UserAgent(),
			"request_id":  param.Request.Header.Get("X-Request-ID"),
			"service":     "gateway",
		}).Info("HTTP Request")

		return ""
	})
}

// Recovery middleware avec logging amÃ©liorÃ©
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logrus.WithFields(logrus.Fields{
			"error":      recovered,
			"path":       c.Request.URL.Path,
			"method":     c.Request.Method,
			"client_ip":  c.ClientIP(),
			"request_id": c.GetHeader("X-Request-ID"),
			"service":    "gateway",
		}).Error("Panic recovered in gateway")

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Internal server error",
			"request_id": c.GetHeader("X-Request-ID"),
		})
	})
}

// CORS middleware configurÃ© pour les jeux
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
			"X-Game-Session",
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

// RequestID ajoute un ID unique Ã  chaque requÃªte
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

// RateLimiter avec support distribuÃ© pour gaming
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mutex    sync.RWMutex
	rpm      int
	burst    int
}

// NewRateLimiter crÃ©e un nouveau rate limiter
func NewRateLimiter(rpm, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rpm:      rpm,
		burst:    burst,
	}
}

// GetLimiter rÃ©cupÃ¨re ou crÃ©e un limiteur pour un client
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

// CleanupOldLimiters nettoie les anciens limiteurs
func (rl *RateLimiter) CleanupOldLimiters() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	// En production, implÃ©menter une logique de nettoyage basÃ©e sur le temps
	// Pour simplifier, on garde tous les limiteurs pour l'instant
}

// RateLimit middleware avec diffÃ©rents niveaux pour gaming
func RateLimit(cfg config.RateLimitConfig) gin.HandlerFunc {
	limiter := NewRateLimiter(cfg.RequestsPerMinute, cfg.BurstSize)

	// Nettoyage pÃ©riodique
	go func() {
		ticker := time.NewTicker(cfg.CleanupInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				limiter.CleanupOldLimiters()
			}
		}
	}()

	return func(c *gin.Context) {
		// Identifier le client (IP + User ID si authentifiÃ©)
		clientID := c.ClientIP()
		if userID := c.GetHeader("X-User-ID"); userID != "" {
			clientID = fmt.Sprintf("%s:%s", clientID, userID)
		}

		// Obtenir le limiteur pour ce client
		clientLimiter := limiter.GetLimiter(clientID)

		// VÃ©rifier la limite
		if !clientLimiter.Allow() {
			// IncrÃ©menter la mÃ©trique
			rateLimitHits.WithLabelValues(c.ClientIP()).Inc()

			logrus.WithFields(logrus.Fields{
				"client_id":  clientID,
				"client_ip":  c.ClientIP(),
				"path":       c.Request.URL.Path,
				"method":     c.Request.Method,
				"request_id": c.GetHeader("X-Request-ID"),
			}).Warn("Rate limit exceeded")

			c.Header("X-Rate-Limit-Remaining", "0")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"message":     "Too many requests, please slow down",
				"retry_after": 60,
				"request_id":  c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		// Ajouter les headers de rate limiting
		remaining := int(clientLimiter.Tokens())
		c.Header("X-Rate-Limit-Remaining", fmt.Sprintf("%d", remaining))

		c.Next()
	}
}

// Metrics middleware pour Prometheus
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		activeConnections.Inc()

		// Extraire le service de destination depuis le path
		service := extractServiceFromPath(c.Request.URL.Path)

		c.Next()

		// Calculer la durÃ©e
		duration := time.Since(start).Seconds()
		status := fmt.Sprintf("%d", c.Writer.Status())

		// Enregistrer les mÃ©triques
		requestsTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			status,
			service,
		).Inc()

		requestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			service,
		).Observe(duration)

		activeConnections.Dec()
	}
}

// HighPerformanceMode middleware pour les routes critiques (combat)
func HighPerformanceMode() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Headers pour optimiser les performances
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")

		// PrioritÃ© Ã©levÃ©e pour les requÃªtes de combat
		c.Header("X-Priority", "high")

		// Timeout rÃ©duit pour les opÃ©rations critiques
		c.Set("timeout", 3*time.Second)

		c.Next()
	}
}

// RequireRole middleware pour vÃ©rifier les rÃ´les utilisateur
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

		// VÃ©rifier si l'utilisateur a l'un des rÃ´les requis
		for _, requiredRole := range requiredRoles {
			if userRole == requiredRole {
				c.Next()
				return
			}
		}

		// Admin a tous les droits
		if userRole == "admin" {
			c.Next()
			return
		}

		logrus.WithFields(logrus.Fields{
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
	}
}

// GameClientValidator valide les clients de jeu authentiques
func GameClientValidator() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientVersion := c.GetHeader("X-Client-Version")
		gameSession := c.GetHeader("X-Game-Session")

		// Validation de la version du client
		if clientVersion == "" {
			logrus.WithFields(logrus.Fields{
				"client_ip":  c.ClientIP(),
				"path":       c.Request.URL.Path,
				"request_id": c.GetHeader("X-Request-ID"),
			}).Warn("Missing client version header")

			c.JSON(http.StatusBadRequest, gin.H{
				"error":      "Client version required",
				"message":    "Please update your game client",
				"request_id": c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		// Validation de la session de jeu (pour certaines routes)
		if c.Request.URL.Path != "/api/v1/auth/login" && c.Request.URL.Path != "/api/v1/auth/register" {
			if gameSession == "" {
				logrus.WithFields(logrus.Fields{
					"client_ip":      c.ClientIP(),
					"client_version": clientVersion,
					"path":           c.Request.URL.Path,
					"request_id":     c.GetHeader("X-Request-ID"),
				}).Warn("Missing game session header")

				c.JSON(http.StatusBadRequest, gin.H{
					"error":      "Game session required",
					"message":    "Invalid game session",
					"request_id": c.GetHeader("X-Request-ID"),
				})
				c.Abort()
				return
			}
		}

		// Ajouter les informations client au contexte
		c.Set("client_version", clientVersion)
		c.Set("game_session", gameSession)

		c.Next()
	}
}

// CircuitBreaker middleware pour gÃ©rer les pannes de service
type CircuitBreakerState int

const (
	CircuitBreakerClosed CircuitBreakerState = iota
	CircuitBreakerOpen
	CircuitBreakerHalfOpen
)

type CircuitBreaker struct {
	state        CircuitBreakerState
	failureCount int
	lastFailTime time.Time
	timeout      time.Duration
	threshold    int
	mutex        sync.RWMutex
}

func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:     CircuitBreakerClosed,
		threshold: threshold,
		timeout:   timeout,
	}
}

func (cb *CircuitBreaker) CanExecute() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	switch cb.state {
	case CircuitBreakerClosed:
		return true
	case CircuitBreakerOpen:
		if time.Since(cb.lastFailTime) > cb.timeout {
			cb.mutex.RUnlock()
			cb.mutex.Lock()
			cb.state = CircuitBreakerHalfOpen
			cb.mutex.Unlock()
			cb.mutex.RLock()
			return true
		}
		return false
	case CircuitBreakerHalfOpen:
		return true
	default:
		return false
	}
}

func (cb *CircuitBreaker) OnSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failureCount = 0
	cb.state = CircuitBreakerClosed
}

func (cb *CircuitBreaker) OnFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failureCount++
	cb.lastFailTime = time.Now()

	if cb.failureCount >= cb.threshold {
		cb.state = CircuitBreakerOpen
	}
}

// CircuitBreakerMiddleware crÃ©e un middleware circuit breaker
func CircuitBreakerMiddleware(serviceName string) gin.HandlerFunc {
	breaker := NewCircuitBreaker(5, 30*time.Second)

	return func(c *gin.Context) {
		if !breaker.CanExecute() {
			logrus.WithFields(logrus.Fields{
				"service":    serviceName,
				"state":      "open",
				"path":       c.Request.URL.Path,
				"request_id": c.GetHeader("X-Request-ID"),
			}).Warn("Circuit breaker open, rejecting request")

			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":      "Service temporarily unavailable",
				"service":    serviceName,
				"message":    "Circuit breaker is open",
				"request_id": c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		c.Next()

		// VÃ©rifier le statut de la rÃ©ponse
		if c.Writer.Status() >= 500 {
			breaker.OnFailure()
		} else {
			breaker.OnSuccess()
		}
	}
}

// SecurityHeaders ajoute les en-tÃªtes de sÃ©curitÃ©
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Protection contre le clickjacking
		c.Header("X-Frame-Options", "DENY")

		// Protection XSS
		c.Header("X-XSS-Protection", "1; mode=block")

		// EmpÃªcher le sniffing MIME
		c.Header("X-Content-Type-Options", "nosniff")

		// Politique de rÃ©fÃ©rent strict
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// CSP pour les API (pas de script inline nÃ©cessaire)
		c.Header("Content-Security-Policy", "default-src 'none'; connect-src 'self'")

		// HSTS (HTTPS only)
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		c.Next()
	}
}

// AntiCheat middleware basique pour dÃ©tecter les comportements suspects
func AntiCheat() gin.HandlerFunc {
	suspiciousIPs := make(map[string]int)
	var mutex sync.RWMutex

	// Nettoyage pÃ©riodique
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				mutex.Lock()
				suspiciousIPs = make(map[string]int) // Reset
				mutex.Unlock()
			}
		}
	}()

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")

		// DÃ©tecter les requÃªtes sans User-Agent (bots basiques)
		if userAgent == "" {
			mutex.Lock()
			suspiciousIPs[clientIP]++
			suspiciousCount := suspiciousIPs[clientIP]
			mutex.Unlock()

			logrus.WithFields(logrus.Fields{
				"client_ip":        clientIP,
				"suspicious_count": suspiciousCount,
				"path":             c.Request.URL.Path,
				"request_id":       c.GetHeader("X-Request-ID"),
			}).Warn("Request without User-Agent detected")

			if suspiciousCount > 10 {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":      "Suspicious activity detected",
					"message":    "Please use a valid game client",
					"request_id": c.GetHeader("X-Request-ID"),
				})
				c.Abort()
				return
			}
		}

		// DÃ©tecter les User-Agents suspects
		suspiciousAgents := []string{"curl", "wget", "python", "bot", "crawler"}
		for _, suspicious := range suspiciousAgents {
			if contains(userAgent, suspicious) {
				logrus.WithFields(logrus.Fields{
					"client_ip":  clientIP,
					"user_agent": userAgent,
					"path":       c.Request.URL.Path,
					"request_id": c.GetHeader("X-Request-ID"),
				}).Warn("Suspicious User-Agent detected")
				break
			}
		}

		c.Next()
	}
}

// Fonctions utilitaires

// extractServiceFromPath extrait le nom du service depuis le chemin de l'URL
func extractServiceFromPath(path string) string {
	// /api/v1/auth/login -> auth
	// /api/v1/player/profile -> player
	// /ws -> websocket

	if path == "/ws" {
		return "websocket"
	}

	parts := strings.Split(path, "/")
	if len(parts) >= 4 && parts[1] == "api" && parts[2] == "v1" {
		return parts[3]
	}

	return "unknown"
}

// contains vÃ©rifie si une chaÃ®ne contient une sous-chaÃ®ne (case insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
