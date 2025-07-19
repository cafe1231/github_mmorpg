package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"

	"github.com/dan-2/github_mmorpg/services/analytics/internal/config"
	"github.com/dan-2/github_mmorpg/services/analytics/internal/handlers"
	"github.com/dan-2/github_mmorpg/services/analytics/internal/repository"
	"github.com/dan-2/github_mmorpg/services/analytics/internal/service"
)

// Metrics
var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "analytics_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "analytics_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
)

// registerMetrics enregistre les métriques Prometheus
func registerMetrics() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
}

// setupDatabase configure la connection à la base de données
func setupDatabase(cfg *config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
		cfg.Database.Password, cfg.Database.DBName, cfg.Database.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de l'ouverture de la base de données: %w", err)
	}

	// Configurer la connection
	const (
		maxOpenConns    = 25
		maxIdleConns    = 5
		connMaxLifetime = 5
	)
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime * time.Minute)

	// Tester la connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("erreur lors du test de connection à la base de données: %w", err)
	}

	return db, nil
}

// setupRouter configure le routeur Gin
func setupRouter(
	analyticsHandler *handlers.AnalyticsHandler,
) *gin.Engine {
	router := gin.Default()

	// Middleware de métriques
	router.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()

		status := fmt.Sprintf("%d", c.Writer.Status())
		httpRequestsTotal.WithLabelValues(c.Request.Method, c.FullPath(), status).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, c.FullPath()).Observe(duration)
	})

	// Middleware CORS
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		const httpStatusNoContent = 204
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(httpStatusNoContent)
			return
		}

		c.Next()
	})

	// Routes de santé
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"service":   "analytics",
			"timestamp": time.Now().UTC(),
		})
	})

	// Métriques Prometheus
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API v1
	v1 := router.Group("/api/v1")
	{
		// Routes des événements
		events := v1.Group("/events")
		{
			events.POST("/", analyticsHandler.TrackEvent)
			events.GET("/", analyticsHandler.GetEvents)
			events.GET("/:id", analyticsHandler.GetEvent)
		}

		// Routes des métriques
		metrics := v1.Group("/metrics")
		{
			metrics.POST("/", analyticsHandler.RecordMetric)
			metrics.POST("/query", analyticsHandler.GetMetrics)
		}

		// Routes des logs
		logs := v1.Group("/logs")
		{
			logs.POST("/", analyticsHandler.Log)
			logs.POST("/query", analyticsHandler.GetLogs)
		}
	}

	return router
}

// startServer démarre le serveur HTTP
func startServer(lifecycle fx.Lifecycle, router *gin.Engine, cfg *config.Config, logger *logrus.Logger) {
	const readHeaderTimeoutSeconds = 30
	server := &http.Server{
		Addr:              fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:           router,
		ReadHeaderTimeout: readHeaderTimeoutSeconds * time.Second,
	}

	lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			logger.Infof("Startup du service Analytics sur %s:%s", cfg.Server.Host, cfg.Server.Port)
			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Fatalf("Erreur lors du startup du serveur: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Arrêt du service Analytics...")
			return server.Shutdown(ctx)
		},
	})
}

func main() {
	// Enregistrer les métriques Prometheus
	registerMetrics()

	// Charger la configuration
	cfg := config.Load()

	// Configurer le logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)

	// Créer l'application FX
	app := fx.New(
		fx.Provide(
			func() *config.Config { return cfg },
			func() *logrus.Logger { return logger },
			setupDatabase,
			// Repositories
			repository.NewEventRepository,
			repository.NewMetricRepository,
			repository.NewLogRepository,
			// Services
			service.NewAnalyticsService,
			service.NewMetricsService,
			service.NewLoggingService,
			// Handlers
			handlers.NewAnalyticsHandler,
			// Router
			setupRouter,
		),
		fx.Invoke(startServer),
	)

	// Démarrer l'application
	if err := app.Start(context.Background()); err != nil {
		log.Fatalf("Erreur lors du startup de l'application: %v", err)
	}

	// Attendre l'interruption
	<-app.Done()
}
