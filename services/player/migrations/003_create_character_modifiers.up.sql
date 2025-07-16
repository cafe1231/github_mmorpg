-- Migration pour créer la table des modificateurs de personnage
-- Version: 003
-- Description: Création de la table character_modifiers pour les buffs/debuffs temporaires

CREATE TABLE IF NOT EXISTS character_modifiers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    modifier_type VARCHAR(50) NOT NULL, -- 'buff', 'debuff', 'equipment', 'skill'
    
    -- Type de statistique affectée
    stat_type VARCHAR(50) NOT NULL, -- 'strength', 'agility', 'health_max', etc.
    
    -- Valeurs du modificateur
    modifier_value INTEGER NOT NULL,
    modifier_percentage FLOAT DEFAULT 0.0, -- Modificateur en pourcentage
    is_percentage BOOLEAN DEFAULT FALSE, -- Si vrai, utilise le pourcentage au lieu de la valeur fixe
    
    -- Timing et durée
    applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE, -- NULL pour les modificateurs permanents
    duration_seconds INTEGER, -- Durée originale en secondes
    
    -- Source du modificateur
    source_type VARCHAR(50), -- 'spell', 'item', 'skill', 'potion', 'environment'
    source_id UUID, -- ID de la source (sort, objet, etc.)
    source_name VARCHAR(100),
    
    -- Stacking et cumul
    stack_count INTEGER DEFAULT 1 CHECK (stack_count >= 1),
    max_stacks INTEGER DEFAULT 1 CHECK (max_stacks >= 1),
    stacks_additively BOOLEAN DEFAULT TRUE, -- Si faux, les stacks se multiplient
    
    -- Métadonnées
    is_active BOOLEAN DEFAULT TRUE,
    can_be_dispelled BOOLEAN DEFAULT TRUE,
    priority INTEGER DEFAULT 0, -- Ordre d'application des modificateurs
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Contraintes
    CONSTRAINT valid_modifier_value CHECK (modifier_value != 0),
    CONSTRAINT valid_percentage CHECK (modifier_percentage BETWEEN -100.0 AND 1000.0),
    CONSTRAINT valid_duration CHECK (duration_seconds IS NULL OR duration_seconds > 0),
    CONSTRAINT valid_stack_logic CHECK (stack_count <= max_stacks)
);

-- Index pour améliorer les performances
CREATE INDEX IF NOT EXISTS idx_character_modifiers_character_id ON character_modifiers(character_id);
CREATE INDEX IF NOT EXISTS idx_character_modifiers_type ON character_modifiers(modifier_type);
CREATE INDEX IF NOT EXISTS idx_character_modifiers_stat_type ON character_modifiers(stat_type);
CREATE INDEX IF NOT EXISTS idx_character_modifiers_expires_at ON character_modifiers(expires_at);
CREATE INDEX IF NOT EXISTS idx_character_modifiers_source ON character_modifiers(source_type, source_id);
CREATE INDEX IF NOT EXISTS idx_character_modifiers_active ON character_modifiers(is_active) WHERE is_active = TRUE;
CREATE INDEX IF NOT EXISTS idx_character_modifiers_priority ON character_modifiers(character_id, priority, applied_at);

-- Index composé pour les requêtes de nettoyage
CREATE INDEX IF NOT EXISTS idx_character_modifiers_cleanup ON character_modifiers(expires_at, is_active) 
    WHERE expires_at IS NOT NULL AND is_active = TRUE;

-- Trigger pour mettre à jour automatiquement updated_at
CREATE TRIGGER update_character_modifiers_updated_at 
    BEFORE UPDATE ON character_modifiers 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Fonction pour nettoyer automatiquement les modificateurs expirés
CREATE OR REPLACE FUNCTION cleanup_expired_modifiers()
RETURNS INTEGER AS $$
DECLARE
    cleaned_count INTEGER;
BEGIN
    UPDATE character_modifiers 
    SET is_active = FALSE, updated_at = CURRENT_TIMESTAMP
    WHERE expires_at IS NOT NULL 
      AND expires_at <= CURRENT_TIMESTAMP 
      AND is_active = TRUE;
    
    GET DIAGNOSTICS cleaned_count = ROW_COUNT;
    RETURN cleaned_count;
END;
$$ language 'plpgsql';

-- Fonction pour appliquer des stacks de modificateurs
CREATE OR REPLACE FUNCTION apply_modifier_stack(
    p_character_id UUID,
    p_name VARCHAR,
    p_stat_type VARCHAR,
    p_modifier_value INTEGER,
    p_duration_seconds INTEGER DEFAULT NULL,
    p_source_type VARCHAR DEFAULT NULL,
    p_source_id UUID DEFAULT NULL,
    p_max_stacks INTEGER DEFAULT 1
)
RETURNS UUID AS $$
DECLARE
    existing_modifier character_modifiers%ROWTYPE;
    modifier_id UUID;
    expires_time TIMESTAMP WITH TIME ZONE;
BEGIN
    -- Calculer le temps d'expiration
    IF p_duration_seconds IS NOT NULL THEN
        expires_time := CURRENT_TIMESTAMP + INTERVAL '1 second' * p_duration_seconds;
    END IF;
    
    -- Chercher un modificateur existant du même type
    SELECT * INTO existing_modifier 
    FROM character_modifiers 
    WHERE character_id = p_character_id 
      AND name = p_name 
      AND stat_type = p_stat_type 
      AND is_active = TRUE
      AND (source_type = p_source_type OR (source_type IS NULL AND p_source_type IS NULL))
    ORDER BY applied_at DESC 
    LIMIT 1;
    
    IF FOUND AND existing_modifier.stack_count < existing_modifier.max_stacks THEN
        -- Augmenter le stack existant
        UPDATE character_modifiers 
        SET stack_count = stack_count + 1,
            expires_at = CASE WHEN p_duration_seconds IS NOT NULL THEN expires_time ELSE expires_at END,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = existing_modifier.id;
        
        modifier_id := existing_modifier.id;
    ELSE
        -- Créer un nouveau modificateur
        INSERT INTO character_modifiers (
            character_id, name, stat_type, modifier_value, 
            duration_seconds, expires_at, source_type, source_id, max_stacks
        ) VALUES (
            p_character_id, p_name, p_stat_type, p_modifier_value,
            p_duration_seconds, expires_time, p_source_type, p_source_id, p_max_stacks
        ) RETURNING id INTO modifier_id;
    END IF;
    
    RETURN modifier_id;
END;
$$ language 'plpgsql';

-- Commentaires pour la documentation
COMMENT ON TABLE character_modifiers IS 'Table des modificateurs temporaires et permanents appliqués aux personnages';
COMMENT ON COLUMN character_modifiers.modifier_type IS 'Type de modificateur: buff, debuff, equipment, skill';
COMMENT ON COLUMN character_modifiers.stat_type IS 'Statistique affectée (strength, agility, health_max, etc.)';
COMMENT ON COLUMN character_modifiers.is_percentage IS 'Si vrai, utilise modifier_percentage au lieu de modifier_value';
COMMENT ON COLUMN character_modifiers.stack_count IS 'Nombre de stacks actuels du modificateur';
COMMENT ON COLUMN character_modifiers.stacks_additively IS 'Si vrai, les stacks s''additionnent, sinon ils se multiplient';
COMMENT ON COLUMN character_modifiers.priority IS 'Ordre d''application (plus élevé = appliqué en dernier)'; 