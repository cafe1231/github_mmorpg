-- Création de la table world_events
CREATE TABLE IF NOT EXISTS world_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    zone_id VARCHAR(255), -- NULL si événement global
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL CHECK (type IN ('boss_spawn', 'treasure_hunt', 'pvp_tournament', 'seasonal', 'maintenance', 'special')),
    
    -- Timing de l'événement
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE,
    duration INTEGER, -- en minutes
    
    -- Configuration de participation
    max_participants INTEGER DEFAULT 0, -- 0 = illimité
    min_level INTEGER NOT NULL DEFAULT 1 CHECK (min_level >= 1),
    max_level INTEGER DEFAULT 0, -- 0 = pas de limite
    
    -- Récompenses (stockées en JSON)
    rewards JSONB NOT NULL DEFAULT '[]',
    
    -- État de l'événement
    status VARCHAR(20) NOT NULL DEFAULT 'scheduled' CHECK (status IN ('scheduled', 'active', 'completed', 'cancelled')),
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Clé étrangère vers zones (optionnelle)
    FOREIGN KEY (zone_id) REFERENCES zones(id) ON DELETE CASCADE
);

-- Index pour améliorer les performances
CREATE INDEX idx_world_events_zone_id ON world_events(zone_id);
CREATE INDEX idx_world_events_type ON world_events(type);
CREATE INDEX idx_world_events_status ON world_events(status);
CREATE INDEX idx_world_events_start_time ON world_events(start_time);
CREATE INDEX idx_world_events_end_time ON world_events(end_time);
CREATE INDEX idx_world_events_level_range ON world_events(min_level, max_level);

-- Index composé pour les requêtes fréquentes
CREATE INDEX idx_world_events_active ON world_events(status, start_time, end_time);
CREATE INDEX idx_world_events_zone_status ON world_events(zone_id, status);

-- Contrainte pour vérifier la cohérence des niveaux
ALTER TABLE world_events ADD CONSTRAINT check_level_range 
    CHECK (max_level = 0 OR max_level >= min_level);

-- Contrainte pour vérifier la cohérence des temps
ALTER TABLE world_events ADD CONSTRAINT check_time_range 
    CHECK (end_time IS NULL OR end_time > start_time);

-- Trigger pour mettre à jour updated_at
CREATE TRIGGER update_world_events_updated_at 
    BEFORE UPDATE ON world_events 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Fonction pour calculer automatiquement end_time si duration est spécifiée
CREATE OR REPLACE FUNCTION calculate_event_end_time()
RETURNS TRIGGER AS $$
BEGIN
    -- Si duration est spécifiée et end_time n'est pas défini, calculer end_time
    IF NEW.duration IS NOT NULL AND NEW.duration > 0 AND NEW.end_time IS NULL THEN
        NEW.end_time = NEW.start_time + (NEW.duration || ' minutes')::INTERVAL;
    END IF;
    
    -- Si end_time est défini et duration n'est pas définie, calculer duration
    IF NEW.end_time IS NOT NULL AND NEW.duration IS NULL THEN
        NEW.duration = EXTRACT(EPOCH FROM (NEW.end_time - NEW.start_time)) / 60;
    END IF;
    
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER calculate_world_events_end_time 
    BEFORE INSERT OR UPDATE ON world_events 
    FOR EACH ROW 
    EXECUTE FUNCTION calculate_event_end_time();

-- Vue pour les événements actifs avec informations de zone
CREATE OR REPLACE VIEW active_events_with_zones AS
SELECT 
    we.*,
    z.name as zone_name,
    z.display_name as zone_display_name,
    z.type as zone_type,
    CASE 
        WHEN we.end_time IS NOT NULL THEN 
            EXTRACT(EPOCH FROM (we.end_time - CURRENT_TIMESTAMP)) / 60
        ELSE NULL 
    END as minutes_remaining
FROM world_events we
LEFT JOIN zones z ON we.zone_id = z.id
WHERE we.status = 'active'
    AND we.start_time <= CURRENT_TIMESTAMP
    AND (we.end_time IS NULL OR we.end_time > CURRENT_TIMESTAMP);

-- Données d'exemple pour le développement
INSERT INTO world_events (name, description, type, zone_id, start_time, duration, min_level, max_level, rewards, status) VALUES
-- Événement global
('Festival du Printemps', 'Célébration annuelle avec récompenses spéciales', 'seasonal', NULL, 
 CURRENT_TIMESTAMP + INTERVAL '1 hour', 1440, 1, 0, 
 '[{"type": "experience", "amount": 1000}, {"type": "gold", "amount": 500}]', 'scheduled'),

-- Événement dans la forêt
('Invasion des Loups', 'Les loups attaquent en meute dans la forêt', 'boss_spawn', 'dark_forest',
 CURRENT_TIMESTAMP + INTERVAL '30 minutes', 120, 5, 15,
 '[{"type": "experience", "amount": 2000}, {"type": "item", "item_id": "wolf_pelt", "rarity": "rare"}]', 'scheduled'),

-- Tournoi PvP
('Tournoi Hebdomadaire', 'Compétition PvP avec prix en or', 'pvp_tournament', 'pvp_arena',
 CURRENT_TIMESTAMP + INTERVAL '2 hours', 180, 10, 0,
 '[{"type": "gold", "amount": 10000}, {"type": "item", "item_id": "champion_trophy", "rarity": "legendary"}]', 'scheduled'),

-- Chasse au trésor active
('Trésor Caché', 'Trouvez le trésor caché quelque part dans la ville', 'treasure_hunt', 'starter_town',
 CURRENT_TIMESTAMP - INTERVAL '15 minutes', 60, 1, 10,
 '[{"type": "gold", "amount": 1000}, {"type": "experience", "amount": 500}]', 'active'); 