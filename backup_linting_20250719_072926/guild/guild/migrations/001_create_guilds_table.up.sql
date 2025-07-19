-- Migration pour créer la table des guildes
-- Up: 001_create_guilds_table.up.sql

CREATE TABLE IF NOT EXISTS guilds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    tag VARCHAR(5) NOT NULL UNIQUE,
    level INTEGER NOT NULL DEFAULT 1,
    experience BIGINT NOT NULL DEFAULT 0,
    max_members INTEGER NOT NULL DEFAULT 50,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Index pour améliorer les performances
CREATE INDEX IF NOT EXISTS idx_guilds_name ON guilds(name);
CREATE INDEX IF NOT EXISTS idx_guilds_tag ON guilds(tag);
CREATE INDEX IF NOT EXISTS idx_guilds_level ON guilds(level);
CREATE INDEX IF NOT EXISTS idx_guilds_experience ON guilds(experience);

-- Trigger pour mettre à jour updated_at automatiquement
CREATE OR REPLACE FUNCTION update_guilds_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_guilds_updated_at
    BEFORE UPDATE ON guilds
    FOR EACH ROW
    EXECUTE FUNCTION update_guilds_updated_at();

-- Contraintes de validation
ALTER TABLE guilds ADD CONSTRAINT check_guild_level CHECK (level >= 1);
ALTER TABLE guilds ADD CONSTRAINT check_guild_experience CHECK (experience >= 0);
ALTER TABLE guilds ADD CONSTRAINT check_guild_max_members CHECK (max_members >= 1 AND max_members <= 100);

-- Données de test
INSERT INTO guilds (name, description, tag, level, experience, max_members) VALUES
('Les Gardiens de la Nuit', 'Une guilde dédiée à la protection des nouveaux joueurs', 'GDN', 5, 15000, 75),
('Les Chasseurs de Dragons', 'Spécialisés dans la chasse aux créatures légendaires', 'CDD', 8, 45000, 50),
('Les Marchands du Royaume', 'Guild de commerce et d''artisanat', 'MDR', 3, 8000, 100),
('Les Explorateurs', 'Découvreurs de secrets et de trésors cachés', 'EXP', 6, 25000, 60),
('Les Guerriers de Fer', 'Combattants d''élite pour les batailles PvP', 'GFI', 10, 75000, 40)
ON CONFLICT (name) DO NOTHING; 