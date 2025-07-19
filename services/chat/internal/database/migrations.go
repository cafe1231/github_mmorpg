package database

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
)

// Migration représente une migration de base de données
type Migration struct {
	Version    int
	Name       string
	UpScript   string
	DownScript string
}

// RunMigrations exécute toutes les migrations en attente
func RunMigrations(db *sql.DB) error {
	logrus.Info("Starting database migrations...")

	// Créer la table des migrations si elle n'existe pas
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Charger les migrations depuis les fichiers
	migrations, err := loadMigrationsFromFiles()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	// Obtenir la version actuelle
	currentVersion, err := getCurrentVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	// Exécuter les migrations en attente
	for _, migration := range migrations {
		if migration.Version > currentVersion {
			logrus.WithFields(logrus.Fields{
				"version": migration.Version,
				"name":    migration.Name,
			}).Info("Running migration...")

			if err := runMigration(db, migration); err != nil {
				return fmt.Errorf("failed to run migration %d: %w", migration.Version, err)
			}

			logrus.WithField("version", migration.Version).Info("Migration completed successfully")
		}
	}

	logrus.Info("All migrations completed successfully")
	return nil
}

// createMigrationsTable crée la table pour suivre les migrations
func createMigrationsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`
	_, err := db.Exec(query)
	return err
}

// getCurrentVersion obtient la version actuelle de la base de données
func getCurrentVersion(db *sql.DB) (int, error) {
	var version int
	err := db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}

// runMigration exécute une migration spécifique
func runMigration(db *sql.DB, migration Migration) error {
	// Commencer une transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			logrus.WithError(err).Warn("Erreur lors du rollback")
		}
	}()

	// Exécuter le script de migration
	if _, err := tx.Exec(migration.UpScript); err != nil {
		return fmt.Errorf("failed to execute migration script: %w", err)
	}

	// Enregistrer la migration
	if _, err := tx.Exec(
		"INSERT INTO schema_migrations (version, name) VALUES ($1, $2)",
		migration.Version, migration.Name,
	); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Valider la transaction
	return tx.Commit()
}

// loadMigrationsFromFiles charge les migrations depuis le dossier migrations/
func loadMigrationsFromFiles() ([]Migration, error) {
	migrationsDir := "migrations"
	files, err := ioutil.ReadDir(migrationsDir)
	if err != nil {
		// Si le dossier n'existe pas, retourner une liste vide
		return []Migration{}, nil
	}

	migrations := make(map[int]Migration)

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		// Parser le nom du fichier (format: 001_name.up.sql ou 001_name.down.sql)
		parts := strings.Split(file.Name(), "_")
		if len(parts) < 2 {
			continue
		}

		var version int
		if _, err := fmt.Sscanf(parts[0], "%d", &version); err != nil {
			continue
		}

		content, err := ioutil.ReadFile(filepath.Join(migrationsDir, file.Name()))
		if err != nil {
			return nil, err
		}

		migration := migrations[version]
		migration.Version = version

		if strings.Contains(file.Name(), ".up.sql") {
			migration.UpScript = string(content)
			// Extraire le nom de la migration
			nameParts := strings.Split(file.Name(), ".up.sql")
			if len(nameParts) > 0 {
				nameWithoutVersion := strings.Join(strings.Split(nameParts[0], "_")[1:], "_")
				migration.Name = nameWithoutVersion
			}
		} else if strings.Contains(file.Name(), ".down.sql") {
			migration.DownScript = string(content)
		}

		migrations[version] = migration
	}

	// Convertir la map en slice et trier
	var result []Migration
	for _, migration := range migrations {
		if migration.UpScript != "" { // Ignorer les migrations sans script up
			result = append(result, migration)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Version < result[j].Version
	})

	return result, nil
}
