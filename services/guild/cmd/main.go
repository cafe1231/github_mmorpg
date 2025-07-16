package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dan-2/github_mmorpg/services/guild/internal/config"
	"github.com/dan-2/github_mmorpg/services/guild/internal/handlers"
	"github.com/dan-2/github_mmorpg/services/guild/internal/models"
	"github.com/dan-2/github_mmorpg/services/guild/internal/repository"
	"github.com/dan-2/github_mmorpg/services/guild/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

// Metrics
var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "guild_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "guild_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
}

// setupDatabase configure la connexion à la base de données
func setupDatabase(cfg *config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
		cfg.Database.Password, cfg.Database.DBName, cfg.Database.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de l'ouverture de la base de données: %w", err)
	}

	// Configurer la connexion
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Tester la connexion
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("erreur lors du test de connexion à la base de données: %w", err)
	}

	return db, nil
}

// setupRouter configure le routeur Gin
func setupRouter(
	guildHandler *handlers.GuildHandler,
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

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Routes de santé
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"service":   "guild",
			"timestamp": time.Now().UTC(),
		})
	})

	// Métriques Prometheus
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API v1
	v1 := router.Group("/api/v1")
	{
		// Routes des guildes
		guilds := v1.Group("/guilds")
		{
			guilds.POST("/", guildHandler.CreateGuild)
			guilds.GET("/", guildHandler.ListGuilds)
			guilds.GET("/search", guildHandler.SearchGuilds)
			guilds.GET("/:id", guildHandler.GetGuild)
			guilds.PUT("/:id", guildHandler.UpdateGuild)
			guilds.DELETE("/:id", guildHandler.DeleteGuild)
			guilds.GET("/:id/stats", guildHandler.GetGuildStats)
		}

		// TODO: Ajouter les autres routes (membres, invitations, etc.)
	}

	return router
}

// startServer démarre le serveur HTTP
func startServer(lifecycle fx.Lifecycle, router *gin.Engine, cfg *config.Config, logger *logrus.Logger) {
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			logger.Infof("Démarrage du service Guild sur %s:%s", cfg.Server.Host, cfg.Server.Port)
			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Fatalf("Erreur lors du démarrage du serveur: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Arrêt du service Guild...")
			return server.Shutdown(ctx)
		},
	})
}

func main() {
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
			repository.NewGuildRepository,
			repository.NewGuildMemberRepository,
			// TODO: Ajouter les autres repositories
			// Services
			func(db *sql.DB) service.GuildPermissionService {
				// TODO: Implémenter le service de permissions
				return &mockPermissionService{}
			},
			func(db *sql.DB) service.GuildLogService {
				// TODO: Implémenter le service de logs
				return &mockLogService{}
			},
			service.NewGuildService,
			// Handlers
			handlers.NewGuildHandler,
			// Router
			setupRouter,
		),
		fx.Invoke(startServer),
	)

	// Démarrer l'application
	if err := app.Start(context.Background()); err != nil {
		log.Fatalf("Erreur lors du démarrage de l'application: %v", err)
	}

	// Attendre l'interruption
	<-app.Done()
}

// Services mock pour les dépendances non encore implémentées
type mockPermissionService struct{}

func (m *mockPermissionService) GetPermissions(ctx context.Context, guildID, playerID uuid.UUID) (*models.GuildPermissionResponse, error) {
	return &models.GuildPermissionResponse{
		CanInvitePlayers:      true,
		CanKickMembers:        true,
		CanPromoteMembers:     true,
		CanDemoteMembers:      true,
		CanManageBank:         true,
		CanDeclareWar:         true,
		CanCreateAlliance:     true,
		CanManageApplications: true,
		CanViewLogs:           true,
		CanEditGuildInfo:      true,
	}, nil
}

func (m *mockPermissionService) HasPermission(ctx context.Context, guildID, playerID uuid.UUID, permission string) (bool, error) {
	return true, nil
}

type mockLogService struct{}

func (m *mockLogService) AddLog(ctx context.Context, guildID, playerID uuid.UUID, action, details string) error {
	return nil
}

func (m *mockLogService) GetLogs(ctx context.Context, guildID uuid.UUID, action *string, page, limit int) ([]*models.GuildLogResponse, int, error) {
	return []*models.GuildLogResponse{}, 0, nil
}

func (m *mockLogService) CleanOldLogs(ctx context.Context, days int) error {
	return nil
}
