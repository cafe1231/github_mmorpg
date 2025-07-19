package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Constantes pour remplacer les magic numbers
const (
	// Constantes de combat
	DefaultBaseDamage          = 10
	DefaultCriticalChance      = 0.05
	DefaultCriticalMultiplier  = 1.5
	DefaultManaCostBasic       = 25
	DefaultCooldownBasic       = 2
	DefaultRangeBasic          = 3
	DefaultBaseDamageSkill     = 35
	DefaultEffectValue         = 5
	DefaultEffectDuration      = 3
	DefaultEffectProbability   = 0.3
	DefaultManaCostHeal        = 20
	DefaultRangeHeal           = 2
	DefaultBaseHealing         = 30
	DefaultManaCostAdvanced    = 30
	DefaultCooldownAdvanced    = 3
	DefaultRangeAdvanced       = 4
	DefaultBaseDamageAdvanced  = 40
	DefaultEffectProbability2  = 0.2
	DefaultManaCostSpecial     = 15
	DefaultCooldownSpecial     = 2
	DefaultBaseDamageSpecial   = 20
	DefaultEffectProbability3  = 0.5
	DefaultManaCostUltimate    = 20
	DefaultCooldownUltimate    = 4
	DefaultBaseDamageUltimate  = 25
	DefaultCriticalChanceBonus = 0.5
	DefaultCriticalMultiplier2 = 2.0

	// Constantes d'effets
	DefaultModifierValue  = 20
	DefaultBaseDuration   = 5
	DefaultMaxStacks      = 3
	DefaultModifierValue2 = 10
	DefaultBaseDuration2  = 3
	DefaultMaxStacks2     = 5
	DefaultModifierValue3 = 15
	DefaultBaseDuration3  = 4
	DefaultMaxStacks3     = 2
	DefaultModifierValue4 = 50
	DefaultBaseDuration4  = 3
	DefaultBaseDuration5  = 2
	DefaultBaseDuration6  = 2
	DefaultBaseDuration7  = 4
	DefaultMaxStacks4     = 2
	DefaultModifierValue5 = 25
	DefaultBaseDuration8  = 6
	DefaultMaxStacks5     = 2

	// Constantes PvP
	DefaultPvPRating            = 1000
	DefaultGoldMultiplier       = 10
	DefaultExperienceMultiplier = 5
	DefaultMaxGold              = 10000
	DefaultMaxItems             = 10
	DefaultExpandedRange        = 500
	DefaultMaxGoldReward        = 10000
	DefaultMaxExperienceReward  = 5000
	DefaultGoldReward2          = 7500
	DefaultExperienceReward2    = 3500
	DefaultGoldReward3          = 5000
	DefaultExperienceReward3    = 2500

	// Constantes de validation
	DefaultTurnTimeLimit   = 300
	DefaultMaxDuration     = 3600
	DefaultMaxLimit        = 100
	DefaultMaxLimit2       = 200
	DefaultMaxDuration2    = 100
	DefaultMaxStacks6      = 10
	DefaultMaxSpeed        = 10.0
	DefaultMaxParticipants = 4
	DefaultTurnTimeLimit2  = 30
	DefaultMaxDuration3    = 300
	DefaultMaxLimit3       = 20

	// Constantes de combat
	DefaultMagicalDamageMultiplier  = 0.8
	DefaultPhysicalDamageMultiplier = 0.8
	DefaultDefenseDivisor           = 100
	DefaultHealingMultiplier        = 0.6
	DefaultAgilityDivisor           = 100.0
	DefaultAgilityMultiplier        = 0.1
	DefaultEffectivenessDivisor     = 10.0
	DefaultHealingDivisor           = 8.0
	DefaultLevelDivisor             = 2
	DefaultRangeMultiplier          = 10
	DefaultHealthPercentage         = 100
	DefaultManaPercentage           = 100

	// Constantes de validation des pourcentages
	DefaultMaxCriticalChance = 0.95
	DefaultMinHitChance      = 0.05
	DefaultMaxHitChance      = 0.95
	DefaultMaxEffectiveness  = 100
	DefaultMaxBlockChance    = 0.75
	DefaultMinFinalChance    = 0.05
	DefaultMaxFinalChance    = 0.95
	DefaultMaxKnockback      = 10.0
	DefaultMinKnockback      = 0.1
	DefaultMaxMultiplier     = 3.0

	// Constantes de combat avancé
	DefaultDamageReduction          = 0.3
	DefaultArmorDivisor             = 2
	DefaultMassRatioDivisor         = 50
	DefaultThreatMultiplier         = 1.2
	DefaultThreatDivisor            = 2
	DefaultPercentDivisor           = 100.0
	DefaultEffectDurationMultiplier = 30

	// Constantes anti-cheat
	DefaultMaxActionsPerSecond     = 5
	DefaultMaxActionsPerSecond2    = 60
	DefaultMaxActions              = 10
	DefaultMaxHighDamageActions    = 3
	DefaultMaxTotalDamageActions   = 5
	DefaultMaxCriticalRate         = 0.5
	DefaultMaxScore                = 30
	DefaultMaxRecentActivities     = 3
	DefaultMaxConsistentHighDamage = 5
	DefaultMaxImpossibleActions    = 2
	DefaultMinProcessingTime       = 50
	DefaultMaxScore2               = 100
	DefaultMaxSuspiciousLogs       = 50
	DefaultMinScore                = 30
	DefaultMinScore2               = 50
	DefaultMinScore3               = 80
	DefaultMinActions              = 2
	DefaultMinIntervals            = 5
	DefaultMinSuspicionScore       = 20
	DefaultMinSuspicionScore2      = 50
	DefaultMinSuspicionScore3      = 80

	// Constantes de combat PvE
	DefaultMinParticipants  = 2
	DefaultMaxDamage        = 100
	DefaultMaxHealing       = 50
	DefaultBaseExperience   = 10
	DefaultBaseGold         = 5
	DefaultImprovementScore = 50.0

	// Constantes de calcul de dégâts
	DefaultElementalPowerFire      = 0.8
	DefaultElementalPowerIce       = 0.7
	DefaultElementalPowerLightning = 0.9
	DefaultElementalPowerEarth     = 0.6
	DefaultElementalPowerWind      = 0.75
	DefaultElementalPowerDark      = 0.85
	DefaultElementalPowerLight     = 0.8

	// Constantes de timing
	DefaultChallengeExpiration = 24
	DefaultQueueTicker         = 30
	DefaultQueueTicker2        = 10
	DefaultBaseWait            = 30
	DefaultWaitMultiplier      = 2
	DefaultRatingMultiplier    = 50

	// Constantes de calcul de rating
	DefaultRatingDivisor = 400.0
	DefaultRatingPower   = 10
	DefaultRatingRange   = 100

	// Constantes de combat PvP
	DefaultMaxParticipantsPvP = 2
	DefaultTurnTimeLimitPvP   = 30
	DefaultMaxDurationPvP     = 300

	// Constantes de mana et récupération
	DefaultManaRecoveryPercent  = 0.1
	DefaultSpeedBonusMultiplier = 10
	DefaultRandomFactor         = 10

	// Constantes de fuite
	DefaultFleeChanceBase    = 0.5
	DefaultFleeChanceDivisor = 1000.0
	DefaultMaxFleeChance     = 0.9

	// Constantes de temps
	DefaultTimeWindow               = 5
	DefaultBaseDurationAntiCheat    = 5
	DefaultScoreDivisor             = 20.0
	DefaultMaxDurationAntiCheat     = 30
	DefaultAverageDamageMultiplier  = 0.9
	DefaultAverageDamageMultiplier2 = 0.1

	// Constantes de pourcentage
	DefaultPercentageMultiplier = 100

	// Constantes pour les calculs de variance
	DefaultVarianceDivisor            = 100
	DefaultVarianceBase               = 0.85
	DefaultAuthParts                  = 2
	DefaultMinuteSeconds              = 60
	DefaultCleanupInterval            = 5
	DefaultHighDamageThreshold        = 3
	DefaultTotalDamageThreshold       = 5
	DefaultCritRateThreshold          = 0.5
	DefaultScoreThreshold             = 30
	DefaultRecentActivitiesThreshold  = 3
	DefaultConsistentDamageThreshold  = 5
	DefaultImpossibleActionsThreshold = 2
	DefaultSuspiciousLogsLimit        = 50
	DefaultMemoryMB                   = 1024 * 1024
	DefaultAllocMB                    = 50
	DefaultTotalAllocMB               = 100
	DefaultSysMB                      = 200
	DefaultNumGC                      = 10
	DefaultGoroutines                 = 100
	DefaultMaxConnections             = 20
	DefaultDatabaseTimeout            = 5
	DefaultShutdownTimeout            = 30
	DefaultBurstRatio                 = 4
	DefaultSecurityBurst              = 3
	DefaultCombatBurstRatio           = 6
	DefaultLoadPenaltyFactor          = 0.2
	DefaultRetryAfterSeconds          = 60
	DefaultServerErrorCode            = 500
	DefaultClientErrorCode            = 400
	DefaultRedirectCode               = 300
	DefaultVarianceRange              = 0.3
	DefaultHealingVarianceBase        = 0.9
	DefaultHealingVarianceRange       = 0.2

	// Constantes pour l'algorithme Elo
	DefaultEloK       = 32.0
	DefaultEloBase    = 10.0
	DefaultEloDivisor = 400.0

	// Constantes pour les ratings PvP
	DefaultGrandMasterRating = 2400
	DefaultMasterRating      = 2100
	DefaultDiamondRating     = 1800
	DefaultPlatinumRating    = 1500
	DefaultGoldRating        = 1200

	// Constantes pour les ports et timeouts
	DefaultServerPort                   = 8083
	DefaultDatabasePort                 = 5432
	DefaultRedisPort                    = 6379
	DefaultPrometheusPort               = 9083
	DefaultReadTimeout                  = 30
	DefaultWriteTimeout                 = 30
	DefaultConnMaxLifetime              = 300
	DefaultJWTExpiration                = 24
	DefaultMaxOpenConns                 = 25
	DefaultMaxIdleConns                 = 5
	DefaultRedisMaxRetries              = 3
	DefaultRedisPoolSize                = 10
	DefaultServiceTimeout               = 10
	DefaultServiceRetries               = 3
	DefaultCombatMaxDuration            = 300
	DefaultCombatTurnTimeout            = 30
	DefaultCombatMaxConcurrent          = 1000
	DefaultCombatCleanupInterval        = 60
	DefaultCombatMaxPartySize           = 4
	DefaultAntiCheatMaxActionsPerSecond = 5
	DefaultAntiCheatMaxDamageMultiplier = 3.0
	DefaultRateLimitRequestsPerMinute   = 100
	DefaultRateLimitBurstSize           = 20
	DefaultRateLimitCleanupInterval     = 5
	DefaultJWTMinSecretLength           = 32
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
			Port:         DefaultServerPort,
			Environment:  "development",
			Debug:        true,
			ReadTimeout:  time.Duration(DefaultReadTimeout) * time.Second,
			WriteTimeout: time.Duration(DefaultWriteTimeout) * time.Second,
		},
		Database: DatabaseConfig{
			Host:            "localhost",
			Port:            DefaultDatabasePort,
			Name:            "gameserver_combat",
			User:            "postgres",
			Password:        "postgres",
			SSLMode:         "disable",
			MaxOpenConns:    DefaultMaxOpenConns,
			MaxIdleConns:    DefaultMaxIdleConns,
			ConnMaxLifetime: time.Duration(DefaultConnMaxLifetime) * time.Second,
		},
		JWT: JWTConfig{
			Secret:         "your-super-secret-jwt-key-change-in-production-minimum-64-characters",
			ExpirationTime: time.Duration(DefaultJWTExpiration) * time.Hour,
		},
		Redis: RedisConfig{
			Host:       "localhost",
			Port:       DefaultRedisPort,
			Password:   "",
			DB:         0,
			MaxRetries: DefaultRedisMaxRetries,
			PoolSize:   DefaultRedisPoolSize,
		},
		Services: ServicesConfig{
			AuthService: ServiceEndpoint{
				URL:     "http://localhost:8081",
				Timeout: time.Duration(DefaultServiceTimeout) * time.Second,
				Retries: DefaultServiceRetries,
			},
			PlayerService: ServiceEndpoint{
				URL:     "http://localhost:8082",
				Timeout: time.Duration(DefaultServiceTimeout) * time.Second,
				Retries: DefaultServiceRetries,
			},
			WorldService: ServiceEndpoint{
				URL:     "http://localhost:8084",
				Timeout: time.Duration(DefaultServiceTimeout) * time.Second,
				Retries: DefaultServiceRetries,
			},
		},
		Combat: CombatConfig{
			MaxDuration:     time.Duration(DefaultCombatMaxDuration) * time.Second,
			TurnTimeout:     time.Duration(DefaultCombatTurnTimeout) * time.Second,
			MaxConcurrent:   DefaultCombatMaxConcurrent,
			CleanupInterval: time.Duration(DefaultCombatCleanupInterval) * time.Second,
			EnablePvP:       true,
			EnablePvE:       true,
			MaxPartySize:    DefaultCombatMaxPartySize,
		},
		AntiCheat: AntiCheatConfig{
			MaxActionsPerSecond:    DefaultAntiCheatMaxActionsPerSecond,
			MaxDamageMultiplier:    DefaultAntiCheatMaxDamageMultiplier,
			ValidateMovement:       true,
			ValidateStatsIntegrity: true,
			LogSuspiciousActivity:  true,
		},
		RateLimit: RateLimitConfig{
			RequestsPerMinute: DefaultRateLimitRequestsPerMinute,
			BurstSize:         DefaultRateLimitBurstSize,
			CleanupInterval:   time.Duration(DefaultRateLimitCleanupInterval) * time.Minute,
		},
		Monitoring: MonitoringConfig{
			PrometheusPort: DefaultPrometheusPort,
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
	if err := viper.BindEnv("server.host", "SERVER_HOST"); err != nil {
		return nil, fmt.Errorf("failed to bind server.host env: %w", err)
	}
	if err := viper.BindEnv("server.port", "SERVER_PORT"); err != nil {
		return nil, fmt.Errorf("failed to bind server.port env: %w", err)
	}
	if err := viper.BindEnv("server.environment", "SERVER_ENVIRONMENT"); err != nil {
		return nil, fmt.Errorf("failed to bind server.environment env: %w", err)
	}
	if err := viper.BindEnv("server.debug", "SERVER_DEBUG"); err != nil {
		return nil, fmt.Errorf("failed to bind server.debug env: %w", err)
	}
	if err := viper.BindEnv("server.read_timeout", "SERVER_READ_TIMEOUT"); err != nil {
		return nil, fmt.Errorf("failed to bind server.read_timeout env: %w", err)
	}
	if err := viper.BindEnv("server.write_timeout", "SERVER_WRITE_TIMEOUT"); err != nil {
		return nil, fmt.Errorf("failed to bind server.write_timeout env: %w", err)
	}

	if err := viper.BindEnv("database.host", "DATABASE_HOST"); err != nil {
		return nil, fmt.Errorf("failed to bind database.host env: %w", err)
	}
	if err := viper.BindEnv("database.port", "DATABASE_PORT"); err != nil {
		return nil, fmt.Errorf("failed to bind database.port env: %w", err)
	}
	if err := viper.BindEnv("database.name", "DATABASE_NAME"); err != nil {
		return nil, fmt.Errorf("failed to bind database.name env: %w", err)
	}
	if err := viper.BindEnv("database.user", "DATABASE_USER"); err != nil {
		return nil, fmt.Errorf("failed to bind database.user env: %w", err)
	}
	if err := viper.BindEnv("database.password", "DATABASE_PASSWORD"); err != nil {
		return nil, fmt.Errorf("failed to bind database.password env: %w", err)
	}
	if err := viper.BindEnv("database.ssl_mode", "DATABASE_SSL_MODE"); err != nil {
		return nil, fmt.Errorf("failed to bind database.ssl_mode env: %w", err)
	}
	if err := viper.BindEnv("database.max_open_conns", "DATABASE_MAX_OPEN_CONNS"); err != nil {
		return nil, fmt.Errorf("failed to bind database.max_open_conns env: %w", err)
	}
	if err := viper.BindEnv("database.max_idle_conns", "DATABASE_MAX_IDLE_CONNS"); err != nil {
		return nil, fmt.Errorf("failed to bind database.max_idle_conns env: %w", err)
	}
	if err := viper.BindEnv("database.conn_max_lifetime", "DATABASE_CONN_MAX_LIFETIME"); err != nil {
		return nil, fmt.Errorf("failed to bind database.conn_max_lifetime env: %w", err)
	}

	if err := viper.BindEnv("jwt.secret", "JWT_SECRET"); err != nil {
		return nil, fmt.Errorf("failed to bind jwt.secret env: %w", err)
	}
	if err := viper.BindEnv("jwt.expiration_time", "JWT_EXPIRATION_TIME"); err != nil {
		return nil, fmt.Errorf("failed to bind jwt.expiration_time env: %w", err)
	}

	if err := viper.BindEnv("redis.host", "REDIS_HOST"); err != nil {
		return nil, fmt.Errorf("failed to bind redis.host env: %w", err)
	}
	if err := viper.BindEnv("redis.port", "REDIS_PORT"); err != nil {
		return nil, fmt.Errorf("failed to bind redis.port env: %w", err)
	}
	if err := viper.BindEnv("redis.password", "REDIS_PASSWORD"); err != nil {
		return nil, fmt.Errorf("failed to bind redis.password env: %w", err)
	}
	if err := viper.BindEnv("redis.db", "REDIS_DB"); err != nil {
		return nil, fmt.Errorf("failed to bind redis.db env: %w", err)
	}
	if err := viper.BindEnv("redis.max_retries", "REDIS_MAX_RETRIES"); err != nil {
		return nil, fmt.Errorf("failed to bind redis.max_retries env: %w", err)
	}
	if err := viper.BindEnv("redis.pool_size", "REDIS_POOL_SIZE"); err != nil {
		return nil, fmt.Errorf("failed to bind redis.pool_size env: %w", err)
	}

	if err := viper.BindEnv("services.auth_service.url", "AUTH_SERVICE_URL"); err != nil {
		return nil, fmt.Errorf("failed to bind services.auth_service.url env: %w", err)
	}
	if err := viper.BindEnv("services.player_service.url", "PLAYER_SERVICE_URL"); err != nil {
		return nil, fmt.Errorf("failed to bind services.player_service.url env: %w", err)
	}
	if err := viper.BindEnv("services.world_service.url", "WORLD_SERVICE_URL"); err != nil {
		return nil, fmt.Errorf("failed to bind services.world_service.url env: %w", err)
	}

	if err := viper.BindEnv("combat.max_duration", "COMBAT_MAX_DURATION"); err != nil {
		return nil, fmt.Errorf("failed to bind combat.max_duration env: %w", err)
	}
	if err := viper.BindEnv("combat.turn_timeout", "COMBAT_TURN_TIMEOUT"); err != nil {
		return nil, fmt.Errorf("failed to bind combat.turn_timeout env: %w", err)
	}
	if err := viper.BindEnv("combat.max_concurrent", "COMBAT_MAX_CONCURRENT"); err != nil {
		return nil, fmt.Errorf("failed to bind combat.max_concurrent env: %w", err)
	}
	if err := viper.BindEnv("combat.cleanup_interval", "COMBAT_CLEANUP_INTERVAL"); err != nil {
		return nil, fmt.Errorf("failed to bind combat.cleanup_interval env: %w", err)
	}

	if err := viper.BindEnv("anticheat.max_actions_per_second", "ANTICHEAT_MAX_ACTIONS_PER_SECOND"); err != nil {
		return nil, fmt.Errorf("failed to bind anticheat.max_actions_per_second env: %w", err)
	}
	if err := viper.BindEnv("anticheat.max_damage_multiplier", "ANTICHEAT_MAX_DAMAGE_MULTIPLIER"); err != nil {
		return nil, fmt.Errorf("failed to bind anticheat.max_damage_multiplier env: %w", err)
	}
	if err := viper.BindEnv("anticheat.validate_movement", "ANTICHEAT_VALIDATE_MOVEMENT"); err != nil {
		return nil, fmt.Errorf("failed to bind anticheat.validate_movement env: %w", err)
	}

	if err := viper.BindEnv("rate_limit.requests_per_minute", "RATE_LIMIT_REQUESTS_PER_MINUTE"); err != nil {
		return nil, fmt.Errorf("failed to bind rate_limit.requests_per_minute env: %w", err)
	}
	if err := viper.BindEnv("rate_limit.burst_size", "RATE_LIMIT_BURST_SIZE"); err != nil {
		return nil, fmt.Errorf("failed to bind rate_limit.burst_size env: %w", err)
	}

	if err := viper.BindEnv("monitoring.prometheus_port", "MONITORING_PROMETHEUS_PORT"); err != nil {
		return nil, fmt.Errorf("failed to bind monitoring.prometheus_port env: %w", err)
	}
	if err := viper.BindEnv("monitoring.metrics_path", "MONITORING_METRICS_PATH"); err != nil {
		return nil, fmt.Errorf("failed to bind monitoring.metrics_path env: %w", err)
	}
	if err := viper.BindEnv("monitoring.health_path", "MONITORING_HEALTH_PATH"); err != nil {
		return nil, fmt.Errorf("failed to bind monitoring.health_path env: %w", err)
	}

	if err := viper.BindEnv("logging.level", "LOG_LEVEL"); err != nil {
		return nil, fmt.Errorf("failed to bind logging.level env: %w", err)
	}
	if err := viper.BindEnv("logging.format", "LOG_FORMAT"); err != nil {
		return nil, fmt.Errorf("failed to bind logging.format env: %w", err)
	}

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
	if len(c.JWT.Secret) < DefaultJWTMinSecretLength {
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
