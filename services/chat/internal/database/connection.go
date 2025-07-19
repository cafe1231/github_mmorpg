package database

import (
	"chat/internal/config"
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

const (
	DefaultDBTimeout = 5
)

// DB structure pour encapsuler la connection
type DB struct {
	*sql.DB
}

// NewConnection crée une nouvelle connection à la base de données
func NewConnection(cfg *config.DatabaseConfig) (*DB, error) {
	logrus.Info("Connecting to PostgreSQL database...")

	db, err := sql.Open("postgres", cfg.GetDatabaseURL())
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configuration du pool de connections
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.MaxLifetime)

	// Test de la connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"host":           cfg.Host,
		"port":           cfg.Port,
		"database":       cfg.Name,
		"max_open_conns": cfg.MaxOpenConns,
		"max_idle_conns": cfg.MaxIdleConns,
	}).Info("Successfully connected to database")

	return &DB{db}, nil
}

// Close ferme la connection à la base de données
func (db *DB) Close() error {
	logrus.Info("Closing database connection...")
	return db.DB.Close()
}

// HealthCheck vérifie la santé de la connection
func (db *DB) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultDBTimeout*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// GetStats retourne les statistiques de la base de données
func (db *DB) GetStats() sql.DBStats {
	return db.Stats()
}
