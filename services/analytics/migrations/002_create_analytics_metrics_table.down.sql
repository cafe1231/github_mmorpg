-- Migration de rollback pour la table des métriques analytics
-- Down: 002_create_analytics_metrics_table.down.sql

-- Supprimer les contraintes
ALTER TABLE analytics_metrics DROP CONSTRAINT IF EXISTS unique_metric_name_date_tags;
ALTER TABLE analytics_metrics DROP CONSTRAINT IF EXISTS check_analytics_metrics_name;
ALTER TABLE analytics_metrics DROP CONSTRAINT IF EXISTS check_analytics_metrics_value;

-- Supprimer les index
DROP INDEX IF EXISTS idx_analytics_metrics_name;
DROP INDEX IF EXISTS idx_analytics_metrics_date;
DROP INDEX IF EXISTS idx_analytics_metrics_name_date;
DROP INDEX IF EXISTS idx_analytics_metrics_tags;

-- Supprimer la table
DROP TABLE IF EXISTS analytics_metrics; 