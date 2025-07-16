-- Migration pour créer la table des membres de guilde
-- Up: 002_create_guild_members_table.up.sql

CREATE TABLE IF NOT EXISTS guild_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    guild_id UUID NOT NULL REFERENCES guilds(id) ON DELETE CASCADE,
    player_id UUID NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'member',
    joined_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_seen TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    contribution BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Index pour améliorer les performances
CREATE INDEX IF NOT EXISTS idx_guild_members_guild_id ON guild_members(guild_id);
CREATE INDEX IF NOT EXISTS idx_guild_members_player_id ON guild_members(player_id);
CREATE INDEX IF NOT EXISTS idx_guild_members_role ON guild_members(role);
CREATE INDEX IF NOT EXISTS idx_guild_members_joined_at ON guild_members(joined_at);
CREATE INDEX IF NOT EXISTS idx_guild_members_last_seen ON guild_members(last_seen);

-- Index unique pour éviter qu'un joueur soit dans plusieurs guildes
CREATE UNIQUE INDEX IF NOT EXISTS idx_guild_members_player_unique ON guild_members(player_id);

-- Trigger pour mettre à jour updated_at automatiquement
CREATE OR REPLACE FUNCTION update_guild_members_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_guild_members_updated_at
    BEFORE UPDATE ON guild_members
    FOR EACH ROW
    EXECUTE FUNCTION update_guild_members_updated_at();

-- Contraintes de validation
ALTER TABLE guild_members ADD CONSTRAINT check_guild_member_role CHECK (role IN ('leader', 'officer', 'member'));
ALTER TABLE guild_members ADD CONSTRAINT check_guild_member_contribution CHECK (contribution >= 0);

-- Données de test
INSERT INTO guild_members (guild_id, player_id, role, contribution) 
SELECT 
    g.id,
    gen_random_uuid(),
    'leader',
    1000
FROM guilds g
WHERE g.name = 'Les Gardiens de la Nuit'
ON CONFLICT (player_id) DO NOTHING;

INSERT INTO guild_members (guild_id, player_id, role, contribution) 
SELECT 
    g.id,
    gen_random_uuid(),
    'officer',
    750
FROM guilds g
WHERE g.name = 'Les Chasseurs de Dragons'
ON CONFLICT (player_id) DO NOTHING;

INSERT INTO guild_members (guild_id, player_id, role, contribution) 
SELECT 
    g.id,
    gen_random_uuid(),
    'member',
    500
FROM guilds g
WHERE g.name = 'Les Marchands du Royaume'
ON CONFLICT (player_id) DO NOTHING; 