-- internal/database/migrations/001_initial.sql
-- Migration initiale pour le service Combat

-- Extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Types ENUM pour plus de cohérence
CREATE TYPE combat_type AS ENUM ('pve', 'pvp', 'raid', 'dungeon', 'arena');
CREATE TYPE combat_status AS ENUM ('waiting', 'active', 'ended', 'cancelled');
CREATE TYPE participant_status AS ENUM ('alive', 'dead', 'fled', 'disconnected');
CREATE TYPE action_type AS ENUM ('attack', 'spell', 'move', 'item', 'defend', 'flee');

-- Table des sessions de combat
CREATE TABLE combat_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type combat_type NOT NULL,
    status combat_status NOT NULL DEFAULT 'waiting',
    zone_id VARCHAR(100) NOT NULL,
    created_by UUID NOT NULL,
    
    -- Configuration
    max_participants INTEGER NOT NULL DEFAULT 8,
    is_private BOOLEAN NOT NULL DEFAULT FALSE,
    level_range JSONB, -- {min: int, max: int}
    rules JSONB NOT NULL DEFAULT '{}', -- CombatRules
    password_hash VARCHAR(255), -- pour les combats privés
    
    -- Timing
    started_at TIMESTAMP,
    ended_at TIMESTAMP,
    last_action_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index pour les requêtes fréquentes
CREATE INDEX idx_combat_sessions_type_status ON combat_sessions(type, status);
CREATE INDEX idx_combat_sessions_zone_id ON combat_sessions(zone_id);
CREATE INDEX idx_combat_sessions_created_by ON combat_sessions(created_by);
CREATE INDEX idx_combat_sessions_status_last_action ON combat_sessions(status, last_action_at);

-- Table des participants au combat
CREATE TABLE combat_participants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES combat_sessions(id) ON DELETE CASCADE,
    character_id UUID NOT NULL,
    player_id UUID NOT NULL,
    
    -- État du participant
    team INTEGER NOT NULL DEFAULT 0, -- 0=neutral, 1=team1, 2=team2
    position JSONB NOT NULL, -- {x: float, y: float, z: float}
    status participant_status NOT NULL DEFAULT 'alive',
    
    -- Stats de combat actuelles
    current_health INTEGER NOT NULL,
    max_health INTEGER NOT NULL,
    current_mana INTEGER NOT NULL,
    max_mana INTEGER NOT NULL,
    
    -- Stats calculées avec les effets
    damage INTEGER NOT NULL,
    defense INTEGER NOT NULL,
    crit_chance DECIMAL(5,4) NOT NULL DEFAULT 0.05, -- 5%
    attack_speed DECIMAL(5,4) NOT NULL DEFAULT 1.0,
    
    -- Timing
    joined_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_action_at TIMESTAMP,
    
    UNIQUE(session_id, character_id)
);

CREATE INDEX idx_combat_participants_session ON combat_participants(session_id);
CREATE INDEX idx_combat_participants_character ON combat_participants(character_id);
CREATE INDEX idx_combat_participants_status ON combat_participants(status);

-- Table des actions de combat
CREATE TABLE combat_actions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES combat_sessions(id) ON DELETE CASCADE,
    actor_id UUID NOT NULL, -- character_id qui fait l'action
    
    -- Type d'action
    type action_type NOT NULL,
    action_data JSONB NOT NULL DEFAULT '{}',
    
    -- Cibles
    targets UUID[] NOT NULL DEFAULT '{}',
    
    -- Résultats
    results JSONB NOT NULL DEFAULT '[]', -- ActionResult[]
    success BOOLEAN NOT NULL DEFAULT TRUE,
    critical_hit BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Timing
    executed_at TIMESTAMP NOT NULL DEFAULT NOW(),
    duration INTERVAL NOT NULL DEFAULT '0 seconds'
);

CREATE INDEX idx_combat_actions_session ON combat_actions(session_id);
CREATE INDEX idx_combat_actions_actor ON combat_actions(actor_id);
CREATE INDEX idx_combat_actions_executed_at ON combat_actions(executed_at);
CREATE INDEX idx_combat_actions_type ON combat_actions(type);

-- Table des logs de combat
CREATE TABLE combat_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES combat_sessions(id) ON DELETE CASCADE,
    action_id UUID REFERENCES combat_actions(id) ON DELETE SET NULL,
    
    -- Participants
    actor_id UUID, -- character_id qui a fait l'action
    target_id UUID, -- character_id qui a reçu l'action
    
    -- Type d'événement
    event_type VARCHAR(50) NOT NULL, -- damage, heal, death, spell_cast, effect_applied, etc.
    
    -- Messages
    message TEXT NOT NULL, -- message formaté pour l'affichage
    raw_data JSONB, -- données brutes de l'événement
    
    -- Valeurs numériques
    value INTEGER DEFAULT 0, -- dégâts, soins, etc.
    old_value INTEGER DEFAULT 0, -- valeur avant l'action
    new_value INTEGER DEFAULT 0, -- valeur après l'action
    
    -- Contexte
    is_critical BOOLEAN NOT NULL DEFAULT FALSE,
    is_resisted BOOLEAN NOT NULL DEFAULT FALSE,
    is_absorbed BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Métadonnées
    color VARCHAR(20) DEFAULT '#FFFFFF', -- couleur pour l'affichage
    icon VARCHAR(100),
    priority INTEGER NOT NULL DEFAULT 0, -- pour le tri d'affichage
    
    timestamp TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_combat_logs_session ON combat_logs(session_id);
CREATE INDEX idx_combat_logs_action ON combat_logs(action_id);
CREATE INDEX idx_combat_logs_actor ON combat_logs(actor_id);
CREATE INDEX idx_combat_logs_target ON combat_logs(target_id);
CREATE INDEX idx_combat_logs_timestamp ON combat_logs(timestamp);
CREATE INDEX idx_combat_logs_event_type ON combat_logs(event_type);

-- Fonction pour mettre à jour updated_at automatiquement
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers pour updated_at
CREATE TRIGGER update_combat_sessions_updated_at 
    BEFORE UPDATE ON combat_sessions 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Fonction pour nettoyer les sessions inactives
CREATE OR REPLACE FUNCTION cleanup_inactive_sessions()
RETURNS INTEGER AS $$
DECLARE
    cleanup_count INTEGER;
BEGIN
    -- Marquer comme terminées les sessions inactives depuis plus de 2 heures
    UPDATE combat_sessions 
    SET status = 'ended', ended_at = NOW()
    WHERE status IN ('waiting', 'active') 
    AND last_action_at < NOW() - INTERVAL '2 hours';
    
    GET DIAGNOSTICS cleanup_count = ROW_COUNT;
    
    -- Log le nettoyage
    INSERT INTO combat_logs (session_id, event_type, message, timestamp)
    SELECT id, 'session_cleanup', 'Session automatiquement terminée pour inactivité', NOW()
    FROM combat_sessions 
    WHERE status = 'ended' AND ended_at > NOW() - INTERVAL '1 minute';
    
    RETURN cleanup_count;
END;
$$ LANGUAGE plpgsql;