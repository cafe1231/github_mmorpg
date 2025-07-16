-- Création de la table player_positions
CREATE TABLE IF NOT EXISTS player_positions (
    character_id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    zone_id VARCHAR(255) NOT NULL,
    
    -- Position actuelle dans le monde 3D
    x DECIMAL(15,6) NOT NULL DEFAULT 0.0,
    y DECIMAL(15,6) NOT NULL DEFAULT 0.0,
    z DECIMAL(15,6) NOT NULL DEFAULT 0.0,
    rotation DECIMAL(8,4) NOT NULL DEFAULT 0.0,
    
    -- Vecteur de mouvement
    velocity_x DECIMAL(10,6) NOT NULL DEFAULT 0.0,
    velocity_y DECIMAL(10,6) NOT NULL DEFAULT 0.0,
    velocity_z DECIMAL(10,6) NOT NULL DEFAULT 0.0,
    is_moving BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- État de connexion
    is_online BOOLEAN NOT NULL DEFAULT FALSE,
    last_update TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Clé étrangère vers zones
    FOREIGN KEY (zone_id) REFERENCES zones(id) ON DELETE CASCADE
);

-- Index pour améliorer les performances
CREATE INDEX idx_player_positions_user_id ON player_positions(user_id);
CREATE INDEX idx_player_positions_zone_id ON player_positions(zone_id);
CREATE INDEX idx_player_positions_is_online ON player_positions(is_online);
CREATE INDEX idx_player_positions_last_update ON player_positions(last_update);
CREATE INDEX idx_player_positions_is_moving ON player_positions(is_moving);

-- Index spatial pour les requêtes de proximité
CREATE INDEX idx_player_positions_location ON player_positions(zone_id, x, y, z);

-- Index composé pour les requêtes fréquentes
CREATE INDEX idx_player_positions_zone_online ON player_positions(zone_id, is_online);

-- Trigger pour mettre à jour updated_at
CREATE TRIGGER update_player_positions_updated_at 
    BEFORE UPDATE ON player_positions 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Fonction pour mettre à jour last_update lors des changements de position
CREATE OR REPLACE FUNCTION update_player_position_last_update()
RETURNS TRIGGER AS $$
BEGIN
    -- Mettre à jour last_update si position ou mouvement change
    IF OLD.x != NEW.x OR 
       OLD.y != NEW.y OR 
       OLD.z != NEW.z OR
       OLD.rotation != NEW.rotation OR
       OLD.velocity_x != NEW.velocity_x OR
       OLD.velocity_y != NEW.velocity_y OR
       OLD.velocity_z != NEW.velocity_z OR
       OLD.is_moving != NEW.is_moving THEN
        NEW.last_update = CURRENT_TIMESTAMP;
    END IF;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_player_positions_last_update 
    BEFORE UPDATE ON player_positions 
    FOR EACH ROW 
    EXECUTE FUNCTION update_player_position_last_update();

-- Vue pour obtenir les statistiques de population par zone
CREATE OR REPLACE VIEW zone_population AS
SELECT 
    z.id as zone_id,
    z.name as zone_name,
    z.display_name,
    z.max_players,
    COUNT(pp.character_id) as current_players,
    COUNT(CASE WHEN pp.is_online THEN 1 END) as online_players,
    ROUND((COUNT(pp.character_id)::DECIMAL / z.max_players) * 100, 2) as occupancy_percentage
FROM zones z
LEFT JOIN player_positions pp ON z.id = pp.zone_id
GROUP BY z.id, z.name, z.display_name, z.max_players
ORDER BY current_players DESC;

-- Index sur la vue pour améliorer les performances
CREATE INDEX idx_zone_population_stats ON player_positions(zone_id, is_online);

-- Données d'exemple pour le développement (positions de test)
-- Note: Ces données seront remplacées par les vraies positions des joueurs
INSERT INTO player_positions (character_id, user_id, zone_id, x, y, z, is_online) VALUES
('123e4567-e89b-12d3-a456-426614174000', '123e4567-e89b-12d3-a456-426614174001', 'starter_town', 0, 0, 5, true),
('223e4567-e89b-12d3-a456-426614174000', '223e4567-e89b-12d3-a456-426614174001', 'starter_town', 25, -15, 5, true),
('323e4567-e89b-12d3-a456-426614174000', '323e4567-e89b-12d3-a456-426614174001', 'dark_forest', -100, 200, 8, false); 