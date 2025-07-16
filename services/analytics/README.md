# Service Analytics

Service de collecte et d'analyse des données pour le MMORPG. Ce service gère les événements, métriques et logs pour fournir des insights sur le comportement des joueurs et les performances du jeu.

## Architecture

Le service Analytics suit une architecture en couches :

```
┌─────────────────┐
│   Handlers      │  ← API REST (Gin)
├─────────────────┤
│   Services      │  ← Logique métier
├─────────────────┤
│  Repositories   │  ← Accès aux données
├─────────────────┤
│   Models        │  ← Structures de données
└─────────────────┘
```

### Composants

- **Handlers** : Gestion des requêtes HTTP et validation des données
- **Services** : Logique métier et orchestration
- **Repositories** : Accès aux données PostgreSQL
- **Models** : Structures de données et validation

## Fonctionnalités

### 1. Gestion des Événements

Le service collecte et stocke tous les événements du jeu :

- **Connexions/Déconnexions** : Suivi de l'activité des joueurs
- **Achats** : Transactions et économie du jeu
- **Combat** : Statistiques de combat et PvP
- **Quêtes** : Progression et accomplissements
- **Guildes** : Activité des guildes et interactions

### 2. Métriques Agrégées

Calcul et stockage de métriques importantes :

- **DAU (Daily Active Users)** : Utilisateurs actifs quotidiens
- **Revenus** : Chiffre d'affaires et transactions
- **PvP Matches** : Statistiques de combat
- **Quests Completed** : Progression des quêtes
- **Guild Activity** : Activité des guildes

### 3. Système de Logs

Centralisation des logs applicatifs :

- **Niveaux** : info, warn, error, debug
- **Contexte** : Métadonnées JSON pour chaque log
- **Recherche** : Filtrage par niveau, période, contexte

## API Endpoints

### Événements

```
POST   /api/v1/events/           # Enregistrer un événement
GET    /api/v1/events/           # Lister les événements (avec filtres)
GET    /api/v1/events/:id        # Récupérer un événement spécifique
```

### Métriques

```
POST   /api/v1/metrics/          # Enregistrer une métrique
POST   /api/v1/metrics/query     # Récupérer des métriques agrégées
```

### Logs

```
POST   /api/v1/logs/             # Enregistrer un log
POST   /api/v1/logs/query        # Récupérer des logs (avec filtres)
```

### Monitoring

```
GET    /health                   # Statut du service
GET    /metrics                  # Métriques Prometheus
```

## Base de Données

### Tables

#### `analytics_events`
```sql
CREATE TABLE analytics_events (
    id UUID PRIMARY KEY,
    type VARCHAR(100) NOT NULL,
    player_id UUID,
    guild_id UUID,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    payload TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);
```

#### `analytics_metrics`
```sql
CREATE TABLE analytics_metrics (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    date DATE NOT NULL,
    tags TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);
```

#### `analytics_logs`
```sql
CREATE TABLE analytics_logs (
    id UUID PRIMARY KEY,
    level VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    context TEXT NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);
```

### Index et Performance

- Index sur les colonnes fréquemment utilisées
- Index composites pour les requêtes complexes
- Index GIN pour les champs JSON
- Contraintes de validation pour l'intégrité des données

## Configuration

### Variables d'Environnement

```bash
# Serveur
ANALYTICS_SERVER_HOST=0.0.0.0
ANALYTICS_SERVER_PORT=8088

# Base de données
ANALYTICS_DB_HOST=localhost
ANALYTICS_DB_PORT=5432
ANALYTICS_DB_USER=auth_user
ANALYTICS_DB_PASSWORD=auth_password
ANALYTICS_DB_NAME=analytics_db
ANALYTICS_DB_SSL_MODE=disable
```

### Configuration par Défaut

```go
type Config struct {
    Server struct {
        Host string `env:"ANALYTICS_SERVER_HOST" envDefault:"0.0.0.0"`
        Port string `env:"ANALYTICS_SERVER_PORT" envDefault:"8088"`
    }
    Database struct {
        Host     string `env:"ANALYTICS_DB_HOST" envDefault:"localhost"`
        Port     string `env:"ANALYTICS_DB_PORT" envDefault:"5432"`
        User     string `env:"ANALYTICS_DB_USER" envDefault:"auth_user"`
        Password string `env:"ANALYTICS_DB_PASSWORD" envDefault:"auth_password"`
        DBName   string `env:"ANALYTICS_DB_NAME" envDefault:"analytics_db"`
        SSLMode  string `env:"ANALYTICS_DB_SSL_MODE" envDefault:"disable"`
    }
}
```

## Installation et Démarrage

### Prérequis

- Go 1.21+
- PostgreSQL 14+
- Accès à la base de données `analytics_db`

### Installation

1. **Cloner le projet** :
```bash
git clone <repository>
cd services/analytics
```

2. **Installer les dépendances** :
```bash
go mod tidy
```

3. **Configurer la base de données** :
```sql
-- Créer l'utilisateur et la base de données
CREATE USER auth_user WITH PASSWORD 'auth_password';
CREATE DATABASE analytics_db OWNER auth_user;
GRANT ALL PRIVILEGES ON DATABASE analytics_db TO auth_user;
```

4. **Exécuter les migrations** :
```bash
# Appliquer les migrations
psql -h localhost -U auth_user -d analytics_db -f migrations/001_create_analytics_events_table.up.sql
psql -h localhost -U auth_user -d analytics_db -f migrations/002_create_analytics_metrics_table.up.sql
psql -h localhost -U auth_user -d analytics_db -f migrations/003_create_analytics_logs_table.up.sql
```

5. **Démarrer le service** :
```bash
go run cmd/main.go
```

### Docker

```bash
# Build de l'image
docker build -t analytics-service .

# Exécution
docker run -p 8088:8088 \
  -e ANALYTICS_DB_HOST=host.docker.internal \
  -e ANALYTICS_DB_USER=auth_user \
  -e ANALYTICS_DB_PASSWORD=auth_password \
  -e ANALYTICS_DB_NAME=analytics_db \
  analytics-service
```

## Utilisation

### Exemples d'API

#### Enregistrer un événement
```bash
curl -X POST http://localhost:8088/api/v1/events/ \
  -H "Content-Type: application/json" \
  -d '{
    "type": "login",
    "player_id": "123e4567-e89b-12d3-a456-426614174000",
    "payload": "{\"ip\":\"192.168.1.1\",\"user_agent\":\"Mozilla/5.0\"}"
  }'
```

#### Récupérer des événements
```bash
curl "http://localhost:8088/api/v1/events/?type=login&from=2024-01-01T00:00:00Z&to=2024-01-31T23:59:59Z&page=1&limit=20"
```

#### Enregistrer une métrique
```bash
curl -X POST http://localhost:8088/api/v1/metrics/ \
  -H "Content-Type: application/json" \
  -d '{
    "name": "dau",
    "value": 1250,
    "tags": {"platform": "web"}
  }'
```

#### Récupérer des métriques
```bash
curl -X POST http://localhost:8088/api/v1/metrics/query \
  -H "Content-Type: application/json" \
  -d '{
    "name": "dau",
    "from": "2024-01-01T00:00:00Z",
    "to": "2024-01-31T23:59:59Z",
    "tags": {"platform": "web"}
  }'
```

#### Enregistrer un log
```bash
curl -X POST http://localhost:8088/api/v1/logs/ \
  -H "Content-Type: application/json" \
  -d '{
    "level": "info",
    "message": "Service démarré avec succès",
    "context": {"service": "analytics", "version": "1.0.0"}
  }'
```

#### Récupérer des logs
```bash
curl -X POST http://localhost:8088/api/v1/logs/query \
  -H "Content-Type: application/json" \
  -d '{
    "level": "error",
    "from": "2024-01-01T00:00:00Z",
    "to": "2024-01-31T23:59:59Z",
    "page": 1,
    "limit": 50
  }'
```

## Monitoring

### Métriques Prometheus

Le service expose des métriques Prometheus sur `/metrics` :

- `analytics_http_requests_total` : Nombre total de requêtes HTTP
- `analytics_http_request_duration_seconds` : Durée des requêtes HTTP

### Health Check

```bash
curl http://localhost:8088/health
```

Réponse :
```json
{
  "status": "ok",
  "service": "analytics",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## Développement

### Structure du Projet

```
analytics/
├── cmd/
│   └── main.go                 # Point d'entrée
├── internal/
│   ├── config/
│   │   └── config.go          # Configuration
│   ├── handlers/
│   │   └── analytics.go       # Handlers HTTP
│   ├── models/
│   │   ├── analytics.go       # Modèles de données
│   │   └── requests.go        # Modèles de requêtes
│   ├── repository/
│   │   ├── interfaces.go      # Interfaces des repositories
│   │   ├── event.go          # Repository des événements
│   │   ├── metric.go         # Repository des métriques
│   │   └── log.go            # Repository des logs
│   └── service/
│       ├── interfaces.go      # Interfaces des services
│       ├── analytics.go       # Service principal
│       ├── metrics.go         # Service des métriques
│       └── logging.go         # Service de logging
├── migrations/                # Migrations SQL
├── go.mod                     # Dépendances Go
└── README.md                  # Documentation
```

### Tests

```bash
# Tests unitaires
go test ./...

# Tests avec couverture
go test -cover ./...

# Tests d'intégration
go test -tags=integration ./...
```

### Linting

```bash
# Golangci-lint
golangci-lint run

# Go vet
go vet ./...
```

## Déploiement

### Production

1. **Variables d'environnement** :
```bash
export ANALYTICS_DB_HOST=analytics-db.prod
export ANALYTICS_DB_PASSWORD=<password-secure>
export ANALYTICS_DB_SSL_MODE=require
```

2. **Base de données** :
```sql
-- Créer la base de données de production
CREATE DATABASE analytics_db_prod;
GRANT ALL PRIVILEGES ON DATABASE analytics_db_prod TO auth_user;
```

3. **Migrations** :
```bash
# Appliquer les migrations de production
psql -h analytics-db.prod -U auth_user -d analytics_db_prod -f migrations/*.up.sql
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: analytics-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: analytics-service
  template:
    metadata:
      labels:
        app: analytics-service
    spec:
      containers:
      - name: analytics
        image: analytics-service:latest
        ports:
        - containerPort: 8088
        env:
        - name: ANALYTICS_DB_HOST
          valueFrom:
            secretKeyRef:
              name: analytics-secrets
              key: db-host
        - name: ANALYTICS_DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: analytics-secrets
              key: db-password
```

## Sécurité

### Bonnes Pratiques

- Validation stricte des données d'entrée
- Échappement des requêtes SQL (utilisation de paramètres)
- Logs sécurisés (pas de données sensibles)
- Rate limiting sur les endpoints
- Authentification pour les endpoints sensibles

### Audit

Le service enregistre automatiquement :
- Tous les événements de sécurité
- Tentatives d'accès non autorisées
- Erreurs de base de données
- Métriques de performance

## Support

Pour toute question ou problème :

1. Consulter les logs du service
2. Vérifier la connectivité à la base de données
3. Tester les endpoints de santé
4. Consulter la documentation des API

---

**Service Analytics** - Version 1.0.0 