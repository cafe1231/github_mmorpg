package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"auth/internal/config"
)

// NewConnection crée une nouvelle connexion à la base de données
func NewConnection(cfg config.DatabaseConfig) (*sqlx.DB, error) {
	// Construction de l'URL de connexion
	dsn := cfg.GetDatabaseURL()

	logrus.WithFields(logrus.Fields{
		"host":     cfg.Host,
		"port":     cfg.Port,
		"database": cfg.Name,
		"username": cfg.Username,
	}).Info("Connecting to PostgreSQL database...")

	// Connexion à la base de données
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configuration de la pool de connexions
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.MaxLifetime)

	// Test de la connexion
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"host":           cfg.Host,
		"port":           cfg.Port,
		"database":       cfg.Name,
		"max_open_conns": cfg.MaxOpenConns,
		"max_idle_conns": cfg.MaxIdleConns,
		"service":        "auth",
	}).Info("Connected to PostgreSQL database successfully")

	return db, nil
}

// RunMigrations exécute les migrations de base de données
func RunMigrations(db *sqlx.DB) error {
	logrus.Info("Running auth database migrations...")

	// Liste des migrations (référencées depuis migrations.go)
	migrationList := []string{
		createUsersTable,
		createUserSessionsTable,
		createOAuthAccountsTable,
		createLoginAttemptsTable,
		createEmailVerificationsTable,
		createPasswordResetsTable,
		createTwoFactorBackupCodesTable,
		createAuditLogsTable,
		createIndexes,
		createTriggers,
		createConstraintsAndViews,
	}

	// Exécuter chaque migration
	for i, migration := range migrationList {
		logrus.WithField("migration", i+1).Debug("Executing migration")

		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("failed to execute migration %d: %w", i+1, err)
		}
	}

	logrus.Info("Auth database migrations completed successfully")
	return nil
}
