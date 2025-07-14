// internal/middleware/logger.go
package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Logger middleware personnalisé pour le service Combat
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
			"service":      "combat",
		}).Info("HTTP Request")

		return ""
	})
}