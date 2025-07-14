// internal/handler/health.go
package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"combat/internal/monitoring"
)

// HealthHandler gère les endpoints de santé
type HealthHandler struct {
	healthChecker *monitoring.HealthChecker
}

// NewHealthHandler crée une nouvelle instance du handler health
func NewHealthHandler(hc *monitoring.HealthChecker) *HealthHandler {
	return &HealthHandler{
		healthChecker: hc,
	}
}

// HealthCheck endpoint de vérification de santé
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	if h.healthChecker != nil {
		h.healthChecker.HealthCheck(c)
	} else {
		// Health check basique si pas de checker
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "combat-service",
			"version":   "1.0.0",
			"timestamp": time.Now().Unix(),
		})
	}
}

// ListRoutes affiche toutes les routes (debug)
func (h *HealthHandler) ListRoutes(router *gin.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		routes := []map[string]string{
			{"method": "POST", "path": "/api/v1/combat/start", "description": "Start combat session"},
			{"method": "POST", "path": "/api/v1/combat/join/:sessionId", "description": "Join combat"},
			{"method": "POST", "path": "/api/v1/combat/leave/:sessionId", "description": "Leave combat"},
			{"method": "POST", "path": "/api/v1/combat/action", "description": "Perform combat action"},
			{"method": "GET", "path": "/api/v1/combat/status/:sessionId", "description": "Get combat status"},
			{"method": "GET", "path": "/api/v1/combat/sessions/active", "description": "Get active combats"},
			{"method": "GET", "path": "/api/v1/spells/character/:characterId", "description": "Get character spells"},
			{"method": "POST", "path": "/api/v1/spells/cast", "description": "Cast spell"},
			{"method": "GET", "path": "/api/v1/spells/cooldowns/:characterId", "description": "Get spell cooldowns"},
			{"method": "GET", "path": "/health", "description": "Health check"},
			{"method": "GET", "path": "/ws", "description": "WebSocket connection"},
		}

		c.JSON(http.StatusOK, gin.H{
			"routes":     routes,
			"total":      len(routes),
			"service":    "combat-service",
			"request_id": c.GetHeader("X-Request-ID"),
		})
	}
}

// ShowConfig affiche la configuration (debug, sans secrets)
func (h *HealthHandler) ShowConfig(c *gin.Context) {
	safeConfig := map[string]interface{}{
		"service": map[string]interface{}{
			"name":    "combat-service",
			"version": "1.0.0",
			"port":    8084,
		},
		"features": map[string]interface{}{
			"pvp_enabled":       true,
			"real_time":         true,
			"anti_cheat":        true,
			"spell_system":      true,
			"status_effects":    true,
		},
		"limits": map[string]interface{}{
			"max_participants":    20,
			"max_combat_duration": "30m",
			"actions_per_minute":  180,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"config":     safeConfig,
		"timestamp":  time.Now().Unix(),
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

// ShowMetrics affiche un résumé des métriques (debug)
func (h *HealthHandler) ShowMetrics(c *gin.Context) {
	metrics := map[string]interface{}{
		"active_sessions":     0,
		"total_actions":       0,
		"spells_cast":         0,
		"active_effects":      0,
		"websocket_connections": 0,
		"uptime_seconds":      time.Since(time.Now().Truncate(24*time.Hour)).Seconds(),
	}

	c.JSON(http.StatusOK, gin.H{
		"metrics":    metrics,
		"timestamp":  time.Now().Unix(),
		"request_id": c.GetHeader("X-Request-ID"),
	})
}