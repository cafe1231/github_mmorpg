-- Création de la table npcs
CREATE TABLE IF NOT EXISTS npcs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    zone_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('merchant', 'guard', 'quest_giver', 'monster')),
    subtype VARCHAR(100),
    
    -- Position dans le monde 3D
    position_x DECIMAL(15,6) NOT NULL,
    position_y DECIMAL(15,6) NOT NULL,
    position_z DECIMAL(15,6) NOT NULL,
    rotation DECIMAL(8,4) NOT NULL DEFAULT 0.0,
    
    -- Apparence
    model VARCHAR(255) NOT NULL DEFAULT 'default',
    texture VARCHAR(255) NOT NULL DEFAULT 'default',
    scale DECIMAL(4,2) NOT NULL DEFAULT 1.0 CHECK (scale > 0),
    
    -- Comportement (stocké en JSON)
    behavior JSONB NOT NULL DEFAULT '{}',
    
    -- Statistiques de combat
    level INTEGER NOT NULL DEFAULT 1 CHECK (level >= 1),
    health INTEGER NOT NULL DEFAULT 100 CHECK (health >= 0),
    max_health INTEGER NOT NULL DEFAULT 100 CHECK (max_health > 0),
    
    -- État du NPC
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'dead')),
    last_seen TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Clé étrangère vers zones
    FOREIGN KEY (zone_id) REFERENCES zones(id) ON DELETE CASCADE
);

-- Index pour améliorer les performances
CREATE INDEX idx_npcs_zone_id ON npcs(zone_id);
CREATE INDEX idx_npcs_type ON npcs(type);
CREATE INDEX idx_npcs_subtype ON npcs(subtype);
CREATE INDEX idx_npcs_status ON npcs(status);
CREATE INDEX idx_npcs_level ON npcs(level);
CREATE INDEX idx_npcs_last_seen ON npcs(last_seen);

-- Index spatial pour les requêtes de proximité
CREATE INDEX idx_npcs_position ON npcs(zone_id, position_x, position_y, position_z);

-- Index composé pour les requêtes fréquentes
CREATE INDEX idx_npcs_zone_type_status ON npcs(zone_id, type, status);

-- Contrainte pour vérifier que la santé ne dépasse pas la santé max
ALTER TABLE npcs ADD CONSTRAINT check_health_max CHECK (health <= max_health);

-- Trigger pour mettre à jour updated_at et last_seen
CREATE TRIGGER update_npcs_updated_at 
    BEFORE UPDATE ON npcs 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Fonction pour mettre à jour last_seen lors de certaines actions
CREATE OR REPLACE FUNCTION update_npc_last_seen()
RETURNS TRIGGER AS $$
BEGIN
    -- Mettre à jour last_seen si position, santé ou statut change
    IF OLD.position_x != NEW.position_x OR 
       OLD.position_y != NEW.position_y OR 
       OLD.position_z != NEW.position_z OR
       OLD.health != NEW.health OR
       OLD.status != NEW.status THEN
        NEW.last_seen = CURRENT_TIMESTAMP;
    END IF;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_npcs_last_seen 
    BEFORE UPDATE ON npcs 
    FOR EACH ROW 
    EXECUTE FUNCTION update_npc_last_seen();

-- Données d'exemple pour le développement
INSERT INTO npcs (zone_id, name, type, subtype, position_x, position_y, position_z, level, health, max_health, behavior) VALUES
-- NPCs dans la ville de départ
('starter_town', 'Marchand Pierre', 'merchant', 'blacksmith', 10, 15, 5, 5, 150, 150, 
 '{"is_stationary": true, "interaction_text": "Bienvenue dans ma forge! Je peux réparer vos équipements.", "is_hostile": false, "faction": "neutral"}'),
 
('starter_town', 'Garde Marcus', 'guard', 'city_guard', -20, 30, 5, 8, 200, 200,
 '{"is_stationary": true, "patrol_route": [], "interaction_text": "La paix règne dans notre ville.", "is_hostile": false, "faction": "city"}'),

('starter_town', 'Sage Elena', 'quest_giver', 'tutorial', 0, -25, 5, 3, 80, 80,
 '{"is_stationary": true, "interaction_text": "J''ai une tâche importante pour toi, jeune aventurier.", "is_hostile": false, "faction": "neutral"}'),

-- Monstres dans la forêt
('dark_forest', 'Loup Sombre', 'monster', 'wolf', -200, 150, 8, 6, 120, 120,
 '{"is_stationary": false, "patrol_route": [{"x": -200, "y": 150, "z": 8}, {"x": -180, "y": 170, "z": 10}], "aggro_range": 15.0, "chase_range": 30.0, "respawn_time": 300, "is_hostile": true, "faction": "wild"}'),

('dark_forest', 'Ours des Cavernes', 'monster', 'bear', 350, -400, 12, 10, 300, 300,
 '{"is_stationary": false, "patrol_route": [], "aggro_range": 20.0, "chase_range": 25.0, "respawn_time": 600, "is_hostile": true, "faction": "wild"}'),

-- NPC dans l'arène PvP
('pvp_arena', 'Maître d''Armes', 'quest_giver', 'pvp_master', 0, 0, 10, 15, 500, 500,
 '{"is_stationary": true, "interaction_text": "Prêt pour le combat? Montre-moi tes compétences!", "is_hostile": false, "faction": "arena"}'); 