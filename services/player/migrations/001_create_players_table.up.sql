-- Migration pour créer la table des joueurs
-- Version: 001
-- Description: Création de la table players avec toutes les colonnes nécessaires

CREATE TABLE IF NOT EXISTS players (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    level INTEGER DEFAULT 1 CHECK (level >= 1 AND level <= 100),
    experience BIGINT DEFAULT 0 CHECK (experience >= 0),
    gold BIGINT DEFAULT 0 CHECK (gold >= 0),
    location_x FLOAT DEFAULT 0.0,
    location_y FLOAT DEFAULT 0.0,
    location_z FLOAT DEFAULT 0.0,
    zone_id UUID,
    last_login TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- Contraintes
    CONSTRAINT valid_coordinates CHECK (
        location_x BETWEEN -10000 AND 10000 AND
        location_y BETWEEN -10000 AND 10000 AND
        location_z BETWEEN -1000 AND 1000
    )
);

-- Index pour améliorer les performances
CREATE INDEX IF NOT EXISTS idx_players_username ON players(username);
CREATE INDEX IF NOT EXISTS idx_players_email ON players(email);
CREATE INDEX IF NOT EXISTS idx_players_level ON players(level);
CREATE INDEX IF NOT EXISTS idx_players_zone_id ON players(zone_id);
CREATE INDEX IF NOT EXISTS idx_players_location ON players(location_x, location_y, zone_id);
CREATE INDEX IF NOT EXISTS idx_players_created_at ON players(created_at);
CREATE INDEX IF NOT EXISTS idx_players_deleted_at ON players(deleted_at) WHERE deleted_at IS NULL;

-- Trigger pour mettre à jour automatiquement updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_players_updated_at 
    BEFORE UPDATE ON players 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Commentaires pour la documentation
COMMENT ON TABLE players IS 'Table principale des joueurs du MMORPG';
COMMENT ON COLUMN players.id IS 'Identifiant unique du joueur';
COMMENT ON COLUMN players.username IS 'Nom d''utilisateur unique (affiché en jeu)';
COMMENT ON COLUMN players.email IS 'Email unique pour l''authentification';
COMMENT ON COLUMN players.level IS 'Niveau du joueur (1-100)';
COMMENT ON COLUMN players.experience IS 'Points d''expérience accumulés';
COMMENT ON COLUMN players.gold IS 'Monnaie du joueur';
COMMENT ON COLUMN players.zone_id IS 'Zone actuelle du joueur'; 