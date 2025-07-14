// internal/monitoring/metrics.go
package monitoring

import (
"net/http"
"time"

"github.com/gin-gonic/gin"
"github.com/prometheus/client_golang/prometheus"
"github.com/prometheus/client_golang/prometheus/promhttp"
"github.com/sirupsen/logrus"
)

// Métriques Prometheus pour le service Combat
var (
CombatSessionsTotal = prometheus.NewCounterVec(
prometheus.CounterOpts{
Name: "combat_sessions_total",
Help: "Total number of combat sessions created",
},
[]string{"type", "zone_id"},
)

HTTPRequestsTotal = prometheus.NewCounterVec(
prometheus.CounterOpts{
Name: "http_requests_total",
Help: "Total number of HTTP requests",
},
[]string{"method", "endpoint", "status_code"},
)

HTTPRequestDuration = prometheus.NewHistogramVec(
prometheus.HistogramOpts{
Name:    "http_request_duration_seconds",
Help:    "Duration of HTTP requests",
Buckets: prometheus.DefBuckets,
},
[]string{"method", "endpoint"},
)
)

// Metrics structure pour gérer les métriques
type Metrics struct {
registry *prometheus.Registry
}

// NewMetrics crée une nouvelle instance de metrics
func NewMetrics() *Metrics {
registry := prometheus.NewRegistry()

// Enregistrer les métriques
registry.MustRegister(CombatSessionsTotal)
registry.MustRegister(HTTPRequestsTotal)
registry.MustRegister(HTTPRequestDuration)

logrus.Info("Prometheus metrics initialized")

return &Metrics{
registry: registry,
}
}

// Handler retourne le handler Prometheus
func (m *Metrics) Handler() http.Handler {
return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

// Middleware Prometheus pour instrumenter les requêtes HTTP
func (m *Metrics) Middleware() gin.HandlerFunc {
return func(c *gin.Context) {
start := time.Now()

// Traiter la requête
c.Next()

// Mesurer et enregistrer les métriques
duration := time.Since(start).Seconds()
statusCode := c.Writer.Status()

HTTPRequestsTotal.WithLabelValues(
c.Request.Method,
c.FullPath(),
http.StatusText(statusCode),
).Inc()

HTTPRequestDuration.WithLabelValues(
c.Request.Method,
c.FullPath(),
).Observe(duration)
}
}
