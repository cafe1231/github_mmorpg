-- Suppression de la table player_positions et des éléments associés

-- Supprimer la vue
DROP VIEW IF EXISTS zone_population;

-- Supprimer les triggers
DROP TRIGGER IF EXISTS update_player_positions_last_update ON player_positions;
DROP TRIGGER IF EXISTS update_player_positions_updated_at ON player_positions;

-- Supprimer la fonction trigger spécifique aux positions
DROP FUNCTION IF EXISTS update_player_position_last_update();

-- Supprimer les index
DROP INDEX IF EXISTS idx_player_positions_user_id;
DROP INDEX IF EXISTS idx_player_positions_zone_id;
DROP INDEX IF EXISTS idx_player_positions_is_online;
DROP INDEX IF EXISTS idx_player_positions_last_update;
DROP INDEX IF EXISTS idx_player_positions_is_moving;
DROP INDEX IF EXISTS idx_player_positions_location;
DROP INDEX IF EXISTS idx_player_positions_zone_online;
DROP INDEX IF EXISTS idx_zone_population_stats;

-- Supprimer la table
DROP TABLE IF EXISTS player_positions; 