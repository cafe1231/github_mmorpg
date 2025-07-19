package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/spf13/viper"
)

// Constantes pour les valeurs par défaut
const (
	// Ports par défaut
	DefaultGatewayPort    = 8080
	DefaultPrometheusPort = 9090

	// Timeouts par défaut (en secondes)
	DefaultServerTimeout      = 30
	DefaultJWTExpiration      = 24 // heures
	DefaultServiceTimeout     = 10
	DefaultCombatTimeout      = 3
	DefaultWorldTimeout       = 5
	DefaultChatTimeout        = 5
	DefaultAnalyticsTimeout   = 15
	DefaultNATSConnectTimeout = 10
	DefaultNATSReconnectDelay = 2

	// Rate Limiting par défaut
	DefaultRPM             = 1000
	DefaultBurstSize       = 100
	DefaultCleanupInterval = 5 // minutes

	// Retry et Health
	DefaultRetries          = 3
	DefaultLightRetries     = 2
	DefaultSingleRetry      = 1
	DefaultNATSMaxReconnect = 10

	// JWT
	MinJWTSecretLength = 32
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
			Port:         DefaultGatewayPort,
			Host:         "0.0.0.0",
			Environment:  "development",
			Debug:        true,
			ReadTimeout:  DefaultServerTimeout * time.Second,
			WriteTimeout: DefaultServerTimeout * time.Second,
		},
		JWT: JWTConfig{
			Secret:         "your-super-secret-jwt-key-change-in-production-minimum-64-characters",
			ExpirationTime: DefaultJWTExpiration * time.Hour,
		},
		Services: ServicesConfig{
			Auth: ServiceEndpoint{
				URL:     "http://localhost:8081", // Changé de auth-service vers localhost
				Timeout: DefaultServiceTimeout * time.Second,
				Retries: DefaultRetries,
			},
			Player: ServiceEndpoint{
				URL:     "http://localhost:8082", // Changé de player-service vers localhost
				Timeout: DefaultServiceTimeout * time.Second,
				Retries: DefaultRetries,
			},
			World: ServiceEndpoint{
				URL:     "http://localhost:8083", // Changé de world-service vers localhost
				Timeout: DefaultWorldTimeout * time.Second,
				Retries: DefaultLightRetries,
			},
			Combat: ServiceEndpoint{
				URL:     "http://localhost:8084", // Changé de combat-service vers localhost
				Timeout: DefaultCombatTimeout * time.Second,
				Retries: DefaultSingleRetry,
			},
			Inventory: ServiceEndpoint{
				URL:     "http://localhost:8085", // Changé de inventory-service vers localhost
				Timeout: DefaultServiceTimeout * time.Second,
				Retries: DefaultRetries,
			},
			Guild: ServiceEndpoint{
				URL:     "http://localhost:8086", // Changé de guild-service vers localhost
				Timeout: DefaultServiceTimeout * time.Second,
				Retries: DefaultRetries,
			},
			Chat: ServiceEndpoint{
				URL:     "http://localhost:8087", // Changé de chat-service vers localhost
				Timeout: DefaultChatTimeout * time.Second,
				Retries: DefaultLightRetries,
			},
			Analytics: ServiceEndpoint{
				URL:     "http://localhost:8088", // Changé de analytics-service vers localhost
				Timeout: DefaultAnalyticsTimeout * time.Second,
				Retries: DefaultSingleRetry,
			},
		},
		RateLimit: RateLimitConfig{
			RequestsPerMinute: DefaultRPM,
			BurstSize:         DefaultBurstSize,
			CleanupInterval:   DefaultCleanupInterval * time.Minute,
		},
		Monitoring: MonitoringConfig{
			PrometheusPort: DefaultPrometheusPort,
			MetricsPath:    "/metrics",
			HealthPath:     "/health",
		},
		NATS: NATSConfig{
			URL:                  "nats://localhost:4222", // Changé de nats vers localhost
			ClusterID:            "mmorpg-cluster",
			ClientID:             "gateway-service",
			ConnectTimeout:       DefaultNATSConnectTimeout * time.Second,
			ReconnectDelay:       DefaultNATSReconnectDelay * time.Second,
			MaxReconnectAttempts: DefaultNATSMaxReconnect,
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
	loadServerConfigFromEnv(config)
	loadJWTConfigFromEnv(config)
	loadServicesConfigFromEnv(config)
	loadNATSConfigFromEnv(config)
	loadRateLimitConfigFromEnv(config)
}

// loadServerConfigFromEnv charge la configuration du serveur
func loadServerConfigFromEnv(config *Config) {
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
}

// loadJWTConfigFromEnv charge la configuration JWT
func loadJWTConfigFromEnv(config *Config) {
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		config.JWT.Secret = secret
	}
}

// loadServicesConfigFromEnv charge la configuration des services
func loadServicesConfigFromEnv(config *Config) {
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
}

// loadNATSConfigFromEnv charge la configuration NATS
func loadNATSConfigFromEnv(config *Config) {
	if natsURL := os.Getenv("NATS_URL"); natsURL != "" {
		config.NATS.URL = natsURL
	}
}

// loadRateLimitConfigFromEnv charge la configuration du rate limiting
func loadRateLimitConfigFromEnv(config *Config) {
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
	if len(config.JWT.Secret) < MinJWTSecretLength {
		return fmt.Errorf("JWT secret must be at least %d characters long", MinJWTSecretLength)
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
