-- Création de la table weather
CREATE TABLE IF NOT EXISTS weather (
    zone_id VARCHAR(255) PRIMARY KEY,
    type VARCHAR(20) NOT NULL CHECK (type IN ('clear', 'rain', 'storm', 'snow', 'fog')),
    intensity DECIMAL(3,2) NOT NULL DEFAULT 0.5 CHECK (intensity >= 0.0 AND intensity <= 1.0),
    temperature DECIMAL(6,2) NOT NULL DEFAULT 20.0, -- en Celsius
    wind_speed DECIMAL(6,2) NOT NULL DEFAULT 0.0 CHECK (wind_speed >= 0.0), -- km/h
    wind_direction DECIMAL(5,2) NOT NULL DEFAULT 0.0 CHECK (wind_direction >= 0.0 AND wind_direction < 360.0), -- degrés
    visibility DECIMAL(8,2) NOT NULL DEFAULT 1000.0 CHECK (visibility >= 0.0), -- mètres
    
    -- Timing de la météo
    start_time TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    end_time TIMESTAMP WITH TIME ZONE,
    
    -- État
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Clé étrangère vers zones
    FOREIGN KEY (zone_id) REFERENCES zones(id) ON DELETE CASCADE
);

-- Index pour améliorer les performances
CREATE INDEX idx_weather_type ON weather(type);
CREATE INDEX idx_weather_is_active ON weather(is_active);
CREATE INDEX idx_weather_start_time ON weather(start_time);
CREATE INDEX idx_weather_end_time ON weather(end_time);
CREATE INDEX idx_weather_temperature ON weather(temperature);

-- Index composé pour les requêtes fréquentes
CREATE INDEX idx_weather_active_time ON weather(is_active, start_time, end_time);

-- Contrainte pour vérifier la cohérence des temps
ALTER TABLE weather ADD CONSTRAINT check_weather_time_range 
    CHECK (end_time IS NULL OR end_time > start_time);

-- Trigger pour mettre à jour updated_at
CREATE TRIGGER update_weather_updated_at 
    BEFORE UPDATE ON weather 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Fonction pour désactiver automatiquement la météo expirée
CREATE OR REPLACE FUNCTION check_weather_expiry()
RETURNS TRIGGER AS $$
BEGIN
    -- Si la météo a expiré, la désactiver
    IF NEW.end_time IS NOT NULL AND NEW.end_time <= CURRENT_TIMESTAMP THEN
        NEW.is_active = FALSE;
    END IF;
    
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER check_weather_expiry_trigger 
    BEFORE UPDATE ON weather 
    FOR EACH ROW 
    EXECUTE FUNCTION check_weather_expiry();

-- Vue pour la météo active avec informations de zone
CREATE OR REPLACE VIEW active_weather_with_zones AS
SELECT 
    w.*,
    z.name as zone_name,
    z.display_name as zone_display_name,
    z.type as zone_type,
    CASE 
        WHEN w.end_time IS NOT NULL THEN 
            EXTRACT(EPOCH FROM (w.end_time - CURRENT_TIMESTAMP)) / 60
        ELSE NULL 
    END as minutes_remaining,
    CASE w.type
        WHEN 'clear' THEN 'Temps clair'
        WHEN 'rain' THEN 'Pluie'
        WHEN 'storm' THEN 'Orage'
        WHEN 'snow' THEN 'Neige'
        WHEN 'fog' THEN 'Brouillard'
        ELSE w.type
    END as type_display
FROM weather w
LEFT JOIN zones z ON w.zone_id = z.id
WHERE w.is_active = TRUE
    AND (w.end_time IS NULL OR w.end_time > CURRENT_TIMESTAMP);

-- Fonction pour générer une météo aléatoire pour une zone
CREATE OR REPLACE FUNCTION generate_random_weather(target_zone_id VARCHAR(255))
RETURNS VOID AS $$
DECLARE
    zone_type VARCHAR(50);
    weather_types VARCHAR(20)[];
    selected_type VARCHAR(20);
    random_intensity DECIMAL(3,2);
    random_temp DECIMAL(6,2);
    random_wind DECIMAL(6,2);
    random_direction DECIMAL(5,2);
    random_visibility DECIMAL(8,2);
BEGIN
    -- Récupérer le type de zone
    SELECT type INTO zone_type FROM zones WHERE id = target_zone_id;
    
    -- Définir les types de météo possibles selon la zone
    CASE zone_type
        WHEN 'city', 'safe' THEN
            weather_types := ARRAY['clear', 'rain', 'fog'];
        WHEN 'wilderness' THEN
            weather_types := ARRAY['clear', 'rain', 'storm', 'fog'];
        WHEN 'dungeon' THEN
            weather_types := ARRAY['fog', 'clear'];
        ELSE
            weather_types := ARRAY['clear', 'rain', 'storm', 'snow', 'fog'];
    END CASE;
    
    -- Sélectionner un type aléatoire
    selected_type := weather_types[1 + floor(random() * array_length(weather_types, 1))];
    
    -- Générer des paramètres selon le type
    CASE selected_type
        WHEN 'clear' THEN
            random_intensity := 0.0;
            random_temp := 15.0 + random() * 20.0; -- 15-35°C
            random_wind := random() * 10.0; -- 0-10 km/h
            random_visibility := 10000.0; -- 10km
        WHEN 'rain' THEN
            random_intensity := 0.3 + random() * 0.4; -- 0.3-0.7
            random_temp := 5.0 + random() * 15.0; -- 5-20°C
            random_wind := 5.0 + random() * 15.0; -- 5-20 km/h
            random_visibility := 1000.0 + random() * 4000.0; -- 1-5km
        WHEN 'storm' THEN
            random_intensity := 0.7 + random() * 0.3; -- 0.7-1.0
            random_temp := 10.0 + random() * 10.0; -- 10-20°C
            random_wind := 20.0 + random() * 30.0; -- 20-50 km/h
            random_visibility := 300.0 + random() * 700.0; -- 300m-1km
        WHEN 'snow' THEN
            random_intensity := 0.2 + random() * 0.6; -- 0.2-0.8
            random_temp := -10.0 + random() * 10.0; -- -10-0°C
            random_wind := random() * 20.0; -- 0-20 km/h
            random_visibility := 500.0 + random() * 2000.0; -- 500m-2.5km
        WHEN 'fog' THEN
            random_intensity := 0.4 + random() * 0.4; -- 0.4-0.8
            random_temp := 5.0 + random() * 20.0; -- 5-25°C
            random_wind := random() * 5.0; -- 0-5 km/h
            random_visibility := 50.0 + random() * 450.0; -- 50-500m
    END CASE;
    
    random_direction := random() * 360.0; -- 0-360°
    
    -- Insérer ou mettre à jour la météo
    INSERT INTO weather (zone_id, type, intensity, temperature, wind_speed, wind_direction, visibility, end_time)
    VALUES (
        target_zone_id, 
        selected_type, 
        random_intensity, 
        random_temp, 
        random_wind, 
        random_direction, 
        random_visibility,
        CURRENT_TIMESTAMP + (1 + floor(random() * 5)) * INTERVAL '1 hour' -- 1-6 heures
    )
    ON CONFLICT (zone_id) 
    DO UPDATE SET 
        type = EXCLUDED.type,
        intensity = EXCLUDED.intensity,
        temperature = EXCLUDED.temperature,
        wind_speed = EXCLUDED.wind_speed,
        wind_direction = EXCLUDED.wind_direction,
        visibility = EXCLUDED.visibility,
        start_time = CURRENT_TIMESTAMP,
        end_time = EXCLUDED.end_time,
        is_active = TRUE,
        updated_at = CURRENT_TIMESTAMP;
END;
$$ language 'plpgsql';

-- Générer une météo par défaut pour chaque zone existante
INSERT INTO weather (zone_id, type, intensity, temperature, wind_speed, wind_direction, visibility, end_time)
SELECT 
    id as zone_id,
    'clear' as type,
    0.3 as intensity,
    22.0 as temperature,
    5.0 as wind_speed,
    180.0 as wind_direction,
    1000.0 as visibility,
    CURRENT_TIMESTAMP + INTERVAL '2 hours' as end_time
FROM zones
ON CONFLICT (zone_id) DO NOTHING;

-- Générer une météo aléatoire pour chaque zone
SELECT generate_random_weather(id) FROM zones; 