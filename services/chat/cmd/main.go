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

	"chat/internal/config"
	"chat/internal/database"
	"chat/internal/handlers"
	"chat/internal/repository"
	"chat/internal/service"
)

// Version du service (à définir lors du build)
var (
	Version   = "1.0.0"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// Initialisation du logger
	initLogger()

	logrus.WithFields(logrus.Fields{
		"service":    "chat",
		"version":    Version,
		"build_time": BuildTime,
		"git_commit": GitCommit,
	}).Info("💬 Starting Chat Service...")

	// Chargement de la configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logrus.Fatal("Failed to load config: ", err)
	}

	// connection à la base de données
	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		logrus.Fatal("Failed to connect to database: ", err)
	}
	defer db.Close()

	// Exécution des migrations
	if err := database.RunMigrations(db.DB); err != nil {
		logrus.Fatal("Failed to run migrations: ", err)
	}

	// Initialisation des repositories
	channelRepo := repository.NewChannelRepository(db.DB)
	messageRepo := repository.NewMessageRepository(db.DB)
	moderationRepo := repository.NewModerationRepository(db.DB)
	userRepo := repository.NewUserRepository(db.DB)

	// Initialisation des services
	chatService := service.NewChatService(channelRepo, messageRepo, moderationRepo, userRepo, cfg)

	// Initialisation des handlers
	chatHandler := handlers.NewChatHandler(chatService)

	// Configuration du mode Gin
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Configuration des routes
	router := setupRoutes(chatHandler, cfg)

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
		}).Info("🚀 Chat Service starting...")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatal("Failed to start server: ", err)
		}
	}()

	// Gestion gracieuse de l'arrêt
	gracefulShutdown(server)
}

// setupRoutes configure toutes les routes du service chat
func setupRoutes(chatHandler *handlers.ChatHandler, cfg *config.Config) *gin.Engine {
	router := gin.New()

	// Middlewares globaux
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// CORS pour le développement
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// Route de santé
	router.GET(cfg.Monitoring.HealthPath, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "chat-service",
			"version": Version,
			"time":    time.Now().Unix(),
		})
	})

	// Routes API v1
	v1 := router.Group("/api/v1")
	{
		// Routes des channels
		channels := v1.Group("/channels")
		{
			channels.POST("", chatHandler.CreateChannel)
			channels.GET("/:id", chatHandler.GetChannel)
			channels.POST("/:id/join", chatHandler.JoinChannel)
			channels.POST("/:id/leave", chatHandler.LeaveChannel)

			// Routes des messages
			channels.POST("/:id/messages", chatHandler.SendMessage)
			channels.GET("/:id/messages", chatHandler.GetMessages)

			// Routes des membres
			channels.GET("/:id/members", chatHandler.GetChannelMembers)
		}
	}

	// TODO: Routes WebSocket
	// router.GET("/ws", websocketHandler.HandleWebSocket)

	return router
}

// initLogger initialize la configuration des logs
func initLogger() {
	// Configuration du format JSON pour la production
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})

	// Niveau de log depuis l'environnement
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		if parsedLevel, err := logrus.ParseLevel(level); err == nil {
			logrus.SetLevel(parsedLevel)
		}
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	// Sortie vers stdout
	logrus.SetOutput(os.Stdout)
}

// gracefulShutdown gère l'arrêt gracieux du serveur
func gracefulShutdown(server *http.Server) {
	// Canal pour capturer les signaux système
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Attendre le signal
	sig := <-quit
	logrus.WithField("signal", sig.String()).Info("Received shutdown signal")

	// Timeout pour l'arrêt gracieux
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Arrêt du serveur HTTP
	logrus.Info("Shutting down HTTP server...")
	if err := server.Shutdown(ctx); err != nil {
		logrus.WithError(err).Error("Failed to shutdown server gracefully")
		return
	}

	logrus.Info("💬 Chat Service stopped gracefully")
}

