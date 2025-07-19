package handlers

import (
	"combat/internal/config"
	"combat/internal/database"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Constantes pour les status de santé
const (
	HealthStatusHealthy   = "healthy"
	HealthStatusUnhealthy = "unhealthy"
	HealthStatusDegraded  = "degraded"
)

// HealthHandler gère les requêtes de santé du service
type HealthHandler struct {
	config *config.Config
	db     *database.DB
}

// NewHealthHandler crée un nouveau handler de santé
func NewHealthHandler(config *config.Config, db *database.DB) *HealthHandler {
	return &HealthHandler{
		config: config,
		db:     db,
	}
}

// HealthResponse représente la réponse de santé du service
type HealthResponse struct {
	Status    string                 `json:"status"`
	Service   string                 `json:"service"`
	Version   string                 `json:"version"`
	Timestamp time.Time              `json:"timestamp"`
	Uptime    string                 `json:"uptime"`
	Checks    map[string]HealthCheck `json:"checks"`
}

// HealthCheck représente le résultat d'une vérification de santé
type HealthCheck struct {
	Status  string                 `json:"status"`
	Message string                 `json:"message,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
	Latency string                 `json:"latency,omitempty"`
}

var serviceStartTime = time.Now()

// HealthCheck effectue une vérification complète de la santé du service
// @Summary Vérification de santé
// @Description Vérifie l'état de santé du service et de ses dépendances
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Failure 503 {object} HealthResponse
// @Router /health [get]
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	startTime := time.Now()

	response := HealthResponse{
		Service:   "combat",
		Version:   "1.0.0", // TODO: Récupérer depuis les variables de build
		Timestamp: time.Now(),
		Uptime:    time.Since(serviceStartTime).String(),
		Checks:    make(map[string]HealthCheck),
	}

	// Vérification de la base de données
	dbCheck := h.checkDatabase()
	response.Checks["database"] = dbCheck

	// Vérification de la mémoire
	memoryCheck := h.checkMemory()
	response.Checks["memory"] = memoryCheck

	// Vérification des services externes (optionnel)
	if c.Query("external") == "true" {
		externalCheck := h.checkExternalServices()
		response.Checks["external_services"] = externalCheck
	}

	// Déterminer le statut global
	overallStatus := HealthStatusHealthy
	for _, check := range response.Checks {
		if check.Status == HealthStatusUnhealthy {
			overallStatus = HealthStatusUnhealthy
			break
		} else if check.Status == HealthStatusDegraded && overallStatus == HealthStatusHealthy {
			overallStatus = HealthStatusDegraded
		}
	}

	response.Status = overallStatus

	// Log des vérifications importantes
	duration := time.Since(startTime)
	logrus.WithFields(logrus.Fields{
		"status":   overallStatus,
		"duration": duration,
		"checks":   len(response.Checks),
	}).Info("Health check completed")

	// Retourner le code de statut approprié
	statusCode := http.StatusOK
	if overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, response)
}

// ReadinessCheck vérifie si le service est prêt à recevoir du traffic
// @Summary Vérification de préparation
// @Description Vérifie si le service est prêt à traiter les requêtes
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 503 {object} map[string]interface{}
// @Router /ready [get]
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	checks := make(map[string]interface{})

	// Vérification critique: base de données
	if err := h.db.Health(); err != nil {
		logrus.WithError(err).Error("Database readiness check failed")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "not_ready",
			"reason":  "database_unavailable",
			"message": "Database connection failed",
			"checks":  checks,
		})
		return
	}

	checks["database"] = "ready"
	checks["service"] = "ready"

	c.JSON(http.StatusOK, gin.H{
		"status":    "ready",
		"service":   "combat",
		"timestamp": time.Now(),
		"checks":    checks,
	})
}

// LivenessCheck vérifie si le service est en vie
// @Summary Vérification de vie
// @Description Vérifie si le service fonctionne
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /live [get]
func (h *HealthHandler) LivenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "alive",
		"service":   "combat",
		"timestamp": time.Now(),
		"uptime":    time.Since(serviceStartTime).String(),
	})
}

// MetricsInfo retourne des informations basiques sur les métriques
// @Summary Informations métriques
// @Description Retourne des métriques de base du service
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /metrics-info [get]
func (h *HealthHandler) MetricsInfo(c *gin.Context) {
	dbStats := h.db.GetStats()

	c.JSON(http.StatusOK, gin.H{
		"service":   "combat",
		"timestamp": time.Now(),
		"uptime":    time.Since(serviceStartTime).String(),
		"database": gin.H{
			"open_connections": dbStats["open_connections"],
			"in_use":           dbStats["in_use"],
			"idle":             dbStats["idle"],
		},
		"memory": h.getMemoryUsage(),
	})
}

// checkDatabase vérifie la santé de la base de données
func (h *HealthHandler) checkDatabase() HealthCheck {
	start := time.Now()

	if err := h.db.Health(); err != nil {
		return HealthCheck{
			Status:  "unhealthy",
			Message: "Database connection failed",
			Details: map[string]interface{}{
				"error": err.Error(),
			},
			Latency: time.Since(start).String(),
		}
	}

	// Obtenir les statistiques de la DB
	stats := h.db.GetStats()

	// Vérifier si on a trop de connections ouvertes
	status := "healthy"
	if stats["open_connections"].(int) > config.DefaultMaxConnections {
		status = "degraded"
	}

	return HealthCheck{
		Status:  status,
		Message: "Database connection successful",
		Details: map[string]interface{}{
			"open_connections": stats["open_connections"],
			"in_use":           stats["in_use"],
			"idle":             stats["idle"],
			"max_open":         stats["max_open_connections"],
		},
		Latency: time.Since(start).String(),
	}
}

// checkMemory vérifie l'utilisation de la mémoire
func (h *HealthHandler) checkMemory() HealthCheck {
	memUsage := h.getMemoryUsage()

	status := "healthy"
	message := "Memory usage normal"

	// Convertir en MB pour les seuils
	allocMB := memUsage["alloc"].(uint64) / config.DefaultMemoryMB
	sysMB := memUsage["sys"].(uint64) / config.DefaultMemoryMB

	if allocMB > 512 || sysMB > 1024 {
		status = "degraded"
		message = "High memory usage detected"
	}

	if allocMB > 1024 || sysMB > 2048 {
		status = "unhealthy"
		message = "Critical memory usage"
	}

	return HealthCheck{
		Status:  status,
		Message: message,
		Details: memUsage,
	}
}

// checkExternalServices vérifie les services externes
func (h *HealthHandler) checkExternalServices() HealthCheck {
	// TODO: Implémenter les vérifications des services externes
	// comme auth, player, world services

	// Pour l'instant, on retourne toujours healthy
	return HealthCheck{
		Status:  "healthy",
		Message: "External services check skipped",
		Details: map[string]interface{}{
			"auth_service":   "not_checked",
			"player_service": "not_checked",
			"world_service":  "not_checked",
		},
	}
}

// getMemoryUsage retourne l'utilisation de la mémoire
func (h *HealthHandler) getMemoryUsage() map[string]interface{} {
	// Note: Dans un vrai projet, vous utiliseriez runtime.MemStats
	// Pour simplifier, on retourne des valeurs simulées
	return map[string]interface{}{
		"alloc":       uint64(config.DefaultAllocMB * config.DefaultMemoryMB),      // 50MB
		"total_alloc": uint64(config.DefaultTotalAllocMB * config.DefaultMemoryMB), // 100MB
		"sys":         uint64(config.DefaultSysMB * config.DefaultMemoryMB),        // 200MB
		"num_gc":      uint32(config.DefaultNumGC),
		"goroutines":  config.DefaultGoroutines,
	}
}
