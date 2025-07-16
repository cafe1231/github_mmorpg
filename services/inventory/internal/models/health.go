package models

import "time"

type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
)

type HealthResponse struct {
	Status    HealthStatus       `json:"status"`
	Timestamp time.Time          `json:"timestamp"`
	Services  map[string]Service `json:"services"`
	Version   string             `json:"version"`
	Uptime    time.Duration      `json:"uptime"`
}

type Service struct {
	Status       HealthStatus  `json:"status"`
	Message      string        `json:"message,omitempty"`
	LastChecked  time.Time     `json:"last_checked"`
	ResponseTime time.Duration `json:"response_time,omitempty"`
}

type ReadinessResponse struct {
	Ready     bool             `json:"ready"`
	Timestamp time.Time        `json:"timestamp"`
	Checks    map[string]Check `json:"checks"`
}

type Check struct {
	Status  bool   `json:"status"`
	Message string `json:"message,omitempty"`
}

// Helper function to create health response
func NewHealthResponse(version string, uptime time.Duration) *HealthResponse {
	return &HealthResponse{
		Status:    HealthStatusHealthy,
		Timestamp: time.Now(),
		Services:  make(map[string]Service),
		Version:   version,
		Uptime:    uptime,
	}
}

// Add service status to health response
func (h *HealthResponse) AddService(name string, status HealthStatus, message string, responseTime time.Duration) {
	h.Services[name] = Service{
		Status:       status,
		Message:      message,
		LastChecked:  time.Now(),
		ResponseTime: responseTime,
	}

	// Update overall status based on service statuses
	h.updateOverallStatus()
}

// Update overall health status based on individual services
func (h *HealthResponse) updateOverallStatus() {
	hasUnhealthy := false
	hasDegraded := false

	for _, service := range h.Services {
		switch service.Status {
		case HealthStatusUnhealthy:
			hasUnhealthy = true
		case HealthStatusDegraded:
			hasDegraded = true
		}
	}

	if hasUnhealthy {
		h.Status = HealthStatusUnhealthy
	} else if hasDegraded {
		h.Status = HealthStatusDegraded
	} else {
		h.Status = HealthStatusHealthy
	}
}

// Helper function to create readiness response
func NewReadinessResponse() *ReadinessResponse {
	return &ReadinessResponse{
		Ready:     true,
		Timestamp: time.Now(),
		Checks:    make(map[string]Check),
	}
}

// Add check to readiness response
func (r *ReadinessResponse) AddCheck(name string, status bool, message string) {
	r.Checks[name] = Check{
		Status:  status,
		Message: message,
	}

	// Update overall readiness
	if !status {
		r.Ready = false
	}
}
