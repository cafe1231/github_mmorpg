-- Suppression de la table npcs et des éléments associés

-- Supprimer les triggers
DROP TRIGGER IF EXISTS update_npcs_last_seen ON npcs;
DROP TRIGGER IF EXISTS update_npcs_updated_at ON npcs;

-- Supprimer la fonction trigger spécifique aux NPCs
DROP FUNCTION IF EXISTS update_npc_last_seen();

-- Supprimer les index
DROP INDEX IF EXISTS idx_npcs_zone_id;
DROP INDEX IF EXISTS idx_npcs_type;
DROP INDEX IF EXISTS idx_npcs_subtype;
DROP INDEX IF EXISTS idx_npcs_status;
DROP INDEX IF EXISTS idx_npcs_level;
DROP INDEX IF EXISTS idx_npcs_last_seen;
DROP INDEX IF EXISTS idx_npcs_position;
DROP INDEX IF EXISTS idx_npcs_zone_type_status;

-- Supprimer la table
DROP TABLE IF EXISTS npcs; 