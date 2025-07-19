-- Suppression de la table guild_logs
DROP TABLE IF EXISTS guild_logs CASCADE;

-- Suppression de la fonction de nettoyage
DROP FUNCTION IF EXISTS clean_old_guild_logs(); 