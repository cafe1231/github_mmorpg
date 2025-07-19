# Configuration de la Base de Données - Service Inventory

## Prérequis
- PostgreSQL installé et démarré
- pgAdmin4 ouvert et connecté
- Accès administrateur à PostgreSQL

## Étapes de configuration

### 1. Configurer l'utilisateur auth_user

**ÉTAPE IMPORTANTE** : D'abord, assurez-vous que l'utilisateur `auth_user` existe et a le bon mot de passe.

Dans pgAdmin4, connectez-vous en tant que `postgres` et exécutez :

```sql
-- Vérifier si l'utilisateur existe
SELECT * FROM pg_user WHERE usename = 'auth_user';

-- Si l'utilisateur n'existe pas, le créer :
-- CREATE USER auth_user WITH PASSWORD 'auth_password';

-- Si l'utilisateur existe déjà, mettre à jour le mot de passe :
ALTER USER auth_user WITH PASSWORD 'auth_password';

-- Accorder les privilèges nécessaires
ALTER USER auth_user CREATEDB;
```

### 2. Créer la base de données

Dans pgAdmin4 :

1. **Clic droit sur "Databases"** dans l'arbre de navigation
2. **Sélectionner "Create" → "Database..."**
3. **Nom de la base** : `inventory_db`
4. **Owner** : `auth_user`
5. **Cliquer sur "Save"**

**OU** exécuter le script SQL (en tant que postgres) :

```sql
CREATE DATABASE inventory_db WITH OWNER auth_user;
GRANT ALL PRIVILEGES ON DATABASE inventory_db TO auth_user;
```

### 3. Vérifier la configuration de connexion

Le service utilise ces paramètres par défaut :

- **Host** : localhost
- **Port** : 5432
- **Database** : inventory_db
- **User** : auth_user
- **Password** : auth_password
- **SSL Mode** : disable

Si vos paramètres sont différents, vous pouvez les modifier avec des variables d'environnement :

```bash
set INVENTORY_DATABASE_HOST=localhost
set INVENTORY_DATABASE_PORT=5432
set INVENTORY_DATABASE_NAME=inventory_db
set INVENTORY_DATABASE_USER=auth_user
set INVENTORY_DATABASE_PASSWORD=votremotdepasse
```

### 4. Démarrer le service

Le service créera automatiquement les tables nécessaires via les migrations Go :

```bash
cd services/inventory
go run cmd/main.go
```

Vous devriez voir des logs indiquant :
- Connexion à la base de données réussie
- Exécution des migrations
- Serveur démarré sur le port 8084

### 5. Vérifier que le service fonctionne

Testez l'endpoint de santé :

```bash
curl http://localhost:8084/health
```

Réponse attendue :
```json
{
  "status": "ok",
  "timestamp": "2024-01-XX...",
  "version": "development"
}
```

### 6. (Optionnel) Ajouter des données de test

Pour tester l'API avec des données, exécutez le script `test_data.sql` dans pgAdmin4 :

1. **Ouvrir la base** `inventory_db`
2. **Clic droit** → "Query Tool"
3. **Copier-coller le contenu** de `test_data.sql`
4. **Exécuter** (F5)

Cela créera :
- 5 items de test (épée, potion, casque, minerai, anneau)
- 1 inventaire de test pour le character ID `550e8400-e29b-41d4-a716-446655440000`
- 2 items dans cet inventaire

### 6. Tester avec Postman

Utilisez les exemples dans `POSTMAN_TESTS.md` avec :

- **Character ID** : `550e8400-e29b-41d4-a716-446655440000`
- **Item IDs disponibles** :
  - `123e4567-e89b-12d3-a456-426614174000` (Épée de fer)
  - `456e7890-e89b-12d3-a456-426614174001` (Potion de santé)
  - `789abcde-e89b-12d3-a456-426614174002` (Casque de cuir)
  - `987fcdeb-e89b-12d3-a456-426614174003` (Minerai de fer)
  - `654321ab-e89b-12d3-a456-426614174004` (Anneau magique)

## Dépannage

### Erreur : "database does not exist"
- Vérifiez que la base `inventory_db` a bien été créée
- Vérifiez la configuration de connexion

### Erreur : "connection refused"
- Vérifiez que PostgreSQL est démarré
- Vérifiez le port (5432 par défaut)

### Erreur : "authentication failed"
- Vérifiez le nom d'utilisateur et mot de passe
- Par défaut : user=auth_user, password=auth_password

### Le service ne démarre pas
- Vérifiez les logs d'erreur dans la console
- Vérifiez que le port 8084 n'est pas déjà utilisé

## Structure des tables créées

Le service créera automatiquement ces tables :

- **items** : Définitions des objets du jeu
- **inventories** : Inventaires des personnages
- **inventory_items** : Objets dans les inventaires
- **equipment** : Équipements portés
- **trades** : Échanges entre joueurs
- **trade_offers** : Offres d'échange
- **effects** : Effets temporaires 