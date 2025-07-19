package models

import (
	"time"
)

// Constantes pour les statuts de santé
const (
	HealthStatusHealthy   = "healthy"
	HealthStatusUnhealthy = "unhealthy"
)

// HealthResponse représente la réponse de santé du service
type HealthResponse struct {
	Status    string                 `json:"status"`
	Service   string                 `json:"service"`
	Version   string                 `json:"version"`
	Timestamp time.Time              `json:"timestamp"`
	Uptime    string                 `json:"uptime"`
	Checks    map[string]interface{} `json:"checks"`
}

// HealthCheck représente un contrôle de santé individual
type HealthCheck struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Details interface{} `json:"details,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// DatabaseHealth représente la santé de la base de données
type DatabaseHealth struct {
	Status         string        `json:"status"`
	ResponseTime   time.Duration `json:"response_time"`
	Connections    int           `json:"connections"`
	MaxConnections int           `json:"max_connections"`
	Error          string        `json:"error,omitempty"`
}

// RedisHealth représente la santé de Redis
type RedisHealth struct {
	Status       string        `json:"status"`
	ResponseTime time.Duration `json:"response_time"`
	Connected    bool          `json:"connected"`
	Version      string        `json:"version,omitempty"`
	Memory       string        `json:"memory,omitempty"`
	Error        string        `json:"error,omitempty"`
}

// SystemHealth représente la santé du système
type SystemHealth struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskUsage   float64 `json:"disk_usage"`
	Goroutines  int     `json:"goroutines"`
	GCCycles    uint32  `json:"gc_cycles"`
}

// CombatHealth représente la santé spécifique au service combat
type CombatHealth struct {
	ActiveCombats       int           `json:"active_combats"`
	QueuedActions       int           `json:"queued_actions"`
	PendingChallenges   int           `json:"pending_challenges"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	ErrorRate           float64       `json:"error_rate"`
	LastCleanup         time.Time     `json:"last_cleanup"`
}

// MetricsResponse représente la réponse des métriques
type MetricsResponse struct {
	Service     string                 `json:"service"`
	Timestamp   time.Time              `json:"timestamp"`
	Metrics     map[string]interface{} `json:"metrics"`
	Performance map[string]interface{} `json:"performance"`
	Combat      *CombatMetrics         `json:"combat,omitempty"`
}

// CombatMetrics représente les métriques spécifiques au combat
type CombatMetrics struct {
	// Métriques des combats
	TotalCombats    int64   `json:"total_combats"`
	ActiveCombats   int     `json:"active_combats"`
	CombatsPerHour  float64 `json:"combats_per_hour"`
	AverageDuration float64 `json:"average_duration_seconds"`

	// Métriques des actions
	TotalActions      int64   `json:"total_actions"`
	ActionsPerSecond  float64 `json:"actions_per_second"`
	FailedActions     int64   `json:"failed_actions"`
	ActionFailureRate float64 `json:"action_failure_rate"`

	// Métriques PvP
	PvPChallenges      int64   `json:"pvp_challenges"`
	PvPMatches         int64   `json:"pvp_matches"`
	AveragePvPDuration float64 `json:"average_pvp_duration_seconds"`

	// Métriques de performance
	ResponseTime      PerformanceMetrics `json:"response_time"`
	ThroughputMetrics ThroughputMetrics  `json:"throughput"`
	ErrorMetrics      ErrorMetrics       `json:"errors"`

	// Cache et base de données
	CacheHitRate    float64 `json:"cache_hit_rate"`
	DatabaseQueries int64   `json:"database_queries"`
	SlowQueries     int64   `json:"slow_queries"`
}

// PerformanceMetrics représente les métriques de performance
type PerformanceMetrics struct {
	Min float64 `json:"min_ms"`
	Max float64 `json:"max_ms"`
	Avg float64 `json:"avg_ms"`
	P50 float64 `json:"p50_ms"`
	P90 float64 `json:"p90_ms"`
	P95 float64 `json:"p95_ms"`
	P99 float64 `json:"p99_ms"`
}

// ThroughputMetrics représente les métriques de débit
type ThroughputMetrics struct {
	RequestsPerSecond     float64 `json:"requests_per_second"`
	RequestsPerMinute     float64 `json:"requests_per_minute"`
	RequestsPerHour       float64 `json:"requests_per_hour"`
	PeakRequestsPerSecond float64 `json:"peak_requests_per_second"`
}

// ErrorMetrics représente les métriques d'erreurs
type ErrorMetrics struct {
	TotalErrors      int64   `json:"total_errors"`
	ErrorRate        float64 `json:"error_rate_percent"`
	ValidationErrors int64   `json:"validation_errors"`
	DatabaseErrors   int64   `json:"database_errors"`
	TimeoutErrors    int64   `json:"timeout_errors"`
	AuthErrors       int64   `json:"auth_errors"`
	AntiCheatFlags   int64   `json:"anti_cheat_flags"`
}

// ServiceStatus représente le statut d'un service externe
type ServiceStatus struct {
	Name         string        `json:"name"`
	Status       string        `json:"status"`
	URL          string        `json:"url"`
	ResponseTime time.Duration `json:"response_time"`
	LastCheck    time.Time     `json:"last_check"`
	Error        string        `json:"error,omitempty"`
}

// MemoryStats représente les statistiques mémoire
type MemoryStats struct {
	Alloc        uint64 `json:"alloc"`
	TotalAlloc   uint64 `json:"total_alloc"`
	Sys          uint64 `json:"sys"`
	Lookups      uint64 `json:"lookups"`
	Mallocs      uint64 `json:"mallocs"`
	Frees        uint64 `json:"frees"`
	HeapAlloc    uint64 `json:"heap_alloc"`
	HeapSys      uint64 `json:"heap_sys"`
	HeapIdle     uint64 `json:"heap_idle"`
	HeapInuse    uint64 `json:"heap_inuse"`
	HeapReleased uint64 `json:"heap_released"`
	HeapObjects  uint64 `json:"heap_objects"`
	StackInuse   uint64 `json:"stack_inuse"`
	StackSys     uint64 `json:"stack_sys"`
	NumGC        uint32 `json:"num_gc"`
	NumForcedGC  uint32 `json:"num_forced_gc"`
}

// ReadinessProbe représente une sonde de préparation
type ReadinessProbe struct {
	Database   bool   `json:"database"`
	Redis      bool   `json:"redis"`
	Services   bool   `json:"external_services"`
	Migrations bool   `json:"migrations"`
	Ready      bool   `json:"ready"`
	Message    string `json:"message,omitempty"`
}

// LivenessProbe représente une sonde de vie
type LivenessProbe struct {
	Healthy      bool   `json:"healthy"`
	Uptime       string `json:"uptime"`
	MemoryOK     bool   `json:"memory_ok"`
	GoroutinesOK bool   `json:"goroutines_ok"`
	Message      string `json:"message,omitempty"`
}

// DebugInfo représente les informations de debug
type DebugInfo struct {
	BuildInfo    BuildInfo              `json:"build_info"`
	Runtime      RuntimeInfo            `json:"runtime"`
	Config       map[string]interface{} `json:"config"`
	Environment  map[string]string      `json:"environment"`
	Dependencies []DependencyInfo       `json:"dependencies"`
}

// BuildInfo représente les informations de build
type BuildInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Branch    string `json:"branch"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
}

// RuntimeInfo représente les informations runtime
type RuntimeInfo struct {
	OS         string        `json:"os"`
	Arch       string        `json:"arch"`
	Goroutines int           `json:"goroutines"`
	CGOCalls   int64         `json:"cgo_calls"`
	StartTime  time.Time     `json:"start_time"`
	Uptime     time.Duration `json:"uptime"`
}

// DependencyInfo représente les informations d'une dépendance
type DependencyInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Status  string `json:"status"`
	Health  string `json:"health,omitempty"`
}

// GetHealthStatus retourne le statut de santé global
func GetHealthStatus(checks map[string]*HealthCheck) string {
	for _, check := range checks {
		if check.Status != HealthStatusHealthy {
			return HealthStatusUnhealthy
		}
	}
	return HealthStatusHealthy
}

// CreateHealthCheck crée un contrôle de santé
func CreateHealthCheck(status, message string, details interface{}, err error) *HealthCheck {
	check := &HealthCheck{
		Status:  status,
		Message: message,
		Details: details,
	}

	if err != nil {
		check.Error = err.Error()
		if status == "" {
			check.Status = "unhealthy"
		}
	} else if status == "" {
		check.Status = HealthStatusHealthy
	}

	return check
}

// IsHealthy vérifie si le service est en bonne santé
func (h *HealthResponse) IsHealthy() bool {
	return h.Status == "healthy"
}

// AddCheck ajoute un contrôle de santé
func (h *HealthResponse) AddCheck(name string, check *HealthCheck) {
	if h.Checks == nil {
		h.Checks = make(map[string]interface{})
	}
	h.Checks[name] = check

	// Mettre à jour le statut global
	if check.Status != "healthy" {
		h.Status = "unhealthy"
	}
}
