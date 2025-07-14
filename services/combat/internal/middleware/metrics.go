package middleware

import (
"time"
"github.com/gin-gonic/gin"
"combat/internal/monitoring"
)

// Metrics middleware pour instrumenter les requêtes
func Metrics(metrics *monitoring.Metrics) gin.HandlerFunc {
if metrics == nil {
// Si pas de metrics, middleware transparent
return func(c *gin.Context) {
c.Next()
}
}

return metrics.Middleware()
}

// MetricsSimple middleware simple sans dépendances
func MetricsSimple() gin.HandlerFunc {
return func(c *gin.Context) {
start := time.Now()
c.Next()
// Log simple de la durée
duration := time.Since(start)
_ = duration // éviter l'erreur de variable non utilisée
}
}
