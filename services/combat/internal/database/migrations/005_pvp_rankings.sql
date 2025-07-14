-- internal/database/migrations/005_pvp_rankings.sql
-- Migration pour le système PvP et classements

-- Types ENUM pour le PvP
CREATE TYPE pvp_match_type AS ENUM ('duel', 'team_vs_team', 'battle_royale', 'arena_1v1', 'arena_2v2', 'arena_3v3');
CREATE TYPE pvp_rating_type AS ENUM ('casual', 'ranked', 'tournament');
CREATE TYPE pvp_match_status AS ENUM ('waiting', 'active', 'completed', 'cancelled');
CREATE TYPE challenge_status AS ENUM ('pending', 'accepted', 'declined', 'expired', 'cancelled');
CREATE TYPE pvp_tier AS ENUM ('bronze', 'silver', 'gold', 'platinum', 'diamond', 'master', 'grandmaster');

-- Table des matches PvP
CREATE TABLE pvp_matches (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES combat_sessions(id) ON DELETE CASCADE,
    type pvp_match_type NOT NULL,
    
    -- Configuration
    match_rules JSONB NOT NULL DEFAULT '{}', -- PvPRules
    rating_type pvp_rating_type NOT NULL DEFAULT 'casual',
    
    -- Résultats
    winner_team INTEGER, -- équipe gagnante
    results JSONB DEFAULT '[]', -- PvPResult[]
    
    -- Timing
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMP,
    duration INTERVAL,
    
    -- État
    status pvp_match_status NOT NULL DEFAULT 'waiting',
    
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_pvp_matches_session ON pvp_matches(session_id);
CREATE INDEX idx_pvp_matches_type ON pvp_matches(type);
CREATE INDEX idx_pvp_matches_rating_type ON pvp_matches(rating_type);
CREATE INDEX idx_pvp_matches_status ON pvp_matches(status);
CREATE INDEX idx_pvp_matches_started_at ON pvp_matches(started_at);

-- Table des participants PvP
CREATE TABLE pvp_participants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    match_id UUID NOT NULL REFERENCES pvp_matches(id) ON DELETE CASCADE,
    character_id UUID NOT NULL,
    player_id UUID NOT NULL,
    
    -- Équipe et position
    team INTEGER NOT NULL,
    position INTEGER NOT NULL DEFAULT 1, -- position dans l'équipe
    
    -- Rating avant le match
    rating_before INTEGER NOT NULL DEFAULT 1000,
    rating_after INTEGER,
    rating_change INTEGER DEFAULT 0,
    
    -- Résultat
    is_winner BOOLEAN DEFAULT FALSE,
    placement INTEGER, -- classement final
    
    -- Statistiques du match
    kills INTEGER DEFAULT 0,
    deaths INTEGER DEFAULT 0,
    assists INTEGER DEFAULT 0,
    damage_dealt BIGINT DEFAULT 0,
    damage_taken BIGINT DEFAULT 0,
    healing_done BIGINT DEFAULT 0,
    
    joined_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_pvp_participants_match ON pvp_participants(match_id);
CREATE INDEX idx_pvp_participants_character ON pvp_participants(character_id);
CREATE INDEX idx_pvp_participants_player ON pvp_participants(player_id);

-- Table des classements PvP
CREATE TABLE pvp_rankings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    character_id UUID NOT NULL,
    player_id UUID NOT NULL,
    season VARCHAR(20) NOT NULL DEFAULT 'season_1',
    
    -- Rating et classement
    rating INTEGER NOT NULL DEFAULT 1000,
    rank INTEGER,
    tier pvp_tier NOT NULL DEFAULT 'bronze',
    division INTEGER NOT NULL DEFAULT 5 CHECK (division BETWEEN 1 AND 5),
    
    -- Statistiques
    wins INTEGER NOT NULL DEFAULT 0,
    losses INTEGER NOT NULL DEFAULT 0,
    win_rate DECIMAL(5,4) NOT NULL DEFAULT 0.0,
    streak INTEGER NOT NULL DEFAULT 0, -- série de victoires/défaites (positif = victoires)
    best_streak INTEGER NOT NULL DEFAULT 0,
    
    -- Historique
    peak_rating INTEGER NOT NULL DEFAULT 1000,
    peak_rank INTEGER,
    
    -- Activité
    last_match_at TIMESTAMP,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    UNIQUE(character_id, season)
);

CREATE INDEX idx_pvp_rankings_season_rating ON pvp_rankings(season, rating DESC);
CREATE INDEX idx_pvp_rankings_character ON pvp_rankings(character_id);
CREATE INDEX idx_pvp_rankings_player ON pvp_rankings(player_id);
CREATE INDEX idx_pvp_rankings_tier_division ON pvp_rankings(tier, division);
CREATE INDEX idx_pvp_rankings_active ON pvp_rankings(is_active, season);

-- Table des défis PvP
CREATE TABLE pvp_challenges (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    challenger_id UUID NOT NULL, -- character_id qui lance le défi
    challenged_id UUID NOT NULL, -- character_id qui reçoit le défi
    
    -- Configuration du défi
    type pvp_match_type NOT NULL DEFAULT 'duel',
    rules JSONB NOT NULL DEFAULT '{}', -- PvPRules
    message TEXT,
    
    -- État
    status challenge_status NOT NULL DEFAULT 'pending',
    
    -- Réponse
    response_message TEXT,
    responded_at TIMESTAMP,
    
    -- Match résultant
    match_id UUID REFERENCES pvp_matches(id) ON DELETE SET NULL,
    
    -- Timing
    expires_at TIMESTAMP NOT NULL DEFAULT (NOW() + INTERVAL '10 minutes'),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_pvp_challenges_challenger ON pvp_challenges(challenger_id);
CREATE INDEX idx_pvp_challenges_challenged ON pvp_challenges(challenged_id);
CREATE INDEX idx_pvp_challenges_status ON pvp_challenges(status);
CREATE INDEX idx_pvp_challenges_expires_at ON pvp_challenges(expires_at);

-- Triggers pour updated_at
CREATE TRIGGER update_pvp_matches_updated_at 
    BEFORE UPDATE ON pvp_matches 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_pvp_rankings_updated_at 
    BEFORE UPDATE ON pvp_rankings 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_pvp_challenges_updated_at 
    BEFORE UPDATE ON pvp_challenges 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Fonction pour calculer le nouveau rating ELO
CREATE OR REPLACE FUNCTION calculate_elo_rating(
    current_rating INTEGER,
    opponent_rating INTEGER,
    is_winner BOOLEAN,
    k_factor INTEGER DEFAULT 32
)
RETURNS INTEGER AS $$
DECLARE
    expected_score DECIMAL;
    actual_score DECIMAL;
    new_rating INTEGER;
BEGIN
    -- Calculer le score attendu
    expected_score := 1.0 / (1.0 + POWER(10.0, (opponent_rating - current_rating) / 400.0));
    
    -- Score réel (1 pour victoire, 0 pour défaite)
    actual_score := CASE WHEN is_winner THEN 1.0 ELSE 0.0 END;
    
    -- Nouveau rating
    new_rating := ROUND(current_rating + k_factor * (actual_score - expected_score));
    
    -- Éviter les ratings négatifs
    IF new_rating < 0 THEN
        new_rating := 0;
    END IF;
    
    RETURN new_rating;
END;
$$ LANGUAGE plpgsql;

-- Fonction pour déterminer le tier basé sur le rating
CREATE OR REPLACE FUNCTION get_tier_from_rating(rating INTEGER)
RETURNS pvp_tier AS $$
BEGIN
    CASE 
        WHEN rating >= 2400 THEN RETURN 'grandmaster';
        WHEN rating >= 2000 THEN RETURN 'master';
        WHEN rating >= 1700 THEN RETURN 'diamond';
        WHEN rating >= 1400 THEN RETURN 'platinum';
        WHEN rating >= 1100 THEN RETURN 'gold';
        WHEN rating >= 800 THEN RETURN 'silver';
        ELSE RETURN 'bronze';
    END CASE;
END;
$$ LANGUAGE plpgsql;

-- Fonction pour obtenir la division basée sur le rating dans un tier
CREATE OR REPLACE FUNCTION get_division_from_rating(rating INTEGER, tier pvp_tier)
RETURNS INTEGER AS $$
DECLARE
    tier_min INTEGER;
    tier_max INTEGER;
    tier_range INTEGER;
    position_in_tier INTEGER;
BEGIN
    -- Définir les bornes de chaque tier
    CASE tier
        WHEN 'bronze' THEN tier_min := 0; tier_max := 799;
        WHEN 'silver' THEN tier_min := 800; tier_max := 1099;
        WHEN 'gold' THEN tier_min := 1100; tier_max := 1399;
        WHEN 'platinum' THEN tier_min := 1400; tier_max := 1699;
        WHEN 'diamond' THEN tier_min := 1700; tier_max := 1999;
        WHEN 'master' THEN tier_min := 2000; tier_max := 2399;
        WHEN 'grandmaster' THEN tier_min := 2400; tier_max := 9999;
    END CASE;
    
    tier_range := tier_max - tier_min;
    position_in_tier := rating - tier_min;
    
    -- Diviser en 5 divisions (5 = la plus basse, 1 = la plus haute)
    RETURN 5 - LEAST(4, FLOOR(position_in_tier * 5.0 / tier_range));
END;
$$ LANGUAGE plpgsql;

-- Fonction pour mettre à jour le classement après un match
CREATE OR REPLACE FUNCTION update_pvp_ranking_after_match(
    p_character_id UUID,
    p_player_id UUID,
    p_old_rating INTEGER,
    p_new_rating INTEGER,
    p_is_winner BOOLEAN,
    p_season VARCHAR DEFAULT 'season_1'
)
RETURNS VOID AS $$
DECLARE
    current_ranking RECORD;
    new_tier pvp_tier;
    new_division INTEGER;
    new_streak INTEGER;
BEGIN
    -- Récupérer le classement actuel
    SELECT * INTO current_ranking 
    FROM pvp_rankings 
    WHERE character_id = p_character_id AND season = p_season;
    
    -- Déterminer le nouveau tier et division
    new_tier := get_tier_from_rating(p_new_rating);
    new_division := get_division_from_rating(p_new_rating, new_tier);
    
    -- Calculer la nouvelle série
    IF p_is_winner THEN
        new_streak := CASE 
            WHEN current_ranking.streak >= 0 THEN current_ranking.streak + 1
            ELSE 1 
        END;
    ELSE
        new_streak := CASE 
            WHEN current_ranking.streak <= 0 THEN current_ranking.streak - 1
            ELSE -1 
        END;
    END IF;
    
    -- Mettre à jour ou créer le classement
    INSERT INTO pvp_rankings (
        character_id, player_id, season, rating, tier, division,
        wins, losses, win_rate, streak, best_streak,
        peak_rating, peak_rank, last_match_at
    ) VALUES (
        p_character_id, p_player_id, p_season, p_new_rating, new_tier, new_division,
        CASE WHEN p_is_winner THEN 1 ELSE 0 END,
        CASE WHEN p_is_winner THEN 0 ELSE 1 END,
        CASE WHEN p_is_winner THEN 1.0 ELSE 0.0 END,
        new_streak,
        ABS(new_streak),
        p_new_rating,
        NULL, -- sera calculé par la fonction de classement
        NOW()
    )
    ON CONFLICT (character_id, season) DO UPDATE SET
        rating = p_new_rating,
        tier = new_tier,
        division = new_division,
        wins = CASE WHEN p_is_winner THEN pvp_rankings.wins + 1 ELSE pvp_rankings.wins END,
        losses = CASE WHEN p_is_winner THEN pvp_rankings.losses ELSE pvp_rankings.losses + 1 END,
        win_rate = CASE 
            WHEN (pvp_rankings.wins + pvp_rankings.losses + 1) > 0 
            THEN (CASE WHEN p_is_winner THEN pvp_rankings.wins + 1 ELSE pvp_rankings.wins END)::DECIMAL / (pvp_rankings.wins + pvp_rankings.losses + 1)
            ELSE 0.0 
        END,
        streak = new_streak,
        best_streak = GREATEST(pvp_rankings.best_streak, ABS(new_streak)),
        peak_rating = GREATEST(pvp_rankings.peak_rating, p_new_rating),
        last_match_at = NOW(),
        updated_at = NOW();
END;
$$ LANGUAGE plpgsql;

-- Fonction pour mettre à jour tous les classements (rang global)
CREATE OR REPLACE FUNCTION update_global_rankings(p_season VARCHAR DEFAULT 'season_1')
RETURNS INTEGER AS $$
DECLARE
    ranking_record RECORD;
    current_rank INTEGER := 1;
BEGIN
    -- Mettre à jour les rangs pour la saison donnée
    FOR ranking_record IN 
        SELECT id FROM pvp_rankings 
        WHERE season = p_season AND is_active = TRUE
        ORDER BY rating DESC, wins DESC, (wins::DECIMAL / GREATEST(wins + losses, 1)) DESC
    LOOP
        UPDATE pvp_rankings 
        SET rank = current_rank, 
            peak_rank = LEAST(COALESCE(peak_rank, current_rank), current_rank),
            updated_at = NOW()
        WHERE id = ranking_record.id;
        
        current_rank := current_rank + 1;
    END LOOP;
    
    RETURN current_rank - 1; -- nombre total de joueurs classés
END;
$$ LANGUAGE plpgsql;

-- Fonction pour nettoyer les défis expirés
CREATE OR REPLACE FUNCTION cleanup_expired_challenges()
RETURNS INTEGER AS $$
DECLARE
    cleanup_count INTEGER;
BEGIN
    UPDATE pvp_challenges 
    SET status = 'expired', updated_at = NOW()
    WHERE status = 'pending' 
    AND expires_at < NOW();
    
    GET DIAGNOSTICS cleanup_count = ROW_COUNT;
    RETURN cleanup_count;
END;
$$ LANGUAGE plpgsql;

-- Vue pour les statistiques PvP globales
CREATE VIEW pvp_global_stats AS
SELECT 
    season,
    COUNT(*) as total_players,
    COUNT(CASE WHEN tier = 'bronze' THEN 1 END) as bronze_players,
    COUNT(CASE WHEN tier = 'silver' THEN 1 END) as silver_players,
    COUNT(CASE WHEN tier = 'gold' THEN 1 END) as gold_players,
    COUNT(CASE WHEN tier = 'platinum' THEN 1 END) as platinum_players,
    COUNT(CASE WHEN tier = 'diamond' THEN 1 END) as diamond_players,
    COUNT(CASE WHEN tier = 'master' THEN 1 END) as master_players,
    COUNT(CASE WHEN tier = 'grandmaster' THEN 1 END) as grandmaster_players,
    AVG(rating) as average_rating,
    MAX(rating) as highest_rating,
    AVG(win_rate) as average_win_rate
FROM pvp_rankings 
WHERE is_active = TRUE
GROUP BY season;

-- Insertion de données de test pour le développement
INSERT INTO pvp_rankings (character_id, player_id, season, rating, tier, division, wins, losses, win_rate) VALUES
-- Quelques exemples de classements pour tester
('11111111-1111-1111-1111-111111111111', '11111111-1111-1111-1111-111111111111', 'season_1', 1500, 'gold', 3, 25, 15, 0.625),
('22222222-2222-2222-2222-222222222222', '22222222-2222-2222-2222-222222222222', 'season_1', 2100, 'master', 2, 45, 20, 0.692),
('33333333-3333-3333-3333-333333333333', '33333333-3333-3333-3333-333333333333', 'season_1', 900, 'silver', 4, 12, 18, 0.400);

-- Supprimer les données de test (elles sont juste pour valider la structure)
DELETE FROM pvp_rankings WHERE character_id LIKE '%1111%' OR character_id LIKE '%2222%' OR character_id LIKE '%3333%';

-- Mise à jour des statistiques
ANALYZE pvp_matches;
ANALYZE pvp_participants;
ANALYZE pvp_rankings;
ANALYZE pvp_challenges;