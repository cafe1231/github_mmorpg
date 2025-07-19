package config

import (
	"os"
)

// Config contient la configuration du service guild
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
}

// ServerConfig contient la configuration du serveur
type ServerConfig struct {
	Port string
	Host string
}

// DatabaseConfig contient la configuration de la base de données
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// AuthConfig contient la configuration d'authentification
type AuthConfig struct {
	JWTSecret string
}

// Load charge la configuration depuis les variables d'environnement
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("GUILD_SERVER_PORT", "8086"),
			Host: getEnv("GUILD_SERVER_HOST", "0.0.0.0"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("GUILD_DB_HOST", "localhost"),
			Port:     getEnv("GUILD_DB_PORT", "5432"),
			User:     getEnv("GUILD_DB_USER", "auth_user"),
			Password: getEnv("GUILD_DB_PASSWORD", "auth_password"),
			DBName:   getEnv("GUILD_DB_NAME", "guild_db"),
			SSLMode:  getEnv("GUILD_DB_SSLMODE", "disable"),
		},
		Auth: AuthConfig{
			JWTSecret: getEnv("JWT_SECRET", "your-secret-key"),
		},
	}
}

// getEnv récupère une variable d'environnement avec une valeur par défaut
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
