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

	"combat/internal/config"
	"combat/internal/database"
	"combat/internal/handler"
	"combat/internal/middleware"
	"combat/internal/monitoring"
	"combat/internal/repository"
	"combat/internal/service"
	"combat/internal/external"
)

func main() {
	// Initialisation du logger
	initLogger()

	// Chargement de la configuration
	cfg, err := config.Load()
	if err != nil {
		logrus.Fatal("Failed to load config: ", err)
	}

	// Connexion à la base de données
	db, err := database.NewDB(cfg)
	if err != nil {
		logrus.Fatal("Failed to connect to database: ", err)
	}
	defer db.Close()

	// Exécution des migrations (optionnel - peut être fait par un outil séparé)
	// if err := database.RunMigrations(db); err != nil {
	//     logrus.Fatal("Failed to run migrations: ", err)
	// }

	// Initialisation des dépendances
	params := setupDependencies(cfg, db)

	// Configuration du mode Gin
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Configuration des routes
	router := setupRoutes(params, cfg)

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
		}).Info("🚀 Combat Service starting...")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatal("Failed to start server: ", err)
		}
	}()

	// Gestion gracieuse de l'arrêt
	gracefulShutdown(server)
}

// DependencyParams contient toutes les dépendances du service
type DependencyParams struct {
	CombatHandler *handler.CombatHandler
	HealthHandler *handler.HealthHandler
}

// setupDependencies configure toutes les dépendances
func setupDependencies(cfg *config.Config, db *database.DB) *DependencyParams {
	// Repositories
	combatRepo := repository.NewCombatRepository(db)
	spellRepo := repository.NewSpellRepository(db)
	effectRepo := repository.NewEffectRepository(db)
	combatLogRepo := repository.NewCombatLogRepository(db)

	// External clients
	playerClient := external.NewPlayerClient(cfg)
	worldClient := external.NewWorldClient(cfg)

	// Services
	damageCalc := service.NewDamageCalculator(cfg)
	effectService := service.NewEffectService(cfg, effectRepo, combatRepo, combatLogRepo)
	combatService := service.NewCombatService(
		cfg,
		combatRepo,
		spellRepo,
		effectRepo,
		combatLogRepo,
		playerClient,
		worldClient,
		damageCalc,
		effectService,
	)

	// Handlers
	combatHandler := handler.NewCombatHandler(combatService)
	healthHandler := handler.NewHealthHandler(nil)

	return &DependencyParams{
		CombatHandler: combatHandler,
		HealthHandler: healthHandler,
	}
}

// setupRoutes configure toutes les routes du service
func setupRoutes(params *DependencyParams, cfg *config.Config) *gin.Engine {
	router := gin.New()

	// Middleware globaux
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())

	// Routes de santé et monitoring
	router.GET("/health", params.HealthHandler.HealthCheck)
	router.GET("/metrics", gin.WrapH(monitoring.NewMetrics().Handler()))

	// API v1
	v1 := router.Group("/api/v1")
	{
		// Routes de combat (avec JWT)
		combat := v1.Group("/combat")
		combat.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			// Gestion des sessions
			combat.POST("/start", params.CombatHandler.StartCombat)
			combat.POST("/join/:sessionId", params.CombatHandler.JoinCombat)
			combat.POST("/leave/:sessionId", params.CombatHandler.LeaveCombat)
			combat.GET("/status/:sessionId", params.CombatHandler.GetCombatStatus)
			
			// Actions de combat
			combat.POST("/action", params.CombatHandler.PerformAction)
			
			// Utilitaires
			combat.GET("/sessions/active", params.CombatHandler.GetActiveCombats)
			combat.GET("/logs/character/:characterId", params.CombatHandler.GetCombatLogs)
			combat.GET("/logs/session/:sessionId", params.CombatHandler.GetSessionLogs)
		}

		// Routes admin (avec JWT + rôle admin)
		admin := v1.Group("/admin")
		admin.Use(middleware.JWTAuth(cfg.JWT.Secret))
		admin.Use(middleware.RequireRole("admin"))
		{
			admin.GET("/combats", params.CombatHandler.GetAllCombatSessions)
			admin.POST("/combats/:sessionId/end", params.CombatHandler.ForceEndCombat)
			admin.GET("/statistics", params.CombatHandler.GetCombatStatistics)
			admin.POST("/simulate", params.CombatHandler.SimulateCombat)
		}

		// Routes PvP (temporairement commentées jusqu'à implémentation complète)
		/*
		pvp := v1.Group("/pvp")
		pvp.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			pvp.POST("/challenge", params.CombatHandler.ChallengePvP)
			pvp.POST("/challenge/:challengeId/accept", params.CombatHandler.AcceptPvPChallenge)
			pvp.POST("/challenge/:challengeId/decline", params.CombatHandler.DeclinePvPChallenge)
			pvp.GET("/rankings", params.CombatHandler.GetPvPRankings)
		}

		// Routes d'effets (temporairement commentées jusqu'à implémentation complète)
		effects := v1.Group("/effects")
		effects.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			effects.GET("/character/:characterId", params.CombatHandler.GetActiveEffects)
			effects.DELETE("/:effectId", params.CombatHandler.RemoveEffect)
		}
		*/
	}

	return router
}

// gracefulShutdown gère l'arrêt gracieux du serveur
func gracefulShutdown(server *http.Server) {
	// Canal pour recevoir les signaux du système
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Attendre le signal
	<-quit
	logrus.Info("🛑 Combat Service shutting down...")

	// Contexte avec timeout pour l'arrêt
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Arrêt gracieux du serveur
	if err := server.Shutdown(ctx); err != nil {
		logrus.Fatal("Server forced to shutdown: ", err)
	}

	logrus.Info("✅ Combat Service stopped")
}

// initLogger initialise le logger
func initLogger() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)

	if os.Getenv("ENVIRONMENT") == "development" {
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
		logrus.SetLevel(logrus.DebugLevel)
	}
}