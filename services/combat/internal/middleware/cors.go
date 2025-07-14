// internal/middleware/cors.go
package middleware

import (
"net/http"
"github.com/gin-gonic/gin"
)

// CORS middleware pour les requêtes cross-origin
func CORS() gin.HandlerFunc {
return func(c *gin.Context) {
c.Header("Access-Control-Allow-Origin", "*")
c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Character-ID, X-Request-ID")
c.Header("Access-Control-Expose-Headers", "Content-Length, X-Request-ID")
c.Header("Access-Control-Allow-Credentials", "true")

if c.Request.Method == "OPTIONS" {
c.AbortWithStatus(http.StatusOK)
return
}

c.Next()
}
}
