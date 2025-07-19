// auth/internal/config/config.go
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/spf13/viper"
)

// Constantes par défaut pour éviter les magic numbers
const (
	DefaultServerPort        = 8081
	DefaultServerTimeout     = 30
	DefaultDatabasePort      = 5432
	DefaultMaxOpenConns      = 25
	DefaultMaxIdleConns      = 10
	DefaultMaxLifetimeMin    = 5
	DefaultAccessTokenMin    = 15
	DefaultEmailVerifHours   = 24
	DefaultBCryptCost        = 12
	DefaultMaxLoginAttempts  = 5
	DefaultLockoutMin        = 15
	DefaultPasswordMinLength = 8
	DefaultSessionHours      = 24
	DefaultSMTPPort          = 587
	DefaultRateLimitReqs     = 5
	DefaultRateLimitWindow   = 15
	DefaultRateLimitBurst    = 2
	DefaultLoginReqs         = 3
	DefaultGeneralReqs       = 5
	DefaultAPIReqs           = 100
	DefaultAPIBurst          = 20
	DefaultPrometheusPort    = 9091
	DefaultRedisPort         = 6379
	DefaultJWTSecretMin      = 32
	DefaultPasswordMin       = 6
	DefaultConnectionTO      = 10
	DefaultShutdownTO        = 30
)

// Config représente la configuration complète du service Auth
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	Security   SecurityConfig   `mapstructure:"security"`
	Email      EmailConfig      `mapstructure:"email"`
	OAuth      OAuthConfig      `mapstructure:"oauth"`
	RateLimit  RateLimitConfig  `mapstructure:"rate_limit"`
	Monitoring MonitoringConfig `mapstructure:"monitoring"`
	Redis      RedisConfig      `mapstructure:"redis"`
}

// ServerConfig configuration du serveur HTTP
type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	Host         string        `mapstructure:"host"`
	Environment  string        `mapstructure:"environment"`
	Debug        bool          `mapstructure:"debug"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	TLSCertFile  string        `mapstructure:"tls_cert_file"`
	TLSKeyFile   string        `mapstructure:"tls_key_file"`
}

// DatabaseConfig configuration de la base de données
type DatabaseConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Name         string        `mapstructure:"name"`
	Username     string        `mapstructure:"username"`
	Password     string        `mapstructure:"password"`
	SSLMode      string        `mapstructure:"ssl_mode"`
	MaxOpenConns int           `mapstructure:"max_open_conns"`
	MaxIdleConns int           `mapstructure:"max_idle_conns"`
	MaxLifetime  time.Duration `mapstructure:"max_lifetime"`
}

// JWTConfig configuration JWT
type JWTConfig struct {
	Secret                      string        `mapstructure:"secret"`
	Issuer                      string        `mapstructure:"issuer"`
	AccessTokenExpiration       time.Duration `mapstructure:"access_token_expiration"`
	RefreshTokenExpiration      time.Duration `mapstructure:"refresh_token_expiration"`
	EmailVerificationExpiration time.Duration `mapstructure:"email_verification_expiration"`
	PasswordResetExpiration     time.Duration `mapstructure:"password_reset_expiration"`
}

// SecurityConfig configuration de sécurité
type SecurityConfig struct {
	BCryptCost             int           `mapstructure:"bcrypt_cost"`
	MaxLoginAttempts       int           `mapstructure:"max_login_attempts"`
	LockoutDuration        time.Duration `mapstructure:"lockout_duration"`
	PasswordMinLength      int           `mapstructure:"password_min_length"`
	PasswordRequireUpper   bool          `mapstructure:"password_require_upper"`
	PasswordRequireLower   bool          `mapstructure:"password_require_lower"`
	PasswordRequireNumber  bool          `mapstructure:"password_require_number"`
	PasswordRequireDigit   bool          `mapstructure:"password_require_digit"` // Alias pour Number
	PasswordRequireSymbol  bool          `mapstructure:"password_require_symbol"`
	PasswordRequireSpecial bool          `mapstructure:"password_require_special"` // Alias pour Symbol
	SessionTimeout         time.Duration `mapstructure:"session_timeout"`
	TwoFactorRequired      bool          `mapstructure:"two_factor_required"`
}

// EmailConfig configuration email
type EmailConfig struct {
	SMTPHost     string `mapstructure:"smtp_host"`
	SMTPPort     int    `mapstructure:"smtp_port"`
	SMTPUser     string `mapstructure:"smtp_user"`
	SMTPPassword string `mapstructure:"smtp_password"`
	FromEmail    string `mapstructure:"from_email"`
	FromName     string `mapstructure:"from_name"`
	UseTLS       bool   `mapstructure:"use_tls"`
	UseSSL       bool   `mapstructure:"use_ssl"`
}

// OAuthConfig configuration OAuth
type OAuthConfig struct {
	Google   OAuthProvider `mapstructure:"google"`
	Discord  OAuthProvider `mapstructure:"discord"`
	GitHub   OAuthProvider `mapstructure:"github"`
	Facebook OAuthProvider `mapstructure:"facebook"`
}

// OAuthProvider configuration d'un provider OAuth
type OAuthProvider struct {
	ClientID     string   `mapstructure:"client_id"`
	ClientSecret string   `mapstructure:"client_secret"`
	RedirectURL  string   `mapstructure:"redirect_url"`
	Scopes       []string `mapstructure:"scopes"`
	Enabled      bool     `mapstructure:"enabled"`
}

// RateLimitConfig configuration du rate limiting
type RateLimitConfig struct {
	LoginAttempts     RateLimit `mapstructure:"login_attempts"`
	Registration      RateLimit `mapstructure:"registration"`
	PasswordReset     RateLimit `mapstructure:"password_reset"`
	EmailVerification RateLimit `mapstructure:"email_verification"`
	Global            RateLimit `mapstructure:"global"`
}

// RateLimit configuration d'une limite de taux
type RateLimit struct {
	Requests int           `mapstructure:"requests"`
	Window   time.Duration `mapstructure:"window"`
	Burst    int           `mapstructure:"burst"`
}

// MonitoringConfig configuration du monitoring
type MonitoringConfig struct {
	Enabled        bool   `mapstructure:"enabled"`
	PrometheusPort int    `mapstructure:"prometheus_port"`
	MetricsPath    string `mapstructure:"metrics_path"`
	HealthPath     string `mapstructure:"health_path"`
	LogLevel       string `mapstructure:"log_level"`
}

// RedisConfig configuration Redis (pour le cache et sessions)
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	Database int    `mapstructure:"database"`
	Enabled  bool   `mapstructure:"enabled"`
}

// GetDatabaseURL construit l'URL de connection PostgreSQL
func (d *DatabaseConfig) GetDatabaseURL() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.Username, d.Password, d.Name, d.SSLMode,
	)
}

// GetRedisAddr retourne l'adresse Redis
func (r RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
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
			ReadTimeout:  DefaultServerTimeout * time.Second,
			WriteTimeout: DefaultServerTimeout * time.Second,
		},
		Database: DatabaseConfig{
			Host:         "localhost",
			Port:         DefaultDatabasePort,
			Name:         "auth_db",
			Username:     "auth_user",
			Password:     "auth_password",
			SSLMode:      "disable",
			MaxOpenConns: DefaultMaxOpenConns,
			MaxIdleConns: DefaultMaxIdleConns,
			MaxLifetime:  DefaultMaxLifetimeMin * time.Minute,
		},
		JWT: JWTConfig{
			Secret:                      "your-super-secret-jwt-key-change-in-production-minimum-64-characters",
			Issuer:                      "mmo-auth-service",
			AccessTokenExpiration:       DefaultAccessTokenMin * time.Minute,
			RefreshTokenExpiration:      7 * DefaultEmailVerifHours * time.Hour, // 7 jours
			EmailVerificationExpiration: DefaultEmailVerifHours * time.Hour,
			PasswordResetExpiration:     1 * time.Hour,
		},
		Security: SecurityConfig{
			BCryptCost:             DefaultBCryptCost,
			MaxLoginAttempts:       DefaultMaxLoginAttempts,
			LockoutDuration:        DefaultLockoutMin * time.Minute,
			PasswordMinLength:      DefaultPasswordMinLength,
			PasswordRequireUpper:   true,
			PasswordRequireLower:   true,
			PasswordRequireNumber:  true,
			PasswordRequireDigit:   true, // Même valeur que Number
			PasswordRequireSymbol:  true,
			PasswordRequireSpecial: true, // Même valeur que Symbol
			SessionTimeout:         DefaultSessionHours * time.Hour,
			TwoFactorRequired:      false,
		},
		Email: EmailConfig{
			SMTPHost: "localhost",
			SMTPPort: DefaultSMTPPort,
			UseTLS:   true,
			UseSSL:   false,
		},
		OAuth: OAuthConfig{
			Google: OAuthProvider{
				Scopes:  []string{"openid", "profile", "email"},
				Enabled: false,
			},
			Discord: OAuthProvider{
				Scopes:  []string{"identify", "email"},
				Enabled: false,
			},
			GitHub: OAuthProvider{
				Scopes:  []string{"user:email"},
				Enabled: false,
			},
		},
		RateLimit: RateLimitConfig{
			LoginAttempts: RateLimit{
				Requests: DefaultRateLimitReqs,
				Window:   DefaultRateLimitWindow * time.Minute,
				Burst:    DefaultRateLimitBurst,
			},
			Registration: RateLimit{
				Requests: DefaultLoginReqs,
				Window:   1 * time.Hour,
				Burst:    1,
			},
			PasswordReset: RateLimit{
				Requests: DefaultLoginReqs,
				Window:   1 * time.Hour,
				Burst:    1,
			},
			EmailVerification: RateLimit{
				Requests: DefaultGeneralReqs,
				Window:   1 * time.Hour,
				Burst:    DefaultRateLimitBurst,
			},
			Global: RateLimit{
				Requests: DefaultAPIReqs,
				Window:   1 * time.Minute,
				Burst:    DefaultAPIBurst,
			},
		},
		Monitoring: MonitoringConfig{
			Enabled:        true,
			PrometheusPort: DefaultPrometheusPort,
			MetricsPath:    "/metrics",
			HealthPath:     "/health",
			LogLevel:       "info",
		},
		Redis: RedisConfig{
			Host:     "localhost",
			Port:     DefaultRedisPort,
			Database: 0,
			Enabled:  false,
		},
	}

	// Configurer Viper
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/auth/")

	// Variables d'environnement
	viper.AutomaticEnv()
	viper.SetEnvPrefix("AUTH")

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
	loadServerEnv(config)
	loadDatabaseEnv(config)
	loadJWTEnv(config)
	loadEmailEnv(config)
	loadOAuthEnv(config)
	loadRedisEnv(config)
}

// loadServerEnv charge la configuration serveur depuis les variables d'environnement
func loadServerEnv(config *Config) {
	if port := os.Getenv("AUTH_SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Server.Port = p
		}
	}
	if host := os.Getenv("AUTH_SERVER_HOST"); host != "" {
		config.Server.Host = host
	}
	if env := os.Getenv("AUTH_ENVIRONMENT"); env != "" {
		config.Server.Environment = env
	}
	if debug := os.Getenv("AUTH_DEBUG"); debug != "" {
		config.Server.Debug = debug == "true"
	}
}

// loadDatabaseEnv charge la configuration base de données depuis les variables d'environnement
func loadDatabaseEnv(config *Config) {
	if dbHost := os.Getenv("AUTH_DB_HOST"); dbHost != "" {
		config.Database.Host = dbHost
	}
	if dbPort := os.Getenv("AUTH_DB_PORT"); dbPort != "" {
		if p, err := strconv.Atoi(dbPort); err == nil {
			config.Database.Port = p
		}
	}
	if dbName := os.Getenv("AUTH_DB_NAME"); dbName != "" {
		config.Database.Name = dbName
	}
	if dbUser := os.Getenv("AUTH_DB_USERNAME"); dbUser != "" {
		config.Database.Username = dbUser
	}
	if dbPass := os.Getenv("AUTH_DB_PASSWORD"); dbPass != "" {
		config.Database.Password = dbPass
	}
}

// loadJWTEnv charge la configuration JWT depuis les variables d'environnement
func loadJWTEnv(config *Config) {
	if jwtSecret := os.Getenv("AUTH_JWT_SECRET"); jwtSecret != "" {
		config.JWT.Secret = jwtSecret
	}
}

// loadEmailEnv charge la configuration email depuis les variables d'environnement
func loadEmailEnv(config *Config) {
	if smtpHost := os.Getenv("AUTH_SMTP_HOST"); smtpHost != "" {
		config.Email.SMTPHost = smtpHost
	}
	if smtpUser := os.Getenv("AUTH_SMTP_USER"); smtpUser != "" {
		config.Email.SMTPUser = smtpUser
	}
	if smtpPass := os.Getenv("AUTH_SMTP_PASSWORD"); smtpPass != "" {
		config.Email.SMTPPassword = smtpPass
	}
}

// loadOAuthEnv charge la configuration OAuth depuis les variables d'environnement
func loadOAuthEnv(config *Config) {
	if googleClientID := os.Getenv("AUTH_GOOGLE_CLIENT_ID"); googleClientID != "" {
		config.OAuth.Google.ClientID = googleClientID
		config.OAuth.Google.Enabled = true
	}
	if googleClientSecret := os.Getenv("AUTH_GOOGLE_CLIENT_SECRET"); googleClientSecret != "" {
		config.OAuth.Google.ClientSecret = googleClientSecret
	}
}

// loadRedisEnv charge la configuration Redis depuis les variables d'environnement
func loadRedisEnv(config *Config) {
	if redisHost := os.Getenv("AUTH_REDIS_HOST"); redisHost != "" {
		config.Redis.Host = redisHost
		config.Redis.Enabled = true
	}
}

// validateConfig valide la configuration
func validateConfig(config *Config) error {
	// Validation du serveur
	if config.Server.Port < 1 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	// Validation JWT
	if len(config.JWT.Secret) < DefaultJWTSecretMin {
		return fmt.Errorf("JWT secret must be at least 32 characters long")
	}

	if config.JWT.AccessTokenExpiration <= 0 {
		return fmt.Errorf("access token expiration must be positive")
	}

	if config.JWT.RefreshTokenExpiration <= 0 {
		return fmt.Errorf("refresh token expiration must be positive")
	}

	// Validation Database
	if config.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if config.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}

	if config.Database.Username == "" {
		return fmt.Errorf("database username is required")
	}

	// Validation Security
	if config.Security.BCryptCost < 4 || config.Security.BCryptCost > 15 {
		return fmt.Errorf("bcrypt cost must be between 4 and 15")
	}

	if config.Security.PasswordMinLength < DefaultPasswordMin {
		return fmt.Errorf("password minimum length must be at least 6")
	}

	return nil
}
