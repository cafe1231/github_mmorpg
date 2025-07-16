-- Création de la table zones
CREATE TABLE IF NOT EXISTS zones (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL CHECK (type IN ('city', 'dungeon', 'wilderness', 'pvp', 'safe')),
    level INTEGER NOT NULL DEFAULT 1 CHECK (level >= 1),
    
    -- Géométrie de la zone (bounding box)
    min_x DECIMAL(15,6) NOT NULL,
    min_y DECIMAL(15,6) NOT NULL,
    min_z DECIMAL(15,6) NOT NULL,
    max_x DECIMAL(15,6) NOT NULL,
    max_y DECIMAL(15,6) NOT NULL,
    max_z DECIMAL(15,6) NOT NULL,
    
    -- Points de spawn par défaut
    spawn_x DECIMAL(15,6) NOT NULL,
    spawn_y DECIMAL(15,6) NOT NULL,
    spawn_z DECIMAL(15,6) NOT NULL,
    
    -- Configuration de la zone
    max_players INTEGER NOT NULL DEFAULT 100 CHECK (max_players > 0),
    is_pvp BOOLEAN NOT NULL DEFAULT FALSE,
    is_safe_zone BOOLEAN NOT NULL DEFAULT FALSE,
    settings JSONB NOT NULL DEFAULT '{}',
    
    -- État de la zone
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'maintenance', 'disabled')),
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Index pour améliorer les performances
CREATE INDEX idx_zones_type ON zones(type);
CREATE INDEX idx_zones_level ON zones(level);
CREATE INDEX idx_zones_status ON zones(status);
CREATE INDEX idx_zones_is_pvp ON zones(is_pvp);
CREATE INDEX idx_zones_is_safe_zone ON zones(is_safe_zone);

-- Index spatial pour les requêtes géométriques
CREATE INDEX idx_zones_bounds ON zones(min_x, min_y, min_z, max_x, max_y, max_z);

-- Contraintes pour vérifier la cohérence des coordonnées
ALTER TABLE zones ADD CONSTRAINT check_zone_bounds_x CHECK (min_x < max_x);
ALTER TABLE zones ADD CONSTRAINT check_zone_bounds_y CHECK (min_y < max_y);
ALTER TABLE zones ADD CONSTRAINT check_zone_bounds_z CHECK (min_z < max_z);

-- Contraintes pour vérifier que le spawn est dans les limites
ALTER TABLE zones ADD CONSTRAINT check_spawn_in_bounds_x CHECK (spawn_x >= min_x AND spawn_x <= max_x);
ALTER TABLE zones ADD CONSTRAINT check_spawn_in_bounds_y CHECK (spawn_y >= min_y AND spawn_y <= max_y);
ALTER TABLE zones ADD CONSTRAINT check_spawn_in_bounds_z CHECK (spawn_z >= min_z AND spawn_z <= max_z);

-- Trigger pour mettre à jour updated_at automatiquement
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_zones_updated_at 
    BEFORE UPDATE ON zones 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Données d'exemple pour le développement
INSERT INTO zones (id, name, display_name, description, type, level, min_x, min_y, min_z, max_x, max_y, max_z, spawn_x, spawn_y, spawn_z, max_players, is_safe_zone) VALUES
('starter_town', 'starter_town', 'Ville du Débutant', 'Zone de départ pour les nouveaux joueurs', 'city', 1, -100, -100, 0, 100, 100, 50, 0, 0, 5, 50, true),
('dark_forest', 'dark_forest', 'Forêt Sombre', 'Une forêt mystérieuse remplie de créatures dangereuses', 'wilderness', 5, -500, -500, 0, 500, 500, 100, 0, 0, 10, 30, false),
('pvp_arena', 'pvp_arena', 'Arène PvP', 'Zone de combat joueur contre joueur', 'pvp', 10, -50, -50, 0, 50, 50, 20, 0, 0, 5, 20, false);

UPDATE zones SET settings = jsonb_build_object(
    'weather', 'clear',
    'time_of_day', 'day',
    'background_music', 'peaceful',
    'allowed_classes', '[]',
    'restricted_items', '[]',
    'experience_multiplier', 1.0,
    'loot_multiplier', 1.0,
    'death_penalty', 'durability'
); 