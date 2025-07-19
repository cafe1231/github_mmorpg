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

// Timeout constants
const (
	HealthCheckTimeout = 5 * time.Second
)

// DB encapsulates the database connection
type DB struct {
	*sqlx.DB
}

// NewConnection creates a new database connection
func NewConnection(cfg *config.DatabaseConfig) (*DB, error) {
	// Connect to PostgreSQL
	db, err := sqlx.Connect("postgres", cfg.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Test connection
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

// Close closes the database connection
func (db *DB) Close() error {
	logrus.Info("Closing database connection")
	return db.DB.Close()
}

// Health checks the database status
func (db *DB) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), HealthCheckTimeout)
	defer cancel()

	return db.PingContext(ctx)
}

// GetStats returns database statistics
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
