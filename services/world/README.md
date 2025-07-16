# Service World - MMORPG

Le service World est responsable de la gestion du monde virtuel, incluant les zones, les NPCs, les événements mondiaux, la météo et les positions des joueurs.

## 🏗️ Architecture

Le service suit une architecture clean avec:
- **Handlers**: Gestion des requêtes HTTP
- **Services**: Logique métier
- **Repositories**: Accès aux données
- **Models**: Structures de données
- **Migrations**: Scripts SQL pour la base de données

## 🚀 Fonctionnalités

### 🗺️ Gestion des Zones
- Création et gestion des zones de jeu
- Types de zones: city, dungeon, wilderness, pvp, safe
- Système de coordonnées 3D avec bounding box
- Configuration avancée (PvP, sécurité, limites de joueurs)
- Statistiques en temps réel de population

### 🤖 NPCs (Personnages Non-Joueurs)
- Création et gestion des NPCs
- Types: merchant, guard, quest_giver, monster
- Système de comportement configurable (patrouille, agression, faction)
- Gestion de la santé et du respawn
- Recherche de proximité en 3D
- Interactions avec les joueurs

### 🌍 Événements Mondiaux
- Système d'événements temporels
- Types: boss_spawn, treasure_hunt, pvp_tournament, seasonal
- Gestion automatique du cycle de vie (scheduled → active → completed)
- Système de récompenses configurable
- Filtrage par niveau et zone

### 🌤️ Système Météorologique
- Météo dynamique par zone
- Types: clear, rain, storm, snow, fog
- Paramètres réalistes (température, vent, visibilité)
- Génération automatique selon le type de zone
- Système d'expiration automatique

### 📍 Positions des Joueurs
- Suivi en temps réel des positions 3D
- Gestion du mouvement et de la vélocité
- État de connexion (online/offline)
- Statistiques de population par zone

## 🛠️ Technologies

- **Go 1.21+**
- **PostgreSQL 14+** avec support JSONB
- **Gin** - Framework HTTP
- **Uber FX** - Injection de dépendances
- **Logrus** - Logging structuré
- **JWT** - Authentification
- **Database/SQL** - Accès base de données

## 📁 Structure du Projet

```
services/world/
├── cmd/
│   └── main.go              # Point d'entrée
├── internal/
│   ├── config/              # Configuration
│   ├── database/            # Connexion DB
│   ├── handlers/            # Handlers HTTP
│   │   ├── zone_handler.go
│   │   ├── player_position_handler.go
│   │   └── stubs.go        # NPCs, Events, Weather
│   ├── middleware/          # Middlewares
│   ├── models/             # Modèles de données
│   ├── repository/         # Accès données
│   │   ├── zone.go
│   │   ├── player_position.go
│   │   └── stubs.go        # NPCs, Events, Weather
│   └── service/            # Logique métier
│       ├── zone_service.go
│       ├── player_position_service.go
│       └── stubs.go        # NPCs, Events, Weather
├── migrations/             # Scripts SQL
├── pkg/
│   └── monitoring/         # Métriques
├── go.mod
├── go.sum
├── Dockerfile
└── README.md
```

## 🔧 Installation et Configuration

### Prérequis
- Go 1.21+
- PostgreSQL 14+
- Variables d'environnement configurées

### Variables d'Environnement
```bash
# Serveur
SERVER_PORT=8083
SERVER_HOST=0.0.0.0
SERVER_ENVIRONMENT=development

# Base de données
DB_HOST=localhost
DB_PORT=5432
DB_NAME=world_db
DB_USER=auth_user
DB_PASSWORD=auth_password
DB_SSL_MODE=disable

# JWT
JWT_SECRET=your-jwt-secret-key

# Rate Limiting
RATE_LIMIT_REQUESTS_PER_MINUTE=60

# Monitoring
MONITORING_HEALTH_PATH=/health
MONITORING_METRICS_PATH=/metrics
```

### Installation
```bash
# Cloner le repository
git clone <repository-url>
cd github_mmorpg/services/world

# Installer les dépendances
go mod tidy

# Configurer la base de données
createdb world_db
psql -d world_db -f migrations/001_create_zones_table.up.sql
psql -d world_db -f migrations/002_create_npcs_table.up.sql
psql -d world_db -f migrations/003_create_player_positions_table.up.sql
psql -d world_db -f migrations/004_create_world_events_table.up.sql
psql -d world_db -f migrations/005_create_weather_table.up.sql

# Compiler et lancer
go build -o world-service ./cmd/main.go
./world-service
```

## 📊 Base de Données

### Tables Principales

#### zones
```sql
- id (VARCHAR, PK) - Identifiant unique
- name (VARCHAR) - Nom technique
- display_name (VARCHAR) - Nom affiché
- type (ENUM) - Type de zone
- level (INTEGER) - Niveau recommandé
- bounds (DECIMAL x6) - Limites 3D
- spawn_point (DECIMAL x3) - Point d'apparition
- settings (JSONB) - Configuration
```

#### npcs
```sql
- id (UUID, PK) - Identifiant unique
- zone_id (VARCHAR, FK) - Zone d'appartenance
- name (VARCHAR) - Nom du NPC
- type (ENUM) - Type de NPC
- position (DECIMAL x4) - Position 3D + rotation
- behavior (JSONB) - Comportement
- health/max_health (INTEGER) - Santé
```

#### world_events
```sql
- id (UUID, PK) - Identifiant unique
- zone_id (VARCHAR, FK) - Zone (optionnelle)
- name (VARCHAR) - Nom de l'événement
- type (ENUM) - Type d'événement
- timing (TIMESTAMP x2) - Début/fin
- rewards (JSONB) - Récompenses
```

#### weather
```sql
- zone_id (VARCHAR, PK/FK) - Zone
- type (ENUM) - Type de météo
- parameters (DECIMAL x5) - Paramètres météo
- timing (TIMESTAMP x2) - Période active
```

#### player_positions
```sql
- character_id (UUID, PK) - Identifiant personnage
- user_id (UUID) - Identifiant utilisateur
- zone_id (VARCHAR, FK) - Zone actuelle
- position (DECIMAL x4) - Position 3D + rotation
- velocity (DECIMAL x3) - Vecteur mouvement
```

## 🔗 API Endpoints

### Routes Publiques
- `GET /health` - Santé du service
- `GET /metrics` - Métriques Prometheus

### Routes Authentifiées (JWT requis)

#### Zones
- `GET /api/v1/zones` - Liste des zones
- `GET /api/v1/zones/:id` - Détails d'une zone
- `POST /api/v1/zones/:id/enter` - Entrer dans une zone
- `POST /api/v1/zones/:id/leave` - Quitter une zone
- `GET /api/v1/zones/:id/players` - Joueurs dans la zone
- `GET /api/v1/zones/:id/npcs` - NPCs de la zone

#### NPCs
- `GET /api/v1/npcs` - Liste des NPCs
- `GET /api/v1/npcs/:id` - Détails d'un NPC
- `POST /api/v1/npcs/:id/interact` - Interagir avec un NPC
- `GET /api/v1/npcs/zone/:zoneId` - NPCs par zone

#### Positions
- `GET /api/v1/positions/character/:characterId` - Position du personnage
- `PUT /api/v1/positions/character/:characterId` - Mettre à jour position
- `GET /api/v1/positions/zone/:zoneId` - Positions dans la zone

#### Événements
- `GET /api/v1/events` - Liste des événements
- `GET /api/v1/events/active` - Événements actifs
- `GET /api/v1/events/zone/:zoneId` - Événements par zone
- `POST /api/v1/events/:id/participate` - Participer à un événement

#### Météo
- `GET /api/v1/weather/zone/:zoneId` - Météo actuelle
- `GET /api/v1/weather/forecast/:zoneId` - Prévisions météo

### Routes Administrateur (Role admin requis)

#### Gestion des Zones
- `POST /api/v1/admin/zones` - Créer une zone
- `PUT /api/v1/admin/zones/:id` - Modifier une zone
- `DELETE /api/v1/admin/zones/:id` - Supprimer une zone

#### Gestion des NPCs
- `POST /api/v1/admin/npcs` - Créer un NPC
- `PUT /api/v1/admin/npcs/:id` - Modifier un NPC
- `DELETE /api/v1/admin/npcs/:id` - Supprimer un NPC

#### Gestion des Événements
- `POST /api/v1/admin/events` - Créer un événement
- `PUT /api/v1/admin/events/:id` - Gérer événement (start/end/cancel)
- `DELETE /api/v1/admin/events/:id` - Supprimer un événement

#### Gestion de la Météo
- `POST /api/v1/admin/weather/:zoneId` - Définir la météo

## 📈 Monitoring

### Métriques Disponibles
- Nombre de requêtes par endpoint
- Temps de réponse par endpoint
- Erreurs par type
- Connexions actives à la base de données
- Population par zone en temps réel

### Health Check
```json
{
  "status": "healthy",
  "service": "world-service",
  "version": "1.0.0",
  "timestamp": 1673123456
}
```

## 🧪 Tests

```bash
# Tests unitaires
go test ./internal/...

# Tests d'intégration
go test -tags=integration ./...

# Coverage
go test -cover ./internal/...
```

## 🚢 Déploiement

### Docker
```bash
# Build
docker build -t world-service .

# Run
docker run -p 8083:8083 --env-file .env world-service
```

### Kubernetes
Voir les manifests dans `/infrastructure/kubernetes/`

## 🔧 Développement

### Ajout d'une Nouvelle Fonctionnalité
1. Définir le modèle dans `/internal/models/`
2. Créer la migration SQL dans `/migrations/`
3. Implémenter le repository dans `/internal/repository/`
4. Ajouter la logique métier dans `/internal/service/`
5. Créer les handlers dans `/internal/handlers/`
6. Ajouter les routes dans `/cmd/main.go`
7. Tester et documenter

### Structure des Migrations
Les migrations suivent le pattern: `XXX_description.up.sql` et `XXX_description.down.sql`

### Logging
Utilisation de Logrus avec des champs structurés:
```go
logrus.WithFields(logrus.Fields{
    "zone_id": zoneID,
    "user_id": userID,
}).Info("Player entered zone")
```

## 🤝 Contribution

1. Fork le projet
2. Créer une branche feature (`git checkout -b feature/amazing-feature`)
3. Commit les changements (`git commit -m 'Add amazing feature'`)
4. Push la branche (`git push origin feature/amazing-feature`)
5. Ouvrir une Pull Request

## 📄 Licence

Ce projet est sous licence MIT. Voir le fichier `LICENSE` pour plus de détails.

## 🔗 Services Liés

- **auth-service** (port 8081) - Authentification
- **player-service** (port 8082) - Gestion des joueurs
- **combat-service** (port 8083) - Système de combat
- **inventory-service** (port 8084) - Gestion inventaire
- **gateway** (port 8080) - API Gateway 