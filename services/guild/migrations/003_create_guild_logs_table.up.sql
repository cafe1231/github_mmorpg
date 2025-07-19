-- Table des logs de guilde
CREATE TABLE guild_logs (
    id UUID PRIMARY KEY,
    guild_id UUID NOT NULL REFERENCES guilds(id) ON DELETE CASCADE,
    player_id UUID NOT NULL,
    action VARCHAR(100) NOT NULL,
    details TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index pour améliorer les performances
CREATE INDEX idx_guild_logs_guild_id ON guild_logs(guild_id);
CREATE INDEX idx_guild_logs_created_at ON guild_logs(created_at);
CREATE INDEX idx_guild_logs_action ON guild_logs(action);

-- Trigger pour nettoyer automatiquement les logs anciens (optionnel)
CREATE OR REPLACE FUNCTION clean_old_guild_logs()
RETURNS void AS $$
BEGIN
    DELETE FROM guild_logs WHERE created_at < NOW() - INTERVAL '30 days';
END;
$$ LANGUAGE plpgsql; 