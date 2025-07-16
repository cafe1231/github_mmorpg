package config

import (
	"fmt"
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

// GetDatabaseURL construit l'URL de connexion à la base de données
func (d *DatabaseConfig) GetDatabaseURL() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		d.Host, d.Port, d.User, d.Password, d.Name)
}

func LoadConfig() (*Config, error) {
	return &Config{
		Server: ServerConfig{
			Port:         8083,
			Host:         "0.0.0.0",
			Environment:  "development",
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
		},
		Database: DatabaseConfig{
			Host:         "localhost",
			Port:         5432,
			Name:         "world_db",
			User:         "auth_user",
			Password:     "auth_password",
			SSLMode:      "disable",
			MaxOpenConns: 25,
			MaxIdleConns: 5,
		},
		JWT: JWTConfig{
			Secret: "your-super-secret-jwt-key-change-in-production-minimum-64-characters",
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
