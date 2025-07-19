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
	"github.com/dan-2/github_mmorpg/services/guild/internal/repository"
	"github.com/dan-2/github_mmorpg/services/guild/internal/service"
	"github.com/gin-gonic/gin"
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
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Tester la connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("erreur lors du test de connection à la base de données: %w", err)
	}

	return db, nil
}

// setupRouter configure le routeur Gin
func setupRouter(
	guildHandler *handlers.GuildHandler,
	guildMemberHandler *handlers.GuildMemberHandler,
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

		// Routes des membres de guilde (séparées pour éviter les conflits)
		members := v1.Group("/guild-members")
		{
			members.POST("/:guild_id/join", guildMemberHandler.JoinGuild)
			members.GET("/:guild_id", guildMemberHandler.GetMembers)
			members.GET("/:guild_id/:player_id", guildMemberHandler.GetMember)
			members.PUT("/:guild_id/:player_id/role", guildMemberHandler.UpdateMemberRole)
			members.DELETE("/:guild_id/:player_id", guildMemberHandler.KickMember)
			members.DELETE("/:guild_id/leave", guildMemberHandler.LeaveGuild)
		}
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
			repository.NewGuildLogRepository,
			// TODO: Ajouter les autres repositories
			// Services
			func(db *sql.DB) service.GuildPermissionService {
				return service.NewGuildPermissionService(db)
			},
			func(db *sql.DB) service.GuildLogService {
				return service.NewGuildLogService(db)
			},
			func(db *sql.DB) service.GuildMemberService {
				return service.NewGuildMemberService(db)
			},
			service.NewGuildService,
			// Handlers
			handlers.NewGuildHandler,
			handlers.NewGuildMemberHandler,
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
