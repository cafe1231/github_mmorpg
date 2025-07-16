# Service Guild

Le service Guild gère toutes les fonctionnalités liées aux guildes dans le MMORPG, incluant la création, la gestion des membres, les invitations, les candidatures, la banque de guilde, les guerres et alliances.

## Architecture

### Structure du Service

```
services/guild/
├── cmd/
│   └── main.go                 # Point d'entrée de l'application
├── internal/
│   ├── config/
│   │   └── config.go          # Configuration du service
│   ├── handlers/
│   │   └── guild.go           # Handlers HTTP pour les guildes
│   ├── models/
│   │   ├── guild.go           # Modèles de données
│   │   ├── requests.go        # Modèles de requêtes
│   │   ├── responses.go       # Modèles de réponses
│   │   └── errors.go          # Définitions d'erreurs
│   ├── repository/
│   │   ├── interfaces.go      # Interfaces des repositories
│   │   ├── guild.go           # Repository des guildes
│   │   └── member.go          # Repository des membres
│   └── service/
│       ├── interfaces.go      # Interfaces des services
│       └── guild.go           # Service principal des guildes
├── migrations/                 # Migrations SQL
├── go.mod                     # Dépendances Go
└── README.md                  # Documentation
```

### Technologies Utilisées

- **Go 1.21** : Langage principal
- **Gin** : Framework web pour les API REST
- **PostgreSQL** : Base de données principale
- **Uber FX** : Injection de dépendances
- **Prometheus** : Métriques et monitoring
- **Logrus** : Logging structuré

## Configuration

### Variables d'Environnement

| Variable | Défaut | Description |
|----------|--------|-------------|
| `GUILD_SERVER_PORT` | `8086` | Port du serveur HTTP |
| `GUILD_SERVER_HOST` | `0.0.0.0` | Host du serveur HTTP |
| `GUILD_DB_HOST` | `localhost` | Host de la base de données |
| `GUILD_DB_PORT` | `5432` | Port de la base de données |
| `GUILD_DB_USER` | `auth_user` | Utilisateur de la base de données |
| `GUILD_DB_PASSWORD` | `auth_password` | Mot de passe de la base de données |
| `GUILD_DB_NAME` | `guild_db` | Nom de la base de données |
| `GUILD_DB_SSLMODE` | `disable` | Mode SSL pour PostgreSQL |
| `JWT_SECRET` | `your-secret-key` | Clé secrète pour JWT |

## API Endpoints

### Guildes

#### Créer une Guilde
```http
POST /api/v1/guilds/
Content-Type: application/json

{
  "name": "Les Gardiens de la Nuit",
  "description": "Une guilde dédiée à la protection",
  "tag": "GDN"
}
```

#### Récupérer une Guilde
```http
GET /api/v1/guilds/{id}
```

#### Mettre à Jour une Guilde
```http
PUT /api/v1/guilds/{id}
Content-Type: application/json

{
  "name": "Nouveau Nom",
  "description": "Nouvelle description",
  "tag": "NN"
}
```

#### Supprimer une Guilde
```http
DELETE /api/v1/guilds/{id}
```

#### Lister les Guildes
```http
GET /api/v1/guilds/?page=1&limit=20
```

#### Rechercher des Guildes
```http
GET /api/v1/guilds/search?name=Gardiens&tag=GDN&min_level=5&max_level=10&page=1&limit=20
```

#### Statistiques d'une Guilde
```http
GET /api/v1/guilds/{id}/stats
```

### Santé et Métriques

#### Vérification de Santé
```http
GET /health
```

#### Métriques Prometheus
```http
GET /metrics
```

## Base de Données

### Tables Principales

#### `guilds`
- `id` : UUID (PK)
- `name` : VARCHAR(50) UNIQUE
- `description` : TEXT
- `tag` : VARCHAR(5) UNIQUE
- `level` : INTEGER DEFAULT 1
- `experience` : BIGINT DEFAULT 0
- `max_members` : INTEGER DEFAULT 50
- `created_at` : TIMESTAMP WITH TIME ZONE
- `updated_at` : TIMESTAMP WITH TIME ZONE

#### `guild_members`
- `id` : UUID (PK)
- `guild_id` : UUID (FK vers guilds)
- `player_id` : UUID
- `role` : VARCHAR(20) DEFAULT 'member'
- `joined_at` : TIMESTAMP WITH TIME ZONE
- `last_seen` : TIMESTAMP WITH TIME ZONE
- `contribution` : BIGINT DEFAULT 0

### Index et Contraintes

- Index sur `guilds.name`, `guilds.tag`, `guilds.level`
- Index sur `guild_members.guild_id`, `guild_members.player_id`
- Contrainte unique sur `guild_members.player_id`
- Contraintes de validation sur les rôles et contributions

## Installation et Démarrage

### Prérequis

1. **PostgreSQL** installé et configuré
2. **Go 1.21+** installé
3. Base de données `guild_db` créée avec l'utilisateur `auth_user`

### Configuration de la Base de Données

```sql
-- Créer la base de données
CREATE DATABASE guild_db OWNER auth_user;

-- Accorder les permissions
GRANT ALL PRIVILEGES ON DATABASE guild_db TO auth_user;
GRANT ALL PRIVILEGES ON SCHEMA public TO auth_user;
```

### Installation

```bash
# Cloner le projet
cd services/guild

# Installer les dépendances
go mod tidy

# Compiler le service
go build -o main.exe cmd/main.go

# Exécuter les migrations
# (à implémenter selon votre gestionnaire de migrations)

# Démarrer le service
./main.exe
```

### Variables d'Environnement

Créez un fichier `.env` à la racine du service :

```env
GUILD_SERVER_PORT=8086
GUILD_DB_HOST=localhost
GUILD_DB_PORT=5432
GUILD_DB_USER=auth_user
GUILD_DB_PASSWORD=auth_password
GUILD_DB_NAME=guild_db
JWT_SECRET=your-secret-key-here
```

## Fonctionnalités

### ✅ Implémentées

- **Gestion des Guildes** : Création, lecture, mise à jour, suppression
- **Recherche et Filtrage** : Recherche par nom, tag, niveau
- **Pagination** : Support de la pagination pour toutes les listes
- **Validation** : Validation des données d'entrée
- **Gestion d'Erreurs** : Erreurs structurées et cohérentes
- **Métriques** : Métriques Prometheus intégrées
- **Logging** : Logging structuré avec Logrus
- **CORS** : Support CORS pour les requêtes cross-origin

### 🚧 En Cours de Développement

- **Gestion des Membres** : Invitations, candidatures, rôles
- **Banque de Guilde** : Dépôts, retraits, transactions
- **Guerres de Guilde** : Déclarations, scores, résultats
- **Alliances** : Création, gestion, dissolution
- **Logs d'Activité** : Historique des actions de guilde
- **Permissions** : Système de permissions granulaire

### 📋 À Implémenter

- **Notifications** : Système de notifications en temps réel
- **Événements** : Événements de guilde et récompenses
- **Rangs Personnalisés** : Création de rangs personnalisés
- **Statistiques Avancées** : Statistiques détaillées des membres
- **Intégration Chat** : Intégration avec le service de chat

## Monitoring et Observabilité

### Métriques Prometheus

- `guild_http_requests_total` : Nombre total de requêtes HTTP
- `guild_http_request_duration_seconds` : Durée des requêtes HTTP

### Endpoints de Monitoring

- `/health` : Vérification de santé du service
- `/metrics` : Métriques Prometheus

## Développement

### Structure des Tests

```bash
# Tests unitaires
go test ./...

# Tests avec couverture
go test -cover ./...

# Tests d'intégration
go test -tags=integration ./...
```

### Linting et Formatage

```bash
# Formater le code
go fmt ./...

# Vérifier le code
golangci-lint run
```

### Build pour Production

```bash
# Build optimisé
go build -ldflags="-s -w" -o guild-service cmd/main.go

# Build pour Docker
docker build -t guild-service .
```

## Déploiement

### Docker

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o guild-service cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/guild-service .
CMD ["./guild-service"]
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: guild-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: guild-service
  template:
    metadata:
      labels:
        app: guild-service
    spec:
      containers:
      - name: guild-service
        image: guild-service:latest
        ports:
        - containerPort: 8086
        env:
        - name: GUILD_DB_HOST
          valueFrom:
            configMapKeyRef:
              name: guild-config
              key: db_host
```

## Sécurité

### Authentification

- Validation JWT pour toutes les opérations sensibles
- Extraction automatique de l'ID du joueur depuis le token
- Vérification des permissions pour les opérations administratives

### Validation des Données

- Validation des entrées JSON avec les tags `binding`
- Sanitisation des paramètres de requête
- Contraintes de base de données pour l'intégrité

### CORS

- Configuration CORS pour les requêtes cross-origin
- Headers de sécurité appropriés

## Performance

### Optimisations

- **Index de Base de Données** : Index optimisés pour les requêtes fréquentes
- **Pool de Connexions** : Configuration optimisée du pool PostgreSQL
- **Pagination** : Pagination efficace pour les grandes listes
- **Métriques** : Monitoring des performances en temps réel

### Recommandations

- Utiliser un load balancer pour la haute disponibilité
- Configurer des timeouts appropriés pour les requêtes
- Monitorer les métriques de base de données
- Implémenter un cache Redis pour les données fréquemment accédées

## Support et Maintenance

### Logs

Les logs sont structurés en JSON et incluent :
- Niveau de log (INFO, WARN, ERROR)
- Timestamp UTC
- Service identifier
- Contexte de la requête
- Détails de l'erreur

### Monitoring

- Métriques Prometheus pour le monitoring
- Endpoint de santé pour les health checks
- Logs structurés pour l'analyse

### Maintenance

- Migrations SQL versionnées
- Scripts de rollback pour chaque migration
- Documentation des changements de schéma 