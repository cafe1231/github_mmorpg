package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/spf13/viper"
)

// Config représente la configuration du Gateway
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	Services   ServicesConfig   `mapstructure:"services"`
	RateLimit  RateLimitConfig  `mapstructure:"rate_limit"`
	Monitoring MonitoringConfig `mapstructure:"monitoring"`
	NATS       NATSConfig       `mapstructure:"nats"`
}

// ServerConfig configuration du serveur Gateway
type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	Host         string        `mapstructure:"host"`
	Environment  string        `mapstructure:"environment"`
	Debug        bool          `mapstructure:"debug"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// JWTConfig configuration JWT
type JWTConfig struct {
	Secret         string        `mapstructure:"secret"`
	ExpirationTime time.Duration `mapstructure:"expiration_time"`
}

// ServicesConfig configuration des services backend
type ServicesConfig struct {
	Auth      ServiceEndpoint `mapstructure:"auth"`
	Player    ServiceEndpoint `mapstructure:"player"`
	World     ServiceEndpoint `mapstructure:"world"`
	Combat    ServiceEndpoint `mapstructure:"combat"`
	Inventory ServiceEndpoint `mapstructure:"inventory"`
	Guild     ServiceEndpoint `mapstructure:"guild"`
	Chat      ServiceEndpoint `mapstructure:"chat"`
	Analytics ServiceEndpoint `mapstructure:"analytics"`
}

// ServiceEndpoint représente un endpoint de service
type ServiceEndpoint struct {
	URL     string        `mapstructure:"url"`
	Timeout time.Duration `mapstructure:"timeout"`
	Retries int           `mapstructure:"retries"`
}

// RateLimitConfig configuration du rate limiting
type RateLimitConfig struct {
	RequestsPerMinute int           `mapstructure:"requests_per_minute"`
	BurstSize         int           `mapstructure:"burst_size"`
	CleanupInterval   time.Duration `mapstructure:"cleanup_interval"`
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

// LoadConfig charge la configuration depuis les variables d'environnement et fichiers
func LoadConfig() (*Config, error) {
	// Configuration par défaut - LOCALHOST pour développement
	config := &Config{
		Server: ServerConfig{
			Port:         8080,
			Host:         "0.0.0.0",
			Environment:  "development",
			Debug:        true,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		JWT: JWTConfig{
			Secret:         "your-super-secret-jwt-key-change-in-production-minimum-64-characters",
			ExpirationTime: 24 * time.Hour,
		},
		Services: ServicesConfig{
			Auth: ServiceEndpoint{
				URL:     "http://localhost:8081", // Changé de auth-service vers localhost
				Timeout: 10 * time.Second,
				Retries: 3,
			},
			Player: ServiceEndpoint{
				URL:     "http://localhost:8082", // Changé de player-service vers localhost
				Timeout: 10 * time.Second,
				Retries: 3,
			},
			World: ServiceEndpoint{
				URL:     "http://localhost:8083", // Changé de world-service vers localhost
				Timeout: 5 * time.Second,
				Retries: 2,
			},
			Combat: ServiceEndpoint{
				URL:     "http://localhost:8084", // Changé de combat-service vers localhost
				Timeout: 3 * time.Second,
				Retries: 1,
			},
			Inventory: ServiceEndpoint{
				URL:     "http://localhost:8085", // Changé de inventory-service vers localhost
				Timeout: 10 * time.Second,
				Retries: 3,
			},
			Guild: ServiceEndpoint{
				URL:     "http://localhost:8086", // Changé de guild-service vers localhost
				Timeout: 10 * time.Second,
				Retries: 3,
			},
			Chat: ServiceEndpoint{
				URL:     "http://localhost:8087", // Changé de chat-service vers localhost
				Timeout: 5 * time.Second,
				Retries: 2,
			},
			Analytics: ServiceEndpoint{
				URL:     "http://localhost:8088", // Changé de analytics-service vers localhost
				Timeout: 15 * time.Second,
				Retries: 1,
			},
		},
		RateLimit: RateLimitConfig{
			RequestsPerMinute: 1000,
			BurstSize:         100,
			CleanupInterval:   5 * time.Minute,
		},
		Monitoring: MonitoringConfig{
			PrometheusPort: 9090,
			MetricsPath:    "/metrics",
			HealthPath:     "/health",
		},
		NATS: NATSConfig{
			URL:                  "nats://localhost:4222", // Changé de nats vers localhost
			ClusterID:            "mmorpg-cluster",
			ClientID:             "gateway-service",
			ConnectTimeout:       10 * time.Second,
			ReconnectDelay:       2 * time.Second,
			MaxReconnectAttempts: 10,
		},
	}

	// Charger depuis les variables d'environnement
	loadFromEnv(config)

	// Tentative de chargement depuis fichier config
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/gateway/")

	if err := viper.ReadInConfig(); err != nil {
		// Si pas de fichier config, utiliser les valeurs par défaut/env
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	} else {
		// Unmarshall de la config depuis le fichier
		if err := viper.Unmarshal(config); err != nil {
			return nil, fmt.Errorf("error unmarshalling config: %w", err)
		}
	}

	// Validation de la configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// loadFromEnv charge la configuration depuis les variables d'environnement
func loadFromEnv(config *Config) {
	// Server
	if port := os.Getenv("GATEWAY_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Server.Port = p
		}
	}
	if host := os.Getenv("GATEWAY_HOST"); host != "" {
		config.Server.Host = host
	}
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		config.Server.Environment = env
	}
	if debug := os.Getenv("DEBUG"); debug != "" {
		config.Server.Debug = debug == "true"
	}

	// JWT
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		config.JWT.Secret = secret
	}

	// Services URLs - Support des deux formats (Docker et localhost)
	if authURL := os.Getenv("AUTH_SERVICE_URL"); authURL != "" {
		config.Services.Auth.URL = authURL
	}
	if playerURL := os.Getenv("PLAYER_SERVICE_URL"); playerURL != "" {
		config.Services.Player.URL = playerURL
	}
	if worldURL := os.Getenv("WORLD_SERVICE_URL"); worldURL != "" {
		config.Services.World.URL = worldURL
	}
	if combatURL := os.Getenv("COMBAT_SERVICE_URL"); combatURL != "" {
		config.Services.Combat.URL = combatURL
	}
	if inventoryURL := os.Getenv("INVENTORY_SERVICE_URL"); inventoryURL != "" {
		config.Services.Inventory.URL = inventoryURL
	}
	if guildURL := os.Getenv("GUILD_SERVICE_URL"); guildURL != "" {
		config.Services.Guild.URL = guildURL
	}
	if chatURL := os.Getenv("CHAT_SERVICE_URL"); chatURL != "" {
		config.Services.Chat.URL = chatURL
	}
	if analyticsURL := os.Getenv("ANALYTICS_SERVICE_URL"); analyticsURL != "" {
		config.Services.Analytics.URL = analyticsURL
	}

	// NATS
	if natsURL := os.Getenv("NATS_URL"); natsURL != "" {
		config.NATS.URL = natsURL
	}

	// Rate Limiting
	if rpm := os.Getenv("RATE_LIMIT_RPM"); rpm != "" {
		if r, err := strconv.Atoi(rpm); err == nil {
			config.RateLimit.RequestsPerMinute = r
		}
	}
}

// validateConfig valide la configuration
func validateConfig(config *Config) error {
	// Validation du serveur
	if config.Server.Port < 1 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	// Validation JWT
	if len(config.JWT.Secret) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 characters long")
	}

	// Validation des services
	services := map[string]ServiceEndpoint{
		"auth":      config.Services.Auth,
		"player":    config.Services.Player,
		"world":     config.Services.World,
		"combat":    config.Services.Combat,
		"inventory": config.Services.Inventory,
		"guild":     config.Services.Guild,
		"chat":      config.Services.Chat,
		"analytics": config.Services.Analytics,
	}

	for name, service := range services {
		if service.URL == "" {
			return fmt.Errorf("service %s URL is required", name)
		}
		if service.Timeout <= 0 {
			return fmt.Errorf("service %s timeout must be positive", name)
		}
		if service.Retries < 0 {
			return fmt.Errorf("service %s retries must be non-negative", name)
		}
	}

	// Validation NATS
	if config.NATS.URL == "" {
		return fmt.Errorf("NATS URL is required")
	}

	// Validation rate limiting
	if config.RateLimit.RequestsPerMinute <= 0 {
		return fmt.Errorf("rate limit requests per minute must be positive")
	}

	return nil
}
