package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/spf13/viper"
)

// Configuration constants
const (
	DefaultServerPort     = 8087
	DefaultDBPort         = 5432
	DefaultTimeout        = 30
	DefaultMaxOpenConns   = 25
	DefaultMaxIdleConns   = 10
	DefaultDBMaxLifetime  = 5
	DefaultMaxMessageLen  = 500
	DefaultRetentionDays  = 30
	DefaultMaxChannels    = 10
	DefaultAntiSpamMax    = 10
	DefaultBufferSize     = 1024
	DefaultHandshakeTO    = 10
	DefaultPongWait       = 60
	DefaultPingPeriod     = 54
	DefaultWriteWait      = 10
	DefaultMaxMsgSize     = 512
	DefaultMsgsPerMin     = 60
	DefaultBurstSize      = 10
	DefaultCleanupInt     = 5
	DefaultPrometheusPort = 9087
	DefaultJWTMinLength   = 32
	DefaultDBTimeout      = 5
	DefaultLogContentLen  = 50
	DefaultShutdownTO     = 30
)

// Config représente la configuration complète du service Chat
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	Chat       ChatConfig       `mapstructure:"chat"`
	WebSocket  WebSocketConfig  `mapstructure:"websocket"`
	RateLimit  RateLimitConfig  `mapstructure:"rate_limit"`
	Monitoring MonitoringConfig `mapstructure:"monitoring"`
}

// ServerConfig configuration du serveur HTTP
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
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Name         string        `mapstructure:"name"`
	User         string        `mapstructure:"user"`
	Password     string        `mapstructure:"password"`
	SSLMode      string        `mapstructure:"ssl_mode"`
	MaxOpenConns int           `mapstructure:"max_open_conns"`
	MaxIdleConns int           `mapstructure:"max_idle_conns"`
	MaxLifetime  time.Duration `mapstructure:"max_lifetime"`
}

// JWTConfig configuration JWT
type JWTConfig struct {
	Secret string `mapstructure:"secret"`
}

// ChatConfig configuration spécifique au chat
type ChatConfig struct {
	MaxMessageLength       int           `mapstructure:"max_message_length"`
	MessageRetentionDays   int           `mapstructure:"message_retention_days"`
	MaxChannelsPerUser     int           `mapstructure:"max_channels_per_user"`
	AntiSpamWindow         time.Duration `mapstructure:"anti_spam_window"`
	AntiSpamMaxMessages    int           `mapstructure:"anti_spam_max_messages"`
	ModerationEnabled      bool          `mapstructure:"moderation_enabled"`
	ProfanityFilterEnabled bool          `mapstructure:"profanity_filter_enabled"`
}

// WebSocketConfig configuration WebSocket
type WebSocketConfig struct {
	ReadBufferSize   int           `mapstructure:"read_buffer_size"`
	WriteBufferSize  int           `mapstructure:"write_buffer_size"`
	HandshakeTimeout time.Duration `mapstructure:"handshake_timeout"`
	PongWait         time.Duration `mapstructure:"pong_wait"`
	PingPeriod       time.Duration `mapstructure:"ping_period"`
	WriteWait        time.Duration `mapstructure:"write_wait"`
	MaxMessageSize   int64         `mapstructure:"max_message_size"`
}

// RateLimitConfig configuration du rate limiting
type RateLimitConfig struct {
	MessagesPerMinute int           `mapstructure:"messages_per_minute"`
	BurstSize         int           `mapstructure:"burst_size"`
	CleanupInterval   time.Duration `mapstructure:"cleanup_interval"`
}

// MonitoringConfig configuration du monitoring
type MonitoringConfig struct {
	PrometheusPort int    `mapstructure:"prometheus_port"`
	MetricsPath    string `mapstructure:"metrics_path"`
	HealthPath     string `mapstructure:"health_path"`
}

// GetDatabaseURL construit l'URL de connection PostgreSQL
func (d *DatabaseConfig) GetDatabaseURL() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

// LoadConfig charge la configuration depuis les variables d'environnement et fichiers
func LoadConfig() (*Config, error) {
	// Configuration par défaut
	config := &Config{
		Server: ServerConfig{
			Port:         DefaultServerPort,
			Host:         "0.0.0.0",
			Environment:  "development",
			Debug:        true,
			ReadTimeout:  DefaultTimeout * time.Second,
			WriteTimeout: DefaultTimeout * time.Second,
		},
		Database: DatabaseConfig{
			Host:         "localhost",
			Port:         DefaultDBPort,
			Name:         "chat_db",
			User:         "auth_user",
			Password:     "auth_password",
			SSLMode:      "disable",
			MaxOpenConns: DefaultMaxOpenConns,
			MaxIdleConns: DefaultMaxIdleConns,
			MaxLifetime:  DefaultDBMaxLifetime * time.Minute,
		},
		JWT: JWTConfig{
			Secret: "your-super-secret-jwt-key-change-in-production-minimum-64-characters",
		},
		Chat: ChatConfig{
			MaxMessageLength:       DefaultMaxMessageLen,
			MessageRetentionDays:   DefaultRetentionDays,
			MaxChannelsPerUser:     DefaultMaxChannels,
			AntiSpamWindow:         1 * time.Minute,
			AntiSpamMaxMessages:    DefaultAntiSpamMax,
			ModerationEnabled:      true,
			ProfanityFilterEnabled: true,
		},
		WebSocket: WebSocketConfig{
			ReadBufferSize:   DefaultBufferSize,
			WriteBufferSize:  DefaultBufferSize,
			HandshakeTimeout: DefaultHandshakeTO * time.Second,
			PongWait:         DefaultPongWait * time.Second,
			PingPeriod:       DefaultPingPeriod * time.Second, // < PongWait
			WriteWait:        DefaultWriteWait * time.Second,
			MaxMessageSize:   DefaultMaxMsgSize,
		},
		RateLimit: RateLimitConfig{
			MessagesPerMinute: DefaultMsgsPerMin,
			BurstSize:         DefaultBurstSize,
			CleanupInterval:   DefaultCleanupInt * time.Minute,
		},
		Monitoring: MonitoringConfig{
			PrometheusPort: DefaultPrometheusPort,
			MetricsPath:    "/metrics",
			HealthPath:     "/health",
		},
	}

	// Configurer Viper
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/chat/")

	// Variables d'environnement
	viper.AutomaticEnv()
	viper.SetEnvPrefix("CHAT")

	// Lecture du fichier de config (optionnel)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	} else {
		if err := viper.Unmarshal(config); err != nil {
			return nil, fmt.Errorf("error unmarshalling config: %w", err)
		}
	}

	// Surcharge avec les variables d'environnement
	loadFromEnv(config)

	// Validation
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// loadFromEnv charge les variables d'environnement
func loadFromEnv(config *Config) {
	// Server
	if port := os.Getenv("CHAT_SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Server.Port = p
		}
	}
	if host := os.Getenv("CHAT_SERVER_HOST"); host != "" {
		config.Server.Host = host
	}
	if env := os.Getenv("CHAT_ENVIRONMENT"); env != "" {
		config.Server.Environment = env
	}

	// Database
	if dbHost := os.Getenv("CHAT_DB_HOST"); dbHost != "" {
		config.Database.Host = dbHost
	}
	if dbPort := os.Getenv("CHAT_DB_PORT"); dbPort != "" {
		if p, err := strconv.Atoi(dbPort); err == nil {
			config.Database.Port = p
		}
	}
	if dbName := os.Getenv("CHAT_DB_NAME"); dbName != "" {
		config.Database.Name = dbName
	}
	if dbUser := os.Getenv("CHAT_DB_USER"); dbUser != "" {
		config.Database.User = dbUser
	}
	if dbPass := os.Getenv("CHAT_DB_PASSWORD"); dbPass != "" {
		config.Database.Password = dbPass
	}

	// JWT
	if jwtSecret := os.Getenv("CHAT_JWT_SECRET"); jwtSecret != "" {
		config.JWT.Secret = jwtSecret
	}
}

// validateConfig valide la configuration
func validateConfig(config *Config) error {
	// Validation du serveur
	if config.Server.Port < 1 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	// Validation JWT
	if len(config.JWT.Secret) < DefaultJWTMinLength {
		return fmt.Errorf("JWT secret must be at least 32 characters long")
	}

	// Validation Database
	if config.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if config.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}
	if config.Database.User == "" {
		return fmt.Errorf("database user is required")
	}

	// Validation Chat
	if config.Chat.MaxMessageLength <= 0 {
		return fmt.Errorf("max message length must be positive")
	}

	return nil
}
