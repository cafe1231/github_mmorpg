// auth/cmd/main.go
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

	"auth/internal/config"
	"auth/internal/database"
	"auth/internal/handlers"
	"auth/internal/middleware"
	"auth/internal/repository"
	"auth/internal/service"
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
		"service":    "auth",
		"version":    Version,
		"build_time": BuildTime,
		"git_commit": GitCommit,
	}).Info("🔐 Starting Auth Service...")

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
	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)

	// Démarrage du nettoyage périodique des sessions
	sessionRepo.ScheduleCleanup(30 * time.Minute)

	// Initialisation des services
	authService := service.NewAuthService(userRepo, sessionRepo, cfg)

	// Initialisation des handlers
	authHandler := handlers.NewAuthHandler(authService, cfg)
	healthHandler := handlers.NewHealthHandler(db)

	// Configuration du mode Gin
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Configuration des routes
	router := setupRoutes(authHandler, healthHandler, cfg)

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
		}).Info("🔐 Auth Service started successfully")

		var err error
		if cfg.Server.TLSCertFile != "" && cfg.Server.TLSKeyFile != "" {
			err = server.ListenAndServeTLS(cfg.Server.TLSCertFile, cfg.Server.TLSKeyFile)
		} else {
			err = server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			logrus.Fatal("Failed to start server: ", err)
		}
	}()

	// Gestion gracieuse de l'arrêt
	gracefulShutdown(server, authService)
}

// setupRoutes configure toutes les routes du service Auth
func setupRoutes(
	authHandler *handlers.AuthHandler,
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

	// Rate limiting si configuré
	if cfg.RateLimit.Global.Requests > 0 {
		router.Use(middleware.GlobalRateLimit(cfg.RateLimit.Global))
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
		}
	}

	// API v1
	v1 := router.Group("/api/v1")
	{
		// Routes publiques d'authentification (pas d'auth JWT requise)
		auth := v1.Group("/auth")
		{
			// Rate limiting spécialisé pour l'auth
			auth.Use(middleware.AuthRateLimit(cfg.RateLimit))

			// Inscription et connexion
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			
			// Récupération de mot de passe
			auth.POST("/forgot-password", authHandler.ForgotPassword)
			auth.POST("/reset-password", authHandler.ResetPassword)
			
			// Vérification d'email
			auth.GET("/verify-email/:token", authHandler.VerifyEmail)
			auth.POST("/resend-verification", authHandler.ResendVerification)

			// OAuth endpoints (si configuré)
			if cfg.OAuth.Google.Enabled || cfg.OAuth.Discord.Enabled || cfg.OAuth.GitHub.Enabled {
				oauth := auth.Group("/oauth")
				{
					oauth.GET("/:provider", authHandler.OAuthRedirect)
					oauth.POST("/:provider/callback", authHandler.OAuthCallback)
				}
			}
		}

		// Routes protégées (authentification JWT requise)
		protected := v1.Group("/")
		protected.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			// Gestion du profil utilisateur
			user := protected.Group("/user")
			{
				user.GET("/profile", authHandler.GetProfile)
				user.PUT("/profile", authHandler.UpdateProfile)
				user.POST("/change-password", authHandler.ChangePassword)
				user.POST("/logout", authHandler.Logout)
				user.POST("/logout-all", authHandler.LogoutAllDevices)
				user.GET("/sessions", authHandler.GetSessions)
				user.DELETE("/sessions/:id", authHandler.RevokeSession)
			}

			// Two-Factor Authentication (si activé)
			if cfg.Security.TwoFactorRequired {
				twofa := protected.Group("/2fa")
				{
					twofa.POST("/enable", authHandler.EnableTwoFactor)
					twofa.POST("/disable", authHandler.DisableTwoFactor)
					twofa.GET("/qr", authHandler.GetTwoFactorQR)
					twofa.POST("/verify", authHandler.VerifyTwoFactor)
					twofa.GET("/backup-codes", authHandler.GetBackupCodes)
					twofa.POST("/regenerate-codes", authHandler.RegenerateBackupCodes)
				}
			}

			// Routes admin (réservées aux admins et superusers)
			admin := protected.Group("/admin")
			admin.Use(middleware.RequireRole("admin", "superuser"))
			{
				// Gestion des utilisateurs
				admin.GET("/users", authHandler.ListUsers)
				admin.GET("/users/:id", authHandler.GetUser)
				admin.PUT("/users/:id", authHandler.UpdateUser)
				admin.POST("/users", authHandler.CreateUser)
				admin.DELETE("/users/:id", authHandler.DeleteUser)
				
				// Gestion des statuts
				admin.POST("/users/:id/suspend", authHandler.SuspendUser)
				admin.POST("/users/:id/activate", authHandler.ActivateUser)
				admin.POST("/users/:id/ban", authHandler.BanUser)
				admin.POST("/users/:id/unban", authHandler.UnbanUser)
				
				// Logs et audit
				admin.GET("/login-attempts", authHandler.GetLoginAttempts)
				admin.GET("/audit-log", authHandler.GetAuditLog)
				admin.GET("/statistics", authHandler.GetStatistics)
				
				// Gestion des sessions
				admin.GET("/sessions", authHandler.GetAllSessions)
				admin.DELETE("/sessions/:id", authHandler.AdminRevokeSession)
				admin.POST("/users/:id/logout-all", authHandler.AdminLogoutUser)
			}

			// Routes moderator (pour les modérateurs)
			moderator := protected.Group("/moderator")
			moderator.Use(middleware.RequireRole("moderator", "admin", "superuser"))
			{
				moderator.GET("/users/search", authHandler.SearchUsers)
				moderator.GET("/users/:id/sessions", authHandler.GetUserSessionsAdmin)
				moderator.POST("/users/:id/warnings", authHandler.AddUserWarning)
				moderator.GET("/users/:id/warnings", authHandler.GetUserWarnings)
			}
		}

		// Routes de validation pour les autres services (authentification de service)
		services := v1.Group("/services")
		services.Use(middleware.ServiceAuth()) // Middleware spécial pour les services internes
		{
			services.POST("/validate", authHandler.ValidateToken)
			services.GET("/user/:id", authHandler.GetUserInfo)
			services.POST("/user/:id/update-activity", authHandler.UpdateUserActivity)
			services.GET("/user/:id/permissions", authHandler.GetUserPermissions)
		}
	}

	// Routes directes pour compatibilité avec le gateway
	router.POST("/register", authHandler.Register)
	router.POST("/login", authHandler.Login)
	router.POST("/refresh", authHandler.RefreshToken)
	router.GET("/health", healthHandler.HealthCheck)
	router.GET("/metrics", healthHandler.Metrics)

	return router
}

// initLogger initialise le logger global
func initLogger() {
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})

	// Niveau de log selon l'environnement
	level := os.Getenv("AUTH_LOG_LEVEL")
	switch level {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		if os.Getenv("AUTH_ENVIRONMENT") == "production" {
			logrus.SetLevel(logrus.InfoLevel)
		} else {
			logrus.SetLevel(logrus.DebugLevel)
		}
	}

	logrus.WithField("service", "auth").Info("Logger initialized")
}

// gracefulShutdown gère l'arrêt gracieux du serveur
func gracefulShutdown(server *http.Server, authService service.AuthServiceInterface) {
	// Canal pour capturer les signaux système
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Attendre un signal
	<-quit
	logrus.Info("🛑 Auth Service shutting down...")

	// Timeout pour l'arrêt gracieux
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Arrêter d'accepter de nouvelles connexions
	if err := server.Shutdown(ctx); err != nil {
		logrus.Error("Auth Service forced to shutdown:", err)
		return
	}

	// Nettoyage des ressources
	if authService != nil {
		logrus.Info("Cleaning up auth service resources...")
		
		// Révoquer les sessions actives si nécessaire
		// authService.CleanupOnShutdown()
	}

	logrus.Info("✅ Auth Service stopped gracefully")
}

// Fonctions utilitaires pour le déploiement

// GetBuildInfo retourne les informations de build
func GetBuildInfo() map[string]string {
	return map[string]string{
		"version":    Version,
		"build_time": BuildTime,
		"git_commit": GitCommit,
		"go_version": fmt.Sprintf("%s %s/%s", 
			os.Getenv("GO_VERSION"), 
			os.Getenv("GOOS"), 
			os.Getenv("GOARCH")),
	}
}

// HealthCheckEndpoint endpoint simple pour les load balancers
func HealthCheckEndpoint(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "auth",
		"version":   Version,
		"timestamp": time.Now().Unix(),
	})
}

// PreflightHandler gère les requêtes OPTIONS pour CORS
func PreflightHandler(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Request-ID")
	c.Header("Access-Control-Max-Age", "86400")
	c.Status(http.StatusNoContent)
}

// NotFoundHandler gère les routes non trouvées
func NotFoundHandler(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"error":      "Not Found",
		"message":    "The requested endpoint does not exist",
		"path":       c.Request.URL.Path,
		"method":     c.Request.Method,
		"timestamp":  time.Now(),
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

// MethodNotAllowedHandler gère les méthodes HTTP non autorisées
func MethodNotAllowedHandler(c *gin.Context) {
	c.JSON(http.StatusMethodNotAllowed, gin.H{
		"error":      "Method Not Allowed",
		"message":    fmt.Sprintf("Method %s is not allowed for this endpoint", c.Request.Method),
		"path":       c.Request.URL.Path,
		"method":     c.Request.Method,
		"timestamp":  time.Now(),
		"request_id": c.GetHeader("X-Request-ID"),
	})
}