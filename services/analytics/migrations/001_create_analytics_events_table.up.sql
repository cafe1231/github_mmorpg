-- Migration pour créer la table des événements analytics
-- Up: 001_create_analytics_events_table.up.sql

CREATE TABLE IF NOT EXISTS analytics_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(100) NOT NULL,
    player_id UUID,
    guild_id UUID,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    payload TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Index pour améliorer les performances
CREATE INDEX IF NOT EXISTS idx_analytics_events_type ON analytics_events(type);
CREATE INDEX IF NOT EXISTS idx_analytics_events_timestamp ON analytics_events(timestamp);
CREATE INDEX IF NOT EXISTS idx_analytics_events_player_id ON analytics_events(player_id);
CREATE INDEX IF NOT EXISTS idx_analytics_events_guild_id ON analytics_events(guild_id);
CREATE INDEX IF NOT EXISTS idx_analytics_events_type_timestamp ON analytics_events(type, timestamp);

-- Contraintes de validation
ALTER TABLE analytics_events ADD CONSTRAINT check_analytics_events_type CHECK (type != '');

-- Données de test
INSERT INTO analytics_events (type, player_id, payload) VALUES
('login', gen_random_uuid(), '{"ip":"192.168.1.1","user_agent":"Mozilla/5.0"}'),
('purchase', gen_random_uuid(), '{"item_id":"sword_001","amount":100,"currency":"gold"}'),
('combat', gen_random_uuid(), '{"enemy":"dragon","damage":150,"result":"victory"}'),
('quest_completed', gen_random_uuid(), '{"quest_id":"main_quest_1","reward":"experience"}'),
('guild_joined', gen_random_uuid(), '{"guild_name":"Les Gardiens","role":"member"}')
ON CONFLICT DO NOTHING; 