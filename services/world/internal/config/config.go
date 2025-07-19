package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	JWT        JWTConfig
	Game       GameConfig
	RateLimit  interface{}
	Monitoring MonitoringConfig
}

type ServerConfig struct {
	Port         int
	Host         string
	Environment  string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DatabaseConfig struct {
	Host         string
	Port         int
	Name         string
	User         string
	Password     string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
}

type JWTConfig struct {
	Secret string
}

type GameConfig struct {
	MaxRenderDistance float64
}

type MonitoringConfig struct {
	HealthPath  string
	MetricsPath string
}

// GetDatabaseURL construit l'URL de connection à la base de données
func (d *DatabaseConfig) GetDatabaseURL() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		d.Host, d.Port, d.User, d.Password, d.Name)
}

// getEnvOrDefault récupère une variable d'environnement ou retourne une valeur par défaut
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvIntOrDefault récupère une variable d'environnement entière ou retourne une valeur par défaut
func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func LoadConfig() (*Config, error) {
	return &Config{
		Server: ServerConfig{
			Port:         getEnvIntOrDefault("WORLD_SERVER_PORT", 8083),
			Host:         getEnvOrDefault("WORLD_SERVER_HOST", "0.0.0.0"),
			Environment:  getEnvOrDefault("ENVIRONMENT", "development"),
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
		},
		Database: DatabaseConfig{
			Host:         getEnvOrDefault("WORLD_DB_HOST", "localhost"),
			Port:         getEnvIntOrDefault("WORLD_DB_PORT", 5432),
			Name:         getEnvOrDefault("WORLD_DB_NAME", "world_db"),
			User:         getEnvOrDefault("WORLD_DB_USER", "auth_user"),
			Password:     getEnvOrDefault("WORLD_DB_PASSWORD", "auth_password"),
			SSLMode:      getEnvOrDefault("WORLD_DB_SSL_MODE", "disable"),
			MaxOpenConns: getEnvIntOrDefault("WORLD_DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getEnvIntOrDefault("WORLD_DB_MAX_IDLE_CONNS", 5),
		},
		JWT: JWTConfig{
			Secret: getEnvOrDefault("JWT_SECRET", "your-super-secret-jwt-key-change-in-production-minimum-64-characters"),
		},
		Game: GameConfig{
			MaxRenderDistance: 100.0,
		},
		Monitoring: MonitoringConfig{
			HealthPath:  "/health",
			MetricsPath: "/metrics",
		},
	}, nil
}
