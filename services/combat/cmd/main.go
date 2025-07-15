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
	"combat/internal/handlers"
	"combat/internal/middleware"
	"combat/internal/repository"
	"combat/internal/service"
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
		"service":    "combat",
		"version":    Version,
		"build_time": BuildTime,
		"git_commit": GitCommit,
	}).Info("⚔️  Starting Combat Service...")

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
	combatRepo := repository.NewCombatRepository(db)
	actionRepo := repository.NewActionRepository(db)
	effectRepo := repository.NewEffectRepository(db)
	pvpRepo := repository.NewPvPRepository(db)

	// Initialisation des services utilitaires
	damageCalc := service.NewDamageCalculator(cfg)
	antiCheat := service.NewAntiCheatService(actionRepo, cfg)

	// Initialisation des services principaux
	effectService := service.NewEffectService(effectRepo, combatRepo, cfg)
	actionService := service.NewActionService(actionRepo, combatRepo, effectRepo, damageCalc, cfg)
	combatService := service.NewCombatService(combatRepo, actionRepo, effectRepo, actionService, effectService, antiCheat, cfg)
	pvpService := service.NewPvPService(pvpRepo, combatRepo, cfg)

	// Démarrage des routines de nettoyage
	combatService.StartCombatCleanupRoutine()
	effectService.StartEffectCleanupRoutine()
	actionService.StartCooldownCleanupRoutine()
	antiCheat.StartCleanupRoutine()

	// Initialisation des handlers
	combatHandler := handlers.NewCombatHandler(combatService, cfg)
	pvpHandler := handlers.NewPvPHandler(pvpService, cfg)
	healthHandler := handlers.NewHealthHandler(cfg, db)

	// Configuration du mode Gin
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Configuration des routes
	router := setupRoutes(combatHandler, pvpHandler, healthHandler, cfg)

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
		}).Info("⚔️  Combat Service started successfully")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatal("Failed to start server: ", err)
		}
	}()

	// Gestion gracieuse de l'arrêt
	gracefulShutdown(server, combatService, effectService, actionService, antiCheat)
}

// setupRoutes configure toutes les routes du service Combat
func setupRoutes(
	combatHandler *handlers.CombatHandler,
	pvpHandler *handlers.PvPHandler,
	healthHandler *handlers.HealthHandler,
	cfg *config.Config,
) *gin.Engine {
	router := gin.New()

	// Middleware globaux
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())
	router.Use(middleware.SecurityHeaders())

	// Rate limiting global si configuré
	if cfg.RateLimit.RequestsPerMinute > 0 {
		router.Use(middleware.RateLimit(cfg.RateLimit))
	}

	// Routes de santé et monitoring (sans auth)
	router.GET(cfg.Monitoring.HealthPath, healthHandler.HealthCheck)
	router.GET("/ready", healthHandler.ReadinessCheck)
	router.GET("/live", healthHandler.LivenessCheck)
	router.GET(cfg.Monitoring.MetricsPath, healthHandler.Metrics)
	router.GET("/stats", healthHandler.Stats)

	// Routes de debug (seulement en développement)
	if cfg.Server.Debug {
		debug := router.Group("/debug")
		{
			debug.GET("/info", healthHandler.Debug)
			debug.GET("/config", healthHandler.ConfigHandler)
			debug.GET("/combats", healthHandler.ActiveCombats)
			debug.GET("/metrics/combat", healthHandler.CombatMetrics)
		}
	}

	// API v1
	v1 := router.Group("/api/v1")
	{
		// Routes protégées (authentification JWT requise)
		protected := v1.Group("/")
		protected.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			// Routes de combat
			combat := protected.Group("/combat")
			combat.Use(middleware.HighPerformanceMode()) // Cache agressif pour les performances
			{
				// CRUD des combats
				combat.POST("/", combatHandler.CreateCombat)
				combat.GET("/:id", combatHandler.GetCombat)
				combat.GET("/:id/status", combatHandler.GetCombatStatus)
				combat.PUT("/:id/start", combatHandler.StartCombat)
				combat.PUT("/:id/end", combatHandler.EndCombat)
				combat.DELETE("/:id", combatHandler.DeleteCombat)

				// Gestion des participants
				combat.POST("/:id/join", combatHandler.JoinCombat)
				combat.POST("/:id/leave", combatHandler.LeaveCombat)
				combat.GET("/:id/participants", combatHandler.GetParticipants)
				combat.PUT("/:id/participants/:participantId", combatHandler.UpdateParticipant)

				// Actions de combat
				combat.POST("/:id/action", combatHandler.ExecuteAction)
				combat.POST("/:id/validate-action", combatHandler.ValidateAction)
				combat.GET("/:id/available-actions", combatHandler.GetAvailableActions)

				// Gestion des tours
				combat.POST("/:id/process-turn", combatHandler.ProcessTurn)
				combat.POST("/:id/advance-turn", combatHandler.AdvanceTurn)
				combat.GET("/:id/turn-info", combatHandler.GetCurrentTurn)

				// Effets
				combat.POST("/:id/effects", combatHandler.ApplyEffect)
				combat.DELETE("/:id/effects/:effectId", combatHandler.RemoveEffect)
				combat.GET("/:id/effects", combatHandler.GetCombatEffects)
			}

			// Routes PvP
			pvp := protected.Group("/pvp")
			{
				// Défis PvP
				pvp.POST("/challenge", pvpHandler.CreateChallenge)
				pvp.GET("/challenges", pvpHandler.GetChallenges)
				pvp.GET("/challenges/:id", pvpHandler.GetChallenge)
				pvp.POST("/challenges/:id/respond", pvpHandler.RespondToChallenge)
				pvp.DELETE("/challenges/:id", pvpHandler.CancelChallenge)

				// Classements et statistiques
				pvp.GET("/rankings", pvpHandler.GetRankings)
				pvp.GET("/statistics/:characterId", pvpHandler.GetPvPStatistics)
				pvp.GET("/season", pvpHandler.GetSeasonInfo)

				// File d'attente
				pvp.POST("/queue", pvpHandler.JoinQueue)
				pvp.DELETE("/queue", pvpHandler.LeaveQueue)
				pvp.GET("/queue/status", pvpHandler.GetQueueStatus)
			}

			// Recherche et historique
			search := protected.Group("/")
			{
				search.GET("/combats", combatHandler.SearchCombats)
				search.GET("/history", combatHandler.GetCombatHistory)
				search.GET("/statistics", combatHandler.GetStatistics)
			}

			// Routes admin
			admin := protected.Group("/admin")
			admin.Use(middleware.RequireRole("admin", "moderator"))
			{
				admin.GET("/combats", combatHandler.ListAllCombats)
				admin.POST("/combats/:id/force-end", combatHandler.ForceEndCombat)
				admin.POST("/combats/:id/admin-action", combatHandler.AdminAction)
				admin.GET("/suspicious-activities", combatHandler.GetSuspiciousActivities)
				admin.POST("/ban/:userId", combatHandler.BanUser)
			}
		}

		// Routes pour les autres services (validation interne)
		services := v1.Group("/services")
		{
			services.GET("/combat/:combatId/status", combatHandler.GetCombatStatusForService)
			services.POST("/validate/character-stats", combatHandler.ValidateCharacterStats)
			services.GET("/active-combats-count", combatHandler.GetActiveCombatCount)
		}
	}

	return router
}

// initLogger initialise le système de logging
func initLogger() {
	// Configuration du format de log selon l'environnement
	if os.Getenv("SERVER_ENVIRONMENT") == "production" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetLevel(logrus.InfoLevel)
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
			ForceColors:   true,
		})
		logrus.SetLevel(logrus.DebugLevel)
	}

	logrus.SetOutput(os.Stdout)
}

// gracefulShutdown gère l'arrêt gracieux du service
func gracefulShutdown(
	server *http.Server,
	combatService service.CombatServiceInterface,
	effectService service.EffectServiceInterface,
	actionService service.ActionServiceInterface,
	antiCheat service.AntiCheatServiceInterface,
) {
	// Canal pour recevoir les signaux d'interruption
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Attendre le signal
	<-quit
	logrus.Info("⚔️  Combat Service is shutting down...")

	// Timeout pour l'arrêt gracieux
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Arrêter les nouvelles connexions
	if err := server.Shutdown(ctx); err != nil {
		logrus.Fatal("Server forced to shutdown: ", err)
	}

	// Nettoyage des ressources
	logrus.Info("Cleaning up resources...")

	// Nettoyer les combats actifs
	activeCount, err := combatService.GetActiveCombatCount()
	if err == nil && activeCount > 0 {
		logrus.WithField("active_combats", activeCount).Warn("Shutting down with active combats")
		// TODO: Sauvegarder l'état des combats actifs ou les mettre en pause
	}

	// Nettoyer les données temporaires de l'anti-cheat
	antiCheat.CleanupOldData()

	logrus.Info("⚔️  Combat Service stopped gracefully")
}

// Fonctions utilitaires pour les handlers

// extractUserID récupère l'ID utilisateur depuis le contexte Gin
func extractUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}

	id, ok := userID.(string)
	return id, ok
}

// extractCharacterID récupère l'ID du personnage depuis les paramètres ou headers
func extractCharacterID(c *gin.Context) (string, bool) {
	// D'abord essayer depuis les paramètres d'URL
	if characterID := c.Param("characterId"); characterID != "" {
		return characterID, true
	}

	// Ensuite essayer depuis les headers
	if characterID := c.GetHeader("X-Character-ID"); characterID != "" {
		return characterID, true
	}

	// Finalement essayer depuis le query
	if characterID := c.Query("character_id"); characterID != "" {
		return characterID, true
	}

	return "", false
}

// respondWithError envoie une réponse d'erreur standardisée
func respondWithError(c *gin.Context, code int, message string, details ...interface{}) {
	response := gin.H{
		"error":      message,
		"timestamp":  time.Now().Unix(),
		"request_id": c.GetHeader("X-Request-ID"),
	}

	if len(details) > 0 {
		response["details"] = details[0]
	}

	c.JSON(code, response)
}

// respondWithSuccess envoie une réponse de succès standardisée
func respondWithSuccess(c *gin.Context, data interface{}, message ...string) {
	response := gin.H{
		"success":   true,
		"data":      data,
		"timestamp": time.Now().Unix(),
	}

	if len(message) > 0 {
		response["message"] = message[0]
	}

	c.JSON(http.StatusOK, response)
}