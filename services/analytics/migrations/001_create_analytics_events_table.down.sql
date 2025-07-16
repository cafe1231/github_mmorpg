-- Migration de rollback pour la table des événements analytics
-- Down: 001_create_analytics_events_table.down.sql

-- Supprimer les contraintes
ALTER TABLE analytics_events DROP CONSTRAINT IF EXISTS check_analytics_events_type;

-- Supprimer les index
DROP INDEX IF EXISTS idx_analytics_events_type;
DROP INDEX IF EXISTS idx_analytics_events_timestamp;
DROP INDEX IF EXISTS idx_analytics_events_player_id;
DROP INDEX IF EXISTS idx_analytics_events_guild_id;
DROP INDEX IF EXISTS idx_analytics_events_type_timestamp;

-- Supprimer la table
DROP TABLE IF EXISTS analytics_events; 