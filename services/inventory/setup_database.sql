-- Script de configuration de la base de données pour le service Inventory
-- À exécuter dans PostgreSQL en tant qu'utilisateur postgres

-- 1. Créer la base de données avec auth_user comme owner
CREATE DATABASE inventory_db WITH OWNER auth_user;

-- 2. Se connecter à la base de données (à faire manuellement dans pgAdmin)
-- \c inventory_db;

-- 3. Accorder tous les privilèges à auth_user
GRANT ALL PRIVILEGES ON DATABASE inventory_db TO auth_user;

-- 4. Les tables seront créées automatiquement par les migrations Go au démarrage du service 