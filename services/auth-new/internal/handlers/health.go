package handlers

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"auth/internal/config"
	"auth/internal/models"
)

var startTime = time.Now()

// HealthChecker interface pour vérifier la santé des composants
type HealthChecker interface {
	Ping() error
}

// HealthHandler gère les endpoints de santé et monitoring
type HealthHandler struct {
	config *config.Config
	db     *sqlx.DB
}

// NewHealthHandler crée un nouveau handler de santé
func NewHealthHandler(config *config.Config, db *sqlx.DB) *HealthHandler {
	return &HealthHandler{
		config: config,
		db:     db,
	}
}

// HealthCheck endpoint de santé du service Auth
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	checks := make(map[string]interface{})
	status := "healthy"

	// Vérification de la base de données
	if h.db != nil {
		start := time.Now()
		if err := h.db.Ping(); err != nil {
			checks["database"] = models.HealthCheck{
				Status:  "unhealthy",
				Message: "Database connection failed",
				Error:   err.Error(),
			}
			status = "unhealthy"
		} else {
			responseTime := time.Since(start)
			checks["database"] = models.HealthCheck{
				Status:  "healthy",
				Message: "Database connection successful",
				Details: map[string]interface{}{
					"response_time": responseTime.String(),
				},
			}
		}
	} else {
		checks["database"] = models.HealthCheck{
			Status:  "unknown",
			Message: "Database connection not available",
		}
		status = "degraded"
	}

	// Informations système
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	checks["system"] = models.HealthCheck{
		Status:  "healthy",
		Message: "System resources normal",
		Details: models.SystemHealth{
			Goroutines: runtime.NumGoroutine(),
			GCCycles:   m.NumGC,
			MemoryUsage: float64(m.Alloc) / 1024 / 1024, // MB
		},
	}

	uptime := time.Since(startTime)

	response := models.HealthResponse{
		Status:    status,
		Service:   "auth",
		Version:   "1.0.0",
		Timestamp: time.Now(),
		Uptime:    uptime.String(),
		Checks:    checks,
	}

	statusCode := http.StatusOK
	if status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	} else if status == "degraded" {
		statusCode = http.StatusPartialContent
	}

	c.JSON(statusCode, response)
}

// ReadinessCheck vérifie si le service est prêt à recevoir des requêtes
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	checks := make(map[string]interface{})
	ready := true

	// Vérifier la base de données
	if h.db != nil {
		if err := h.db.Ping(); err != nil {
			checks["database"] = models.HealthCheck{
				Status: "not_ready",
				Error:  err.Error(),
			}
			ready = false
		} else {
			checks["database"] = models.HealthCheck{
				Status: "ready",
			}
		}
	}

	response := models.HealthResponse{
		Status:    map[bool]string{true: "ready", false: "not_ready"}[ready],
		Service:   "auth",
		Timestamp: time.Now(),
		Checks:    checks,
	}

	statusCode := http.StatusOK
	if !ready {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, response)
}

// LivenessCheck vérifie si le service est vivant
func (h *HealthHandler) LivenessCheck(c *gin.Context) {
	response := models.HealthResponse{
		Status:    "alive",
		Service:   "auth",
		Timestamp: time.Now(),
		Uptime:    time.Since(startTime).String(),
		Checks: map[string]interface{}{
			"service": models.HealthCheck{
				Status:  "alive",
				Message: "Service is running",
			},
		},
	}

	c.JSON(http.StatusOK, response)
}

// Metrics endpoint pour les métriques Prometheus
func (h *HealthHandler) Metrics(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metrics := models.MetricsResponse{
		Service:   "auth",
		Timestamp: time.Now(),
		Metrics: map[string]interface{}{
			"uptime_seconds":     time.Since(startTime).Seconds(),
			"goroutines_total":   runtime.NumGoroutine(),
			"memory_alloc_bytes": m.Alloc,
			"memory_sys_bytes":   m.Sys,
			"gc_cycles_total":    m.NumGC,
		},
		Performance: map[string]interface{}{
			"gc_pause_ns": m.PauseNs[(m.NumGC+255)%256],
		},
	}

	c.JSON(http.StatusOK, metrics)
}

// bToMb converts bytes to megabytes
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

// Stats endpoint pour les statistiques détaillées
func (h *HealthHandler) Stats(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	stats := gin.H{
		"service": "auth",
		"version": "1.0.0",
		"uptime":  time.Since(startTime).String(),
		"memory": gin.H{
			"alloc":       bToMb(m.Alloc),
			"total_alloc": bToMb(m.TotalAlloc),
			"sys":         bToMb(m.Sys),
			"num_gc":      m.NumGC,
		},
		"runtime": gin.H{
			"goroutines": runtime.NumGoroutine(),
			"cpus":       runtime.NumCPU(),
			"go_version": runtime.Version(),
		},
	}

	if h.db != nil {
		if err := h.db.Ping(); err != nil {
			stats["database"] = gin.H{
				"status": "error",
				"error":  err.Error(),
			}
		} else {
			stats["database"] = gin.H{
				"status": "connected",
			}
		}
	}

	c.JSON(http.StatusOK, stats)
}

// Debug endpoint pour les informations de debug
func (h *HealthHandler) Debug(c *gin.Context) {
	if h.config.Server.Environment == "production" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Debug endpoint disabled in production",
		})
		return
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	debug := gin.H{
		"service":     "auth",
		"environment": h.config.Server.Environment,
		"timestamp":   time.Now(),
		"memory": gin.H{
			"alloc":        m.Alloc,
			"total_alloc":  m.TotalAlloc,
			"sys":          m.Sys,
			"lookups":      m.Lookups,
			"mallocs":      m.Mallocs,
			"frees":        m.Frees,
			"heap_alloc":   m.HeapAlloc,
			"heap_sys":     m.HeapSys,
			"heap_idle":    m.HeapIdle,
			"heap_inuse":   m.HeapInuse,
			"heap_objects": m.HeapObjects,
			"num_gc":       m.NumGC,
		},
		"runtime": gin.H{
			"goroutines": runtime.NumGoroutine(),
			"cpus":       runtime.NumCPU(),
			"go_version": runtime.Version(),
		},
	}

	c.JSON(http.StatusOK, debug)
}

// ConfigHandler affiche la configuration (sans secrets)
func (h *HealthHandler) ConfigHandler(c *gin.Context) {
	if h.config.Server.Environment == "production" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Config endpoint disabled in production",
		})
		return
	}

	safeConfig := gin.H{
		"server": gin.H{
			"port":        h.config.Server.Port,
			"host":        h.config.Server.Host,
			"environment": h.config.Server.Environment,
		},
		"database": gin.H{
			"host":            h.config.Database.Host,
			"port":            h.config.Database.Port,
			"name":            h.config.Database.Name,
			"max_open_conns":  h.config.Database.MaxOpenConns,
			"max_idle_conns":  h.config.Database.MaxIdleConns,
		},
		"jwt": gin.H{
			"issuer": h.config.JWT.Issuer,
		},
	}

	c.JSON(http.StatusOK, safeConfig)
}