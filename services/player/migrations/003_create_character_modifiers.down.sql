-- Migration de retour pour supprimer la table des modificateurs de personnage
-- Version: 003
-- Description: Suppression de la table character_modifiers et de ses fonctions

-- Supprimer les triggers
DROP TRIGGER IF EXISTS update_character_modifiers_updated_at ON character_modifiers;

-- Supprimer les fonctions
DROP FUNCTION IF EXISTS cleanup_expired_modifiers();
DROP FUNCTION IF EXISTS apply_modifier_stack(UUID, VARCHAR, VARCHAR, INTEGER, INTEGER, VARCHAR, UUID, INTEGER);

-- Supprimer les index
DROP INDEX IF EXISTS idx_character_modifiers_character_id;
DROP INDEX IF EXISTS idx_character_modifiers_type;
DROP INDEX IF EXISTS idx_character_modifiers_stat_type;
DROP INDEX IF EXISTS idx_character_modifiers_expires_at;
DROP INDEX IF EXISTS idx_character_modifiers_source;
DROP INDEX IF EXISTS idx_character_modifiers_active;
DROP INDEX IF EXISTS idx_character_modifiers_priority;
DROP INDEX IF EXISTS idx_character_modifiers_cleanup;

-- Supprimer la table
DROP TABLE IF EXISTS character_modifiers; 