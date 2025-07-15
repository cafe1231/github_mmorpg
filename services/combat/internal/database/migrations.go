package database

import "context"

// Migration 1: Table des instances de combat
const createCombatInstancesTable = `
CREATE TABLE IF NOT EXISTS combat_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    combat_type VARCHAR(20) NOT NULL CHECK (combat_type IN ('pve', 'pvp', 'dungeon', 'raid')),
    status VARCHAR(20) NOT NULL DEFAULT 'waiting' CHECK (status IN ('waiting', 'active', 'paused', 'finished', 'cancelled')),
    zone_id VARCHAR(255),
    max_participants INTEGER NOT NULL DEFAULT 2,
    current_turn INTEGER NOT NULL DEFAULT 0,
    turn_time_limit INTEGER NOT NULL DEFAULT 30,
    max_duration INTEGER NOT NULL DEFAULT 300,
    
    -- Configuration du combat
    settings JSONB DEFAULT '{}',
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP WITH TIME ZONE,
    ended_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);`

// Migration 2: Table des participants aux combats
const createCombatParticipantsTable = `
CREATE TABLE IF NOT EXISTS combat_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    combat_id UUID NOT NULL REFERENCES combat_instances(id) ON DELETE CASCADE,
    character_id UUID NOT NULL,
    user_id UUID NOT NULL,
    team INTEGER NOT NULL DEFAULT 1,
    position INTEGER NOT NULL DEFAULT 1,
    
    -- Stats de combat au moment du combat
    health INTEGER NOT NULL,
    max_health INTEGER NOT NULL,
    mana INTEGER NOT NULL,
    max_mana INTEGER NOT NULL,
    
    -- Stats calculées
    physical_damage INTEGER NOT NULL DEFAULT 0,
    magical_damage INTEGER NOT NULL DEFAULT 0,
    physical_defense INTEGER NOT NULL DEFAULT 0,
    magical_defense INTEGER NOT NULL DEFAULT 0,
    critical_chance DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    attack_speed DECIMAL(5,2) NOT NULL DEFAULT 1.00,
    
    -- État du participant
    is_alive BOOLEAN NOT NULL DEFAULT true,
    is_ready BOOLEAN NOT NULL DEFAULT false,
    last_action_at TIMESTAMP WITH TIME ZONE,
    
    -- Résultats
    damage_dealt INTEGER NOT NULL DEFAULT 0,
    damage_taken INTEGER NOT NULL DEFAULT 0,
    healing_done INTEGER NOT NULL DEFAULT 0,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(combat_id, character_id)
);`

// Migration 3: Table des actions de combat
const createCombatActionsTable = `
CREATE TABLE IF NOT EXISTS combat_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    combat_id UUID NOT NULL REFERENCES combat_instances(id) ON DELETE CASCADE,
    actor_id UUID NOT NULL REFERENCES combat_participants(id) ON DELETE CASCADE,
    target_id UUID REFERENCES combat_participants(id) ON DELETE SET NULL,
    
    -- Détails de l'action
    action_type VARCHAR(30) NOT NULL CHECK (action_type IN ('attack', 'skill', 'item', 'defend', 'flee', 'wait')),
    skill_id VARCHAR(255),
    item_id VARCHAR(255),
    
    -- Résultats de l'action
    damage_dealt INTEGER DEFAULT 0,
    healing_done INTEGER DEFAULT 0,
    mana_used INTEGER DEFAULT 0,
    is_critical BOOLEAN DEFAULT false,
    is_miss BOOLEAN DEFAULT false,
    is_blocked BOOLEAN DEFAULT false,
    
    -- Métadonnées
    turn_number INTEGER NOT NULL,
    action_order INTEGER NOT NULL,
    processing_time_ms INTEGER,
    
    -- Validation anti-cheat
    client_timestamp TIMESTAMP WITH TIME ZONE,
    server_timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    is_validated BOOLEAN DEFAULT true,
    validation_notes TEXT,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);`

// Migration 4: Table des effets de combat
const createCombatEffectsTable = `
CREATE TABLE IF NOT EXISTS combat_effects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    combat_id UUID NOT NULL REFERENCES combat_instances(id) ON DELETE CASCADE,
    target_id UUID NOT NULL REFERENCES combat_participants(id) ON DELETE CASCADE,
    caster_id UUID REFERENCES combat_participants(id) ON DELETE SET NULL,
    
    -- Détails de l'effet
    effect_type VARCHAR(30) NOT NULL CHECK (effect_type IN ('buff', 'debuff', 'dot', 'hot', 'shield', 'stun', 'silence')),
    effect_name VARCHAR(100) NOT NULL,
    effect_description TEXT,
    
    -- Propriétés de l'effet
    stat_affected VARCHAR(50),
    modifier_value INTEGER NOT NULL DEFAULT 0,
    modifier_type VARCHAR(20) NOT NULL DEFAULT 'flat' CHECK (modifier_type IN ('flat', 'percentage')),
    
    -- Durée et stacks
    duration_turns INTEGER NOT NULL DEFAULT 1,
    remaining_turns INTEGER NOT NULL,
    max_stacks INTEGER NOT NULL DEFAULT 1,
    current_stacks INTEGER NOT NULL DEFAULT 1,
    
    -- État
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_dispellable BOOLEAN NOT NULL DEFAULT true,
    
    applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);`

// Migration 5: Table des défis PvP
const createPvPChallengesTable = `
CREATE TABLE IF NOT EXISTS pvp_challenges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    challenger_id UUID NOT NULL,
    challenged_id UUID NOT NULL,
    combat_id UUID REFERENCES combat_instances(id) ON DELETE SET NULL,
    
    -- Détails du défi
    challenge_type VARCHAR(20) NOT NULL DEFAULT 'duel' CHECK (challenge_type IN ('duel', 'arena', 'tournament')),
    message TEXT,
    stakes JSONB DEFAULT '{}',
    
    -- État du défi
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'declined', 'cancelled', 'expired', 'completed')),
    
    -- Résultat
    winner_id UUID,
    loser_id UUID,
    result_type VARCHAR(20) CHECK (result_type IN ('victory', 'defeat', 'draw', 'forfeit', 'timeout')),
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    responded_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE DEFAULT (CURRENT_TIMESTAMP + INTERVAL '5 minutes'),
    completed_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT different_players CHECK (challenger_id != challenged_id)
);`

// Migration 6: Table des logs de combat
const createCombatLogsTable = `
CREATE TABLE IF NOT EXISTS combat_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    combat_id UUID NOT NULL REFERENCES combat_instances(id) ON DELETE CASCADE,
    
    -- Détails du log
    log_type VARCHAR(30) NOT NULL CHECK (log_type IN ('action', 'effect', 'death', 'resurrection', 'system', 'chat')),
    actor_name VARCHAR(100),
    target_name VARCHAR(100),
    message TEXT NOT NULL,
    
    -- Données structurées
    data JSONB DEFAULT '{}',
    
    -- Timestamps
    turn_number INTEGER,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);`

// Migration 7: Table des statistiques de combat
const createCombatStatsTable = `
CREATE TABLE IF NOT EXISTS combat_statistics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID NOT NULL,
    user_id UUID NOT NULL,
    
    -- Statistiques PvE
    pve_battles_won INTEGER NOT NULL DEFAULT 0,
    pve_battles_lost INTEGER NOT NULL DEFAULT 0,
    monsters_killed INTEGER NOT NULL DEFAULT 0,
    bosses_killed INTEGER NOT NULL DEFAULT 0,
    
    -- Statistiques PvP
    pvp_battles_won INTEGER NOT NULL DEFAULT 0,
    pvp_battles_lost INTEGER NOT NULL DEFAULT 0,
    pvp_draws INTEGER NOT NULL DEFAULT 0,
    pvp_rating INTEGER NOT NULL DEFAULT 1000,
    
    -- Statistiques générales
    total_damage_dealt BIGINT NOT NULL DEFAULT 0,
    total_damage_taken BIGINT NOT NULL DEFAULT 0,
    total_healing_done BIGINT NOT NULL DEFAULT 0,
    total_deaths INTEGER NOT NULL DEFAULT 0,
    
    -- Records
    highest_damage_dealt INTEGER NOT NULL DEFAULT 0,
    longest_combat_duration INTEGER NOT NULL DEFAULT 0,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(character_id)
);`

// Migration 8: Index pour les performances
const createIndexes = `
-- Index pour combat_instances
CREATE INDEX IF NOT EXISTS idx_combat_instances_status ON combat_instances(status);
CREATE INDEX IF NOT EXISTS idx_combat_instances_type ON combat_instances(combat_type);
CREATE INDEX IF NOT EXISTS idx_combat_instances_created_at ON combat_instances(created_at);
CREATE INDEX IF NOT EXISTS idx_combat_instances_zone_id ON combat_instances(zone_id);

-- Index pour combat_participants
CREATE INDEX IF NOT EXISTS idx_combat_participants_combat_id ON combat_participants(combat_id);
CREATE INDEX IF NOT EXISTS idx_combat_participants_character_id ON combat_participants(character_id);
CREATE INDEX IF NOT EXISTS idx_combat_participants_user_id ON combat_participants(user_id);
CREATE INDEX IF NOT EXISTS idx_combat_participants_team ON combat_participants(combat_id, team);

-- Index pour combat_actions
CREATE INDEX IF NOT EXISTS idx_combat_actions_combat_id ON combat_actions(combat_id);
CREATE INDEX IF NOT EXISTS idx_combat_actions_actor_id ON combat_actions(actor_id);
CREATE INDEX IF NOT EXISTS idx_combat_actions_turn_number ON combat_actions(combat_id, turn_number);
CREATE INDEX IF NOT EXISTS idx_combat_actions_timestamp ON combat_actions(server_timestamp);

-- Index pour combat_effects
CREATE INDEX IF NOT EXISTS idx_combat_effects_combat_id ON combat_effects(combat_id);
CREATE INDEX IF NOT EXISTS idx_combat_effects_target_id ON combat_effects(target_id);
CREATE INDEX IF NOT EXISTS idx_combat_effects_active ON combat_effects(target_id, is_active);

-- Index pour pvp_challenges
CREATE INDEX IF NOT EXISTS idx_pvp_challenges_challenger ON pvp_challenges(challenger_id);
CREATE INDEX IF NOT EXISTS idx_pvp_challenges_challenged ON pvp_challenges(challenged_id);
CREATE INDEX IF NOT EXISTS idx_pvp_challenges_status ON pvp_challenges(status);
CREATE INDEX IF NOT EXISTS idx_pvp_challenges_expires_at ON pvp_challenges(expires_at);

-- Index pour combat_logs
CREATE INDEX IF NOT EXISTS idx_combat_logs_combat_id ON combat_logs(combat_id);
CREATE INDEX IF NOT EXISTS idx_combat_logs_timestamp ON combat_logs(combat_id, timestamp);

-- Index pour combat_statistics
CREATE INDEX IF NOT EXISTS idx_combat_statistics_character_id ON combat_statistics(character_id);
CREATE INDEX IF NOT EXISTS idx_combat_statistics_user_id ON combat_statistics(user_id);
CREATE INDEX IF NOT EXISTS idx_combat_statistics_pvp_rating ON combat_statistics(pvp_rating DESC);`