package config

import (
	"os"
	"strconv"
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
			Port: getEnv("ANALYTICS_SERVER_PORT", "8088"),
			Host: getEnv("ANALYTICS_SERVER_HOST", "0.0.0.0"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("ANALYTICS_DB_HOST", "localhost"),
			Port:     getEnv("ANALYTICS_DB_PORT", "5432"),
			User:     getEnv("ANALYTICS_DB_USER", "auth_user"),
			Password: getEnv("ANALYTICS_DB_PASSWORD", "auth_password"),
			DBName:   getEnv("ANALYTICS_DB_NAME", "analytics_db"),
			SSLMode:  getEnv("ANALYTICS_DB_SSLMODE", "disable"),
		},
		Auth: AuthConfig{
			JWTSecret: getEnv("JWT_SECRET", "your-secret-key"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
