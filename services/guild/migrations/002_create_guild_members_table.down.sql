-- Migration de rollback pour la table des membres de guilde
-- Down: 002_create_guild_members_table.down.sql

-- Supprimer les triggers
DROP TRIGGER IF EXISTS trigger_update_guild_members_updated_at ON guild_members;

-- Supprimer les fonctions
DROP FUNCTION IF EXISTS update_guild_members_updated_at();

-- Supprimer les contraintes
ALTER TABLE guild_members DROP CONSTRAINT IF EXISTS check_guild_member_role;
ALTER TABLE guild_members DROP CONSTRAINT IF EXISTS check_guild_member_contribution;

-- Supprimer les index
DROP INDEX IF EXISTS idx_guild_members_guild_id;
DROP INDEX IF EXISTS idx_guild_members_player_id;
DROP INDEX IF EXISTS idx_guild_members_role;
DROP INDEX IF EXISTS idx_guild_members_joined_at;
DROP INDEX IF EXISTS idx_guild_members_last_seen;
DROP INDEX IF EXISTS idx_guild_members_player_unique;

-- Supprimer la table
DROP TABLE IF EXISTS guild_members; 