package middleware

import (
	"bytes"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Constantes pour les méthodes HTTP
const (
	HTTPMethodPOST   = "POST"
	HTTPMethodPUT    = "PUT"
	HTTPMethodDELETE = "DELETE"
	HTTPMethodPATCH  = "PATCH"
)

// LoggingConfig configuration pour le middleware de logging
type LoggingConfig struct {
	SkipPaths      []string `json:"skip_paths"`
	LogRequestBody bool     `json:"log_request_body"`
	LogResponse    bool     `json:"log_response"`
	MaxBodySize    int      `json:"max_body_size"`
}

// responseWriter wrapper pour capturer la réponse
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// StructuredLogging middleware pour un logging structuré avancé
func StructuredLogging(config LoggingConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		// Vérifier si on doit ignorer ce chemin
		for _, skipPath := range config.SkipPaths {
			if strings.HasPrefix(path, skipPath) {
				c.Next()
				return
			}
		}

		// Capturer le corps de la requête si nécessaire
		var requestBody []byte
		if config.LogRequestBody && shouldLogBody(c.Request.Method) {
			if c.Request.Body != nil {
				requestBody, _ = io.ReadAll(c.Request.Body)
				c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
			}
		}

		// Wrapper pour capturer la réponse
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		if config.LogResponse {
			c.Writer = writer
		}

		// Traiter la requête
		c.Next()

		// Calculer la durée
		duration := time.Since(start)

		// Préparer les champs de log
		fields := logrus.Fields{
			"method":     c.Request.Method,
			"path":       path,
			"query":      c.Request.URL.RawQuery,
			"status":     c.Writer.Status(),
			"user_agent": c.Request.UserAgent(),
			"client_ip":  c.ClientIP(),
			"latency":    duration,
			"latency_ms": duration.Milliseconds(),
			"bytes_in":   c.Request.ContentLength,
			"bytes_out":  c.Writer.Size(),
			"request_id": c.GetHeader("X-Request-ID"),
			"referer":    c.Request.Referer(),
			"protocol":   c.Request.Proto,
		}

		// Ajouter les informations utilisateur si disponibles
		if userID := c.GetString("user_id"); userID != "" {
			fields["user_id"] = userID
		}
		if characterID := c.GetString("character_id"); characterID != "" {
			fields["character_id"] = characterID
		}
		if username := c.GetString("username"); username != "" {
			fields["username"] = username
		}
		if role := c.GetString("user_role"); role != "" {
			fields["user_role"] = role
		}

		// Ajouter le corps de la requête si configuré
		if config.LogRequestBody && len(requestBody) > 0 {
			bodyStr := string(requestBody)
			if len(bodyStr) > config.MaxBodySize {
				bodyStr = bodyStr[:config.MaxBodySize] + "...[truncated]"
			}
			fields["request_body"] = bodyStr
		}

		// Ajouter la réponse si configurée
		if config.LogResponse && writer.body.Len() > 0 {
			responseStr := writer.body.String()
			if len(responseStr) > config.MaxBodySize {
				responseStr = responseStr[:config.MaxBodySize] + "...[truncated]"
			}
			fields["response_body"] = responseStr
		}

		// Ajouter les erreurs s'il y en a
		if len(c.Errors) > 0 {
			fields["errors"] = c.Errors.Errors()
		}

		// Déterminer le niveau de log basé sur le code de statut
		statusCode := c.Writer.Status()
		logEntry := logrus.WithFields(fields)

		switch {
		case statusCode >= 500:
			logEntry.Error("Server error")
		case statusCode >= 400:
			logEntry.Warn("Client error")
		case statusCode >= 300:
			logEntry.Info("Redirection")
		default:
			if duration > 1*time.Second {
				logEntry.Warn("Slow request")
			} else {
				logEntry.Info("Request completed")
			}
		}
	}
}

// RequestLogging middleware de logging simple
func RequestLogging() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logrus.WithFields(logrus.Fields{
			"method":     param.Method,
			"path":       param.Path,
			"status":     param.StatusCode,
			"latency":    param.Latency,
			"client_ip":  param.ClientIP,
			"user_agent": param.Request.UserAgent(),
		}).Info("HTTP Request")

		return "" // Retourner une chaîne vide car on utilize logrus
	})
}

// SecurityLogging middleware pour logger les événements de sécurité
func SecurityLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Variables pour détecter les événements suspects
		suspicious := false
		securityEvents := []string{}

		// Vérifier les en-têtes suspects
		userAgent := c.Request.UserAgent()
		if userAgent == "" {
			suspicious = true
			securityEvents = append(securityEvents, "missing_user_agent")
		}

		// Vérifier les tentatives d'injection SQL basiques
		queryParams := c.Request.URL.Query()
		for key, values := range queryParams {
			for _, value := range values {
				if containsSQLInjection(value) {
					suspicious = true
					securityEvents = append(securityEvents, "sql_injection_attempt")
					logrus.WithFields(logrus.Fields{
						"param":      key,
						"value":      value,
						"client_ip":  c.ClientIP(),
						"path":       c.Request.URL.Path,
						"request_id": c.GetHeader("X-Request-ID"),
					}).Warn("SQL injection attempt detected")
				}
			}
		}

		// Vérifier les tentatives XSS basiques
		for key, values := range queryParams {
			for _, value := range values {
				if containsXSS(value) {
					suspicious = true
					securityEvents = append(securityEvents, "xss_attempt")
					logrus.WithFields(logrus.Fields{
						"param":      key,
						"value":      value,
						"client_ip":  c.ClientIP(),
						"path":       c.Request.URL.Path,
						"request_id": c.GetHeader("X-Request-ID"),
					}).Warn("XSS attempt detected")
				}
			}
		}

		// Vérifier les tentatives d'accès à des fichiers système
		if containsPathTraversal(c.Request.URL.Path) {
			suspicious = true
			securityEvents = append(securityEvents, "path_traversal_attempt")
			logrus.WithFields(logrus.Fields{
				"path":       c.Request.URL.Path,
				"client_ip":  c.ClientIP(),
				"request_id": c.GetHeader("X-Request-ID"),
			}).Warn("Path traversal attempt detected")
		}

		// Stocker les informations de sécurité dans le contexte
		if suspicious {
			c.Set("security_suspicious", true)
			c.Set("security_events", securityEvents)
		}

		c.Next()

		// Logger les événements de sécurité après traitement
		if suspicious {
			logrus.WithFields(logrus.Fields{
				"client_ip":       c.ClientIP(),
				"user_agent":      userAgent,
				"path":            c.Request.URL.Path,
				"method":          c.Request.Method,
				"security_events": securityEvents,
				"status_code":     c.Writer.Status(),
				"request_id":      c.GetHeader("X-Request-ID"),
				"user_id":         c.GetString("user_id"),
			}).Warn("Suspicious security activity detected")
		}
	}
}

// ErrorLogging middleware pour logger les erreurs détaillées
func ErrorLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Logger toutes les erreurs qui se sont produites
		for _, err := range c.Errors {
			logrus.WithFields(logrus.Fields{
				"error":      err.Error(),
				"type":       err.Type,
				"meta":       err.Meta,
				"method":     c.Request.Method,
				"path":       c.Request.URL.Path,
				"client_ip":  c.ClientIP(),
				"user_id":    c.GetString("user_id"),
				"request_id": c.GetHeader("X-Request-ID"),
			}).Error("Request error")
		}
	}
}

// AuditLogging middleware pour l'audit des actions importantes
func AuditLogging(auditPaths []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method

		// Vérifier si cette route nécessite un audit
		needsAudit := false
		for _, auditPath := range auditPaths {
			if strings.HasPrefix(path, auditPath) {
				needsAudit = true
				break
			}
		}

		// Ou auditer toutes les opérations de modification
		if method == HTTPMethodPOST || method == HTTPMethodPUT || method == HTTPMethodDELETE || method == HTTPMethodPATCH {
			needsAudit = true
		}

		if needsAudit {
			start := time.Now()

			c.Next()

			duration := time.Since(start)

			logrus.WithFields(logrus.Fields{
				"audit":        true,
				"action":       method + " " + path,
				"user_id":      c.GetString("user_id"),
				"character_id": c.GetString("character_id"),
				"username":     c.GetString("username"),
				"user_role":    c.GetString("user_role"),
				"client_ip":    c.ClientIP(),
				"user_agent":   c.Request.UserAgent(),
				"status_code":  c.Writer.Status(),
				"duration":     duration,
				"request_id":   c.GetHeader("X-Request-ID"),
				"timestamp":    time.Now(),
			}).Info("Audit log entry")
		} else {
			c.Next()
		}
	}
}

// Fonctions utilitaires pour la détection de sécurité

func shouldLogBody(method string) bool {
	return method == HTTPMethodPOST || method == HTTPMethodPUT || method == HTTPMethodPATCH
}

func containsSQLInjection(input string) bool {
	sqlPatterns := []string{
		"'", "\"", ";", "--", "/*", "*/", "xp_", "sp_",
		"union", "select", "insert", "delete", "update",
		"drop", "create", "alter", "exec", "execute",
	}

	lowerInput := strings.ToLower(input)
	for _, pattern := range sqlPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}
	return false
}

func containsXSS(input string) bool {
	xssPatterns := []string{
		"<script", "</script>", "javascript:", "onerror=",
		"onload=", "onclick=", "onmouseover=", "onfocus=",
		"alert(", "document.cookie", "document.write",
	}

	lowerInput := strings.ToLower(input)
	for _, pattern := range xssPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}
	return false
}

func containsPathTraversal(path string) bool {
	traversalPatterns := []string{
		"../", "..\\", "..%2f", "..%5c",
		"%2e%2e%2f", "%2e%2e%5c",
	}

	lowerPath := strings.ToLower(path)
	for _, pattern := range traversalPatterns {
		if strings.Contains(lowerPath, pattern) {
			return true
		}
	}
	return false
}
