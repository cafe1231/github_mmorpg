-- Migration de retour pour supprimer la table des personnages
-- Version: 002
-- Description: Suppression de la table characters et de ses triggers/fonctions

-- Supprimer les triggers
DROP TRIGGER IF EXISTS update_characters_updated_at ON characters;
DROP TRIGGER IF EXISTS calculate_character_stats_trigger ON characters;

-- Supprimer les fonctions
DROP FUNCTION IF EXISTS calculate_character_stats();

-- Supprimer les index
DROP INDEX IF EXISTS idx_characters_player_id;
DROP INDEX IF EXISTS idx_characters_name;
DROP INDEX IF EXISTS idx_characters_class;
DROP INDEX IF EXISTS idx_characters_race;
DROP INDEX IF EXISTS idx_characters_level;
DROP INDEX IF EXISTS idx_characters_zone_id;
DROP INDEX IF EXISTS idx_characters_location;
DROP INDEX IF EXISTS idx_characters_last_played;
DROP INDEX IF EXISTS idx_characters_stats;
DROP INDEX IF EXISTS idx_characters_appearance;

-- Supprimer la table
DROP TABLE IF EXISTS characters; 