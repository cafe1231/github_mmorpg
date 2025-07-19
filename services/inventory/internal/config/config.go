package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Default value constants
const (
	DefaultServerPort     = 8084
	DefaultDatabasePort   = 5432
	DefaultMaxOpenConns   = 25
	DefaultMaxIdleConns   = 5
	DefaultRedisPort      = 6379
	DefaultRedisDB        = 2
	DefaultInventorySlots = 50
	MaxInventorySlots     = 200
	MaxStackSize          = 99
	DefaultRequestsPerMin = 100
	DefaultBurstSize      = 20
)

// Config represents the complete inventory service configuration
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Redis      RedisConfig      `mapstructure:"redis"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	Inventory  InventoryConfig  `mapstructure:"inventory"`
	RateLimit  RateLimitConfig  `mapstructure:"rate_limit"`
	Monitoring MonitoringConfig `mapstructure:"monitoring"`
}

// ServerConfig represents HTTP server configuration
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Environment  string        `mapstructure:"environment"`
	Debug        bool          `mapstructure:"debug"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Name            string        `mapstructure:"name"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	MigrationsPath  string        `mapstructure:"migrations_path"`
}

// RedisConfig represents Redis configuration
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// JWTConfig represents JWT configuration
type JWTConfig struct {
	Secret         string        `mapstructure:"secret"`
	ExpirationTime time.Duration `mapstructure:"expiration_time"`
	RefreshTime    time.Duration `mapstructure:"refresh_time"`
}

// InventoryConfig represents inventory-specific configuration
type InventoryConfig struct {
	DefaultSlots    int           `mapstructure:"default_slots"`
	MaxSlots        int           `mapstructure:"max_slots"`
	MaxStackSize    int           `mapstructure:"max_stack_size"`
	TradeTimeout    time.Duration `mapstructure:"trade_timeout"`
	CraftingEnabled bool          `mapstructure:"crafting_enabled"`
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int           `mapstructure:"requests_per_minute"`
	BurstSize         int           `mapstructure:"burst_size"`
	CleanupInterval   time.Duration `mapstructure:"cleanup_interval"`
}

// MonitoringConfig represents monitoring configuration
type MonitoringConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	MetricsPath string `mapstructure:"metrics_path"`
	HealthPath  string `mapstructure:"health_path"`
}

// GetDSN returns PostgreSQL connection string
func (d *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

// Load loads configuration from environment variables and files
func Load() (*Config, error) {
	// Default configuration
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", DefaultServerPort)
	viper.SetDefault("server.environment", "development")
	viper.SetDefault("server.debug", true)
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")

	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", DefaultDatabasePort)
	viper.SetDefault("database.name", "inventory_db")
	viper.SetDefault("database.user", "auth_user")
	viper.SetDefault("database.password", "auth_password")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_open_conns", DefaultMaxOpenConns)
	viper.SetDefault("database.max_idle_conns", DefaultMaxIdleConns)
	viper.SetDefault("database.conn_max_lifetime", "15m")
	viper.SetDefault("database.migrations_path", "./migrations")

	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", DefaultRedisPort)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", DefaultRedisDB)

	viper.SetDefault("jwt.secret", "inventory-secret-key")
	viper.SetDefault("jwt.expiration_time", "24h")
	viper.SetDefault("jwt.refresh_time", "168h")

	viper.SetDefault("inventory.default_slots", DefaultInventorySlots)
	viper.SetDefault("inventory.max_slots", MaxInventorySlots)
	viper.SetDefault("inventory.max_stack_size", MaxStackSize)
	viper.SetDefault("inventory.trade_timeout", "5m")
	viper.SetDefault("inventory.crafting_enabled", true)

	viper.SetDefault("rate_limit.requests_per_minute", DefaultRequestsPerMin)
	viper.SetDefault("rate_limit.burst_size", DefaultBurstSize)
	viper.SetDefault("rate_limit.cleanup_interval", "1h")

	viper.SetDefault("monitoring.enabled", true)
	viper.SetDefault("monitoring.metrics_path", "/metrics")
	viper.SetDefault("monitoring.health_path", "/health")

	// Read environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix("INVENTORY")

	// Explicit environment variable mapping
	if err := viper.BindEnv("server.host", "INVENTORY_SERVER_HOST"); err != nil {
		return nil, fmt.Errorf("failed to bind server.host env: %w", err)
	}
	if err := viper.BindEnv("server.port", "INVENTORY_SERVER_PORT"); err != nil {
		return nil, fmt.Errorf("failed to bind server.port env: %w", err)
	}
	if err := viper.BindEnv("database.host", "INVENTORY_DB_HOST"); err != nil {
		return nil, fmt.Errorf("failed to bind database.host env: %w", err)
	}
	if err := viper.BindEnv("database.port", "INVENTORY_DB_PORT"); err != nil {
		return nil, fmt.Errorf("failed to bind database.port env: %w", err)
	}
	if err := viper.BindEnv("database.name", "INVENTORY_DB_NAME"); err != nil {
		return nil, fmt.Errorf("failed to bind database.name env: %w", err)
	}
	if err := viper.BindEnv("database.user", "INVENTORY_DB_USER"); err != nil {
		return nil, fmt.Errorf("failed to bind database.user env: %w", err)
	}
	if err := viper.BindEnv("database.password", "INVENTORY_DB_PASSWORD"); err != nil {
		return nil, fmt.Errorf("failed to bind database.password env: %w", err)
	}
	if err := viper.BindEnv("database.ssl_mode", "INVENTORY_DB_SSL_MODE"); err != nil {
		return nil, fmt.Errorf("failed to bind database.ssl_mode env: %w", err)
	}

	// Read configuration file (optional)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/inventory")

	// Read config file (no error if absent)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Unmarshal into Config struct
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	return &config, nil
}
