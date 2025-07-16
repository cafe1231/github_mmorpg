-- Migration pour créer la table des logs analytics
-- Up: 003_create_analytics_logs_table.up.sql

CREATE TABLE IF NOT EXISTS analytics_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    level VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    context TEXT NOT NULL DEFAULT '{}',
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Index pour améliorer les performances
CREATE INDEX IF NOT EXISTS idx_analytics_logs_level ON analytics_logs(level);
CREATE INDEX IF NOT EXISTS idx_analytics_logs_timestamp ON analytics_logs(timestamp);
CREATE INDEX IF NOT EXISTS idx_analytics_logs_level_timestamp ON analytics_logs(level, timestamp);
CREATE INDEX IF NOT EXISTS idx_analytics_logs_context ON analytics_logs USING GIN (context::jsonb);

-- Contraintes de validation
ALTER TABLE analytics_logs ADD CONSTRAINT check_analytics_logs_level CHECK (level IN ('info', 'warn', 'error', 'debug'));
ALTER TABLE analytics_logs ADD CONSTRAINT check_analytics_logs_message CHECK (message != '');

-- Données de test
INSERT INTO analytics_logs (level, message, context) VALUES
('info', 'Service analytics démarré', '{"service":"analytics","version":"1.0.0"}'),
('warn', 'Tentative de connexion échouée', '{"ip":"192.168.1.100","attempts":3}'),
('error', 'Erreur de base de données', '{"table":"players","operation":"insert"}'),
('info', 'Métrique enregistrée', '{"metric":"dau","value":1250}'),
('debug', 'Requête SQL exécutée', '{"query":"SELECT * FROM events","duration":"15ms"}')
ON CONFLICT DO NOTHING; 