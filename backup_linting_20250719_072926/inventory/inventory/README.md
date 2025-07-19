# Inventory Service

Le service d'inventaire gère tous les aspects liés aux objets, inventaires, équipements, échanges et fabrication dans le jeu MMORPG.

## Fonctionnalités

### Gestion des inventaires
- Création et suppression d'inventaires de personnages
- Ajout et suppression d'objets
- Déplacement et division des piles d'objets
- Gestion des slots et du poids maximum
- Expansion de l'espace de stockage

### Gestion des équipements
- Équipement et déséquipement d'objets
- Calcul des statistiques totales
- Bonus de set d'équipement
- Validation des prérequis d'équipement

### Système d'échange
- Création et gestion des échanges entre joueurs
- Système d'offres et de confirmation
- Expiration automatique des échanges
- Historique des transactions

### Système de fabrication
- Recettes de fabrication d'objets
- Validation des matériaux requis
- Consommation automatique des ressources
- Gain d'expérience de fabrication

## API Endpoints

### Inventaires
- `GET /:characterId` - Récupérer l'inventaire d'un personnage
- `POST /:characterId/items` - Ajouter un objet à l'inventaire
- `PUT /:characterId/items/:itemId` - Modifier un objet de l'inventaire
- `DELETE /:characterId/items/:itemId` - Supprimer un objet de l'inventaire
- `POST /:characterId/items/move` - Déplacer un objet entre slots
- `POST /:characterId/items/split` - Diviser une pile d'objets
- `GET /:characterId/items` - Lister les objets avec filtres
- `POST /:characterId/items/bulk` - Ajouter plusieurs objets
- `DELETE /:characterId/items/bulk` - Supprimer plusieurs objets

### Équipements
- `GET /:characterId/equipment` - Récupérer l'équipement d'un personnage
- `POST /:characterId/equipment/equip` - Équiper un objet
- `POST /:characterId/equipment/unequip` - Déséquiper un objet
- `GET /:characterId/equipment/stats` - Calculer les statistiques d'équipement

### Échanges
- `POST /:characterId/trades` - Créer un échange
- `GET /:characterId/trades` - Lister les échanges du joueur
- `GET /trades/:tradeId` - Détails d'un échange
- `PUT /trades/:tradeId/offer` - Mettre à jour l'offre
- `PUT /trades/:tradeId/ready` - Marquer prêt pour l'échange
- `POST /trades/:tradeId/complete` - Finaliser l'échange
- `DELETE /trades/:tradeId` - Annuler l'échange

### Santé du service
- `GET /health` - Statut de santé du service
- `GET /ready` - Statut de préparation du service
- `GET /live` - Test de vie du service

## Configuration

Le service utilise les variables d'environnement suivantes :

```env
# Serveur
PORT=8080
HOST=0.0.0.0

# Base de données
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=inventory_db
DATABASE_USER=inventory_user
DATABASE_PASSWORD=inventory_password
DATABASE_SSL_MODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT
JWT_SECRET=your-jwt-secret-key

# Inventaire
INVENTORY_DEFAULT_SLOTS=30
INVENTORY_DEFAULT_MAX_WEIGHT=100.0
INVENTORY_MAX_SLOTS=200
INVENTORY_EXPANSION_COST_PER_SLOT=1000

# Échanges
TRADE_EXPIRATION_HOURS=24
TRADE_CLEANUP_INTERVAL_MINUTES=60
```

## Structure du projet

```
services/inventory/
├── cmd/
│   └── main.go                 # Point d'entrée principal
├── internal/
│   ├── config/
│   │   └── config.go          # Configuration du service
│   ├── database/
│   │   ├── connection.go      # Connexion à la base de données
│   │   └── migrations.go      # Gestion des migrations
│   ├── handlers/
│   │   ├── inventory_handler.go # Gestionnaires d'inventaire
│   │   └── health_handler.go   # Gestionnaires de santé
│   ├── middleware/
│   │   └── middleware.go      # Middlewares communs
│   ├── models/
│   │   ├── item.go            # Modèles d'objets
│   │   ├── inventory.go       # Modèles d'inventaire
│   │   ├── equipment.go       # Modèles d'équipement
│   │   ├── trade.go           # Modèles d'échange
│   │   ├── requests.go        # Modèles de requêtes
│   │   ├── responses.go       # Modèles de réponses
│   │   ├── health.go          # Modèles de santé
│   │   └── errors.go          # Gestion des erreurs
│   ├── repository/
│   │   ├── interfaces.go      # Interfaces des repositories
│   │   ├── item_repository.go # Repository des objets
│   │   ├── inventory_repository.go # Repository des inventaires
│   │   ├── equipment_repository.go # Repository des équipements
│   │   └── trade_repository.go # Repository des échanges
│   └── service/
│       ├── interfaces.go      # Interfaces des services
│       └── inventory_service.go # Logique métier des inventaires
├── migrations/
│   ├── 001_create_inventory_tables.up.sql
│   └── 001_create_inventory_tables.down.sql
├── go.mod
├── go.sum
├── Dockerfile
└── README.md
```

## Démarrage rapide

1. **Installer les dépendances :**
   ```bash
   cd services/inventory
   go mod tidy
   ```

2. **Configurer la base de données PostgreSQL :**
   ```bash
   # Créer la base de données
   createdb inventory_db
   
   # Exécuter les migrations
   go run cmd/main.go migrate
   ```

3. **Démarrer le service :**
   ```bash
   go run cmd/main.go
   ```

Le service sera disponible sur `http://localhost:8080`

## Tests

```bash
# Exécuter tous les tests
go test ./...

# Tests avec couverture
go test -cover ./...

# Tests d'intégration
go test -tags=integration ./...
```

## Déploiement

### Docker

```bash
# Construire l'image
docker build -t mmorpg/inventory-service .

# Exécuter le conteneur
docker run -p 8080:8080 \
  -e DATABASE_HOST=postgres \
  -e REDIS_HOST=redis \
  mmorpg/inventory-service
```

### Kubernetes

Le service peut être déployé sur Kubernetes en utilisant les manifestes dans le dossier `infrastructure/kubernetes/`.

## Monitoring

Le service expose les métriques Prometheus sur `/metrics` et fournit des logs structurés au format JSON.

### Métriques importantes
- `inventory_operations_total` - Nombre total d'opérations d'inventaire
- `inventory_items_total` - Nombre total d'objets en inventaire
- `trade_operations_total` - Nombre total d'opérations d'échange
- `database_queries_duration` - Durée des requêtes de base de données

## Sécurité

- Authentification JWT requise pour toutes les opérations
- Validation des permissions de propriété des inventaires
- Rate limiting sur les endpoints critiques
- Audit des opérations sensibles (échanges, modifications d'inventaire)

## Performance

- Cache Redis pour les requêtes fréquentes
- Pagination automatique des résultats
- Index de base de données optimisés
- Pool de connexions configurables

## Maintenance

### Nettoyage automatique
- Suppression des échanges expirés toutes les heures
- Archivage des logs anciens
- Optimisation des index de base de données

### Sauvegarde
- Sauvegarde quotidienne de la base de données
- Réplication en temps réel vers un serveur secondaire
- Tests de restauration mensuels 