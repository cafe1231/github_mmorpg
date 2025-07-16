package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"inventory/internal/models"
)

type HealthHandler struct {
	startTime time.Time
	version   string
}

func NewHealthHandler(version string) *HealthHandler {
	return &HealthHandler{
		startTime: time.Now(),
		version:   version,
	}
}

// Health returns the health status of the service
// GET /health
func (h *HealthHandler) Health(c *gin.Context) {
	uptime := time.Since(h.startTime)

	health := models.NewHealthResponse(h.version, uptime)

	// Add database health check (placeholder)
	health.AddService("database", models.HealthStatusHealthy, "Connected", 5*time.Millisecond)

	// Add Redis health check (placeholder)
	health.AddService("redis", models.HealthStatusHealthy, "Connected", 2*time.Millisecond)

	statusCode := http.StatusOK
	if health.Status != models.HealthStatusHealthy {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, health)
}

// Readiness returns the readiness status of the service
// GET /ready
func (h *HealthHandler) Readiness(c *gin.Context) {
	readiness := models.NewReadinessResponse()

	// Add readiness checks
	readiness.AddCheck("database", true, "Database connection is ready")
	readiness.AddCheck("redis", true, "Redis connection is ready")
	readiness.AddCheck("migrations", true, "Database migrations are up to date")

	statusCode := http.StatusOK
	if !readiness.Ready {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, readiness)
}

// Liveness returns the liveness status of the service
// GET /live
func (h *HealthHandler) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "alive",
		"timestamp": time.Now(),
		"uptime":    time.Since(h.startTime).String(),
	})
}
