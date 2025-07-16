-- Migration pour créer la table des métriques analytics
-- Up: 002_create_analytics_metrics_table.up.sql

CREATE TABLE IF NOT EXISTS analytics_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    date DATE NOT NULL,
    tags TEXT NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Index pour améliorer les performances
CREATE INDEX IF NOT EXISTS idx_analytics_metrics_name ON analytics_metrics(name);
CREATE INDEX IF NOT EXISTS idx_analytics_metrics_date ON analytics_metrics(date);
CREATE INDEX IF NOT EXISTS idx_analytics_metrics_name_date ON analytics_metrics(name, date);
CREATE INDEX IF NOT EXISTS idx_analytics_metrics_tags ON analytics_metrics USING GIN (tags::jsonb);

-- Contrainte unique pour éviter les doublons
ALTER TABLE analytics_metrics ADD CONSTRAINT unique_metric_name_date_tags UNIQUE (name, date, tags);

-- Contraintes de validation
ALTER TABLE analytics_metrics ADD CONSTRAINT check_analytics_metrics_name CHECK (name != '');
ALTER TABLE analytics_metrics ADD CONSTRAINT check_analytics_metrics_value CHECK (value >= 0);

-- Données de test
INSERT INTO analytics_metrics (name, value, date, tags) VALUES
('dau', 1250, CURRENT_DATE, '{"platform":"web"}'),
('revenue', 5000.50, CURRENT_DATE, '{"currency":"usd"}'),
('pvp_matches', 150, CURRENT_DATE, '{"mode":"ranked"}'),
('quests_completed', 89, CURRENT_DATE, '{"type":"daily"}'),
('guild_activity', 75.5, CURRENT_DATE, '{"metric":"percentage"}')
ON CONFLICT (name, date, tags) DO NOTHING; 