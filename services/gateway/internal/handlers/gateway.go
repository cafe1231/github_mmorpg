package handlers

import (
	"net/http"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	startTime   = time.Now()
	reloadCount int32
)

type ServiceStatus struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	Status string `json:"status"`
}

type GatewayHandler struct {
	Services map[string]string // nom -> url
	Version  string
	Commit   string
	Build    string
}

func NewGatewayHandler(services map[string]string, version, commit, build string) *GatewayHandler {
	return &GatewayHandler{
		Services: services,
		Version:  version,
		Commit:   commit,
		Build:    build,
	}
}

// /gateway/status
func (h *GatewayHandler) Status(c *gin.Context) {
	uptime := time.Since(startTime)
	c.JSON(http.StatusOK, gin.H{
		"status":        "ok",
		"uptime":        uptime.String(),
		"version":       h.Version,
		"commit":        h.Commit,
		"build":         h.Build,
		"go_version":    runtime.Version(),
		"num_goroutine": runtime.NumGoroutine(),
		"reloads":       atomic.LoadInt32(&reloadCount),
		"services":      len(h.Services),
	})
}

// /gateway/services
func (h *GatewayHandler) ServicesList(c *gin.Context) {
	statuses := make([]ServiceStatus, 0, len(h.Services))
	for name, url := range h.Services {
		status := "unknown"
		client := http.Client{Timeout: 1 * time.Second}
		resp, err := client.Get(url + "/health")
		if err == nil && resp.StatusCode == 200 {
			status = "up"
		} else {
			status = "down"
		}
		statuses = append(statuses, ServiceStatus{Name: name, URL: url, Status: status})
	}
	c.JSON(http.StatusOK, statuses)
}

// /gateway/version
func (h *GatewayHandler) VersionInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version": h.Version,
		"commit":  h.Commit,
		"build":   h.Build,
	})
}

// /gateway/info
func (h *GatewayHandler) Info(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"env":     os.Environ(),
		"pid":     os.Getpid(),
		"ppid":    os.Getppid(),
		"goarch":  runtime.GOARCH,
		"goos":    runtime.GOOS,
		"num_cpu": runtime.NumCPU(),
	})
}

// /gateway/health/all
func (h *GatewayHandler) HealthAll(c *gin.Context) {
	results := make(map[string]string)
	for name, url := range h.Services {
		client := http.Client{Timeout: 1 * time.Second}
		resp, err := client.Get(url + "/health")
		if err == nil && resp.StatusCode == 200 {
			results[name] = "up"
		} else {
			results[name] = "down"
		}
	}
	c.JSON(http.StatusOK, results)
}

// /gateway/reload
func (h *GatewayHandler) Reload(c *gin.Context) {
	atomic.AddInt32(&reloadCount, 1)
	// Ici, on pourrait recharger la config dynamiquement
	c.JSON(http.StatusOK, gin.H{"message": "Reload effectué", "reloads": atomic.LoadInt32(&reloadCount)})
}
