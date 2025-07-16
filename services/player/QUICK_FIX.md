# 🚀 Solution Rapide - Permissions Base de Données

## 🎯 Problème Actuel

Le service Player ne peut pas créer les tables car l'utilisateur `auth_user` n'a pas les permissions.

## ⚡ Solutions Rapides

### Option 1 : Utiliser une Interface Graphique PostgreSQL

Si tu as **pgAdmin** ou **DBeaver** installé :

1. **Se connecter** en tant que `postgres` (superutilisateur)

2. **Exécuter ces commandes SQL** :
```sql
-- Créer l'utilisateur auth_user
CREATE USER auth_user WITH PASSWORD 'auth_password';
GRANT CREATEDB TO auth_user;

-- Créer la base de données
CREATE DATABASE player_db OWNER auth_user;

-- Se connecter à player_db et exécuter :
GRANT ALL PRIVILEGES ON SCHEMA public TO auth_user;
GRANT CREATE ON SCHEMA public TO auth_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO auth_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO auth_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON FUNCTIONS TO auth_user;
```

### Option 2 : Trouver psql manuellement

1. **Chercher PostgreSQL** dans les programmes installés :
```powershell
# Chercher dans les répertoires courants
dir "C:\PostgreSQL" -ErrorAction SilentlyContinue
dir "C:\Program Files\PostgreSQL" -ErrorAction SilentlyContinue  
dir "C:\Program Files (x86)\PostgreSQL" -ErrorAction SilentlyContinue
```

2. **Si trouvé**, ajouter au PATH temporairement :
```powershell
# Exemple si PostgreSQL est dans C:\PostgreSQL\15\bin
$env:PATH += ";C:\PostgreSQL\15\bin"
.\setup_database.ps1
```

### Option 3 : Configuration simplifiée (temporaire)

**Modifier temporairement** la configuration du service pour utiliser l'utilisateur `postgres` :

1. **Éditer** `services/player/internal/config/config.go` :
```go
// Changer temporairement :
User:     "postgres",
Password: "ton_mot_de_passe_postgres",
```

2. **Démarrer le service** - il créera les tables automatiquement

3. **Remettre** `auth_user` une fois les tables créées

### Option 4 : Docker PostgreSQL

Si tu préfères Docker :
```bash
# Démarrer PostgreSQL avec Docker
docker run --name postgres-mmorpg -e POSTGRES_PASSWORD=postgres -p 5432:5432 -d postgres:13

# Se connecter et configurer
docker exec -it postgres-mmorpg psql -U postgres
```

## 🧪 Test Rapide

Après avoir configuré les permissions, teste :

```bash
cd services/player
go run ./cmd/main.go
```

**Logs de succès attendus** :
```
{"level":"info","msg":"Connected to PostgreSQL database","database":"player_db"}
{"level":"info","msg":"Running player database migrations..."}
{"level":"info","msg":"Migration 1 executed successfully"}
{"level":"info","msg":"Player service started on :8082"}
```

## 🎯 Résultat Final

Une fois configuré, le service Player aura :
- ✅ **3 tables créées** : `players`, `characters`, `character_modifiers`
- ✅ **Triggers automatiques** pour le calcul des stats
- ✅ **API REST complète** sur le port 8082
- ✅ **Ready for production** ! 🚀

**Quelle option préfères-tu essayer ?** 