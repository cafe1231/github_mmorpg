# Gateway Service

Le Gateway est le point d’entrée unique (API Gateway) de l’architecture microservices du MMORPG. Il gère le routage, la sécurité, l’authentification, le monitoring et l’agrégation des services.

## Fonctionnalités principales
- Reverse proxy vers tous les microservices (auth, player, world, combat, inventory, guild, chat, analytics)
- Sécurité : JWT, rate limiting, CORS, logging, rôles
- Monitoring : endpoints /health, /metrics, Prometheus
- Handlers spécifiques pour l’orchestration et l’admin

## Endpoints principaux

### Endpoints Gateway spécifiques

| Méthode | Endpoint              | Description                                      |
|---------|----------------------|--------------------------------------------------|
| GET     | /gateway/status      | Statut global du Gateway (uptime, version, etc.) |
| GET     | /gateway/services    | Liste et état de tous les microservices          |
| GET     | /gateway/version     | Version, commit, build du Gateway                |
| GET     | /gateway/info        | Infos d’environnement et système                 |
| GET     | /gateway/health/all  | Healthcheck agrégé de tous les services          |
| POST    | /gateway/reload      | Reload dynamique de la config (admin)            |

### Exemples de réponse

#### /gateway/status
```json
{
  "status": "ok",
  "uptime": "1h23m45s",
  "version": "1.0.0",
  "commit": "dev",
  "build": "2024-06-01T12:34:56Z",
  "go_version": "go1.21.0",
  "num_goroutine": 18,
  "reloads": 0,
  "services": 8
}
```

#### /gateway/services
```json
[
  {"name": "auth", "url": "http://localhost:8081", "status": "up"},
  {"name": "player", "url": "http://localhost:8082", "status": "up"},
  ...
]
```

#### /gateway/health/all
```json
{
  "auth": "up",
  "player": "up",
  "world": "down",
  ...
}
```

#### /gateway/version
```json
{
  "version": "1.0.0",
  "commit": "dev",
  "build": "2024-06-01T12:34:56Z"
}
```

#### /gateway/info
```json
{
  "env": ["ANALYTICS_DB_HOST=localhost", ...],
  "pid": 12345,
  "ppid": 1,
  "goarch": "amd64",
  "goos": "windows",
  "num_cpu": 8
}
```

#### /gateway/reload
```json
{
  "message": "Reload effectué",
  "reloads": 1
}
```

## Reverse Proxy et Sécurité
- Toutes les routes /api/v1/* sont routées vers les microservices correspondants
- Authentification JWT sur les routes protégées
- Rate limiting configurable
- Logging structuré (logrus)

## Monitoring
- /health : Statut du Gateway
- /metrics : Métriques Prometheus

## Démarrage

```bash
cd services/gateway
go run cmd/main.go
```

## Configuration
- Voir `internal/config/config.go` pour toutes les options
- Les services sont configurés dans le main (nom -> URL)

---

**Gateway Service** - Version 1.0.0 