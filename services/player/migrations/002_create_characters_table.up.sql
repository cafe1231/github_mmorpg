-- Migration pour créer la table des personnages
-- Version: 002
-- Description: Création de la table characters avec système de statistiques avancé

CREATE TABLE IF NOT EXISTS characters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    class VARCHAR(50) NOT NULL,
    race VARCHAR(50) NOT NULL,
    level INTEGER DEFAULT 1 CHECK (level >= 1 AND level <= 100),
    experience BIGINT DEFAULT 0 CHECK (experience >= 0),
    
    -- Statistiques de base
    health_current INTEGER DEFAULT 100 CHECK (health_current >= 0),
    health_max INTEGER DEFAULT 100 CHECK (health_max > 0),
    mana_current INTEGER DEFAULT 50 CHECK (mana_current >= 0),
    mana_max INTEGER DEFAULT 50 CHECK (mana_max > 0),
    energy_current INTEGER DEFAULT 100 CHECK (energy_current >= 0),
    energy_max INTEGER DEFAULT 100 CHECK (energy_max > 0),
    
    -- Statistiques principales
    strength INTEGER DEFAULT 10 CHECK (strength >= 1 AND strength <= 1000),
    agility INTEGER DEFAULT 10 CHECK (agility >= 1 AND agility <= 1000),
    intellect INTEGER DEFAULT 10 CHECK (intellect >= 1 AND intellect <= 1000),
    vitality INTEGER DEFAULT 10 CHECK (vitality >= 1 AND vitality <= 1000),
    endurance INTEGER DEFAULT 10 CHECK (endurance >= 1 AND endurance <= 1000),
    luck INTEGER DEFAULT 10 CHECK (luck >= 1 AND luck <= 1000),
    
    -- Statistiques calculées (pour optimisation)
    attack_power INTEGER DEFAULT 0,
    spell_power INTEGER DEFAULT 0,
    defense_rating INTEGER DEFAULT 0,
    critical_chance FLOAT DEFAULT 0.05 CHECK (critical_chance >= 0 AND critical_chance <= 1),
    critical_damage FLOAT DEFAULT 1.5 CHECK (critical_damage >= 1 AND critical_damage <= 10),
    movement_speed FLOAT DEFAULT 1.0 CHECK (movement_speed >= 0.1 AND movement_speed <= 5.0),
    
    -- Position et apparence
    location_x FLOAT DEFAULT 0.0,
    location_y FLOAT DEFAULT 0.0,
    location_z FLOAT DEFAULT 0.0,
    zone_id UUID,
    appearance_data JSONB DEFAULT '{}',
    
    -- Métadonnées
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_played TIMESTAMP WITH TIME ZONE,
    play_time_seconds BIGINT DEFAULT 0,
    
    -- Contraintes
    CONSTRAINT valid_character_coordinates CHECK (
        location_x BETWEEN -10000 AND 10000 AND
        location_y BETWEEN -10000 AND 10000 AND
        location_z BETWEEN -1000 AND 1000
    ),
    CONSTRAINT valid_health CHECK (health_current <= health_max),
    CONSTRAINT valid_mana CHECK (mana_current <= mana_max),
    CONSTRAINT valid_energy CHECK (energy_current <= energy_max),
    CONSTRAINT unique_character_name_per_player UNIQUE(player_id, name)
);

-- Index pour améliorer les performances
CREATE INDEX IF NOT EXISTS idx_characters_player_id ON characters(player_id);
CREATE INDEX IF NOT EXISTS idx_characters_name ON characters(name);
CREATE INDEX IF NOT EXISTS idx_characters_class ON characters(class);
CREATE INDEX IF NOT EXISTS idx_characters_race ON characters(race);
CREATE INDEX IF NOT EXISTS idx_characters_level ON characters(level);
CREATE INDEX IF NOT EXISTS idx_characters_zone_id ON characters(zone_id);
CREATE INDEX IF NOT EXISTS idx_characters_location ON characters(location_x, location_y, zone_id);
CREATE INDEX IF NOT EXISTS idx_characters_last_played ON characters(last_played);
CREATE INDEX IF NOT EXISTS idx_characters_stats ON characters(strength, agility, intellect, vitality);

-- Index JSONB pour l'apparence
CREATE INDEX IF NOT EXISTS idx_characters_appearance ON characters USING GIN (appearance_data);

-- Trigger pour mettre à jour automatiquement updated_at
CREATE TRIGGER update_characters_updated_at 
    BEFORE UPDATE ON characters 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Fonction pour recalculer les statistiques dérivées
CREATE OR REPLACE FUNCTION calculate_character_stats()
RETURNS TRIGGER AS $$
BEGIN
    -- Calculer l'attaque physique basée sur la force
    NEW.attack_power := NEW.strength * 2 + NEW.level;
    
    -- Calculer la puissance magique basée sur l'intellect
    NEW.spell_power := NEW.intellect * 2 + NEW.level;
    
    -- Calculer la défense basée sur la vitalité et l'endurance
    NEW.defense_rating := (NEW.vitality + NEW.endurance) + NEW.level;
    
    -- Calculer les chances de critique basées sur l'agilité et la chance
    NEW.critical_chance := LEAST(0.75, 0.05 + (NEW.agility + NEW.luck) * 0.001);
    
    -- Calculer la vitesse de déplacement basée sur l'agilité
    NEW.movement_speed := LEAST(3.0, 1.0 + NEW.agility * 0.01);
    
    -- Ajuster les PV max basés sur la vitalité
    NEW.health_max := 100 + NEW.vitality * 5 + NEW.level * 10;
    
    -- Ajuster le mana max basé sur l'intellect
    NEW.mana_max := 50 + NEW.intellect * 3 + NEW.level * 5;
    
    -- Ajuster l'énergie max basée sur l'endurance
    NEW.energy_max := 100 + NEW.endurance * 2 + NEW.level * 3;
    
    -- S'assurer que les valeurs actuelles ne dépassent pas les maximums
    NEW.health_current := LEAST(NEW.health_current, NEW.health_max);
    NEW.mana_current := LEAST(NEW.mana_current, NEW.mana_max);
    NEW.energy_current := LEAST(NEW.energy_current, NEW.energy_max);
    
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger pour recalculer les stats à chaque mise à jour
CREATE TRIGGER calculate_character_stats_trigger
    BEFORE INSERT OR UPDATE ON characters
    FOR EACH ROW
    EXECUTE FUNCTION calculate_character_stats();

-- Commentaires pour la documentation
COMMENT ON TABLE characters IS 'Table des personnages jouables avec système de statistiques complet';
COMMENT ON COLUMN characters.player_id IS 'Référence vers le joueur propriétaire';
COMMENT ON COLUMN characters.name IS 'Nom du personnage (unique par joueur)';
COMMENT ON COLUMN characters.class IS 'Classe du personnage (Warrior, Mage, Archer, etc.)';
COMMENT ON COLUMN characters.race IS 'Race du personnage (Human, Elf, Dwarf, etc.)';
COMMENT ON COLUMN characters.appearance_data IS 'Données JSON pour l''apparence personnalisée';
COMMENT ON COLUMN characters.play_time_seconds IS 'Temps de jeu total en secondes'; 