package models

import "time"

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

// SystemHealth représente la santé du système
type SystemHealth struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskUsage   float64 `json:"disk_usage"`
	Goroutines  int     `json:"goroutines"`
	GCCycles    uint32  `json:"gc_cycles"`
}

// MetricsResponse représente la réponse des métriques
type MetricsResponse struct {
	Service     string                 `json:"service"`
	Timestamp   time.Time              `json:"timestamp"`
	Metrics     map[string]interface{} `json:"metrics"`
	Performance map[string]interface{} `json:"performance"`
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
