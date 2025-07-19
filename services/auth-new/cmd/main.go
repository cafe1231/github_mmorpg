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
	healthHandler := handlers.NewHealthHandler(cfg, db) // ← CORRECTION ICI

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
		}
	}

	return router
}

// initLogger configure le logger
func initLogger() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)
}

// gracefulShutdown gère l'arrêt propre du serveur
func gracefulShutdown(server *http.Server, authService *service.AuthService) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("🔐 Shutting down Auth Service...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logrus.Fatal("Server forced to shutdown: ", err)
	}

	logrus.Info("🔐 Auth Service stopped gracefully")
}
