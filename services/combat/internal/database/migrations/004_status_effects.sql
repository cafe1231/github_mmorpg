-- internal/database/migrations/004_status_effects.sql
-- Migration pour les effets de statut (buffs/debuffs)

-- Types ENUM pour les effets
CREATE TYPE effect_type AS ENUM ('buff', 'debuff', 'poison', 'burn', 'freeze', 'stun', 'slow', 'haste', 'shield', 'regeneration');
CREATE TYPE effect_source AS ENUM ('spell', 'item', 'environment', 'ability', 'consumable');
CREATE TYPE dispel_type AS ENUM ('magic', 'curse', 'poison', 'disease', 'physical', 'none');

-- Table des effets de statut actifs
CREATE TABLE status_effects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    character_id UUID NOT NULL,
    session_id UUID REFERENCES combat_sessions(id) ON DELETE CASCADE, -- null si hors combat
    
    -- Type et source
    type effect_type NOT NULL,
    source effect_source NOT NULL,
    source_id UUID, -- ID du sort/item qui a causé l'effet
    caster_id UUID, -- qui a lancé l'effet
    
    -- Propriétés
    name VARCHAR(100) NOT NULL,
    description TEXT,
    icon VARCHAR(255),
    
    -- Effets sur les stats
    stat_modifiers JSONB NOT NULL DEFAULT '{}', -- StatModifiers
    
    -- Effets périodiques
    periodic_effect JSONB, -- PeriodicEffect
    next_tick_at TIMESTAMP, -- quand le prochain tick doit se produire
    
    -- Timing
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    ends_at TIMESTAMP, -- null pour les effets permanents
    duration INTERVAL,
    
    -- Stack
    stack_count INTEGER NOT NULL DEFAULT 1,
    max_stacks INTEGER NOT NULL DEFAULT 1,
    
    -- Propriétés spéciales
    is_dispellable BOOLEAN NOT NULL DEFAULT TRUE,
    dispel_type dispel_type NOT NULL DEFAULT 'magic',
    priority INTEGER NOT NULL DEFAULT 0, -- pour l'ordre d'application
    
    -- État
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_status_effects_character ON status_effects(character_id);
CREATE INDEX idx_status_effects_session ON status_effects(session_id);
CREATE INDEX idx_status_effects_ends_at ON status_effects(ends_at);
CREATE INDEX idx_status_effects_next_tick ON status_effects(next_tick_at);
CREATE INDEX idx_status_effects_active ON status_effects(character_id, is_active);
CREATE INDEX idx_status_effects_type ON status_effects(type);
CREATE INDEX idx_status_effects_dispel_type ON status_effects(dispel_type);

-- Trigger pour updated_at
CREATE TRIGGER update_status_effects_updated_at 
    BEFORE UPDATE ON status_effects 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Fonction pour appliquer un effet de statut
CREATE OR REPLACE FUNCTION apply_status_effect(
    p_character_id UUID,
    p_session_id UUID,
    p_type effect_type,
    p_source effect_source,
    p_source_id UUID,
    p_caster_id UUID,
    p_name VARCHAR,
    p_description TEXT,
    p_stat_modifiers JSONB,
    p_periodic_effect JSONB,
    p_duration INTERVAL,
    p_max_stacks INTEGER DEFAULT 1,
    p_is_dispellable BOOLEAN DEFAULT TRUE,
    p_dispel_type dispel_type DEFAULT 'magic',
    p_priority INTEGER DEFAULT 0
)
RETURNS UUID AS $$
DECLARE
    effect_id UUID;
    existing_effect RECORD;
    new_ends_at TIMESTAMP;
BEGIN
    -- Calculer la date de fin
    new_ends_at := CASE 
        WHEN p_duration IS NOT NULL THEN NOW() + p_duration 
        ELSE NULL 
    END;
    
    -- Vérifier s'il existe déjà un effet similaire
    SELECT * INTO existing_effect
    FROM status_effects 
    WHERE character_id = p_character_id 
    AND name = p_name 
    AND is_active = TRUE;
    
    IF existing_effect.id IS NOT NULL THEN
        -- Si l'effet existe déjà
        IF existing_effect.stack_count < p_max_stacks THEN
            -- Ajouter une stack
            UPDATE status_effects 
            SET stack_count = stack_count + 1,
                ends_at = new_ends_at,
                duration = p_duration,
                updated_at = NOW()
            WHERE id = existing_effect.id;
            
            effect_id := existing_effect.id;
        ELSE
            -- Renouveler la durée
            UPDATE status_effects 
            SET ends_at = new_ends_at,
                duration = p_duration,
                updated_at = NOW()
            WHERE id = existing_effect.id;
            
            effect_id := existing_effect.id;
        END IF;
    ELSE
        -- Créer un nouvel effet
        INSERT INTO status_effects (
            character_id, session_id, type, source, source_id, caster_id,
            name, description, stat_modifiers, periodic_effect,
            ends_at, duration, max_stacks, is_dispellable, dispel_type, priority,
            next_tick_at
        ) VALUES (
            p_character_id, p_session_id, p_type, p_source, p_source_id, p_caster_id,
            p_name, p_description, p_stat_modifiers, p_periodic_effect,
            new_ends_at, p_duration, p_max_stacks, p_is_dispellable, p_dispel_type, p_priority,
            CASE 
                WHEN p_periodic_effect IS NOT NULL AND p_periodic_effect->>'interval' IS NOT NULL 
                THEN NOW() + (p_periodic_effect->>'interval')::INTERVAL
                ELSE NULL 
            END
        ) RETURNING id INTO effect_id;
    END IF;
    
    RETURN effect_id;
END;
$$ LANGUAGE plpgsql;

-- Fonction pour supprimer un effet de statut
CREATE OR REPLACE FUNCTION remove_status_effect(
    p_effect_id UUID
)
RETURNS BOOLEAN AS $$
DECLARE
    effect_exists BOOLEAN;
BEGIN
    UPDATE status_effects 
    SET is_active = FALSE, updated_at = NOW()
    WHERE id = p_effect_id AND is_active = TRUE;
    
    GET DIAGNOSTICS effect_exists = FOUND;
    RETURN effect_exists;
END;
$$ LANGUAGE plpgsql;

-- Fonction pour traiter les effets périodiques
CREATE OR REPLACE FUNCTION process_periodic_effects()
RETURNS INTEGER AS $$
DECLARE
    effect_record RECORD;
    processed_count INTEGER := 0;
    next_tick TIMESTAMP;
    tick_interval INTERVAL;
BEGIN
    -- Traiter tous les effets périodiques prêts
    FOR effect_record IN 
        SELECT * FROM status_effects 
        WHERE is_active = TRUE 
        AND periodic_effect IS NOT NULL 
        AND next_tick_at IS NOT NULL 
        AND next_tick_at <= NOW()
    LOOP
        -- Appliquer l'effet périodique
        -- (Cette partie sera gérée par le service Go pour plus de flexibilité)
        
        -- Calculer le prochain tick
        tick_interval := (effect_record.periodic_effect->>'interval')::INTERVAL;
        next_tick := NOW() + tick_interval;
        
        -- Décrémenter les ticks restants si défini
        IF effect_record.periodic_effect ? 'ticks_left' THEN
            UPDATE status_effects 
            SET periodic_effect = jsonb_set(
                periodic_effect, 
                '{ticks_left}', 
                ((periodic_effect->>'ticks_left')::INTEGER - 1)::TEXT::JSONB
            ),
            next_tick_at = CASE 
                WHEN ((periodic_effect->>'ticks_left')::INTEGER - 1) > 0 THEN next_tick
                ELSE NULL 
            END,
            updated_at = NOW()
            WHERE id = effect_record.id;
            
            -- Si plus de ticks, désactiver l'effet
            IF ((effect_record.periodic_effect->>'ticks_left')::INTEGER - 1) <= 0 THEN
                UPDATE status_effects 
                SET is_active = FALSE 
                WHERE id = effect_record.id;
            END IF;
        ELSE
            -- Pas de limite de ticks, juste programmer le prochain
            UPDATE status_effects 
            SET next_tick_at = next_tick,
                updated_at = NOW()
            WHERE id = effect_record.id;
        END IF;
        
        processed_count := processed_count + 1;
    END LOOP;
    
    RETURN processed_count;
END;
$$ LANGUAGE plpgsql;

-- Fonction pour nettoyer les effets expirés
CREATE OR REPLACE FUNCTION cleanup_expired_effects()
RETURNS INTEGER AS $$
DECLARE
    cleanup_count INTEGER;
BEGIN
    UPDATE status_effects 
    SET is_active = FALSE, updated_at = NOW()
    WHERE is_active = TRUE 
    AND ends_at IS NOT NULL 
    AND ends_at < NOW();
    
    GET DIAGNOSTICS cleanup_count = ROW_COUNT;
    
    -- Supprimer définitivement les effets inactifs depuis plus de 24h
    DELETE FROM status_effects 
    WHERE is_active = FALSE 
    AND updated_at < NOW() - INTERVAL '24 hours';
    
    RETURN cleanup_count;
END;
$$ LANGUAGE plpgsql;

-- Fonction pour calculer les stats totales d'un personnage avec les effets
CREATE OR REPLACE FUNCTION calculate_character_stats_with_effects(
    p_character_id UUID
)
RETURNS JSONB AS $$
DECLARE
    base_stats JSONB;
    total_modifiers JSONB;
    effect_record RECORD;
    result_stats JSONB;
BEGIN
    -- Stats de base (à récupérer depuis le service Player)
    base_stats := '{
        "health": 100, "mana": 50, "strength": 10, "agility": 10,
        "intelligence": 10, "vitality": 10, "damage": 20, "defense": 10,
        "crit_chance": 0.05, "attack_speed": 1.0
    }'::JSONB;
    
    -- Initialiser les modificateurs totaux
    total_modifiers := '{
        "health_bonus": 0, "mana_bonus": 0, "strength_bonus": 0, "agility_bonus": 0,
        "intelligence_bonus": 0, "vitality_bonus": 0, "damage_bonus": 0, "defense_bonus": 0,
        "health_multiplier": 1.0, "mana_multiplier": 1.0, "damage_multiplier": 1.0,
        "defense_multiplier": 1.0, "speed_multiplier": 1.0, "crit_chance_bonus": 0.0
    }'::JSONB;
    
    -- Additionner tous les modificateurs actifs
    FOR effect_record IN 
        SELECT stat_modifiers FROM status_effects 
        WHERE character_id = p_character_id 
        AND is_active = TRUE 
        ORDER BY priority DESC
    LOOP
        -- Additionner les bonus additifs
        total_modifiers := jsonb_set(total_modifiers, '{health_bonus}', 
            ((total_modifiers->>'health_bonus')::INTEGER + COALESCE((effect_record.stat_modifiers->>'health_bonus')::INTEGER, 0))::TEXT::JSONB);
        
        -- Multiplier les modificateurs multiplicatifs
        total_modifiers := jsonb_set(total_modifiers, '{damage_multiplier}', 
            ((total_modifiers->>'damage_multiplier')::DECIMAL * COALESCE((effect_record.stat_modifiers->>'damage_multiplier')::DECIMAL, 1.0))::TEXT::JSONB);
        
        -- (Ajouter d'autres calculs selon les besoins)
    END LOOP;
    
    -- Calculer les stats finales
    result_stats := jsonb_build_object(
        'final_health', (base_stats->>'health')::INTEGER + (total_modifiers->>'health_bonus')::INTEGER,
        'final_mana', (base_stats->>'mana')::INTEGER + (total_modifiers->>'mana_bonus')::INTEGER,
        'final_damage', ((base_stats->>'damage')::INTEGER + (total_modifiers->>'damage_bonus')::INTEGER) * (total_modifiers->>'damage_multiplier')::DECIMAL,
        'final_defense', ((base_stats->>'defense')::INTEGER + (total_modifiers->>'defense_bonus')::INTEGER) * (total_modifiers->>'defense_multiplier')::DECIMAL,
        'modifiers_applied', total_modifiers
    );
    
    RETURN result_stats;
END;
$$ LANGUAGE plpgsql;

-- Effets de statut prédéfinis pour les sorts de base
INSERT INTO status_effects (character_id, type, source, name, description, stat_modifiers, duration, is_active) VALUES
-- Exemples d'effets (avec un character_id fictif pour la structure)
('00000000-0000-0000-0000-000000000000', 'buff', 'spell', 'Force augmentée', 'La force est temporairement augmentée', 
 '{"strength_bonus": 5, "damage_multiplier": 1.2}'::JSONB, '300 seconds', FALSE),

('00000000-0000-0000-0000-000000000000', 'debuff', 'spell', 'Ralenti', 'Les mouvements sont ralentis', 
 '{"speed_multiplier": 0.5}'::JSONB, '10 seconds', FALSE),

('00000000-0000-0000-0000-000000000000', 'poison', 'spell', 'Empoisonné', 'Perd des points de vie périodiquement', 
 '{}'::JSONB, '30 seconds', FALSE);

-- Supprimer les exemples (ils étaient juste pour la structure)
DELETE FROM status_effects WHERE character_id = '00000000-0000-0000-0000-000000000000';

-- Mise à jour des statistiques
ANALYZE status_effects;