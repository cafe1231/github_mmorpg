package main

import (
	"context"
	"fmt"
	"gateway/internal/config"
	"gateway/internal/gateway"
	"gateway/internal/handlers"
	"gateway/internal/middleware"
	"gateway/internal/monitoring"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Constantes d'environnement
const (
	EnvProduction   = "production"
	ShutdownTimeout = 30 // secondes
)

func main() {
	// Initialisation du logger
	initLogger()

	// Chargement de la configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logrus.Fatal("Failed to load config: ", err)
	}

	// Configuration du mode Gin
	if cfg.Server.Environment == EnvProduction {
		gin.SetMode(gin.ReleaseMode)
	}

	// Création du serveur gateway
	gatewayServer, err := gateway.NewServer(cfg)
	if err != nil {
		logrus.Fatal("Failed to create gateway server: ", err)
	}

	// Initialisation du monitoring
	monitoring.Init(cfg.Monitoring.PrometheusPort)

	// Initialisation des métriques middleware
	middleware.InitMetrics()

	// Définition des services connus (nom -> URL)
	services := map[string]string{
		"auth":      "http://localhost:8081",
		"player":    "http://localhost:8082",
		"world":     "http://localhost:8083",
		"inventory": "http://localhost:8084",
		"guild":     "http://localhost:8086",
		"chat":      "http://localhost:8087",
		"analytics": "http://localhost:8088",
		"combat":    "http://localhost:8085",
	}
	version := "1.0.0"
	commit := "dev"
	build := time.Now().Format(time.RFC3339)

	gatewayHandler := handlers.NewGatewayHandler(services, version, commit, build)

	// Configuration des routes
	router := setupRoutes(gatewayServer, cfg, gatewayHandler)

	// Configuration du serveur HTTP
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Démarriage du serveur en arrière-plan
	go func() {
		logrus.WithFields(logrus.Fields{
			"host": cfg.Server.Host,
			"port": cfg.Server.Port,
			"env":  cfg.Server.Environment,
		}).Info("🚀 Gateway Service starting...")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatal("Failed to start server: ", err)
		}
	}()

	// Gestion gracieuse de l'arrêt
	gracefulShutdown(server, gatewayServer)
}

// setupRoutes configure toutes les routes du gateway
func setupRoutes(gatewayServer *gateway.Server, cfg *config.Config, gatewayHandler *handlers.GatewayHandler) *gin.Engine {
	router := gin.New()

	// Middleware globaux
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())
	router.Use(middleware.RateLimit(cfg.RateLimit))
	router.Use(middleware.Metrics())

	// Routes de santé et monitoring
	router.GET(cfg.Monitoring.HealthPath, gatewayServer.HealthCheck)
	router.GET(cfg.Monitoring.MetricsPath, gin.WrapH(monitoring.Handler()))

	// Routes directes pour le service Auth (niveau racine)
	authDirect := router.Group("/auth")
	{
		// Routes publiques du service Auth (pas de JWT requis)
		authDirect.GET("/health", gatewayServer.ProxyTo("auth"))
		authDirect.GET("/metrics", gatewayServer.ProxyTo("auth"))
		authDirect.POST("/register", gatewayServer.ProxyTo("auth"))
		authDirect.POST("/login", gatewayServer.ProxyTo("auth"))
		authDirect.POST("/refresh", gatewayServer.ProxyTo("auth"))
		authDirect.POST("/forgot-password", gatewayServer.ProxyTo("auth"))
		authDirect.POST("/reset-password", gatewayServer.ProxyTo("auth"))
		authDirect.GET("/verify-email/:token", gatewayServer.ProxyTo("auth"))

		// OAuth endpoints
		authDirect.GET("/oauth/:provider", gatewayServer.ProxyTo("auth"))
		authDirect.POST("/oauth/:provider/callback", gatewayServer.ProxyTo("auth"))
	}

	// Handlers spécifiques Gateway
	gw := router.Group("/gateway")
	{
		gw.GET("/status", gatewayHandler.Status)
		gw.GET("/services", gatewayHandler.ServicesList)
		gw.GET("/version", gatewayHandler.VersionInfo)
		gw.GET("/info", gatewayHandler.Info)
		gw.GET("/health/all", gatewayHandler.HealthAll)
		gw.POST("/reload", gatewayHandler.Reload)
	}

	// API Gateway routes (structure complète)
	api := router.Group("/api/v1")
	{
		// Routes d'authentification (pas de JWT requis)
		auth := api.Group("/auth")
		{
			auth.POST("/register", gatewayServer.ProxyTo("auth"))
			auth.POST("/login", gatewayServer.ProxyTo("auth"))
			auth.POST("/refresh", gatewayServer.ProxyTo("auth"))
		}

		// Routes protégées (JWT requis)
		protected := api.Group("/")
		protected.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			// Player Service
			player := protected.Group("/player")
			{
				player.GET("/profile", gatewayServer.ProxyTo("player"))
				player.PUT("/profile", gatewayServer.ProxyTo("player"))
				player.GET("/characters", gatewayServer.ProxyTo("player"))
				player.POST("/characters", gatewayServer.ProxyTo("player"))
				player.GET("/characters/:id", gatewayServer.ProxyTo("player"))
				player.PUT("/characters/:id", gatewayServer.ProxyTo("player"))
				player.DELETE("/characters/:id", gatewayServer.ProxyTo("player"))
			}

			// User management via Auth service
			user := protected.Group("/user")
			{
				user.GET("/profile", gatewayServer.ProxyTo("auth"))
				user.PUT("/profile", gatewayServer.ProxyTo("auth"))
				user.POST("/change-password", gatewayServer.ProxyTo("auth"))
				user.POST("/logout", gatewayServer.ProxyTo("auth"))
				user.POST("/logout-all", gatewayServer.ProxyTo("auth"))
				user.GET("/sessions", gatewayServer.ProxyTo("auth"))
				user.DELETE("/sessions/:id", gatewayServer.ProxyTo("auth"))
			}

			// World Service
			world := protected.Group("/world")
			{
				world.GET("/zones", gatewayServer.ProxyTo("world"))
				world.GET("/zones/:id", gatewayServer.ProxyTo("world"))
				world.POST("/zones/:id/enter", gatewayServer.ProxyTo("world"))
				world.POST("/zones/:id/leave", gatewayServer.ProxyTo("world"))
				world.GET("/events", gatewayServer.ProxyTo("world"))
			}

			// Combat Service (performance critique)
			combat := protected.Group("/combat")
			combat.Use(middleware.HighPerformanceMode()) // Cache agressif
			{
				combat.POST("/action", gatewayServer.ProxyTo("combat"))
				combat.GET("/status/:characterId", gatewayServer.ProxyTo("combat"))
				combat.POST("/pvp/challenge", gatewayServer.ProxyTo("combat"))
			}

			// Inventory Service
			inventory := protected.Group("/inventory")
			{
				inventory.GET("/:characterId", gatewayServer.ProxyTo("inventory"))
				inventory.POST("/:characterId/items", gatewayServer.ProxyTo("inventory"))
				inventory.PUT("/:characterId/items/:itemId", gatewayServer.ProxyTo("inventory"))
				inventory.DELETE("/:characterId/items/:itemId", gatewayServer.ProxyTo("inventory"))
				inventory.POST("/:characterId/trade", gatewayServer.ProxyTo("inventory"))
			}

			// Guild Service
			guild := protected.Group("/guild")
			{
				guild.GET("/", gatewayServer.ProxyTo("guild"))
				guild.POST("/", gatewayServer.ProxyTo("guild"))
				guild.GET("/:id", gatewayServer.ProxyTo("guild"))
				guild.POST("/:id/join", gatewayServer.ProxyTo("guild"))
				guild.POST("/:id/leave", gatewayServer.ProxyTo("guild"))
				guild.GET("/:id/members", gatewayServer.ProxyTo("guild"))
			}

			// Chat Service (WebSocket sera géré séparément)
			chat := protected.Group("/chat")
			{
				chat.GET("/channels", gatewayServer.ProxyTo("chat"))
				chat.POST("/channels", gatewayServer.ProxyTo("chat"))
				chat.GET("/messages/:channelId", gatewayServer.ProxyTo("chat"))
				chat.POST("/messages", gatewayServer.ProxyTo("chat"))
			}

			// Analytics Service (admin seulement)
			analytics := protected.Group("/analytics")
			analytics.Use(middleware.RequireRole("admin"))
			{
				analytics.GET("/dashboard", gatewayServer.ProxyTo("analytics"))
				analytics.GET("/metrics", gatewayServer.ProxyTo("analytics"))
				analytics.GET("/reports", gatewayServer.ProxyTo("analytics"))
			}
		}

		// Routes de validation (pour les autres services)
		validation := api.Group("/validate")
		{
			validation.POST("/token", gatewayServer.ProxyTo("auth"))
			validation.GET("/user/:id", gatewayServer.ProxyTo("auth"))
		}
	}

	// WebSocket endpoint pour le chat temps réel
	router.GET("/ws", gatewayServer.HandleWebSocket)

	// Routes de debug (développement seulement)
	if cfg.Server.Debug {
		debug := router.Group("/debug")
		{
			debug.GET("/routes", gatewayServer.ListRoutes)
			debug.GET("/config", gatewayServer.ShowConfig)
			debug.GET("/services", gatewayServer.ServiceStatus)
		}
	}

	return router
}

// initLogger initialize le logger global
func initLogger() {
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})

	// Niveau de log selon l'environnement
	if os.Getenv("ENVIRONMENT") == EnvProduction {
		logrus.SetLevel(logrus.InfoLevel)
	} else {
		logrus.SetLevel(logrus.DebugLevel)
	}

	logrus.WithField("service", "gateway").Info("Logger initialized")
}

// gracefulShutdown gère l'arrêt gracieux du serveur
func gracefulShutdown(server *http.Server, gatewayServer *gateway.Server) {
	// Canal pour capturer les signaux système
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Attendre un signal
	<-quit
	logrus.Info("🛑 Gateway Service shutting down...")

	// Timeout pour l'arrêt gracieux
	ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout*time.Second)
	defer cancel()

	// Arrêter le serveur HTTP
	if err := server.Shutdown(ctx); err != nil {
		logrus.Error("Gateway forced to shutdown:", err)
	}

	// Fermer les connections du gateway (NATS, services, etc.)
	if err := gatewayServer.Close(); err != nil {
		logrus.Error("Error closing gateway server:", err)
	}

	logrus.Info("✅ Gateway Service stopped")
}
