package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config structure principale de configuration
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	Redis      RedisConfig      `mapstructure:"redis"`
	Services   ServicesConfig   `mapstructure:"services"`
	Combat     CombatConfig     `mapstructure:"combat"`
	AntiCheat  AntiCheatConfig  `mapstructure:"anticheat"`
	RateLimit  RateLimitConfig  `mapstructure:"rate_limit"`
	Monitoring MonitoringConfig `mapstructure:"monitoring"`
	Logging    LoggingConfig    `mapstructure:"logging"`
}

// ServerConfig configuration du serveur HTTP
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Environment  string        `mapstructure:"environment"`
	Debug        bool          `mapstructure:"debug"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// DatabaseConfig configuration de la base de données
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
}

// JWTConfig configuration JWT
type JWTConfig struct {
	Secret         string        `mapstructure:"secret"`
	ExpirationTime time.Duration `mapstructure:"expiration_time"`
}

// RedisConfig configuration Redis
type RedisConfig struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	Password   string `mapstructure:"password"`
	DB         int    `mapstructure:"db"`
	MaxRetries int    `mapstructure:"max_retries"`
	PoolSize   int    `mapstructure:"pool_size"`
}

// ServicesConfig configuration des services externes
type ServicesConfig struct {
	AuthService   ServiceEndpoint `mapstructure:"auth_service"`
	PlayerService ServiceEndpoint `mapstructure:"player_service"`
	WorldService  ServiceEndpoint `mapstructure:"world_service"`
}

// ServiceEndpoint configuration d'un service externe
type ServiceEndpoint struct {
	URL     string        `mapstructure:"url"`
	Timeout time.Duration `mapstructure:"timeout"`
	Retries int           `mapstructure:"retries"`
}

// CombatConfig configuration spécifique au combat
type CombatConfig struct {
	MaxDuration     time.Duration `mapstructure:"max_duration"`
	TurnTimeout     time.Duration `mapstructure:"turn_timeout"`
	MaxConcurrent   int           `mapstructure:"max_concurrent"`
	CleanupInterval time.Duration `mapstructure:"cleanup_interval"`
	EnablePvP       bool          `mapstructure:"enable_pvp"`
	EnablePvE       bool          `mapstructure:"enable_pve"`
	MaxPartySize    int           `mapstructure:"max_party_size"`
}

// AntiCheatConfig configuration anti-triche
type AntiCheatConfig struct {
	MaxActionsPerSecond    int     `mapstructure:"max_actions_per_second"`
	MaxDamageMultiplier    float64 `mapstructure:"max_damage_multiplier"`
	ValidateMovement       bool    `mapstructure:"validate_movement"`
	ValidateStatsIntegrity bool    `mapstructure:"validate_stats_integrity"`
	LogSuspiciousActivity  bool    `mapstructure:"log_suspicious_activity"`
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

// LoggingConfig configuration des logs
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// LoadConfig charge la configuration depuis les variables d'environnement
func LoadConfig() (*Config, error) {
	// Configuration par défaut
	config := &Config{
		Server: ServerConfig{
			Host:         "0.0.0.0",
			Port:         8083,
			Environment:  "development",
			Debug:        true,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		Database: DatabaseConfig{
			Host:            "localhost",
			Port:            5432,
			Name:            "gameserver_combat",
			User:            "postgres",
			Password:        "postgres",
			SSLMode:         "disable",
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: 300 * time.Second,
		},
		JWT: JWTConfig{
			Secret:         "your-super-secret-jwt-key-change-in-production-minimum-64-characters",
			ExpirationTime: 24 * time.Hour,
		},
		Redis: RedisConfig{
			Host:       "localhost",
			Port:       6379,
			Password:   "",
			DB:         0,
			MaxRetries: 3,
			PoolSize:   10,
		},
		Services: ServicesConfig{
			AuthService: ServiceEndpoint{
				URL:     "http://localhost:8081",
				Timeout: 10 * time.Second,
				Retries: 3,
			},
			PlayerService: ServiceEndpoint{
				URL:     "http://localhost:8082",
				Timeout: 10 * time.Second,
				Retries: 3,
			},
			WorldService: ServiceEndpoint{
				URL:     "http://localhost:8084",
				Timeout: 10 * time.Second,
				Retries: 3,
			},
		},
		Combat: CombatConfig{
			MaxDuration:     300 * time.Second,
			TurnTimeout:     30 * time.Second,
			MaxConcurrent:   1000,
			CleanupInterval: 60 * time.Second,
			EnablePvP:       true,
			EnablePvE:       true,
			MaxPartySize:    4,
		},
		AntiCheat: AntiCheatConfig{
			MaxActionsPerSecond:    5,
			MaxDamageMultiplier:    3.0,
			ValidateMovement:       true,
			ValidateStatsIntegrity: true,
			LogSuspiciousActivity:  true,
		},
		RateLimit: RateLimitConfig{
			RequestsPerMinute: 100,
			BurstSize:         20,
			CleanupInterval:   5 * time.Minute,
		},
		Monitoring: MonitoringConfig{
			PrometheusPort: 9083,
			MetricsPath:    "/metrics",
			HealthPath:     "/health",
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}

	// Configuration Viper
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	// Mapping des variables d'environnement
	viper.BindEnv("server.host", "SERVER_HOST")
	viper.BindEnv("server.port", "SERVER_PORT")
	viper.BindEnv("server.environment", "SERVER_ENVIRONMENT")
	viper.BindEnv("server.debug", "SERVER_DEBUG")
	viper.BindEnv("server.read_timeout", "SERVER_READ_TIMEOUT")
	viper.BindEnv("server.write_timeout", "SERVER_WRITE_TIMEOUT")

	viper.BindEnv("database.host", "DATABASE_HOST")
	viper.BindEnv("database.port", "DATABASE_PORT")
	viper.BindEnv("database.name", "DATABASE_NAME")
	viper.BindEnv("database.user", "DATABASE_USER")
	viper.BindEnv("database.password", "DATABASE_PASSWORD")
	viper.BindEnv("database.ssl_mode", "DATABASE_SSL_MODE")
	viper.BindEnv("database.max_open_conns", "DATABASE_MAX_OPEN_CONNS")
	viper.BindEnv("database.max_idle_conns", "DATABASE_MAX_IDLE_CONNS")
	viper.BindEnv("database.conn_max_lifetime", "DATABASE_CONN_MAX_LIFETIME")

	viper.BindEnv("jwt.secret", "JWT_SECRET")
	viper.BindEnv("jwt.expiration_time", "JWT_EXPIRATION_TIME")

	viper.BindEnv("redis.host", "REDIS_HOST")
	viper.BindEnv("redis.port", "REDIS_PORT")
	viper.BindEnv("redis.password", "REDIS_PASSWORD")
	viper.BindEnv("redis.db", "REDIS_DB")
	viper.BindEnv("redis.max_retries", "REDIS_MAX_RETRIES")
	viper.BindEnv("redis.pool_size", "REDIS_POOL_SIZE")

	viper.BindEnv("services.auth_service.url", "AUTH_SERVICE_URL")
	viper.BindEnv("services.player_service.url", "PLAYER_SERVICE_URL")
	viper.BindEnv("services.world_service.url", "WORLD_SERVICE_URL")

	viper.BindEnv("combat.max_duration", "COMBAT_MAX_DURATION")
	viper.BindEnv("combat.turn_timeout", "COMBAT_TURN_TIMEOUT")
	viper.BindEnv("combat.max_concurrent", "COMBAT_MAX_CONCURRENT")
	viper.BindEnv("combat.cleanup_interval", "COMBAT_CLEANUP_INTERVAL")

	viper.BindEnv("anticheat.max_actions_per_second", "ANTICHEAT_MAX_ACTIONS_PER_SECOND")
	viper.BindEnv("anticheat.max_damage_multiplier", "ANTICHEAT_MAX_DAMAGE_MULTIPLIER")
	viper.BindEnv("anticheat.validate_movement", "ANTICHEAT_VALIDATE_MOVEMENT")

	viper.BindEnv("rate_limit.requests_per_minute", "RATE_LIMIT_REQUESTS_PER_MINUTE")
	viper.BindEnv("rate_limit.burst_size", "RATE_LIMIT_BURST_SIZE")

	viper.BindEnv("monitoring.prometheus_port", "MONITORING_PROMETHEUS_PORT")
	viper.BindEnv("monitoring.metrics_path", "MONITORING_METRICS_PATH")
	viper.BindEnv("monitoring.health_path", "MONITORING_HEALTH_PATH")

	viper.BindEnv("logging.level", "LOG_LEVEL")
	viper.BindEnv("logging.format", "LOG_FORMAT")

	// Charger le fichier de configuration s'il existe
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Merger avec la configuration par défaut
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validation
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// Validate valide la configuration
func (c *Config) Validate() error {
	// Validation serveur
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	// Validation JWT
	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 characters long")
	}

	// Validation database
	if c.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}

	// Validation des services externes
	if c.Services.AuthService.URL == "" {
		return fmt.Errorf("auth service URL is required")
	}
	if c.Services.PlayerService.URL == "" {
		return fmt.Errorf("player service URL is required")
	}

	// Validation anti-cheat
	if c.AntiCheat.MaxActionsPerSecond <= 0 {
		return fmt.Errorf("max actions per second must be positive")
	}
	if c.AntiCheat.MaxDamageMultiplier <= 0 {
		return fmt.Errorf("max damage multiplier must be positive")
	}

	return nil
}

// GetDSN retourne la chaîne de connection PostgreSQL
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

// GetRedisAddr retourne l'adresse Redis
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

