package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"

	"world/internal/config"
	"world/internal/database"
	"world/internal/handlers"
	"world/internal/middleware"
	"world/internal/repository"
	"world/internal/service"
	"world/pkg/monitoring"
)

func main() {
	// Configuration des logs
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)

	app := fx.New(
		// Configuration
		fx.Provide(config.LoadConfig),

		// Database
		fx.Provide(database.NewConnection),

		// Repositories
		fx.Provide(repository.NewZoneRepository),
		fx.Provide(repository.NewNPCRepository),
		fx.Provide(repository.NewPlayerPositionRepository),
		fx.Provide(repository.NewWorldEventRepository),
		fx.Provide(repository.NewWeatherRepository),

		// Services
		fx.Provide(service.NewZoneService),
		fx.Provide(service.NewNPCService),
		fx.Provide(service.NewPlayerPositionService),
		fx.Provide(service.NewWorldEventService),
		fx.Provide(service.NewWeatherService),

		// Handlers
		fx.Provide(handlers.NewZoneHandler),
		fx.Provide(handlers.NewNPCHandler),
		fx.Provide(handlers.NewPlayerPositionHandler),
		fx.Provide(handlers.NewWorldEventHandler),
		fx.Provide(handlers.NewWeatherHandler),

		// HTTP Server
		fx.Provide(NewHTTPServer),

		// Lifecycle
		fx.Invoke(RunServer),
	)

	app.Run()
}

// NewHTTPServer crée et configure le serveur HTTP
func NewHTTPServer(
	cfg *config.Config,
	zoneHandler *handlers.ZoneHandler,
	npcHandler *handlers.NPCHandler,
	positionHandler *handlers.PlayerPositionHandler,
	eventHandler *handlers.WorldEventHandler,
	weatherHandler *handlers.WeatherHandler,
) *gin.Engine {
	// Mode Gin selon l'environnement
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Middleware globaux
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())
	router.Use(middleware.RateLimit(cfg.RateLimit))
	router.Use(middleware.Metrics())

	// Routes de santé et monitoring
	router.GET(cfg.Monitoring.HealthPath, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "world-service",
			"version":   "1.0.0",
			"timestamp": time.Now().Unix(),
		})
	})
	router.GET(cfg.Monitoring.MetricsPath, gin.WrapH(monitoring.Handler()))

	// API routes
	api := router.Group("/api/v1")
	{
		// Routes des zones
		zones := api.Group("/zones")
		zones.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			zones.GET("", zoneHandler.ListZones)
			zones.GET("/:id", zoneHandler.GetZone)
			zones.POST("/:id/enter", zoneHandler.EnterZone)
			zones.POST("/:id/leave", zoneHandler.LeaveZone)
			zones.GET("/:id/players", zoneHandler.GetPlayersInZone)
			zones.GET("/:id/npcs", zoneHandler.GetNPCsInZone)
		}

		// Routes des NPCs
		npcs := api.Group("/npcs")
		npcs.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			npcs.GET("", npcHandler.ListNPCs)
			npcs.GET("/:id", npcHandler.GetNPC)
			npcs.POST("/:id/interact", npcHandler.InteractWithNPC)
			npcs.GET("/zone/:zoneId", npcHandler.GetNPCsByZone)
		}

		// Routes des positions
		positions := api.Group("/positions")
		positions.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			positions.GET("/character/:characterId", positionHandler.GetCharacterPosition)
			positions.PUT("/character/:characterId", positionHandler.UpdateCharacterPosition)
			positions.GET("/zone/:zoneId", positionHandler.GetZonePositions)
		}

		// Routes des événements du monde
		events := api.Group("/events")
		events.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			events.GET("", eventHandler.ListEvents)
			events.GET("/active", eventHandler.GetActiveEvents)
			events.GET("/zone/:zoneId", eventHandler.GetZoneEvents)
			events.POST("/:id/participate", eventHandler.ParticipateInEvent)
		}

		// Routes de la météo
		weather := api.Group("/weather")
		weather.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			weather.GET("/zone/:zoneId", weatherHandler.GetZoneWeather)
			weather.GET("/forecast/:zoneId", weatherHandler.GetWeatherForecast)
		}

		// Routes administratives
		admin := api.Group("/admin")
		admin.Use(middleware.JWTAuth(cfg.JWT.Secret))
		admin.Use(middleware.RequireRole("admin"))
		{
			admin.POST("/zones", zoneHandler.CreateZone)
			admin.PUT("/zones/:id", zoneHandler.UpdateZone)
			admin.DELETE("/zones/:id", zoneHandler.DeleteZone)
			
			admin.POST("/npcs", npcHandler.CreateNPC)
			admin.PUT("/npcs/:id", npcHandler.UpdateNPC)
			admin.DELETE("/npcs/:id", npcHandler.DeleteNPC)
			
			admin.POST("/events", eventHandler.CreateEvent)
			admin.PUT("/events/:id", eventHandler.UpdateEvent)
			admin.DELETE("/events/:id", eventHandler.DeleteEvent)
			
			admin.POST("/weather/:zoneId", weatherHandler.SetWeather)
		}
	}

	// Routes de debug (seulement en développement)
	if cfg.Server.Environment == "development" {
		debug := router.Group("/debug")
		{
			debug.GET("/routes", func(c *gin.Context) {
				routes := []map[string]string{
					{"method": "GET", "path": "/api/v1/zones"},
					{"method": "GET", "path": "/api/v1/zones/:id"},
					{"method": "POST", "path": "/api/v1/zones/:id/enter"},
					{"method": "GET", "path": "/api/v1/npcs"},
					{"method": "GET", "path": "/api/v1/positions/character/:characterId"},
					{"method": "GET", "path": "/api/v1/events"},
					{"method": "GET", "path": "/api/v1/weather/zone/:zoneId"},
				}
				c.JSON(http.StatusOK, gin.H{"routes": routes})
			})
		}
	}

	return router
}

// RunServer démarre le serveur HTTP
func RunServer(lc fx.Lifecycle, cfg *config.Config, router *gin.Engine, db *database.DB) {
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Exécuter les migrations
			if err := database.RunMigrations(db); err != nil {
				return fmt.Errorf("failed to run migrations: %w", err)
			}

			// Démarrer le serveur
			go func() {
				logrus.WithFields(logrus.Fields{
					"host":        cfg.Server.Host,
					"port":        cfg.Server.Port,
					"environment": cfg.Server.Environment,
				}).Info("Starting World service")

				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logrus.WithError(err).Fatal("Server failed to start")
				}
			}()

			// Attendre les signaux d'arrêt
			go func() {
				quit := make(chan os.Signal, 1)
				signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
				<-quit

				logrus.Info("Shutting down World service...")

				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				if err := server.Shutdown(ctx); err != nil {
					logrus.WithError(err).Error("Server forced to shutdown")
				}

				if err := db.Close(); err != nil {
					logrus.WithError(err).Error("Failed to close database connection")
				}

				logrus.Info("World service stopped")
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return server.Shutdown(ctx)
		},
	})
}