package config

import (
	"os"
)

const (
	// Health status constants
	HealthStatusHealthy   = "healthy"
	HealthStatusUnhealthy = "unhealthy"

	// Damage type constants
	DamageTypeMagical = "magical"

	// HTTP method constants
	HTTPMethodPOST = "POST"

	// Action type constants
	ActionTypeUnknown = "unknown"

	// JSON constants
	EmptyJSON = "{}"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
}

type ServerConfig struct {
	Port string
	Host string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type AuthConfig struct {
	JWTSecret string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnvString("ANALYTICS_SERVER_PORT", "8088"),
			Host: getEnvString("ANALYTICS_SERVER_HOST", "0.0.0.0"),
		},
		Database: DatabaseConfig{
			Host:     getEnvString("ANALYTICS_DB_HOST", "localhost"),
			Port:     getEnvString("ANALYTICS_DB_PORT", "5432"),
			User:     getEnvString("ANALYTICS_DB_USER", "auth_user"),
			Password: getEnvString("ANALYTICS_DB_PASSWORD", "auth_password"),
			DBName:   getEnvString("ANALYTICS_DB_NAME", "analytics_db"),
			SSLMode:  getEnvString("ANALYTICS_DB_SSLMODE", "disable"),
		},
		Auth: AuthConfig{
			JWTSecret: getEnvString("JWT_SECRET", "your-secret-key"),
		},
	}
}

func getEnvString(key string, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
