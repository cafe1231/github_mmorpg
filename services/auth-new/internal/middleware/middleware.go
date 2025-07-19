package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"

	"auth/internal/config"
)

// Logger middleware de logging
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logrus.WithFields(logrus.Fields{
			"client_ip":   param.ClientIP,
			"method":      param.Method,
			"path":        param.Path,
			"status_code": param.StatusCode,
			"latency":     param.Latency,
			"user_agent":  param.Request.UserAgent(),
			"request_id":  param.Request.Header.Get("X-Request-ID"),
		}).Info("HTTP Request")
		return ""
	})
}

// Recovery middleware de récupération des panics
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logrus.WithFields(logrus.Fields{
			"panic":      recovered,
			"path":       c.Request.URL.Path,
			"method":     c.Request.Method,
			"request_id": c.GetHeader("X-Request-ID"),
		}).Error("Panic recovered")

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Internal server error",
			"message":    "An unexpected error occurred",
			"request_id": c.GetHeader("X-Request-ID"),
		})
	})
}

// CORS middleware pour les requêtes cross-origin
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Liste des origines autorisées (à configurer selon l'environnement)
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:3001",
			"http://localhost:8080",
			"https://yourdomain.com",
		}

		// Vérifier si l'origine est autorisée
		originAllowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				originAllowed = true
				break
			}
		}

		if originAllowed {
			c.Header("Access-Control-Allow-Origin", origin)
		} else {
			c.Header("Access-Control-Allow-Origin", "*") // Pour le développement
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Request-ID, X-Character-ID")
		c.Header("Access-Control-Expose-Headers", "Content-Length, X-Request-ID")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400") // 24 heures

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RequestID middleware pour générer un ID unique par requête
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)

		logrus.WithField("request_id", requestID).Debug("Request ID assigned")

		c.Next()
	}
}

// SecurityHeaders middleware pour ajouter des en-têtes de sécurité
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Protection contre le clickjacking
		c.Header("X-Frame-Options", "DENY")

		// Protection contre le sniffing MIME
		c.Header("X-Content-Type-Options", "nosniff")

		// Protection XSS
		c.Header("X-XSS-Protection", "1; mode=block")

		// Référer Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy (à adapter selon vos besoins)
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self';")

		// HSTS (uniquement en HTTPS)
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		c.Next()
	}
}

// Rate limiting

// GlobalRateLimit middleware de limitation de taux global
func GlobalRateLimit(config config.RateLimit) gin.HandlerFunc {
	limiter := rate.NewLimiter(rate.Every(config.Window/time.Duration(config.Requests)), config.Burst)

	return func(c *gin.Context) {
		if !limiter.Allow() {
			logrus.WithFields(logrus.Fields{
				"client_ip":  c.ClientIP(),
				"path":       c.Request.URL.Path,
				"method":     c.Request.Method,
				"request_id": c.GetHeader("X-Request-ID"),
			}).Warn("Rate limit exceeded")

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"message":     "Too many requests, please try again later",
				"retry_after": int(config.Window.Seconds()),
				"request_id":  c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// AuthRateLimit middleware de limitation spécifique pour l'authentification
func AuthRateLimit(rateLimitConfig config.RateLimitConfig) gin.HandlerFunc {
	// Limiteurs par type d'opération
	loginLimiter := rate.NewLimiter(
		rate.Every(rateLimitConfig.LoginAttempts.Window/time.Duration(rateLimitConfig.LoginAttempts.Requests)),
		rateLimitConfig.LoginAttempts.Burst,
	)

	registerLimiter := rate.NewLimiter(
		rate.Every(rateLimitConfig.Registration.Window/time.Duration(rateLimitConfig.Registration.Requests)),
		rateLimitConfig.Registration.Burst,
	)

	return func(c *gin.Context) {
		path := c.Request.URL.Path

		var limiter *rate.Limiter
		var operation string

		// Choisir le bon limiteur selon l'endpoint
		if strings.Contains(path, "/login") {
			limiter = loginLimiter
			operation = "login"
		} else if strings.Contains(path, "/register") {
			limiter = registerLimiter
			operation = "registration"
		} else {
			// Pas de limitation pour les autres endpoints auth
			c.Next()
			return
		}

		if !limiter.Allow() {
			logrus.WithFields(logrus.Fields{
				"client_ip":  c.ClientIP(),
				"operation":  operation,
				"path":       path,
				"request_id": c.GetHeader("X-Request-ID"),
			}).Warn("Auth rate limit exceeded")

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":      "Rate limit exceeded",
				"message":    fmt.Sprintf("Too many %s attempts, please try again later", operation),
				"request_id": c.GetHeader("X-Request-ID"),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ServiceAuth middleware pour l'authentification entre services
func ServiceAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Vérifier le token de service dans les en-têtes
		serviceToken := c.GetHeader("X-Service-Token")
		if serviceToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Service token required",
			})
			c.Abort()
			return
		}

		// Vérifier que le token correspond au secret partagé
		// En production, utiliser un système plus sophistiqué
		expectedToken := "service-secret-token" // À configurer via env
		if serviceToken != expectedToken {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid service token",
			})
			c.Abort()
			return
		}

		// Identifier le service appelant
		serviceName := c.GetHeader("X-Service-Name")
		if serviceName != "" {
			c.Set("calling_service", serviceName)
		}

		c.Next()
	}
}

// ValidateContentType middleware pour valider le Content-Type
func ValidateContentType(expectedTypes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" || c.Request.Method == "DELETE" {
			c.Next()
			return
		}

		contentType := c.GetHeader("Content-Type")
		if contentType == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Content-Type header required",
				"message": "Please specify a valid Content-Type",
			})
			c.Abort()
			return
		}

		// Vérifier si le Content-Type est accepté
		for _, expectedType := range expectedTypes {
			if strings.Contains(contentType, expectedType) {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusUnsupportedMediaType, gin.H{
			"error":           "Unsupported Content-Type",
			"message":         "Content-Type not supported",
			"supported_types": expectedTypes,
			"provided_type":   contentType,
		})
		c.Abort()
	}
}

// Timeout middleware pour limiter la durée des requêtes
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Créer un contexte avec timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// Remplacer le contexte de la requête
		c.Request = c.Request.WithContext(ctx)

		// Canal pour la completion
		finished := make(chan struct{})

		go func() {
			c.Next()
			finished <- struct{}{}
		}()

		select {
		case <-finished:
			// Requête terminée dans les temps
			return
		case <-ctx.Done():
			// Timeout atteint
			logrus.WithFields(logrus.Fields{
				"timeout":    timeout,
				"path":       c.Request.URL.Path,
				"method":     c.Request.Method,
				"request_id": c.GetHeader("X-Request-ID"),
			}).Warn("Request timeout")

			c.JSON(http.StatusRequestTimeout, gin.H{
				"error":      "Request timeout",
				"message":    "Request took too long to process",
				"timeout":    timeout.String(),
				"request_id": c.GetHeader("X-Request-ID"),
			})
			c.Abort()
		}
	}
}

// HealthCheck middleware simple pour les load balancers
func HealthCheck(endpoint string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == endpoint {
			c.JSON(http.StatusOK, gin.H{
				"status":    "healthy",
				"timestamp": time.Now().Unix(),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// LogSensitiveData middleware pour logger les données sensibles (debug seulement)
func LogSensitiveData(debug bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if debug && logrus.GetLevel() == logrus.DebugLevel {
			// Logger les en-têtes (sans Authorization)
			headers := make(map[string]string)
			for key, values := range c.Request.Header {
				if key != "Authorization" && key != "Cookie" {
					headers[key] = strings.Join(values, ", ")
				}
			}

			logrus.WithFields(logrus.Fields{
				"method":     c.Request.Method,
				"path":       c.Request.URL.Path,
				"query":      c.Request.URL.RawQuery,
				"headers":    headers,
				"request_id": c.GetHeader("X-Request-ID"),
			}).Debug("Request details")
		}
		c.Next()
	}
}
