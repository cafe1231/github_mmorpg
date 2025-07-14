-- internal/database/migrations/002_spells.sql
-- Migration pour les sorts et système de magie

-- Types ENUM pour les sorts
CREATE TYPE spell_school AS ENUM ('fire', 'water', 'earth', 'air', 'dark', 'light', 'nature', 'arcane');
CREATE TYPE spell_type AS ENUM ('damage', 'heal', 'buff', 'debuff', 'utility', 'summon', 'teleport');
CREATE TYPE target_type AS ENUM ('self', 'single', 'aoe', 'line', 'cone', 'chain');

-- Table des sorts disponibles
CREATE TABLE spells (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    
    -- Classification
    school spell_school NOT NULL,
    type spell_type NOT NULL,
    target_type target_type NOT NULL,
    
    -- Pré-requis
    required_level INTEGER NOT NULL DEFAULT 1,
    required_class TEXT[], -- classes autorisées
    required_stats JSONB, -- StatRequirement
    
    -- Coûts
    mana_cost INTEGER NOT NULL DEFAULT 0,
    health_cost INTEGER NOT NULL DEFAULT 0,
    material_cost JSONB, -- MaterialCost[]
    
    -- Timing
    cast_time INTERVAL NOT NULL DEFAULT '0 seconds',
    cooldown INTERVAL NOT NULL DEFAULT '0 seconds',
    duration INTERVAL, -- pour les sorts persistants
    
    -- Effets
    effects JSONB NOT NULL, -- SpellEffect[]
    
    -- Propriétés
    range DECIMAL(8,2) NOT NULL DEFAULT 1.0, -- portée en mètres
    radius DECIMAL(8,2) NOT NULL DEFAULT 0.0, -- rayon d'effet pour AoE
    accuracy DECIMAL(5,4) NOT NULL DEFAULT 1.0, -- chance de réussir (0-1)
    can_crit BOOLEAN NOT NULL DEFAULT TRUE,
    is_channeled BOOLEAN NOT NULL DEFAULT FALSE,
    requires_target BOOLEAN NOT NULL DEFAULT TRUE,
    
    -- Métadonnées
    icon VARCHAR(255),
    sound_effect VARCHAR(255),
    visual_effect VARCHAR(255),
    
    -- État
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_spells_school ON spells(school);
CREATE INDEX idx_spells_type ON spells(type);
CREATE INDEX idx_spells_required_level ON spells(required_level);
CREATE INDEX idx_spells_is_active ON spells(is_active);

-- Table des sorts appris par les personnages
CREATE TABLE character_spells (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    character_id UUID NOT NULL,
    spell_id UUID NOT NULL REFERENCES spells(id) ON DELETE CASCADE,
    
    -- Progression
    level INTEGER NOT NULL DEFAULT 1, -- niveau de maîtrise du sort
    experience INTEGER NOT NULL DEFAULT 0,
    
    -- Personnalisation
    customizations JSONB DEFAULT '{}',
    
    -- Raccourcis
    hotkey VARCHAR(10),
    slot_number INTEGER, -- position dans la barre de sorts
    
    learned_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    UNIQUE(character_id, spell_id),
    UNIQUE(character_id, slot_number) -- un seul sort par slot
);

CREATE INDEX idx_character_spells_character ON character_spells(character_id);
CREATE INDEX idx_character_spells_spell ON character_spells(spell_id);
CREATE INDEX idx_character_spells_slot ON character_spells(character_id, slot_number);

-- Table des cooldowns de sorts
CREATE TABLE spell_cooldowns (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    character_id UUID NOT NULL,
    spell_id UUID NOT NULL REFERENCES spells(id) ON DELETE CASCADE,
    
    -- Timing
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    ends_at TIMESTAMP NOT NULL,
    duration INTERVAL NOT NULL,
    
    -- État
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    
    UNIQUE(character_id, spell_id)
);

CREATE INDEX idx_spell_cooldowns_character ON spell_cooldowns(character_id);
CREATE INDEX idx_spell_cooldowns_ends_at ON spell_cooldowns(ends_at);
CREATE INDEX idx_spell_cooldowns_active ON spell_cooldowns(character_id, is_active);

-- Trigger pour updated_at sur spells
CREATE TRIGGER update_spells_updated_at 
    BEFORE UPDATE ON spells 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Fonction pour nettoyer les cooldowns expirés
CREATE OR REPLACE FUNCTION cleanup_expired_cooldowns()
RETURNS INTEGER AS $$
DECLARE
    cleanup_count INTEGER;
BEGIN
    DELETE FROM spell_cooldowns 
    WHERE ends_at < NOW();
    
    GET DIAGNOSTICS cleanup_count = ROW_COUNT;
    RETURN cleanup_count;
END;
$$ LANGUAGE plpgsql;

-- Fonction pour appliquer un cooldown
CREATE OR REPLACE FUNCTION apply_spell_cooldown(
    p_character_id UUID,
    p_spell_id UUID,
    p_duration INTERVAL
)
RETURNS VOID AS $$
BEGIN
    INSERT INTO spell_cooldowns (character_id, spell_id, ends_at, duration)
    VALUES (p_character_id, p_spell_id, NOW() + p_duration, p_duration)
    ON CONFLICT (character_id, spell_id) 
    DO UPDATE SET 
        started_at = NOW(),
        ends_at = NOW() + p_duration,
        duration = p_duration,
        is_active = TRUE;
END;
$$ LANGUAGE plpgsql;

-- Fonction pour vérifier si un sort est en cooldown
CREATE OR REPLACE FUNCTION is_spell_on_cooldown(
    p_character_id UUID,
    p_spell_id UUID
)
RETURNS BOOLEAN AS $$
DECLARE
    cooldown_exists BOOLEAN;
BEGIN
    SELECT EXISTS(
        SELECT 1 FROM spell_cooldowns 
        WHERE character_id = p_character_id 
        AND spell_id = p_spell_id 
        AND ends_at > NOW()
        AND is_active = TRUE
    ) INTO cooldown_exists;
    
    RETURN cooldown_exists;
END;
$$ LANGUAGE plpgsql;

-- Insertion de sorts de base pour les tests
INSERT INTO spells (name, description, school, type, target_type, required_level, mana_cost, cast_time, cooldown, effects, range, accuracy) VALUES
-- Sorts de feu
('Boule de feu', 'Lance une boule de feu qui inflige des dégâts de feu', 'fire', 'damage', 'single', 1, 20, '2 seconds', '5 seconds', 
 '[{"type": "damage", "value": 50, "scaling": {"intelligence": 0.8}, "damage_type": "magical", "element": "fire"}]'::jsonb, 
 15.0, 0.95),

('Explosion de flammes', 'Crée une explosion de flammes dans une zone', 'fire', 'damage', 'aoe', 5, 40, '3 seconds', '8 seconds',
 '[{"type": "damage", "value": 80, "scaling": {"intelligence": 1.0}, "damage_type": "magical", "element": "fire"}]'::jsonb,
 12.0, 0.90),

-- Sorts de soin
('Soin mineur', 'Soigne une petite quantité de points de vie', 'light', 'heal', 'single', 1, 15, '1.5 seconds', '3 seconds',
 '[{"type": "heal", "value": 60, "scaling": {"intelligence": 0.6}}]'::jsonb,
 8.0, 1.0),

('Soin de groupe', 'Soigne tous les alliés dans une zone', 'light', 'heal', 'aoe', 8, 60, '4 seconds', '15 seconds',
 '[{"type": "heal", "value": 100, "scaling": {"intelligence": 0.8}}]'::jsonb,
 10.0, 1.0),

-- Sorts de buff
('Bénédiction de force', 'Augmente la force de la cible', 'light', 'buff', 'single', 3, 25, '2 seconds', '0 seconds',
 '[{"type": "buff", "status_effect": "strength_boost", "duration": "300 seconds"}]'::jsonb,
 6.0, 1.0),

-- Sorts de contrôle
('Ralentissement', 'Ralentit les mouvements de la cible', 'arcane', 'debuff', 'single', 4, 30, '1 second', '6 seconds',
 '[{"type": "debuff", "status_effect": "slow", "duration": "10 seconds"}]'::jsonb,
 12.0, 0.85),

-- Sorts utilitaires
('Téléportation', 'Téléporte le lanceur à un endroit', 'arcane', 'teleport', 'self', 10, 50, '3 seconds', '30 seconds',
 '[{"type": "teleport"}]'::jsonb,
 20.0, 1.0);

-- Mise à jour des statistiques
ANALYZE spells;
ANALYZE character_spells;
ANALYZE spell_cooldowns;