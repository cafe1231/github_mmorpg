package database

import (
"context"
"fmt"
"time"

"github.com/jmoiron/sqlx"
_ "github.com/lib/pq"
"github.com/sirupsen/logrus"

"world/internal/config"
)

type DB struct {
*sqlx.DB
Config *config.DatabaseConfig
}

func NewConnection(cfg *config.Config) (*DB, error) {
dsn := cfg.Database.GetDatabaseURL()

db, err := sqlx.Connect("postgres", dsn)
if err != nil {
return nil, fmt.Errorf("failed to connect to database: %w", err)
}

db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
db.SetConnMaxLifetime(time.Hour)

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := db.PingContext(ctx); err != nil {
return nil, fmt.Errorf("failed to ping database: %w", err)
}

logrus.WithFields(logrus.Fields{
"host":     cfg.Database.Host,
"port":     cfg.Database.Port,
"database": cfg.Database.Name,
"service":  "world",
}).Info("Connected to PostgreSQL database")

return &DB{
DB:     db,
Config: &cfg.Database,
}, nil
}

func (db *DB) Close() error {
if db.DB != nil {
logrus.Info("Closing world database connection")
return db.DB.Close()
}
return nil
}

func (db *DB) HealthCheck() error {
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := db.PingContext(ctx); err != nil {
return fmt.Errorf("world database health check failed: %w", err)
}

return nil
}

func RunMigrations(db *DB) error {
logrus.Info("Running world database migrations...")

migrations := []string{
createZonesTable,
createPlayerPositionsTable,
createWeatherTable,
createZoneTransitionsTable,
createIndexes,
insertDefaultData,
}

for i, migration := range migrations {
logrus.WithField("migration", i+1).Debug("Executing migration")

if _, err := db.Exec(migration); err != nil {
return fmt.Errorf("failed to execute migration %d: %w", i+1, err)
}
}

logrus.Info("World database migrations completed successfully")
return nil
}

const createZonesTable = `
CREATE TABLE IF NOT EXISTS zones (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    type VARCHAR(20) NOT NULL CHECK (type IN ('city', 'dungeon', 'wilderness', 'pvp', 'safe')),
    level INTEGER DEFAULT 1 CHECK (level >= 1 AND level <= 100),
    
    min_x DOUBLE PRECISION DEFAULT 0,
    min_y DOUBLE PRECISION DEFAULT 0,
    min_z DOUBLE PRECISION DEFAULT 0,
    max_x DOUBLE PRECISION DEFAULT 100,
    max_y DOUBLE PRECISION DEFAULT 100,
    max_z DOUBLE PRECISION DEFAULT 100,
    
    spawn_x DOUBLE PRECISION DEFAULT 50,
    spawn_y DOUBLE PRECISION DEFAULT 0,
    spawn_z DOUBLE PRECISION DEFAULT 50,
    
    max_players INTEGER DEFAULT 100 CHECK (max_players > 0),
    is_pvp BOOLEAN DEFAULT FALSE,
    is_safe_zone BOOLEAN DEFAULT FALSE,
    settings JSONB DEFAULT '{}',
    
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'maintenance', 'disabled')),
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);`

const createPlayerPositionsTable = `
CREATE TABLE IF NOT EXISTS player_positions (
    character_id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    zone_id VARCHAR(50) NOT NULL REFERENCES zones(id),
    
    x DOUBLE PRECISION NOT NULL,
    y DOUBLE PRECISION NOT NULL,
    z DOUBLE PRECISION NOT NULL,
    rotation DOUBLE PRECISION DEFAULT 0,
    
    velocity_x DOUBLE PRECISION DEFAULT 0,
    velocity_y DOUBLE PRECISION DEFAULT 0,
    velocity_z DOUBLE PRECISION DEFAULT 0,
    is_moving BOOLEAN DEFAULT FALSE,
    
    is_online BOOLEAN DEFAULT TRUE,
    last_update TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT valid_position CHECK (
        x IS NOT NULL AND y IS NOT NULL AND z IS NOT NULL
    )
);`

const createWeatherTable = `
CREATE TABLE IF NOT EXISTS weather (
    zone_id VARCHAR(50) PRIMARY KEY REFERENCES zones(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL CHECK (type IN ('clear', 'rain', 'storm', 'snow', 'fog')),
    intensity DOUBLE PRECISION DEFAULT 0.5 CHECK (intensity >= 0 AND intensity <= 1),
    temperature DOUBLE PRECISION DEFAULT 20 CHECK (temperature >= -50 AND temperature <= 50),
    wind_speed DOUBLE PRECISION DEFAULT 0 CHECK (wind_speed >= 0),
    wind_direction DOUBLE PRECISION DEFAULT 0 CHECK (wind_direction >= 0 AND wind_direction < 360),
    visibility DOUBLE PRECISION DEFAULT 1000 CHECK (visibility >= 0),
    
    start_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    end_time TIMESTAMP WITH TIME ZONE,
    
    is_active BOOLEAN DEFAULT TRUE,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);`

const createZoneTransitionsTable = `
CREATE TABLE IF NOT EXISTS zone_transitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_zone_id VARCHAR(50) NOT NULL REFERENCES zones(id) ON DELETE CASCADE,
    to_zone_id VARCHAR(50) NOT NULL REFERENCES zones(id) ON DELETE CASCADE,
    
    trigger_x DOUBLE PRECISION NOT NULL,
    trigger_y DOUBLE PRECISION NOT NULL,
    trigger_z DOUBLE PRECISION NOT NULL,
    trigger_radius DOUBLE PRECISION DEFAULT 2.0 CHECK (trigger_radius > 0),
    
    destination_x DOUBLE PRECISION NOT NULL,
    destination_y DOUBLE PRECISION NOT NULL,
    destination_z DOUBLE PRECISION NOT NULL,
    
    required_level INTEGER DEFAULT 1 CHECK (required_level >= 1),
    required_quest VARCHAR(100),
    
    is_active BOOLEAN DEFAULT TRUE,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);`

const createIndexes = `
CREATE INDEX IF NOT EXISTS idx_zones_type ON zones(type);
CREATE INDEX IF NOT EXISTS idx_zones_level ON zones(level);
CREATE INDEX IF NOT EXISTS idx_zones_status ON zones(status);

CREATE INDEX IF NOT EXISTS idx_player_positions_zone_id ON player_positions(zone_id);
CREATE INDEX IF NOT EXISTS idx_player_positions_user_id ON player_positions(user_id);
CREATE INDEX IF NOT EXISTS idx_player_positions_online ON player_positions(is_online);
CREATE INDEX IF NOT EXISTS idx_player_positions_last_update ON player_positions(last_update);

CREATE INDEX IF NOT EXISTS idx_weather_type ON weather(type);
CREATE INDEX IF NOT EXISTS idx_weather_active ON weather(is_active);

CREATE INDEX IF NOT EXISTS idx_zone_transitions_from_zone ON zone_transitions(from_zone_id);
CREATE INDEX IF NOT EXISTS idx_zone_transitions_to_zone ON zone_transitions(to_zone_id);
CREATE INDEX IF NOT EXISTS idx_zone_transitions_active ON zone_transitions(is_active);
`

const insertDefaultData = `
INSERT INTO zones (id, name, display_name, description, type, level, min_x, min_y, min_z, max_x, max_y, max_z, spawn_x, spawn_y, spawn_z, max_players, is_pvp, is_safe_zone, settings) VALUES
('starting_zone', 'starting_zone', 'Newbie Village', 'A peaceful village where new players begin their journey', 'safe', 1, -100, -10, -100, 100, 50, 100, 0, 0, 0, 100, false, true, '{"weather": "clear", "time_of_day": "day", "background_music": "peaceful_village", "allowed_classes": ["warrior", "mage", "archer", "rogue"], "restricted_items": [], "experience_multiplier": 1.0, "loot_multiplier": 1.0, "death_penalty": "none"}'),
('forest_plains', 'forest_plains', 'Forest Plains', 'Vast plains dotted with ancient trees and roaming wildlife', 'wilderness', 5, -500, -20, -500, 500, 100, 500, 0, 0, 0, 150, false, false, '{"weather": "clear", "time_of_day": "day", "background_music": "forest_ambient", "allowed_classes": ["warrior", "mage", "archer", "rogue"], "restricted_items": [], "experience_multiplier": 1.2, "loot_multiplier": 1.1, "death_penalty": "durability"}'),
('capital_city', 'capital_city', 'Royal Capital', 'The grand capital city with markets, guilds, and royal palace', 'city', 1, -200, -5, -200, 200, 100, 200, 0, 0, 0, 500, false, true, '{"weather": "clear", "time_of_day": "day", "background_music": "city_bustle", "allowed_classes": ["warrior", "mage", "archer", "rogue"], "restricted_items": [], "experience_multiplier": 0.0, "loot_multiplier": 0.0, "death_penalty": "none"}')
ON CONFLICT (id) DO NOTHING;

INSERT INTO weather (zone_id, type, intensity, temperature, wind_speed, wind_direction, visibility) VALUES
('starting_zone', 'clear', 0.2, 22, 5, 90, 1000),
('forest_plains', 'clear', 0.3, 18, 8, 180, 800),
('capital_city', 'clear', 0.2, 20, 6, 135, 1000)
ON CONFLICT (zone_id) DO NOTHING;

INSERT INTO zone_transitions (from_zone_id, to_zone_id, trigger_x, trigger_y, trigger_z, trigger_radius, destination_x, destination_y, destination_z, required_level) VALUES
('starting_zone', 'forest_plains', 95, 0, 0, 5, -95, 0, 0, 3),
('forest_plains', 'starting_zone', -95, 0, 0, 5, 95, 0, 0, 1),
('starting_zone', 'capital_city', -95, 0, 0, 5, 95, 0, 0, 1),
('capital_city', 'starting_zone', 95, 0, 0, 5, -95, 0, 0, 1),
('forest_plains', 'capital_city', 0, 0, -95, 5, 0, 0, 95, 1),
('capital_city', 'forest_plains', 0, 0, 95, 5, 0, 0, -95, 1)
ON CONFLICT DO NOTHING;
`