// internal/middleware/recovery.go
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Recovery middleware avec logging amélioré
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logrus.WithFields(logrus.Fields{
			"error":      recovered,
			"path":       c.Request.URL.Path,
			"method":     c.Request.Method,
			"client_ip":  c.ClientIP(),
			"request_id": c.GetHeader("X-Request-ID"),
			"service":    "combat",
		}).Error("Panic recovered in combat service")

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Internal server error",
			"request_id": c.GetHeader("X-Request-ID"),
		})
	})
}