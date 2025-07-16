package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"inventory/internal/config"
)

// DB encapsule la connexion à la base de données
type DB struct {
	*sqlx.DB
}

// NewConnection crée une nouvelle connexion à la base de données
func NewConnection(cfg config.DatabaseConfig) (*DB, error) {
	// Connexion à PostgreSQL
	db, err := sqlx.Connect("postgres", cfg.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configuration de la pool de connexions
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Test de la connexion
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"host":              cfg.Host,
		"port":              cfg.Port,
		"database":          cfg.Name,
		"max_open_conns":    cfg.MaxOpenConns,
		"max_idle_conns":    cfg.MaxIdleConns,
		"conn_max_lifetime": cfg.ConnMaxLifetime,
	}).Info("Connected to database")

	return &DB{db}, nil
}

// Close ferme la connexion à la base de données
func (db *DB) Close() error {
	logrus.Info("Closing database connection")
	return db.DB.Close()
}

// Health vérifie l'état de la base de données
func (db *DB) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return db.PingContext(ctx)
}

// GetStats retourne les statistiques de la base de données
func (db *DB) GetStats() map[string]interface{} {
	stats := db.DB.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}
}
