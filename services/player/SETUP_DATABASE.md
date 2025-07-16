# 🔧 Guide de Configuration - Base de Données Player

## 🎯 Problème Résolu

Le service Player avait une erreur de permissions :
```
pq: droit refusé pour le schéma public
```

Cela signifie que l'utilisateur `auth_user` n'avait pas les droits pour créer des tables dans PostgreSQL.

## 🚀 Solution Rapide (PowerShell)

### Étape 1 : Exécuter le script automatique
```powershell
# Dans le dossier services/player
.\setup_database.ps1
```

Ce script va :
- ✅ Créer l'utilisateur `auth_user` avec mot de passe `auth_password`
- ✅ Créer la base de données `player_db`
- ✅ Configurer tous les privilèges nécessaires
- ✅ Tester la connexion

### Étape 2 : Démarrer le service
```bash
go run ./cmd/main.go
```

## 🛠️ Solution Manuelle (si le script ne fonctionne pas)

### 1. Se connecter à PostgreSQL en tant que superutilisateur
```bash
psql -U postgres
```

### 2. Créer l'utilisateur et la base de données
```sql
-- Créer l'utilisateur s'il n'existe pas
CREATE USER auth_user WITH PASSWORD 'auth_password';
GRANT CREATEDB TO auth_user;

-- Créer la base de données
CREATE DATABASE player_db OWNER auth_user;

-- Se connecter à la nouvelle base
\c player_db;
```

### 3. Configurer les permissions
```sql
-- Donner tous les privilèges sur le schéma public
GRANT ALL PRIVILEGES ON SCHEMA public TO auth_user;
GRANT CREATE ON SCHEMA public TO auth_user;

-- Privilèges sur les objets existants et futurs
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO auth_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO auth_user;
GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO auth_user;

-- Privilèges par défaut pour les futurs objets
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO auth_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO auth_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON FUNCTIONS TO auth_user;
```

### 4. Tester la connexion
```bash
psql -U auth_user -d player_db -c "SELECT version();"
```

## 🔍 Vérification

Après la configuration, le service Player devrait :
1. ✅ Se connecter à `player_db`
2. ✅ Exécuter les migrations automatiquement
3. ✅ Créer les tables : `players`, `characters`, `character_modifiers`
4. ✅ Démarrer sur le port `8082`

## 📋 Logs de Succès Attendus

```
{"level":"info","msg":"Logger initialized","service":"player"}
{"level":"info","msg":"Connected to PostgreSQL database","database":"player_db"}
{"level":"info","msg":"Running player database migrations..."}
{"level":"info","msg":"Migration 1 executed successfully"}
{"level":"info","msg":"Migration 2 executed successfully"}
{"level":"info","msg":"Migration 3 executed successfully"}
{"level":"info","msg":"Player service started on :8082"}
```

## 🐛 Résolution des Problèmes

### Erreur : "role auth_user does not exist"
```bash
# Se connecter en tant que postgres et créer l'utilisateur
psql -U postgres -c "CREATE USER auth_user WITH PASSWORD 'auth_password';"
```

### Erreur : "database player_db does not exist"
```bash
# Créer la base de données
psql -U postgres -c "CREATE DATABASE player_db OWNER auth_user;"
```

### Erreur : PostgreSQL non trouvé
1. Installer PostgreSQL 13+
2. Ajouter `psql` au PATH système
3. Redémarrer PowerShell

## 🎯 Notes Importantes

- **Utilisateur** : `auth_user` (convention du projet)
- **Mot de passe** : `auth_password` 
- **Base de données** : `player_db` (avec suffix `_db`)
- **Port** : `8082` (service Player)

Cette configuration suit les conventions établies dans le projet MMORPG. 