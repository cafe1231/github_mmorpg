-- Script pour créer/configurer l'utilisateur auth_user
-- À exécuter en tant qu'utilisateur postgres dans pgAdmin4

-- 1. Vérifier si l'utilisateur existe déjà
-- SELECT * FROM pg_user WHERE usename = 'auth_user';

-- 2. Créer l'utilisateur s'il n'existe pas (décommentez si nécessaire)
-- CREATE USER auth_user WITH PASSWORD 'auth_password';

-- 3. Ou modifier le mot de passe s'il existe déjà
ALTER USER auth_user WITH PASSWORD 'auth_password';

-- 4. Accorder les privilèges nécessaires
ALTER USER auth_user CREATEDB;
GRANT ALL PRIVILEGES ON DATABASE inventory_db TO auth_user; 