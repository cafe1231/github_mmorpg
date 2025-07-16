-- Suppression de la table world_events et des éléments associés

-- Supprimer la vue
DROP VIEW IF EXISTS active_events_with_zones;

-- Supprimer les triggers
DROP TRIGGER IF EXISTS calculate_world_events_end_time ON world_events;
DROP TRIGGER IF EXISTS update_world_events_updated_at ON world_events;

-- Supprimer la fonction trigger spécifique aux événements
DROP FUNCTION IF EXISTS calculate_event_end_time();

-- Supprimer les index
DROP INDEX IF EXISTS idx_world_events_zone_id;
DROP INDEX IF EXISTS idx_world_events_type;
DROP INDEX IF EXISTS idx_world_events_status;
DROP INDEX IF EXISTS idx_world_events_start_time;
DROP INDEX IF EXISTS idx_world_events_end_time;
DROP INDEX IF EXISTS idx_world_events_level_range;
DROP INDEX IF EXISTS idx_world_events_active;
DROP INDEX IF EXISTS idx_world_events_zone_status;

-- Supprimer la table
DROP TABLE IF EXISTS world_events; 