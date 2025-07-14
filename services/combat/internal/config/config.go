// internal/config/config.go
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/spf13/viper"
)

// Config représente la configuration du service Combat
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	Combat     CombatConfig     `mapstructure:"combat"`
	Services   ServicesConfig   `mapstructure:"services"`
	RateLimit  RateLimitConfig  `mapstructure:"rate_limit"`
	Monitoring MonitoringConfig `mapstructure:"monitoring"`
	NATS       NATSConfig       `mapstructure:"nats"`
	WebSocket  WebSocketConfig  `mapstructure:"websocket"`
}

// ServerConfig configuration du serveur
type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	Host         string        `mapstructure:"host"`
	Environment  string        `mapstructure:"environment"`
	Debug        bool          `mapstructure:"debug"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// DatabaseConfig configuration de la base de données
type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	Database     string `mapstructure:"database"`
	SSLMode      string `mapstructure:"ssl_mode"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxLifetime  time.Duration `mapstructure:"max_lifetime"`
}

// JWTConfig configuration JWT
type JWTConfig struct {
	Secret string `mapstructure:"secret"`
}

// CombatConfig configuration spécifique au combat
type CombatConfig struct {
	// Durées de combat
	MaxCombatDuration    time.Duration `mapstructure:"max_combat_duration"`
	TurnTimeout          time.Duration `mapstructure:"turn_timeout"`
	ActionTimeout        time.Duration `mapstructure:"action_timeout"`
	
	// Limites de combat
	MaxParticipants      int           `mapstructure:"max_participants"`
	MaxSimultaneousCombats int         `mapstructure:"max_simultaneous_combats"`
	
	// Calculs de dégâts
	BaseDamageMultiplier float64       `mapstructure:"base_damage_multiplier"`
	CriticalDamageBonus  float64       `mapstructure:"critical_damage_bonus"`
	LevelDifferenceBonus float64       `mapstructure:"level_difference_bonus"`
	
	// PvP
	PvPEnabled           bool          `mapstructure:"pvp_enabled"`
	PvPLevelDifference   int           `mapstructure:"pvp_level_difference"`
	PvPCooldown          time.Duration `mapstructure:"pvp_cooldown"`
	
	// Anti-cheat
	MaxActionsPerSecond  int           `mapstructure:"max_actions_per_second"`
	ValidatePositions    bool          `mapstructure:"validate_positions"`
	AntiCheatEnabled     bool          `mapstructure:"anti_cheat_enabled"`
}

// ServicesConfig configuration des services externes
type ServicesConfig struct {
	Auth   ServiceEndpoint `mapstructure:"auth"`
	Player ServiceEndpoint `mapstructure:"player"`
	World  ServiceEndpoint `mapstructure:"world"`
}

// ServiceEndpoint représente un endpoint de service
type ServiceEndpoint struct {
	URL     string        `mapstructure:"url"`
	Timeout time.Duration `mapstructure:"timeout"`
	Retries int           `mapstructure:"retries"`
}

// RateLimitConfig configuration du rate limiting
type RateLimitConfig struct {
	CombatActionsPerMinute int           `mapstructure:"combat_actions_per_minute"`
	SpellCastsPerMinute    int           `mapstructure:"spell_casts_per_minute"`
	BurstSize              int           `mapstructure:"burst_size"`
	CleanupInterval        time.Duration `mapstructure:"cleanup_interval"`
}

// MonitoringConfig configuration du monitoring
type MonitoringConfig struct {
	PrometheusPort int    `mapstructure:"prometheus_port"`
	MetricsPath    string `mapstructure:"metrics_path"`
	HealthPath     string `mapstructure:"health_path"`
}

// NATSConfig configuration NATS
type NATSConfig struct {
	URL                  string        `mapstructure:"url"`
	ClusterID            string        `mapstructure:"cluster_id"`
	ClientID             string        `mapstructure:"client_id"`
	ConnectTimeout       time.Duration `mapstructure:"connect_timeout"`
	ReconnectDelay       time.Duration `mapstructure:"reconnect_delay"`
	MaxReconnectAttempts int           `mapstructure:"max_reconnect_attempts"`
}

// WebSocketConfig configuration WebSocket
type WebSocketConfig struct {
	ReadBufferSize  int           `mapstructure:"read_buffer_size"`
	WriteBufferSize int           `mapstructure:"write_buffer_size"`
	CheckOrigin     bool          `mapstructure:"check_origin"`
	PingPeriod      time.Duration `mapstructure:"ping_period"`
	PongWait        time.Duration `mapstructure:"pong_wait"`
	WriteWait       time.Duration `mapstructure:"write_wait"`
}

// Load charge la configuration
func Load() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Port:         8084,
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
			Password:     "auth_pass",
			Database:     "combat_db",
			SSLMode:      "disable",
			MaxOpenConns: 25,
			MaxIdleConns: 5,
			MaxLifetime:  30 * time.Minute,
		},
		JWT: JWTConfig{
			Secret: "your-super-secret-jwt-key-change-in-production-minimum-64-characters",
		},
		Combat: CombatConfig{
			MaxCombatDuration:      30 * time.Minute,
			TurnTimeout:            30 * time.Second,
			ActionTimeout:          10 * time.Second,
			MaxParticipants:        20,
			MaxSimultaneousCombats: 1000,
			BaseDamageMultiplier:   1.0,
			CriticalDamageBonus:    0.5,
			LevelDifferenceBonus:   0.1,
			PvPEnabled:             true,
			PvPLevelDifference:     10,
			PvPCooldown:            5 * time.Minute,
			MaxActionsPerSecond:    3,
			ValidatePositions:      true,
			AntiCheatEnabled:       true,
		},
		Services: ServicesConfig{
			Auth: ServiceEndpoint{
				URL:     "http://auth-service:8081",
				Timeout: 5 * time.Second,
				Retries: 3,
			},
			Player: ServiceEndpoint{
				URL:     "http://player-service:8082",
				Timeout: 5 * time.Second,
				Retries: 3,
			},
			World: ServiceEndpoint{
				URL:     "http://world-service:8083",
				Timeout: 3 * time.Second,
				Retries: 2,
			},
		},
		RateLimit: RateLimitConfig{
			CombatActionsPerMinute: 180, // 3 par seconde max
			SpellCastsPerMinute:    120, // 2 par seconde max
			BurstSize:              10,
			CleanupInterval:        5 * time.Minute,
		},
		Monitoring: MonitoringConfig{
			PrometheusPort: 9094,
			MetricsPath:    "/metrics",
			HealthPath:     "/health",
		},
		NATS: NATSConfig{
			URL:                  "nats://nats:4222",
			ClusterID:            "mmorpg-cluster",
			ClientID:             "combat-service",
			ConnectTimeout:       10 * time.Second,
			ReconnectDelay:       2 * time.Second,
			MaxReconnectAttempts: 10,
		},
		WebSocket: WebSocketConfig{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     false,
			PingPeriod:      54 * time.Second,
			PongWait:        60 * time.Second,
			WriteWait:       10 * time.Second,
		},
	}

	// Charger depuis les variables d'environnement
	loadFromEnv(config)

	// Tentative de chargement depuis fichier config
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("/etc/combat/")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Fichier non trouvé, utiliser la config par défaut
	} else {
		if err := viper.Unmarshal(config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}

	return config, nil
}

// loadFromEnv charge la configuration depuis les variables d'environnement
func loadFromEnv(config *Config) {
	// Serveur
	if port := os.Getenv("COMBAT_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Server.Port = p
		}
	}

	if host := os.Getenv("COMBAT_HOST"); host != "" {
		config.Server.Host = host
	}

	if env := os.Getenv("COMBAT_ENVIRONMENT"); env != "" {
		config.Server.Environment = env
	}

	// Base de données
	if dbHost := os.Getenv("COMBAT_DB_HOST"); dbHost != "" {
		config.Database.Host = dbHost
	}

	if dbPort := os.Getenv("COMBAT_DB_PORT"); dbPort != "" {
		if p, err := strconv.Atoi(dbPort); err == nil {
			config.Database.Port = p
		}
	}

	if dbUser := os.Getenv("COMBAT_DB_USER"); dbUser != "" {
		config.Database.User = dbUser
	}

	if dbPass := os.Getenv("COMBAT_DB_PASSWORD"); dbPass != "" {
		config.Database.Password = dbPass
	}

	if dbName := os.Getenv("COMBAT_DB_NAME"); dbName != "" {
		config.Database.Database = dbName
	}

	// JWT
	if jwtSecret := os.Getenv("COMBAT_JWT_SECRET"); jwtSecret != "" {
		config.JWT.Secret = jwtSecret
	}

	// Services
	if authURL := os.Getenv("AUTH_SERVICE_URL"); authURL != "" {
		config.Services.Auth.URL = authURL
	}

	if playerURL := os.Getenv("PLAYER_SERVICE_URL"); playerURL != "" {
		config.Services.Player.URL = playerURL
	}

	if worldURL := os.Getenv("WORLD_SERVICE_URL"); worldURL != "" {
		config.Services.World.URL = worldURL
	}

	// NATS
	if natsURL := os.Getenv("NATS_URL"); natsURL != "" {
		config.NATS.URL = natsURL
	}
}