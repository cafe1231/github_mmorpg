// internal/monitoring/health.go
package monitoring

import (
"context"
"net/http"
"time"

"github.com/gin-gonic/gin"

"combat/internal/database"
)

// HealthStatus représente l'état de santé du service
type HealthStatus struct {
Status    string            `json:"status"`
Service   string            `json:"service"`
Version   string            `json:"version"`
Timestamp int64             `json:"timestamp"`
Checks    map[string]Check  `json:"checks"`
}

// Check représente une vérification de santé
type Check struct {
Status  string        `json:"status"`
Message string        `json:"message,omitempty"`
Latency time.Duration `json:"latency"`
}

// HealthChecker gère les vérifications de santé
type HealthChecker struct {
db *database.DB
}

// NewHealthChecker crée un nouveau checker de santé
func NewHealthChecker(db *database.DB) *HealthChecker {
return &HealthChecker{
db: db,
}
}

// HealthCheck endpoint de vérification de santé
func (h *HealthChecker) HealthCheck(c *gin.Context) {
status := h.GetHealthStatus()

httpStatus := http.StatusOK
if status.Status != "healthy" {
httpStatus = http.StatusServiceUnavailable
}

c.JSON(httpStatus, status)
}

// GetHealthStatus effectue toutes les vérifications de santé
func (h *HealthChecker) GetHealthStatus() HealthStatus {
checks := make(map[string]Check)
overallStatus := "healthy"

// Vérification de la base de données
dbCheck := h.checkDatabase()
checks["database"] = dbCheck
if dbCheck.Status != "healthy" {
overallStatus = "unhealthy"
}

return HealthStatus{
Status:    overallStatus,
Service:   "combat-service",
Version:   "1.0.0",
Timestamp: time.Now().Unix(),
Checks:    checks,
}
}

// checkDatabase vérifie la santé de la base de données
func (h *HealthChecker) checkDatabase() Check {
start := time.Now()

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := h.db.PingContext(ctx); err != nil {
return Check{
Status:  "unhealthy",
Message: "Database connection failed: " + err.Error(),
Latency: time.Since(start),
}
}

return Check{
Status:  "healthy",
Message: "Database is responsive",
Latency: time.Since(start),
}
}
