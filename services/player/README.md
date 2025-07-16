# 🏃‍♂️ Service Player - MMORPG

## 📋 Vue d'ensemble

Le service **Player** est un microservice central du MMORPG responsable de la gestion des joueurs et de leurs personnages. Il gère l'authentification, les profils de joueurs, la création et gestion des personnages, ainsi que le système de statistiques avancé.

## 🏗️ Architecture

```
services/player/
├── cmd/
│   └── main.go                 # Point d'entrée du service
├── internal/
│   ├── config/
│   │   └── config.go          # Configuration du service
│   ├── database/
│   │   ├── connection.go      # Connexion à la base de données
│   │   └── migrations.go      # Gestion des migrations
│   ├── handlers/
│   │   ├── character.go       # Endpoints des personnages
│   │   ├── health.go         # Endpoint de santé
│   │   └── player.go         # Endpoints des joueurs
│   ├── middleware/
│   │   └── middleware.go     # Middlewares (auth, logging)
│   ├── models/
│   │   ├── character.go      # Modèle des personnages
│   │   ├── player.go         # Modèle des joueurs
│   │   └── stats.go          # Modèle des statistiques
│   ├── repository/
│   │   ├── character.go      # Repository des personnages
│   │   └── player.go         # Repository des joueurs
│   └── service/
│       ├── character.go      # Logique métier des personnages
│       └── player.go         # Logique métier des joueurs
├── migrations/
│   ├── 001_create_players_table.up.sql
│   ├── 001_create_players_table.down.sql
│   ├── 002_create_characters_table.up.sql
│   ├── 002_create_characters_table.down.sql
│   ├── 003_create_character_modifiers.up.sql
│   └── 003_create_character_modifiers.down.sql
├── go.mod
├── go.sum
└── README.md
```

## 🚀 Fonctionnalités

### 👤 Gestion des Joueurs
- **Création** et **mise à jour** des profils de joueurs
- **Authentification** et **autorisation**
- **Gestion des sessions** de jeu
- **Statistiques globales** du joueur

### 🧙‍♂️ Gestion des Personnages
- **Création** de personnages avec classes et races
- **Système de statistiques** avancé (Force, Agilité, Intelligence, etc.)
- **Calcul automatique** des statistiques dérivées
- **Gestion des niveaux** et de l'expérience
- **Positionnement** dans le monde

### 🎯 Système de Modificateurs
- **Buffs et debuffs** temporaires
- **Modificateurs d'équipement**
- **Système de stacking** intelligent
- **Nettoyage automatique** des modificateurs expirés

## 🗄️ Base de Données

### Configuration
- **Base de données** : `player_db`
- **Utilisateur** : `auth_user` / `auth_password`
- **Port** : `5432` (PostgreSQL)

### Tables Principales

#### `players`
Table principale des joueurs avec informations de base :
- Identifiants (UUID, username, email)
- Niveau et expérience globaux
- Position dans le monde
- Métadonnées (création, dernière connexion)

#### `characters`
Table des personnages jouables :
- Informations de base (nom, classe, race)
- Statistiques principales (Force, Agilité, Intelligence, etc.)
- Statistiques dérivées (calculées automatiquement)
- Position et apparence
- Temps de jeu

#### `character_modifiers`
Table des modificateurs temporaires :
- Buffs/debuffs avec durée
- Système de stacking
- Source et priorité
- Nettoyage automatique

## 📡 API Endpoints

### 🏥 Santé du Service
```http
GET /health
```
Retourne l'état de santé du service et de ses dépendances.

### 👤 Joueurs
```http
GET    /api/v1/players              # Liste des joueurs
POST   /api/v1/players              # Créer un joueur
GET    /api/v1/players/{id}         # Détails d'un joueur
PUT    /api/v1/players/{id}         # Mettre à jour un joueur
DELETE /api/v1/players/{id}         # Supprimer un joueur
GET    /api/v1/players/{id}/stats   # Statistiques du joueur
```

### 🧙‍♂️ Personnages
```http
GET    /api/v1/characters                    # Liste des personnages
POST   /api/v1/characters                    # Créer un personnage
GET    /api/v1/characters/{id}               # Détails d'un personnage
PUT    /api/v1/characters/{id}               # Mettre à jour un personnage
DELETE /api/v1/characters/{id}               # Supprimer un personnage
GET    /api/v1/characters/{id}/stats         # Statistiques du personnage
PUT    /api/v1/characters/{id}/stats         # Mettre à jour les stats
GET    /api/v1/characters/{id}/modifiers     # Modificateurs actifs
POST   /api/v1/characters/{id}/modifiers     # Ajouter un modificateur
DELETE /api/v1/characters/{id}/modifiers/{modId} # Supprimer un modificateur
```

## ⚙️ Configuration

### Variables d'Environnement
```bash
# Service
PLAYER_PORT=8082
PLAYER_ENV=development

# Base de données
DB_HOST=localhost
DB_PORT=5432
DB_USER=auth_user
DB_PASSWORD=auth_password
DB_NAME=player_db
DB_SSLMODE=disable

# JWT
JWT_SECRET=your_jwt_secret_key

# Logging
LOG_LEVEL=info
```

### Configuration par Défaut
Le service utilise les configurations suivantes par défaut :
- **Port** : `8082`
- **Base de données** : PostgreSQL sur `localhost:5432`
- **Connexions DB** : 25 max, 5 idle
- **Timeout** : 30 secondes

## 🚀 Démarrage

### Prérequis
- **Go 1.21+**
- **PostgreSQL 13+**
- **Base de données** `player_db` configurée

### Installation
```bash
# Naviguer dans le dossier du service
cd services/player

# Installer les dépendances
go mod tidy

# Compiler le service
go build -o player ./cmd/main.go
```

### Lancement
```bash
# Démarrer le service
./player

# Ou directement avec Go
go run ./cmd/main.go
```

### Migration de la Base de Données
```bash
# Les migrations se lancent automatiquement au démarrage
# Ou manuellement avec votre outil de migration préféré
```

## 🧪 Tests

### Tests Unitaires
```bash
# Lancer tous les tests
go test ./...

# Tests avec couverture
go test -cover ./...

# Tests verbeux
go test -v ./...
```

### Tests d'Intégration
```bash
# Assurer que la base de données de test est disponible
go test -tags=integration ./...
```

## 📊 Monitoring et Métriques

Le service expose des métriques Prometheus sur `/metrics` :
- **player_requests_total** : Nombre total de requêtes
- **player_request_duration** : Durée des requêtes
- **player_db_connections** : Connexions à la base de données
- **player_characters_active** : Nombre de personnages actifs

## 🔧 Dépendances

### Services Externes
- **Service Auth** (`auth-new:8081`) : Authentification JWT
- **Service World** (`world:8083`) : Validation des zones
- **PostgreSQL** : Stockage principal
- **Redis** (optionnel) : Cache

### Bibliothèques Go
- **Gin** : Framework web
- **GORM** : ORM pour PostgreSQL
- **JWT-Go** : Gestion des tokens JWT
- **Testify** : Framework de tests
- **Logrus** : Logging structuré

## 🐛 Résolution des Problèmes

### Problèmes Courants

#### Service ne démarre pas
1. Vérifier la connexion à PostgreSQL
2. S'assurer que la base `player_db` existe
3. Vérifier les credentials `auth_user`

#### Erreurs de migration
1. Vérifier les permissions de l'utilisateur
2. S'assurer que les migrations sont dans l'ordre
3. Consulter les logs détaillés

#### Problèmes de performance
1. Vérifier les index de base de données
2. Analyser les requêtes lentes
3. Surveiller les métriques de connexion

### Logs et Debug
```bash
# Activer le mode debug
export LOG_LEVEL=debug

# Voir les logs en temps réel
tail -f /var/log/player-service.log
```

## 🔒 Sécurité

- **Authentification** JWT obligatoire sur tous les endpoints
- **Validation** stricte des entrées utilisateur
- **Rate limiting** configurable
- **Logs d'audit** pour toutes les actions sensibles
- **Chiffrement** des mots de passe

## 📈 Performance

### Optimisations Implémentées
- **Index de base de données** optimisés
- **Calcul en lot** des statistiques
- **Cache Redis** pour les données fréquemment accédées
- **Connection pooling** PostgreSQL
- **Pagination** automatique des grandes listes

### Métriques de Performance
- **< 50ms** : Temps de réponse moyen
- **1000+ RPS** : Capacité de traitement
- **99.9%** : Disponibilité cible

## 🤝 Contribution

1. **Fork** le repository
2. **Créer** une branche feature (`git checkout -b feature/ma-feature`)
3. **Commit** les changements (`git commit -am 'Ajouter ma feature'`)
4. **Push** vers la branche (`git push origin feature/ma-feature`)
5. **Ouvrir** une Pull Request

## 📄 Licence

Ce projet est sous licence MIT. Voir le fichier `LICENSE` pour plus de détails.

---

**Version** : 1.0.0  
**Dernière mise à jour** : 2024  
**Auteur** : Équipe MMORPG Dev
