package handlers

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"player/internal/config"
	"player/internal/database"
)

var startTime = time.Now()

// HealthChecker interface pour vérifier la santé des composants
type HealthChecker interface {
	HealthCheck() error
}

// HealthHandler gère les endpoints de santé et monitoring
type HealthHandler struct {
	config *config.Config
	db     HealthChecker
}

// NewHealthHandler crée un nouveau handler de santé
func NewHealthHandler(config *config.Config, db HealthChecker) *HealthHandler {
	return &HealthHandler{
		config: config,
		db:     db,
	}
}

// HealthCheck endpoint de santé du service Player
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	checks := make(map[string]interface{})
	status := "healthy"

	// Vérification de la base de données
	if h.db != nil {
		if err := h.db.HealthCheck(); err != nil {
			checks["database"] = map[string]interface{}{
				"status": "unhealthy",
				"error":  err.Error(),
			}
			status = "unhealthy"
		} else {
			checks["database"] = map[string]interface{}{
				"status": "healthy",
			}
		}
	} else {
		checks["database"] = map[string]interface{}{
			"status": "unknown",
			"error":  "database connection not available",
		}
		status = "degraded"
	}

	// Informations système
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	systemInfo := map[string]interface{}{
		"goroutines":   runtime.NumGoroutine(),
		"memory_alloc": bToMb(m.Alloc),
		"memory_total": bToMb(m.TotalAlloc),
		"memory_sys":   bToMb(m.Sys),
		"gc_cycles":    m.NumGC,
	}

	health := map[string]interface{}{
		"status":      status,
		"service":     "player",
		"version":     "1.0.0",
		"timestamp":   time.Now().Unix(),
		"uptime":      time.Since(startTime).Seconds(),
		"environment": h.config.Server.Environment,
		"checks":      checks,
		"system":      systemInfo,
	}

	// Définir le code de statut HTTP selon l'état de santé
	httpStatus := http.StatusOK
	if status == "unhealthy" {
		httpStatus = http.StatusServiceUnavailable
	} else if status == "degraded" {
		httpStatus = http.StatusPartialContent
	}

	c.JSON(httpStatus, health)
}

// DetailedHealthCheck endpoint de santé détaillé (admin)
func (h *HealthHandler) DetailedHealthCheck(c *gin.Context) {
	checks := make(map[string]interface{})
	status := "healthy"

	// Vérification détaillée de la base de données
	if h.db != nil {
		dbStart := time.Now()
		if err := h.db.HealthCheck(); err != nil {
			checks["database"] = map[string]interface{}{
				"status":        "unhealthy",
				"error":         err.Error(),
				"response_time": time.Since(dbStart).Milliseconds(),
			}
			status = "unhealthy"
		} else {
			checks["database"] = map[string]interface{}{
				"status":        "healthy",
				"response_time": time.Since(dbStart).Milliseconds(),
				"connection":    "active",
			}
		}
	}

	// Vérifications de configuration
	configChecks := map[string]interface{}{
		"server_port":       h.config.Server.Port,
		"database_host":     h.config.Database.Host,
		"max_characters":    h.config.Game.MaxCharactersPerPlayer,
		"available_classes": len(h.config.Game.AvailableClasses),
		"available_races":   len(h.config.Game.AvailableRaces),
	}
	checks["configuration"] = configChecks

	// Informations détaillées du système
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	systemInfo := map[string]interface{}{
		"go_version": runtime.Version(),
		"goroutines": runtime.NumGoroutine(),
		"memory": map[string]interface{}{
			"alloc_mb":       bToMb(m.Alloc),
			"total_alloc_mb": bToMb(m.TotalAlloc),
			"sys_mb":         bToMb(m.Sys),
			"heap_alloc_mb":  bToMb(m.HeapAlloc),
			"heap_sys_mb":    bToMb(m.HeapSys),
		},
		"gc": map[string]interface{}{
			"num_gc":      m.NumGC,
			"pause_total": m.PauseTotalNs,
			"last_gc":     time.Unix(0, int64(m.LastGC)).Format(time.RFC3339),
		},
	}

	health := map[string]interface{}{
		"status":      status,
		"service":     "player",
		"version":     "1.0.0",
		"timestamp":   time.Now().Unix(),
		"uptime":      time.Since(startTime).Seconds(),
		"environment": h.config.Server.Environment,
		"checks":      checks,
		"system":      systemInfo,
		"dependencies": map[string]interface{}{
			"auth_service": h.config.Auth.ServiceURL,
		},
	}

	httpStatus := http.StatusOK
	if status == "unhealthy" {
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, health)
}

// Readiness endpoint de préparation (Kubernetes)
func (h *HealthHandler) Readiness(c *gin.Context) {
	ready := true
	checks := make(map[string]interface{})

	// Vérifier que la base de données est prête
	if h.db != nil {
		if err := h.db.HealthCheck(); err != nil {
			ready = false
			checks["database"] = "not ready"
		} else {
			checks["database"] = "ready"
		}
	}

	// Vérifier d'autres dépendances critiques si nécessaire
	// Par example, connection au service Auth, etc.

	status := "ready"
	if !ready {
		status = "not ready"
	}

	readiness := map[string]interface{}{
		"status":    status,
		"service":   "player",
		"timestamp": time.Now().Unix(),
		"checks":    checks,
	}

	httpStatus := http.StatusOK
	if !ready {
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, readiness)
}

// Liveness endpoint de vivacité (Kubernetes)
func (h *HealthHandler) Liveness(c *gin.Context) {
	// Test simple pour vérifier que le service répond
	// En général, on vérifie que le processus est vivant
	// et qu'il n'est pas dans un état de deadlock

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Vérification simple: si le service utilize trop de mémoire, il pourrait être en problème
	maxMemoryMB := uint64(1024) // 1GB limite par example
	currentMemoryMB := bToMb(m.Alloc)

	alive := true
	status := "alive"

	if currentMemoryMB > maxMemoryMB {
		alive = false
		status = "memory_limit_exceeded"
	}

	liveness := map[string]interface{}{
		"status":     status,
		"service":    "player",
		"timestamp":  time.Now().Unix(),
		"uptime":     time.Since(startTime).Seconds(),
		"memory_mb":  currentMemoryMB,
		"goroutines": runtime.NumGoroutine(),
	}

	httpStatus := http.StatusOK
	if !alive {
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, liveness)
}

// Metrics endpoint pour Prometheus
func (h *HealthHandler) Metrics(c *gin.Context) {
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}

// Status endpoint simple pour load balancer
func (h *HealthHandler) Status(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "player",
		"time":    time.Now().Unix(),
	})
}

// Info endpoint avec informations sur le service
func (h *HealthHandler) Info(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	info := map[string]interface{}{
		"service": map[string]interface{}{
			"name":        "player",
			"version":     "1.0.0",
			"environment": h.config.Server.Environment,
			"port":        h.config.Server.Port,
			"started_at":  startTime.Format(time.RFC3339),
			"uptime":      time.Since(startTime).Seconds(),
		},
		"game_config": map[string]interface{}{
			"max_characters_per_player": h.config.Game.MaxCharactersPerPlayer,
			"max_level":                 h.config.Game.MaxLevel,
			"starting_level":            h.config.Game.StartingLevel,
			"available_classes":         h.config.Game.AvailableClasses,
			"available_races":           h.config.Game.AvailableRaces,
		},
		"runtime": map[string]interface{}{
			"go_version": runtime.Version(),
			"num_cpu":    runtime.NumCPU(),
			"goroutines": runtime.NumGoroutine(),
			"memory_mb":  bToMb(m.Alloc),
		},
	}

	c.JSON(http.StatusOK, info)
}

// Version endpoint
func (h *HealthHandler) Version(c *gin.Context) {
	version := map[string]interface{}{
		"service":     "player",
		"version":     "1.0.0",
		"build_time":  startTime.Format(time.RFC3339),
		"go_version":  runtime.Version(),
		"environment": h.config.Server.Environment,
	}

	c.JSON(http.StatusOK, version)
}

// Ping endpoint simple
func (h *HealthHandler) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
		"time":    time.Now().Unix(),
	})
}

// ListRoutes affiche toutes les routes disponibles (debug)
func (h *HealthHandler) ListRoutes(router *gin.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.config.Server.Environment == "production" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Debug endpoints disabled in production"})
			return
		}

		routes := []gin.H{}

		for _, route := range router.Routes() {
			routes = append(routes, gin.H{
				"method": route.Method,
				"path":   route.Path,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"service": "player",
			"routes":  routes,
			"total":   len(routes),
		})
	}
}

// ShowConfig affiche la configuration (debug, sans secrets)
func (h *HealthHandler) ShowConfig(c *gin.Context) {
	if h.config.Server.Environment == "production" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Debug endpoints disabled in production"})
		return
	}

	safeConfig := map[string]interface{}{
		"server": map[string]interface{}{
			"port":        h.config.Server.Port,
			"host":        h.config.Server.Host,
			"environment": h.config.Server.Environment,
			"debug":       h.config.Server.Debug,
		},
		"database": map[string]interface{}{
			"host":           h.config.Database.Host,
			"port":           h.config.Database.Port,
			"name":           h.config.Database.Name,
			"max_open_conns": h.config.Database.MaxOpenConns,
			"max_idle_conns": h.config.Database.MaxIdleConns,
		},
		"game":       h.config.Game,
		"rate_limit": h.config.RateLimit,
		"monitoring": h.config.Monitoring,
	}

	c.JSON(http.StatusOK, safeConfig)
}

// DatabaseStats affiche les statistiques de la base de données (debug)
func (h *HealthHandler) DatabaseStats(c *gin.Context) {
	if h.config.Server.Environment == "production" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Debug endpoints disabled in production"})
		return
	}

	// Si le db implémente une interface pour les stats
	if db, ok := h.db.(*database.DB); ok {
		stats := db.Stats()

		dbStats := map[string]interface{}{
			"max_open_connections": stats.MaxOpenConnections,
			"open_connections":     stats.OpenConnections,
			"in_use":               stats.InUse,
			"idle":                 stats.Idle,
			"wait_count":           stats.WaitCount,
			"wait_duration":        stats.WaitDuration.Milliseconds(),
			"max_idle_closed":      stats.MaxIdleClosed,
			"max_idle_time_closed": stats.MaxIdleTimeClosed,
			"max_lifetime_closed":  stats.MaxLifetimeClosed,
		}

		c.JSON(http.StatusOK, gin.H{
			"database_stats": dbStats,
			"timestamp":      time.Now().Unix(),
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "Database statistics not available",
		})
	}
}

// MemoryProfile endpoint pour profiling mémoire (debug)
func (h *HealthHandler) MemoryProfile(c *gin.Context) {
	if h.config.Server.Environment == "production" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Debug endpoints disabled in production"})
		return
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Force un garbage collection pour avoir des stats plus précises
	runtime.GC()
	runtime.ReadMemStats(&m)

	memProfile := map[string]interface{}{
		"general": map[string]interface{}{
			"alloc_mb":       bToMb(m.Alloc),
			"total_alloc_mb": bToMb(m.TotalAlloc),
			"sys_mb":         bToMb(m.Sys),
			"num_gc":         m.NumGC,
		},
		"heap": map[string]interface{}{
			"heap_alloc_mb":    bToMb(m.HeapAlloc),
			"heap_sys_mb":      bToMb(m.HeapSys),
			"heap_idle_mb":     bToMb(m.HeapIdle),
			"heap_inuse_mb":    bToMb(m.HeapInuse),
			"heap_released_mb": bToMb(m.HeapReleased),
			"heap_objects":     m.HeapObjects,
		},
		"stack": map[string]interface{}{
			"stack_inuse_mb": bToMb(m.StackInuse),
			"stack_sys_mb":   bToMb(m.StackSys),
		},
		"gc": map[string]interface{}{
			"next_gc_mb":     bToMb(m.NextGC),
			"last_gc":        time.Unix(0, int64(m.LastGC)).Format(time.RFC3339),
			"pause_total_ms": m.PauseTotalNs / 1000000,
			"num_forced_gc":  m.NumForcedGC,
		},
	}

	c.JSON(http.StatusOK, memProfile)
}

// GCStats endpoint pour les statistiques du garbage collector
func (h *HealthHandler) GCStats(c *gin.Context) {
	if h.config.Server.Environment == "production" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Debug endpoints disabled in production"})
		return
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	gcInfo := map[string]interface{}{
		"num_gc":          m.NumGC,
		"num_forced_gc":   m.NumForcedGC,
		"gc_cpu_fraction": m.GCCPUFraction,
		"pause_total_ns":  m.PauseTotalNs,
		"pause_ns":        m.PauseNs,
		"pause_end":       m.PauseEnd,
		"last_gc":         time.Unix(0, int64(m.LastGC)).Format(time.RFC3339),
		"next_gc":         m.NextGC,
		"gc_sys":          bToMb(m.GCSys),
	}

	c.JSON(http.StatusOK, gcInfo)
}

// bToMb convertit les bytes en megabytes
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

