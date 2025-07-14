# 🎮 Combat Service - MMORPG Microservice

Service de combat en temps réel pour le MMORPG, gérant les combats PvE/PvP, sorts, effets et classements.

## 🚀 Fonctionnalités

- **Combat PvE/PvP** avec sessions temps réel
- **Système de magie** complet avec sorts et cooldowns
- **Effets de statut** (buffs/debuffs) avec stacks
- **Classements PvP** avec système ELO
- **WebSocket** pour les mises à jour temps réel
- **Anti-cheat** et validation des actions
- **Monitoring** avec Prometheus et logs structurés

## 🏗️ Architecture

`
combat-service/
├── cmd/main.go              # Point d'entrée
├── internal/
│   ├── config/              # Configuration
│   ├── database/            # Connexion DB et migrations
│   ├── models/              # Modèles de données
│   ├── repository/          # Couche d'accès aux données
│   ├── service/             # Logique métier
│   ├── handler/             # Handlers HTTP/WebSocket
│   ├── middleware/          # Middlewares Gin
│   ├── external/            # Clients services externes
│   └── monitoring/          # Métriques et health checks
├── configs/                 # Fichiers de configuration
├── scripts/                 # Scripts de déploiement
└── docker/                  # Configuration Docker
`

## 🛠️ Installation

1. **Cloner et configurer**
`ash
cd services/combat
go mod tidy
`

2. **Configurer la base de données**
`ash
# Créer la base de données
createdb combat_db

# Exécuter les migrations
./scripts/migrate.sh
`

3. **Lancer le service**
`ash
go run cmd/main.go
`

## 🔧 Configuration

Le service utilise Viper pour la configuration avec priorité :
1. Variables d'environnement
2. Fichier config.yaml
3. Valeurs par défaut

### Variables d'environnement principales :
- COMBAT_PORT=8084
- COMBAT_DB_HOST=localhost
- COMBAT_DB_PASSWORD=auth_pass
- COMBAT_JWT_SECRET=your-jwt-secret

## 🌐 API Endpoints

### Combat
- POST /api/v1/combat/start - Créer un combat
- POST /api/v1/combat/join/:sessionId - Rejoindre un combat
- POST /api/v1/combat/action - Effectuer une action
- GET /api/v1/combat/status/:sessionId - Statut du combat

### Sorts
- GET /api/v1/spells/character/:characterId - Sorts du personnage
- POST /api/v1/spells/cast - Lancer un sort
- GET /api/v1/spells/cooldowns/:characterId - Cooldowns actifs

### PvP
- POST /api/v1/combat/pvp/challenge - Défier en PvP
- GET /api/v1/combat/pvp/rankings - Classements PvP

### WebSocket
- GET /ws - Connexion temps réel

## 🏥 Health & Monitoring

- Health Check: GET /health
- Métriques: GET /metrics (port 9094)
- Debug: GET /debug/* (développement uniquement)

## 🐳 Docker

`ash
# Build
docker build -f docker/Dockerfile -t combat-service .

# Run
docker-compose -f docker/docker-compose.yml up
`

## 🧪 Tests

`ash
# Tests unitaires
go test ./...

# Tests d'intégration
go test -tags=integration ./...

# Coverage
go test -cover ./...
`

## 📊 Monitoring

Le service expose des métriques Prometheus :
- combat_sessions_total - Nombre total de sessions
- combat_actions_total - Nombre d'actions par type
- spell_casts_total - Sorts lancés par école
- pvp_matches_total - Matches PvP par type

## 🔗 Intégrations

- **Auth Service** (8081) - Validation JWT
- **Player Service** (8082) - Stats des personnages
- **World Service** (8083) - Validation des zones
- **Gateway** (8080) - Routage des requêtes

## 📝 Logs

Logs structurés JSON avec niveaux :
- INFO - Actions normales
- WARN - Actions suspectes
- ERROR - Erreurs système
- DEBUG - Détails techniques

## 🛡️ Sécurité

- **JWT** pour l'authentification
- **Rate limiting** par utilisateur
- **Validation** stricte des actions
- **Anti-cheat** avec vérification des positions
- **Sanitisation** des entrées utilisateur

## 🚀 Déploiement

1. **Production**
`ash
./scripts/setup.sh production
`

2. **Staging**
`ash
./scripts/setup.sh staging
`

## 📈 Performance

- **Connexions DB** : Pool de 25 connexions max
- **Rate limits** : 180 actions/min par utilisateur
- **WebSocket** : Support de 1000+ connexions simultanées
- **Cache** : Stats calculées avec TTL

## 🤝 Contribution

1. Fork le projet
2. Créer une branche feature
3. Commiter les changements
4. Pousser vers la branche
5. Ouvrir une Pull Request

## 📄 License

MIT License - voir le fichier LICENSE pour plus de détails.
