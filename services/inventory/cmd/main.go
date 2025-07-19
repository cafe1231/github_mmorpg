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

func main() {
	// Initialisation du logging
	initLogger()

	// Chargement de la configuration
	cfg, err := config.Load()
	if err != nil {
		logrus.Fatal("Failed to load configuration: ", err)
	}

	// connection à la base de données
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
	itemRepo := repository.NewItemRepository(db.DB)
	inventoryRepo := repository.NewInventoryRepository(db.DB)

	// Initialisation des services
	inventoryService := service.NewInventoryService(inventoryRepo, itemRepo)

	// Initialisation des handlers
	inventoryHandler := handlers.NewInventoryHandler(inventoryService)
	healthHandler := handlers.NewHealthHandler("development")

	// Configuration du mode Gin
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Création du routeur
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Configuration CORS pour le développement
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

	// Routes de santé
	router.GET("/health", healthHandler.Health)
	router.GET("/health/ready", healthHandler.Readiness)
	router.GET("/health/live", healthHandler.Liveness)

	// Routes d'API
	apiV1 := router.Group("/api/v1")
	{
		// Routes d'inventaire
		inventory := apiV1.Group("/inventory")
		{
			inventory.GET("/:characterId", inventoryHandler.GetInventory)

			// Gestion des objets
			inventory.POST("/:characterId/items", inventoryHandler.AddItem)
			inventory.DELETE("/:characterId/items/:itemId", inventoryHandler.RemoveItem)
			inventory.PUT("/:characterId/items/:itemId", inventoryHandler.UpdateItem)
			inventory.GET("/:characterId/items", inventoryHandler.ListItems)

			// Opérations sur les slots
			inventory.POST("/:characterId/move", inventoryHandler.MoveItem)
			inventory.POST("/:characterId/split", inventoryHandler.SplitStack)

			// Opérations en lot
			inventory.POST("/:characterId/items/bulk/add", inventoryHandler.AddBulkItems)
			inventory.POST("/:characterId/items/bulk/remove", inventoryHandler.RemoveBulkItems)
		}
	}

	// Démarrage du serveur
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router,
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

	// Attendre le signal d'arrêt
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down server...")

	// Arrêt gracieux avec timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
