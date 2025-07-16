-- Suppression de la table zones et des éléments associés

-- Supprimer le trigger
DROP TRIGGER IF EXISTS update_zones_updated_at ON zones;

-- Supprimer la fonction trigger
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Supprimer les index
DROP INDEX IF EXISTS idx_zones_type;
DROP INDEX IF EXISTS idx_zones_level;
DROP INDEX IF EXISTS idx_zones_status;
DROP INDEX IF EXISTS idx_zones_is_pvp;
DROP INDEX IF EXISTS idx_zones_is_safe_zone;
DROP INDEX IF EXISTS idx_zones_bounds;

-- Supprimer la table
DROP TABLE IF EXISTS zones; 