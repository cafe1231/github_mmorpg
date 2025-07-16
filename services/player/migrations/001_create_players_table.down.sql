-- Migration de retour pour supprimer la table des joueurs
-- Version: 001
-- Description: Suppression de la table players et de ses triggers

-- Supprimer le trigger et sa fonction
DROP TRIGGER IF EXISTS update_players_updated_at ON players;
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Supprimer les index
DROP INDEX IF EXISTS idx_players_username;
DROP INDEX IF EXISTS idx_players_email;
DROP INDEX IF EXISTS idx_players_level;
DROP INDEX IF EXISTS idx_players_zone_id;
DROP INDEX IF EXISTS idx_players_location;
DROP INDEX IF EXISTS idx_players_created_at;
DROP INDEX IF EXISTS idx_players_deleted_at;

-- Supprimer la table
DROP TABLE IF EXISTS players; 