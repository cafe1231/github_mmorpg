package middleware

import (
	"combat/internal/config"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Constantes pour les types d'actions
const (
	ActionTypeUnknown = "unknown"
)

var (
	// Métriques HTTP
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "combat_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "combat_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	httpRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "combat_http_request_size_bytes",
			Help:    "HTTP request size in bytes",
			Buckets: []float64{100, 1000, 10000, 100000, 1000000},
		},
		[]string{"method", "endpoint"},
	)

	httpResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "combat_http_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: []float64{100, 1000, 10000, 100000, 1000000},
		},
		[]string{"method", "endpoint"},
	)

	activeConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "combat_active_connections",
			Help: "Number of active HTTP connections",
		},
	)

	// Métriques spécifiques au combat
	combatInstancesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "combat_instances_total",
			Help: "Total number of combat instances created",
		},
		[]string{"combat_type"},
	)

	activeCombats = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "combat_active_instances",
			Help: "Number of currently active combat instances",
		},
		[]string{"combat_type", "status"},
	)

	combatActionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "combat_actions_total",
			Help: "Total number of combat actions executed",
		},
		[]string{"action_type", "result"},
	)

	combatActionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "combat_action_duration_seconds",
			Help:    "Combat action processing duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 2, 5},
		},
		[]string{"action_type"},
	)

	// combatParticipants = promauto.NewGaugeVec(
	// 	prometheus.GaugeOpts{
	// 		Name: "combat_participants_total",
	// 		Help: "Total number of participants in active combats",
	// 	},
	// 	[]string{"combat_type"},
	// )

	// Métriques d'erreurs et anti-cheat
	antiCheatDetections = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "combat_anticheat_detections_total",
			Help: "Total number of anti-cheat detections",
		},
		[]string{"detection_type", "severity"},
	)

	errorCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "combat_errors_total",
			Help: "Total number of errors",
		},
		[]string{"error_type", "endpoint"},
	)

	// Métriques d'authentification
	authAttempts = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "combat_auth_attempts_total",
			Help: "Total number of authentication attempts",
		},
		[]string{"result"},
	)

	// Métriques de base de données
	dbConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "combat_db_connections_active",
			Help: "Number of active database connections",
		},
	)

	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "combat_db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1},
		},
		[]string{"operation"},
	)
)

// PrometheusMetrics middleware pour collecter les métriques Prometheus
func PrometheusMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Incrémenter les connections actives
		activeConnections.Inc()
		defer activeConnections.Dec()

		// Traiter la requête
		c.Next()

		// Calculer la durée
		duration := time.Since(start)

		// Normaliser l'endpoint pour éviter la cardinalité élevée
		endpoint := normalizeEndpoint(c.FullPath())
		method := c.Request.Method
		statusCode := strconv.Itoa(c.Writer.Status())

		// Enregistrer les métriques HTTP
		httpRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
		httpRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())

		// Taille de la requête
		if c.Request.ContentLength > 0 {
			httpRequestSize.WithLabelValues(method, endpoint).Observe(float64(c.Request.ContentLength))
		}

		// Taille de la réponse
		httpResponseSize.WithLabelValues(method, endpoint).Observe(float64(c.Writer.Size()))

		// Métriques d'erreurs
		if c.Writer.Status() >= 400 {
			errorType := getErrorType(c.Writer.Status())
			errorCount.WithLabelValues(errorType, endpoint).Inc()
		}

		// Métriques d'authentification
		if endpoint == "/api/v1/auth/login" || endpoint == "/api/v1/auth/register" {
			result := "success"
			if c.Writer.Status() >= 400 {
				result = "failure"
			}
			authAttempts.WithLabelValues(result).Inc()
		}
	}
}

// CombatMetrics middleware pour les métriques spécifiques au combat
func CombatMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		path := c.FullPath()

		// Métriques pour les actions de combat
		if path == "/api/v1/combat/:id/action" && c.Request.Method == "POST" {
			actionType := ActionTypeUnknown
			result := "success"

			if c.Writer.Status() >= 400 {
				result = "failure"
			}

			// Essayer d'extraire le type d'action du contexte
			if action := c.GetString("action_type"); action != "" {
				actionType = action
			}

			combatActionsTotal.WithLabelValues(actionType, result).Inc()
			combatActionDuration.WithLabelValues(actionType).Observe(duration.Seconds())
		}

		// Métriques pour la création de combats
		if path == "/api/v1/combat" && c.Request.Method == "POST" && c.Writer.Status() == 201 {
			combatType := c.GetString("combat_type")
			if combatType == "" {
				combatType = ActionTypeUnknown
			}
			combatInstancesTotal.WithLabelValues(combatType).Inc()
		}
	}
}

// AntiCheatMetrics middleware pour les métriques anti-cheat
func AntiCheatMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Vérifier si des événements anti-cheat ont été détectés
		if suspicious := c.GetBool("security_suspicious"); suspicious {
			if events, exists := c.Get("security_events"); exists {
				if eventList, ok := events.([]string); ok {
					for _, event := range eventList {
						antiCheatDetections.WithLabelValues(event, "medium").Inc()
					}
				}
			}
		}

		// Métriques anti-cheat spécifiques au combat
		if suspicion := c.GetFloat64("anticheat_suspicion_score"); suspicion > 0 {
			severity := "low"
			if suspicion > config.DefaultScoreThreshold {
				severity = "medium"
			}
			if suspicion > config.DefaultHighScoreThreshold {
				severity = "high"
			}

			antiCheatDetections.WithLabelValues("combat_suspicious", severity).Inc()
		}
	}
}

// DatabaseMetrics met à jour les métriques de base de données
func UpdateDatabaseMetrics(activeConns int) {
	dbConnectionsActive.Set(float64(activeConns))
}

// RecordDatabaseQuery enregistre une requête de base de données
func RecordDatabaseQuery(operation string, duration time.Duration) {
	dbQueryDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// RecordCombatEvent enregistre un événement de combat
func RecordCombatEvent(eventType string, combatType string) {
	switch eventType {
	case "combat_created":
		combatInstancesTotal.WithLabelValues(combatType).Inc()
	case "combat_started":
		activeCombats.WithLabelValues(combatType, "active").Inc()
	case "combat_ended":
		activeCombats.WithLabelValues(combatType, "active").Dec()
		activeCombats.WithLabelValues(combatType, "finished").Inc()
	}
}

// RecordAntiCheatDetection enregistre une détection anti-cheat
func RecordAntiCheatDetection(detectionType string, severity string) {
	antiCheatDetections.WithLabelValues(detectionType, severity).Inc()
}

// normalizeEndpoint normalise les endpoints pour éviter la cardinalité élevée
func normalizeEndpoint(path string) string {
	if path == "" {
		return "unknown"
	}

	// Remplacer les IDs par des placeholders
	normalized := path

	// Appliquer les patterns de normalisation
	// Note: Dans un vrai projet, vous utiliseriez regexp pour cela
	// Ici on simplifie en retournant le path tel quel

	return normalized
}

// getErrorType retourne le type d'erreur basé sur le code de statut
func getErrorType(statusCode int) string {
	switch {
	case statusCode >= 500:
		return "server_error"
	case statusCode == 429:
		return "rate_limit"
	case statusCode == 401:
		return "auth_error"
	case statusCode == 403:
		return "permission_error"
	case statusCode == 404:
		return "not_found"
	case statusCode >= 400:
		return "client_error"
	default:
		return "unknown"
	}
}

// CustomMetrics structure pour les métriques personnalisées
type CustomMetrics struct {
	PvPMatchesTotal  prometheus.Counter
	PvPMatchDuration prometheus.Histogram
	QueueWaitTime    prometheus.Histogram
	CombatTurnTime   prometheus.Histogram
}

// NewCustomMetrics crée les métriques personnalisées
func NewCustomMetrics() *CustomMetrics {
	return &CustomMetrics{
		PvPMatchesTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "combat_pvp_matches_total",
				Help: "Total number of PvP matches completed",
			},
		),
		PvPMatchDuration: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "combat_pvp_match_duration_seconds",
				Help:    "PvP match duration in seconds",
				Buckets: []float64{30, 60, 120, 300, 600, 1200, 1800},
			},
		),
		QueueWaitTime: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "combat_queue_wait_time_seconds",
				Help:    "Time spent waiting in PvP queue",
				Buckets: []float64{1, 5, 10, 30, 60, 120, 300, 600},
			},
		),
		CombatTurnTime: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "combat_turn_duration_seconds",
				Help:    "Combat turn processing duration",
				Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30},
			},
		),
	}
}

// HealthMetrics met à jour les métriques de santé
func UpdateHealthMetrics(healthy bool) {
	healthStatus := promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "combat_service_healthy",
			Help: "Whether the combat service is healthy (1) or not (0)",
		},
	)

	if healthy {
		healthStatus.Set(1)
	} else {
		healthStatus.Set(0)
	}
}
