-- Migration de rollback pour la table des guildes
-- Down: 001_create_guilds_table.down.sql

-- Supprimer les triggers
DROP TRIGGER IF EXISTS trigger_update_guilds_updated_at ON guilds;

-- Supprimer les fonctions
DROP FUNCTION IF EXISTS update_guilds_updated_at();

-- Supprimer les contraintes
ALTER TABLE guilds DROP CONSTRAINT IF EXISTS check_guild_level;
ALTER TABLE guilds DROP CONSTRAINT IF EXISTS check_guild_experience;
ALTER TABLE guilds DROP CONSTRAINT IF EXISTS check_guild_max_members;

-- Supprimer les index
DROP INDEX IF EXISTS idx_guilds_name;
DROP INDEX IF EXISTS idx_guilds_tag;
DROP INDEX IF EXISTS idx_guilds_level;
DROP INDEX IF EXISTS idx_guilds_experience;

-- Supprimer la table
DROP TABLE IF EXISTS guilds; 