package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"player/internal/config"
)

// DB représente la connection à la base de données
type DB struct {
	*sqlx.DB
	Config *config.DatabaseConfig
}

// NewConnection crée une nouvelle connection à la base de données
func NewConnection(cfg config.DatabaseConfig) (*DB, error) {
	// Construction de l'URL de connection
	dsn := cfg.GetDatabaseURL()

	// connection à la base de données
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configuration de la pool de connections
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Hour)

	// Test de la connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"host":     cfg.Host,
		"port":     cfg.Port,
		"database": cfg.Name,
		"service":  "player",
	}).Info("Connected to PostgreSQL database")

	return &DB{
		DB:     db,
		Config: &cfg,
	}, nil
}

// Close ferme la connection à la base de données
func (db *DB) Close() error {
	if db.DB != nil {
		logrus.Info("Closing player database connection")
		return db.DB.Close()
	}
	return nil
}

// HealthCheck vérifie l'état de la base de données
func (db *DB) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("player database health check failed: %w", err)
	}

	return nil
}

// RunMigrations exécute les migrations de base de données
func RunMigrations(db *DB) error {
	logrus.Info("Running player database migrations...")

	// ANCIEN SYSTÈME - Désactivé car nous utilisons maintenant les migrations dans migrations/
	// Les migrations sont maintenant dans le dossier migrations/ et doivent être exécutées séparément
	/*
		// Migrations SQL
		migrations := []string{
			createPlayersTable,
			createCharactersTable,
			createCharacterStatsTable,
			createCombatStatsTable,
			createStatModifiersTable,
			createIndexes,
			createTriggers,
		}

		// Exécuter chaque migration
		for i, migration := range migrations {
			logrus.WithField("migration", i+1).Debug("Executing migration")

			if _, err := db.Exec(migration); err != nil {
				return fmt.Errorf("failed to execute migration %d: %w", i+1, err)
			}
		}
	*/

	logrus.Info("Player database migrations completed successfully - using migrations/ folder")
	return nil
}

// Migrations SQL
const createPlayersTable = `
CREATE TABLE IF NOT EXISTS players (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE NOT NULL, -- Référence vers le service Auth
    display_name VARCHAR(20) UNIQUE NOT NULL,
    avatar VARCHAR(255),
    title VARCHAR(50),
    guild_id UUID,
    total_play_time INTEGER DEFAULT 0,
    last_seen TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    preferences JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);`

const createCharactersTable = `
CREATE TABLE IF NOT EXISTS characters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    name VARCHAR(20) UNIQUE NOT NULL,
    class VARCHAR(20) NOT NULL CHECK (class IN ('warrior', 'mage', 'archer', 'rogue')),
    race VARCHAR(20) NOT NULL CHECK (race IN ('human', 'elf', 'dwarf', 'orc')),
    gender VARCHAR(10) NOT NULL CHECK (gender IN ('male', 'female', 'other')),
    appearance JSONB NOT NULL DEFAULT '{}',
    level INTEGER DEFAULT 1 CHECK (level >= 1 AND level <= 100),
    experience BIGINT DEFAULT 0,
    zone_id VARCHAR(50) DEFAULT 'starting_zone',
    position_x DOUBLE PRECISION DEFAULT 0,
    position_y DOUBLE PRECISION DEFAULT 0,
    position_z DOUBLE PRECISION DEFAULT 0,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'deleted')),
    last_played TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);`

const createCharacterStatsTable = `
CREATE TABLE IF NOT EXISTS character_stats (
    character_id UUID PRIMARY KEY REFERENCES characters(id) ON DELETE CASCADE,
    health INTEGER DEFAULT 100,
    max_health INTEGER DEFAULT 100,
    mana INTEGER DEFAULT 50,
    max_mana INTEGER DEFAULT 50,
    strength INTEGER DEFAULT 10,
    agility INTEGER DEFAULT 10,
    intelligence INTEGER DEFAULT 10,
    vitality INTEGER DEFAULT 10,
    stat_points INTEGER DEFAULT 0,
    skill_points INTEGER DEFAULT 0,
    physical_damage INTEGER DEFAULT 20,
    magical_damage INTEGER DEFAULT 25,
    physical_defense INTEGER DEFAULT 7,
    magical_defense INTEGER DEFAULT 7,
    critical_chance INTEGER DEFAULT 5,
    attack_speed INTEGER DEFAULT 100,
    movement_speed INTEGER DEFAULT 100,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);`

const createCombatStatsTable = `
CREATE TABLE IF NOT EXISTS combat_stats (
    character_id UUID PRIMARY KEY REFERENCES characters(id) ON DELETE CASCADE,
    monsters_killed INTEGER DEFAULT 0,
    bosses_killed INTEGER DEFAULT 0,
    deaths INTEGER DEFAULT 0,
    damage_dealt BIGINT DEFAULT 0,
    damage_taken BIGINT DEFAULT 0,
    healing_done BIGINT DEFAULT 0,
    pvp_kills INTEGER DEFAULT 0,
    pvp_deaths INTEGER DEFAULT 0,
    pvp_damage_dealt BIGINT DEFAULT 0,
    pvp_damage_taken BIGINT DEFAULT 0,
    quests_completed INTEGER DEFAULT 0,
    items_looted INTEGER DEFAULT 0,
    gold_earned BIGINT DEFAULT 0,
    distance_traveled DOUBLE PRECISION DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);`

const createStatModifiersTable = `
CREATE TABLE IF NOT EXISTS stat_modifiers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL CHECK (type IN ('buff', 'debuff', 'equipment')),
    source VARCHAR(100) NOT NULL,
    stat_name VARCHAR(50) NOT NULL,
    value INTEGER NOT NULL,
    duration INTEGER DEFAULT 0, -- 0 = permanent
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);`

const createIndexes = `
-- Index pour optimiser les requêtes
CREATE INDEX IF NOT EXISTS idx_players_user_id ON players(user_id);
CREATE INDEX IF NOT EXISTS idx_players_display_name ON players(display_name);
CREATE INDEX IF NOT EXISTS idx_players_guild_id ON players(guild_id);
CREATE INDEX IF NOT EXISTS idx_players_last_seen ON players(last_seen);

CREATE INDEX IF NOT EXISTS idx_characters_player_id ON characters(player_id);
CREATE INDEX IF NOT EXISTS idx_characters_name ON characters(name);
CREATE INDEX IF NOT EXISTS idx_characters_class ON characters(class);
CREATE INDEX IF NOT EXISTS idx_characters_race ON characters(race);
CREATE INDEX IF NOT EXISTS idx_characters_level ON characters(level);
CREATE INDEX IF NOT EXISTS idx_characters_zone_id ON characters(zone_id);
CREATE INDEX IF NOT EXISTS idx_characters_status ON characters(status);
CREATE INDEX IF NOT EXISTS idx_characters_last_played ON characters(last_played);

CREATE INDEX IF NOT EXISTS idx_stat_modifiers_character_id ON stat_modifiers(character_id);
CREATE INDEX IF NOT EXISTS idx_stat_modifiers_expires_at ON stat_modifiers(expires_at);
CREATE INDEX IF NOT EXISTS idx_stat_modifiers_type ON stat_modifiers(type);
`

const createTriggers = `
-- Trigger pour mettre à jour updated_at automatiquement
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Appliquer le trigger aux tables nécessaires
DROP TRIGGER IF EXISTS update_players_updated_at ON players;
CREATE TRIGGER update_players_updated_at 
    BEFORE UPDATE ON players 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_characters_updated_at ON characters;
CREATE TRIGGER update_characters_updated_at 
    BEFORE UPDATE ON characters 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_character_stats_updated_at ON character_stats;
CREATE TRIGGER update_character_stats_updated_at 
    BEFORE UPDATE ON character_stats 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_combat_stats_updated_at ON combat_stats;
CREATE TRIGGER update_combat_stats_updated_at 
    BEFORE UPDATE ON combat_stats 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Trigger pour créer automatiquement les stats d'un nouveau personnage
CREATE OR REPLACE FUNCTION create_character_stats()
RETURNS TRIGGER AS $$
DECLARE
    base_health INTEGER := 100;
    base_mana INTEGER := 50;
    base_str INTEGER := 10;
    base_agi INTEGER := 10;
    base_int INTEGER := 10;
    base_vit INTEGER := 10;
BEGIN
    -- Ajuster les stats de base selon la classe
    CASE NEW.class
        WHEN 'warrior' THEN
            base_str := base_str + 5;
            base_vit := base_vit + 3;
        WHEN 'mage' THEN
            base_int := base_int + 5;
            base_mana := base_mana + 20;
        WHEN 'archer' THEN
            base_agi := base_agi + 5;
            base_str := base_str + 2;
        WHEN 'rogue' THEN
            base_agi := base_agi + 4;
            base_int := base_int + 3;
    END CASE;

    -- Ajuster selon la race
    CASE NEW.race
        WHEN 'human' THEN
            base_str := base_str + 1;
            base_agi := base_agi + 1;
            base_int := base_int + 1;
            base_vit := base_vit + 1;
        WHEN 'elf' THEN
            base_agi := base_agi + 3;
            base_int := base_int + 2;
        WHEN 'dwarf' THEN
            base_str := base_str + 3;
            base_vit := base_vit + 3;
        WHEN 'orc' THEN
            base_str := base_str + 4;
            base_vit := base_vit + 2;
    END CASE;

    -- Calculer les stats dérivées
    base_health := 100 + (base_vit * 10);
    base_mana := 50 + (base_int * 5);

    -- Insérer les stats de base
    INSERT INTO character_stats (
        character_id, health, max_health, mana, max_mana,
        strength, agility, intelligence, vitality,
        stat_points, skill_points,
        physical_damage, magical_damage, physical_defense, magical_defense,
        critical_chance, attack_speed, movement_speed
    ) VALUES (
        NEW.id, base_health, base_health, base_mana, base_mana,
        base_str, base_agi, base_int, base_vit,
        0, 0,
        (base_str * 2) + (base_agi / 2),
        FLOOR(base_int * 2.5),
        (base_str + base_vit) / 3,
        (base_int + base_vit) / 3,
        LEAST((base_agi / 10) + 5, 50),
        LEAST(100 + (base_agi / 5), 200),
        LEAST(100 + (base_agi / 10), 150)
    );

    -- Insérer les stats de combat vides
    INSERT INTO combat_stats (character_id) VALUES (NEW.id);

    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS create_character_stats_trigger ON characters;
CREATE TRIGGER create_character_stats_trigger
    AFTER INSERT ON characters
    FOR EACH ROW EXECUTE FUNCTION create_character_stats();

-- Fonction pour nettoyer les modificateurs expirés
CREATE OR REPLACE FUNCTION cleanup_expired_modifiers()
RETURNS void AS $$
BEGIN
    DELETE FROM stat_modifiers 
    WHERE expires_at IS NOT NULL AND expires_at < CURRENT_TIMESTAMP;
END;
$$ language 'plpgsql';
`
