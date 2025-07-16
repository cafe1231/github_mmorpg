-- Migration de rollback pour la table des logs analytics
-- Down: 003_create_analytics_logs_table.down.sql

-- Supprimer les contraintes
ALTER TABLE analytics_logs DROP CONSTRAINT IF EXISTS check_analytics_logs_level;
ALTER TABLE analytics_logs DROP CONSTRAINT IF EXISTS check_analytics_logs_message;

-- Supprimer les index
DROP INDEX IF EXISTS idx_analytics_logs_level;
DROP INDEX IF EXISTS idx_analytics_logs_timestamp;
DROP INDEX IF EXISTS idx_analytics_logs_level_timestamp;
DROP INDEX IF EXISTS idx_analytics_logs_context;

-- Supprimer la table
DROP TABLE IF EXISTS analytics_logs; 