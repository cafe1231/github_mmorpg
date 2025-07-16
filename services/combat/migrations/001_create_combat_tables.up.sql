-- Table principale des combats
CREATE TABLE combats (
    id UUID PRIMARY KEY,
    combat_type VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL,
    zone_id UUID,
    max_participants INT NOT NULL,
    current_turn INT NOT NULL,
    turn_time_limit INT NOT NULL,
    max_duration INT NOT NULL,
    settings JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Table des participants à un combat
CREATE TABLE combat_participants (
    id UUID PRIMARY KEY,
    combat_id UUID NOT NULL REFERENCES combats(id) ON DELETE CASCADE,
    character_id UUID NOT NULL,
    user_id UUID,
    team INT NOT NULL,
    position INT NOT NULL,
    health INT NOT NULL,
    max_health INT NOT NULL,
    mana INT NOT NULL,
    max_mana INT NOT NULL,
    physical_damage INT NOT NULL,
    magical_damage INT NOT NULL,
    physical_defense INT NOT NULL,
    magical_defense INT NOT NULL,
    critical_chance FLOAT NOT NULL,
    attack_speed FLOAT NOT NULL,
    is_alive BOOLEAN NOT NULL,
    is_ready BOOLEAN NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Table des actions de combat
CREATE TABLE combat_actions (
    id UUID PRIMARY KEY,
    combat_id UUID NOT NULL REFERENCES combats(id) ON DELETE CASCADE,
    actor_id UUID NOT NULL,
    action_type VARCHAR(64) NOT NULL,
    payload JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Table des effets appliqués lors d'un combat
CREATE TABLE combat_effects (
    id UUID PRIMARY KEY,
    combat_id UUID NOT NULL REFERENCES combats(id) ON DELETE CASCADE,
    participant_id UUID NOT NULL REFERENCES combat_participants(id) ON DELETE CASCADE,
    effect_type VARCHAR(64) NOT NULL,
    value INT,
    duration INT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Table des logs de combat
CREATE TABLE combat_logs (
    id UUID PRIMARY KEY,
    combat_id UUID NOT NULL REFERENCES combats(id) ON DELETE CASCADE,
    log_type VARCHAR(64) NOT NULL,
    message TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
