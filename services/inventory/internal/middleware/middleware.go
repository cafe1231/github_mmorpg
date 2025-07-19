package middleware

import (
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Default value constants
const (
	DefaultCORSMaxAge = 12 * time.Hour
	InternalErrorCode = 500
)

// CORS returns a CORS middleware
func CORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // In production, specify allowed origins
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
		MaxAge:           DefaultCORSMaxAge,
	})
}

// RequestID adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			c.String(InternalErrorCode, "Internal Server Error: %s", err)
		}
		c.AbortWithStatus(InternalErrorCode)
	})
}

// Logger returns a gin logger middleware
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] %q %d %s %q %q\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method+" "+param.Path+" "+param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}
