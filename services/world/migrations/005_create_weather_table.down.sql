-- Suppression de la table weather et des éléments associés

-- Supprimer la vue
DROP VIEW IF EXISTS active_weather_with_zones;

-- Supprimer les triggers
DROP TRIGGER IF EXISTS check_weather_expiry_trigger ON weather;
DROP TRIGGER IF EXISTS update_weather_updated_at ON weather;

-- Supprimer les fonctions spécifiques à la météo
DROP FUNCTION IF EXISTS check_weather_expiry();
DROP FUNCTION IF EXISTS generate_random_weather(VARCHAR(255));

-- Supprimer les index
DROP INDEX IF EXISTS idx_weather_type;
DROP INDEX IF EXISTS idx_weather_is_active;
DROP INDEX IF EXISTS idx_weather_start_time;
DROP INDEX IF EXISTS idx_weather_end_time;
DROP INDEX IF EXISTS idx_weather_temperature;
DROP INDEX IF EXISTS idx_weather_active_time;

-- Supprimer la table
DROP TABLE IF EXISTS weather; 