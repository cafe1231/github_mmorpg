package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config représente la configuration complète du service inventory
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Redis      RedisConfig      `mapstructure:"redis"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	Inventory  InventoryConfig  `mapstructure:"inventory"`
	RateLimit  RateLimitConfig  `mapstructure:"rate_limit"`
	Monitoring MonitoringConfig `mapstructure:"monitoring"`
}

// ServerConfig représente la configuration du serveur HTTP
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Environment  string        `mapstructure:"environment"`
	Debug        bool          `mapstructure:"debug"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// DatabaseConfig représente la configuration de la base de données
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

// RedisConfig représente la configuration Redis
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// JWTConfig représente la configuration JWT
type JWTConfig struct {
	Secret         string        `mapstructure:"secret"`
	ExpirationTime time.Duration `mapstructure:"expiration_time"`
	RefreshTime    time.Duration `mapstructure:"refresh_time"`
}

// InventoryConfig représente la configuration spécifique à l'inventaire
type InventoryConfig struct {
	DefaultSlots    int           `mapstructure:"default_slots"`
	MaxSlots        int           `mapstructure:"max_slots"`
	MaxStackSize    int           `mapstructure:"max_stack_size"`
	TradeTimeout    time.Duration `mapstructure:"trade_timeout"`
	CraftingEnabled bool          `mapstructure:"crafting_enabled"`
}

// RateLimitConfig représente la configuration du rate limiting
type RateLimitConfig struct {
	RequestsPerMinute int           `mapstructure:"requests_per_minute"`
	BurstSize         int           `mapstructure:"burst_size"`
	CleanupInterval   time.Duration `mapstructure:"cleanup_interval"`
}

// MonitoringConfig représente la configuration du monitoring
type MonitoringConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	MetricsPath string `mapstructure:"metrics_path"`
	HealthPath  string `mapstructure:"health_path"`
}

// GetDSN retourne la chaîne de connection PostgreSQL
func (d DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

// Load charge la configuration depuis les variables d'environnement et fichiers
func Load() (*Config, error) {
	// Configuration par défaut
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", 8084)
	viper.SetDefault("server.environment", "development")
	viper.SetDefault("server.debug", true)
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")

	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.name", "inventory_db")
	viper.SetDefault("database.user", "auth_user")
	viper.SetDefault("database.password", "auth_password")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 5)
	viper.SetDefault("database.conn_max_lifetime", "15m")
	viper.SetDefault("database.migrations_path", "./migrations")

	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 2)

	viper.SetDefault("jwt.secret", "inventory-secret-key")
	viper.SetDefault("jwt.expiration_time", "24h")
	viper.SetDefault("jwt.refresh_time", "168h")

	viper.SetDefault("inventory.default_slots", 50)
	viper.SetDefault("inventory.max_slots", 200)
	viper.SetDefault("inventory.max_stack_size", 99)
	viper.SetDefault("inventory.trade_timeout", "5m")
	viper.SetDefault("inventory.crafting_enabled", true)

	viper.SetDefault("rate_limit.requests_per_minute", 100)
	viper.SetDefault("rate_limit.burst_size", 20)
	viper.SetDefault("rate_limit.cleanup_interval", "1h")

	viper.SetDefault("monitoring.enabled", true)
	viper.SetDefault("monitoring.metrics_path", "/metrics")
	viper.SetDefault("monitoring.health_path", "/health")

	// Lecture des variables d'environnement
	viper.AutomaticEnv()
	viper.SetEnvPrefix("INVENTORY")

	// Mapping explicite des variables d'environnement
	viper.BindEnv("server.host", "INVENTORY_SERVER_HOST")
	viper.BindEnv("server.port", "INVENTORY_SERVER_PORT")
	viper.BindEnv("database.host", "INVENTORY_DB_HOST")
	viper.BindEnv("database.port", "INVENTORY_DB_PORT")
	viper.BindEnv("database.name", "INVENTORY_DB_NAME")
	viper.BindEnv("database.user", "INVENTORY_DB_USER")
	viper.BindEnv("database.password", "INVENTORY_DB_PASSWORD")
	viper.BindEnv("database.ssl_mode", "INVENTORY_DB_SSL_MODE")

	// Lecture du fichier de configuration (optionnel)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/inventory")

	// Lecture du fichier de config (pas d'erreur si absent)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Unmarshaling dans la struct Config
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	return &config, nil
}

