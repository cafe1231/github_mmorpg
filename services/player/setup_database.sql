-- Script de configuration de la base de données player_db
-- À exécuter en tant que superutilisateur PostgreSQL (postgres)

-- Se connecter à PostgreSQL en tant que postgres et exécuter :
-- psql -U postgres -c "\i setup_database.sql"

-- Créer la base de données si elle n'existe pas
CREATE DATABASE player_db OWNER auth_user;

-- Se connecter à la base player_db
\c player_db;

-- Donner tous les privilèges sur le schéma public à auth_user
GRANT ALL PRIVILEGES ON SCHEMA public TO auth_user;

-- Donner les privilèges de création sur le schéma public
GRANT CREATE ON SCHEMA public TO auth_user;

-- Donner tous les privilèges sur toutes les tables existantes et futures
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO auth_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO auth_user;
GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO auth_user;

-- Configurer les privilèges par défaut pour les futurs objets
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO auth_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO auth_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON FUNCTIONS TO auth_user;

-- Permettre à auth_user de créer des extensions si nécessaire
ALTER USER auth_user CREATEDB;

-- Vérifier que l'utilisateur auth_user existe et a un mot de passe
-- Si l'utilisateur n'existe pas, le créer :
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'auth_user') THEN
        CREATE USER auth_user WITH PASSWORD 'auth_password';
        GRANT ALL PRIVILEGES ON DATABASE player_db TO auth_user;
    END IF;
END
$$;

-- Afficher les privilèges accordés
SELECT 
    schemaname,
    tablename,
    tableowner,
    hasinserts,
    hasupdates,
    hasdeletes,
    hasreferences
FROM pg_tables 
WHERE schemaname = 'public';

-- Afficher les privilèges sur le schéma
SELECT 
    schema_name,
    schema_owner,
    default_character_set_name
FROM information_schema.schemata 
WHERE schema_name = 'public';

PRINT 'Configuration terminée pour player_db avec auth_user';
PRINT 'L''utilisateur auth_user a maintenant tous les privilèges sur player_db'; 