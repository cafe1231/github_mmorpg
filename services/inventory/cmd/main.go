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

	"inventory/internal/config"
	"inventory/internal/database"
	"inventory/internal/handlers"
	"inventory/internal/repository"
	"inventory/internal/service"
)

// Constantes pour les timeouts du serveur
const (
	ReadHeaderTimeout = 10 * time.Second
	ReadTimeout       = 30 * time.Second
	WriteTimeout      = 30 * time.Second
	IdleTimeout       = 60 * time.Second
	ShutdownTimeout   = 30 * time.Second
)

func main() {
	// Initialize logging
	initLogger()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logrus.Fatal("Failed to load configuration: ", err)
	}

	// Database connection
	db, err := database.NewConnection(&cfg.Database)
	if err != nil {
		logrus.Fatal("Failed to connect to database: ", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		logrus.Fatal("Failed to run migrations: ", err)
	}

	// Initialize repositories
	itemRepo := repository.NewItemRepository(db.DB)
	inventoryRepo := repository.NewInventoryRepository(db.DB)

	// Initialize services
	inventoryService := service.NewInventoryService(inventoryRepo, itemRepo)

	// Initialize handlers
	inventoryHandler := handlers.NewInventoryHandler(inventoryService)
	healthHandler := handlers.NewHealthHandler("development")

	// Configure Gin mode
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// CORS configuration for development
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

	// Health routes
	router.GET("/health", healthHandler.Health)
	router.GET("/health/ready", healthHandler.Readiness)
	router.GET("/health/live", healthHandler.Liveness)

	// API routes
	apiV1 := router.Group("/api/v1")
	{
		// Inventory routes
		inventory := apiV1.Group("/inventory")
		{
			inventory.GET("/:characterId", inventoryHandler.GetInventory)

			// Item management
			inventory.POST("/:characterId/items", inventoryHandler.AddItem)
			inventory.DELETE("/:characterId/items/:itemId", inventoryHandler.RemoveItem)
			inventory.PUT("/:characterId/items/:itemId", inventoryHandler.UpdateItem)
			inventory.GET("/:characterId/items", inventoryHandler.ListItems)

			// Slot operations
			inventory.POST("/:characterId/move", inventoryHandler.MoveItem)
			inventory.POST("/:characterId/split", inventoryHandler.SplitStack)

			// Bulk operations
			inventory.POST("/:characterId/items/bulk/add", inventoryHandler.AddBulkItems)
			inventory.POST("/:characterId/items/bulk/remove", inventoryHandler.RemoveBulkItems)
		}
	}

	// Server startup
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:           router,
		ReadHeaderTimeout: ReadHeaderTimeout,
		ReadTimeout:       ReadTimeout,
		WriteTimeout:      WriteTimeout,
		IdleTimeout:       IdleTimeout,
	}

	go func() {
		logrus.WithFields(logrus.Fields{
			"port":        cfg.Server.Port,
			"environment": cfg.Server.Environment,
		}).Info("Starting inventory service")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatal("Failed to start server: ", err)
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logrus.Fatal("Server forced to shutdown: ", err)
	}

	logrus.Info("Server exiting")
}

func initLogger() {
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})
}
