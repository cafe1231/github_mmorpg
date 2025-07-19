package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/spf13/viper"
)

// Config représente la configuration du service Player
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Auth       AuthConfig       `mapstructure:"auth"`
	RateLimit  RateLimitConfig  `mapstructure:"rate_limit"`
	Monitoring MonitoringConfig `mapstructure:"monitoring"`
	Game       GameConfig       `mapstructure:"game"`
}

// ServerConfig configuration du serveur Player
type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	Host         string        `mapstructure:"host"`
	Environment  string        `mapstructure:"environment"`
	Debug        bool          `mapstructure:"debug"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// DatabaseConfig configuration PostgreSQL
type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	Name         string `mapstructure:"name"`
	SSLMode      string `mapstructure:"ssl_mode"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

// AuthConfig configuration pour la communication avec le service Auth
type AuthConfig struct {
	ServiceURL string `mapstructure:"service_url"`
	JWTSecret  string `mapstructure:"jwt_secret"`
}

// RateLimitConfig configuration rate limiting
type RateLimitConfig struct {
	RequestsPerMinute int           `mapstructure:"requests_per_minute"`
	BurstSize         int           `mapstructure:"burst_size"`
	CleanupInterval   time.Duration `mapstructure:"cleanup_interval"`
}

// MonitoringConfig configuration monitoring
type MonitoringConfig struct {
	PrometheusPort int    `mapstructure:"prometheus_port"`
	MetricsPath    string `mapstructure:"metrics_path"`
	HealthPath     string `mapstructure:"health_path"`
}

// GameConfig configuration spécifique au jeu
type GameConfig struct {
	MaxCharactersPerPlayer int        `mapstructure:"max_characters_per_player"`
	AvailableClasses       []string   `mapstructure:"available_classes"`
	AvailableRaces         []string   `mapstructure:"available_races"`
	MaxLevel               int        `mapstructure:"max_level"`
	StartingLevel          int        `mapstructure:"starting_level"`
	StartingStats          StatConfig `mapstructure:"starting_stats"`
}

// StatConfig configuration des statistiques de base
type StatConfig struct {
	Health       int `mapstructure:"health"`
	Mana         int `mapstructure:"mana"`
	Strength     int `mapstructure:"strength"`
	Agility      int `mapstructure:"agility"`
	Intelligence int `mapstructure:"intelligence"`
	Vitality     int `mapstructure:"vitality"`
}

// LoadConfig charge la configuration
func LoadConfig() (*Config, error) {
	// Configuration par défaut
	config := &Config{
		Server: ServerConfig{
			Port:         8082,
			Host:         "0.0.0.0",
			Environment:  "development",
			Debug:        true,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		Database: DatabaseConfig{
			Host:         "localhost",
			Port:         5432,
			User:         "auth_user",
			Password:     "auth_password",
			Name:         "player_db",
			SSLMode:      "disable",
			MaxOpenConns: 25,
			MaxIdleConns: 5,
		},
		Auth: AuthConfig{
			ServiceURL: "http://localhost:8081",
			JWTSecret:  "your-super-secret-jwt-key-change-in-production-minimum-64-characters",
		},
		RateLimit: RateLimitConfig{
			RequestsPerMinute: 100,
			BurstSize:         20,
			CleanupInterval:   5 * time.Minute,
		},
		Monitoring: MonitoringConfig{
			PrometheusPort: 9092,
			MetricsPath:    "/metrics",
			HealthPath:     "/health",
		},
		Game: GameConfig{
			MaxCharactersPerPlayer: 5,
			AvailableClasses:       []string{"warrior", "mage", "archer", "rogue"},
			AvailableRaces:         []string{"human", "elf", "dwarf", "orc"},
			MaxLevel:               100,
			StartingLevel:          1,
			StartingStats: StatConfig{
				Health:       100,
				Mana:         50,
				Strength:     10,
				Agility:      10,
				Intelligence: 10,
				Vitality:     10,
			},
		},
	}

	// Charger depuis les variables d'environnement
	loadFromEnv(config)

	// Tentative de chargement depuis fichier config
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/player/")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	} else {
		if err := viper.Unmarshal(config); err != nil {
			return nil, fmt.Errorf("error unmarshalling config: %w", err)
		}
	}

	// Validation
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// loadFromEnv charge depuis les variables d'environnement
func loadFromEnv(config *Config) {
	// Server
	if port := os.Getenv("PLAYER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Server.Port = p
		}
	}
	if host := os.Getenv("PLAYER_HOST"); host != "" {
		config.Server.Host = host
	}
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		config.Server.Environment = env
	}

	// Database
	if host := os.Getenv("PLAYER_DB_HOST"); host != "" {
		config.Database.Host = host
	}
	if port := os.Getenv("PLAYER_DB_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Database.Port = p
		}
	}
	if user := os.Getenv("PLAYER_DB_USER"); user != "" {
		config.Database.User = user
	}
	if password := os.Getenv("PLAYER_DB_PASSWORD"); password != "" {
		config.Database.Password = password
	}
	if name := os.Getenv("PLAYER_DB_NAME"); name != "" {
		config.Database.Name = name
	}

	// Auth
	if authURL := os.Getenv("AUTH_SERVICE_URL"); authURL != "" {
		config.Auth.ServiceURL = authURL
	}
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		config.Auth.JWTSecret = secret
	}
}

// validateConfig valide la configuration
func validateConfig(config *Config) error {
	// Validation serveur
	if config.Server.Port < 1 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	// Validation base de données
	if config.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if config.Database.User == "" {
		return fmt.Errorf("database user is required")
	}
	if config.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}

	// Validation Auth
	if config.Auth.ServiceURL == "" {
		return fmt.Errorf("auth service URL is required")
	}
	if len(config.Auth.JWTSecret) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 characters long")
	}

	// Validation Game
	if config.Game.MaxCharactersPerPlayer < 1 {
		return fmt.Errorf("max characters per player must be at least 1")
	}
	if config.Game.MaxLevel < 1 {
		return fmt.Errorf("max level must be at least 1")
	}

	return nil
}

// GetDatabaseURL retourne l'URL de connection PostgreSQL
func (c *DatabaseConfig) GetDatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Name, c.SSLMode)
}
