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

	"player/internal/config"
	"player/internal/database"
	"player/internal/handlers"
	"player/internal/middleware"
	"player/internal/repository"
	"player/internal/service"
)

func main() {
	// Initialisation du logger
	initLogger()

	// Chargement de la configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logrus.Fatal("Failed to load config: ", err)
	}

	// Connexion à la base de données
	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		logrus.Fatal("Failed to connect to database: ", err)
	}
	defer db.Close()

	// Exécution des migrations
	if err := database.RunMigrations(db); err != nil {
		logrus.Fatal("Failed to run migrations: ", err)
	}

	// Initialisation des repositories
	playerRepo := repository.NewPlayerRepository(db)
	characterRepo := repository.NewCharacterRepository(db)

	// Initialisation des services
	playerService := service.NewPlayerService(playerRepo, characterRepo, cfg)
	characterService := service.NewCharacterService(characterRepo, playerRepo, cfg)

	// Configuration du mode Gin
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Configuration des routes
	router := setupRoutes(playerService, characterService, cfg, db)

	// Configuration du serveur HTTP
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Démarrage du serveur en arrière-plan
	go func() {
		logrus.WithFields(logrus.Fields{
			"host": cfg.Server.Host,
			"port": cfg.Server.Port,
			"env":  cfg.Server.Environment,
		}).Info("🎮 Player Service starting...")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatal("Failed to start server: ", err)
		}
	}()

	// Démarrer le nettoyage périodique des modificateurs expirés
	go startPeriodicCleanup(characterService)

	// Gestion gracieuse de l'arrêt
	gracefulShutdown(server, playerService, characterService)
}

// setupRoutes configure toutes les routes du service Player
func setupRoutes(
	playerService *service.PlayerService, 
	characterService *service.CharacterService, 
	cfg *config.Config,
	db *database.DB,
) *gin.Engine {
	router := gin.New()

	// Middleware globaux
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())
	router.Use(middleware.SecurityHeaders())

	// Handlers
	healthHandler := handlers.NewHealthHandler(cfg, db)
	playerHandler := handlers.NewPlayerHandler(playerService, cfg)
	characterHandler := handlers.NewCharacterHandler(characterService, cfg)

	// Routes de santé et monitoring
	router.GET(cfg.Monitoring.HealthPath, healthHandler.HealthCheck)
	router.GET("/health/detailed", healthHandler.DetailedHealthCheck)
	router.GET("/health/ready", healthHandler.Readiness)
	router.GET("/health/live", healthHandler.Liveness)
	router.GET(cfg.Monitoring.MetricsPath, healthHandler.Metrics)
	router.GET("/status", healthHandler.Status)
	router.GET("/info", healthHandler.Info)
	router.GET("/version", healthHandler.Version)
	router.GET("/ping", healthHandler.Ping)

	// API v1
	v1 := router.Group("/api/v1")
	v1.Use(middleware.RateLimit(cfg.RateLimit))
	{
		// Routes publiques (informations sur le jeu)
		public := v1.Group("/")
		{
			public.GET("/characters/game-info", characterHandler.GetGameInfo)
		}

		// Routes protégées (authentification JWT requise)
		protected := v1.Group("/")
		protected.Use(middleware.JWTAuth(cfg.Auth.JWTSecret))
		{
			// Routes joueur
			player := protected.Group("/player")
			{
				player.POST("/profile", playerHandler.CreatePlayer)
				player.GET("/profile", playerHandler.GetPlayer)
				player.PUT("/profile", playerHandler.UpdatePlayer)
				player.GET("/stats", playerHandler.GetPlayerStats)
				player.POST("/playtime", playerHandler.UpdatePlayTime)
				player.POST("/ping", playerHandler.UpdateLastSeen)
				player.GET("/can-create-character", playerHandler.CanCreateCharacter)
				player.GET("/search", playerHandler.SearchPlayers)
				player.GET("/preferences", playerHandler.GetPlayerPreferences)
				player.PUT("/preferences", playerHandler.UpdatePlayerPreferences)
			}

			// Routes personnages
			characters := protected.Group("/characters")
			{
				characters.POST("/", characterHandler.CreateCharacter)
				characters.GET("/", characterHandler.GetCharacters)
				characters.GET("/:id", characterHandler.GetCharacter)
				characters.PUT("/:id", characterHandler.UpdateCharacter)
				characters.DELETE("/:id", characterHandler.DeleteCharacter)
				characters.GET("/:id/stats", characterHandler.GetCharacterStats)
				characters.PUT("/:id/stats", characterHandler.UpdateCharacterStats)
			}

			// Routes admin
			admin := protected.Group("/admin")
			admin.Use(middleware.RequireRole("admin", "superuser"))
			{
				admin.GET("/players", playerHandler.ListPlayers)
				admin.GET("/players/:id", playerHandler.GetPlayerByID)
				admin.DELETE("/players/:id", playerHandler.DeletePlayerProfile)
			}
		}

		// Routes pour les autres services (validation interne)
		services := v1.Group("/services")
		{
			services.GET("/player/:userID", playerHandler.GetPlayerSummary)
			services.GET("/player/:userID/characters", playerHandler.GetPlayerCharactersSummary)
			services.POST("/validate/display-name", playerHandler.ValidateDisplayName)
			services.GET("/online-players", playerHandler.GetOnlinePlayers)
		}
	}

	// Routes de debug (développement seulement)
	if cfg.Server.Debug {
		debug := router.Group("/debug")
		{
			debug.GET("/config", healthHandler.ShowConfig)
			debug.GET("/routes", healthHandler.ListRoutes(router))
			debug.GET("/database", healthHandler.DatabaseStats)
			debug.GET("/memory", healthHandler.MemoryProfile)
			debug.GET("/gc", healthHandler.GCStats)
			debug.GET("/player", playerHandler.ShowPlayerDebug)
			debug.GET("/character/:id", characterHandler.ShowCharacterDebug)
			debug.POST("/characters/:id/experience", characterHandler.AddExperience)
		}
	}

	return router
}

// startPeriodicCleanup démarre le nettoyage périodique
func startPeriodicCleanup(characterService *service.CharacterService) {
	ticker := time.NewTicker(1 * time.Hour) // Nettoyer toutes les heures
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := characterService.CleanupExpiredModifiers(); err != nil {
				logrus.WithError(err).Error("Failed to cleanup expired modifiers")
			} else {
				logrus.Debug("Periodic cleanup of expired modifiers completed")
			}
		}
	}
}

// initLogger initialise le logger global
func initLogger() {
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})

	// Niveau de log selon l'environnement
	if os.Getenv("ENVIRONMENT") == "production" {
		logrus.SetLevel(logrus.InfoLevel)
	} else {
		logrus.SetLevel(logrus.DebugLevel)
	}

	logrus.WithField("service", "player").Info("Logger initialized")
}

// gracefulShutdown gère l'arrêt gracieux du serveur
func gracefulShutdown(server *http.Server, playerService *service.PlayerService, characterService *service.CharacterService) {
	// Canal pour capturer les signaux système
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Attendre un signal
	<-quit
	logrus.Info("🛑 Player Service shutting down...")

	// Timeout pour l'arrêt gracieux
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Arrêter le serveur HTTP
	if err := server.Shutdown(ctx); err != nil {
		logrus.Error("Player service forced to shutdown:", err)
	}

	// Fermer les services
	if err := playerService.Close(); err != nil {
		logrus.Error("Error closing player service:", err)
	}

	if err := characterService.Close(); err != nil {
		logrus.Error("Error closing character service:", err)
	}

	logrus.Info("✅ Player Service stopped")
}