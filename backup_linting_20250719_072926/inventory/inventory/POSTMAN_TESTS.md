# Test du Service d'Inventaire avec Postman

## Configuration de base

- **URL de base** : `http://localhost:8084` (Service Inventory)
- **Content-Type** : `application/json`

## Architecture des ports du projet

- **Gateway** : `http://localhost:8080`
- **Auth** : `http://localhost:8081` 
- **Player** : `http://localhost:8082`
- **World/Combat** : `http://localhost:8083`
- **Inventory** : `http://localhost:8084`

## Tests de Santé du Service

### 1. Health Check
```
GET http://localhost:8084/health
```

### 2. Readiness Check
```
GET http://localhost:8084/health/ready
```

### 3. Liveness Check
```
GET http://localhost:8084/health/live
```

## Tests d'Inventaire

### Variables à utiliser
- `{{baseUrl}}` = `http://localhost:8084`
- `{{characterId}}` = `550e8400-e29b-41d4-a716-446655440000` (exemple d'UUID)
- `{{itemId}}` = `123e4567-e89b-12d3-a456-426614174000` (exemple d'UUID d'item)

### 1. Obtenir l'inventaire d'un personnage
```
GET {{baseUrl}}/api/v1/inventory/{{characterId}}
```

**Réponse attendue** (si l'inventaire n'existe pas) :
```json
{
  "success": false,
  "error": {
    "code": "inventory_not_found",
    "message": "Inventory not found"
  }
}
```

### 2. Ajouter un objet à l'inventaire
```
POST {{baseUrl}}/api/v1/inventory/{{characterId}}/items
Content-Type: application/json

{
  "item_id": "{{itemId}}",
  "quantity": 5
}
```

**Note** : Cette requête échouera probablement car l'item n'existe pas en base. Nous devons d'abord créer des items de test.

### 3. Lister les objets d'un inventaire
```
GET {{baseUrl}}/api/v1/inventory/{{characterId}}/items
```

### 4. Mettre à jour un objet dans l'inventaire
```
PUT {{baseUrl}}/api/v1/inventory/{{characterId}}/items/{{itemId}}
Content-Type: application/json

{
  "quantity": 10
}
```

### 5. Supprimer un objet de l'inventaire
```
DELETE {{baseUrl}}/api/v1/inventory/{{characterId}}/items/{{itemId}}?quantity=2
```

### 6. Déplacer un objet entre slots
```
POST {{baseUrl}}/api/v1/inventory/{{characterId}}/move
Content-Type: application/json

{
  "from_slot": 0,
  "to_slot": 5
}
```

### 7. Diviser une pile d'objets
```
POST {{baseUrl}}/api/v1/inventory/{{characterId}}/split
Content-Type: application/json

{
  "from_slot": 0,
  "to_slot": 1,
  "quantity": 3
}
```

### 8. Ajouter plusieurs objets en lot
```
POST {{baseUrl}}/api/v1/inventory/{{characterId}}/items/bulk/add
Content-Type: application/json

{
  "items": [
    {
      "item_id": "{{itemId}}",
      "quantity": 3
    },
    {
      "item_id": "456e7890-e89b-12d3-a456-426614174001",
      "quantity": 1
    }
  ]
}
```

### 9. Supprimer plusieurs objets en lot
```
POST {{baseUrl}}/api/v1/inventory/{{characterId}}/items/bulk/remove
Content-Type: application/json

{
  "items": [
    {
      "item_id": "{{itemId}}",
      "quantity": 1
    }
  ]
}
```

## Collection Postman complète

Voici une collection Postman au format JSON que vous pouvez importer :

```json
{
  "info": {
    "name": "Inventory Service Tests",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "variable": [
         {
       "key": "baseUrl",
       "value": "http://localhost:8084"
     },
    {
      "key": "characterId",
      "value": "550e8400-e29b-41d4-a716-446655440000"
    },
    {
      "key": "itemId",
      "value": "123e4567-e89b-12d3-a456-426614174000"
    }
  ],
  "item": [
    {
      "name": "Health Checks",
      "item": [
        {
          "name": "Health Check",
          "request": {
            "method": "GET",
            "header": [],
            "url": "{{baseUrl}}/health"
          }
        },
        {
          "name": "Readiness Check",
          "request": {
            "method": "GET",
            "header": [],
            "url": "{{baseUrl}}/health/ready"
          }
        },
        {
          "name": "Liveness Check",
          "request": {
            "method": "GET",
            "header": [],
            "url": "{{baseUrl}}/health/live"
          }
        }
      ]
    },
    {
      "name": "Inventory Operations",
      "item": [
        {
          "name": "Get Inventory",
          "request": {
            "method": "GET",
            "header": [],
            "url": "{{baseUrl}}/api/v1/inventory/{{characterId}}"
          }
        },
        {
          "name": "Add Item",
          "request": {
            "method": "POST",
            "header": [
              {
                "key": "Content-Type",
                "value": "application/json"
              }
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"item_id\": \"{{itemId}}\",\n  \"quantity\": 5\n}"
            },
            "url": "{{baseUrl}}/api/v1/inventory/{{characterId}}/items"
          }
        },
        {
          "name": "List Items",
          "request": {
            "method": "GET",
            "header": [],
            "url": "{{baseUrl}}/api/v1/inventory/{{characterId}}/items"
          }
        },
        {
          "name": "Update Item",
          "request": {
            "method": "PUT",
            "header": [
              {
                "key": "Content-Type",
                "value": "application/json"
              }
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"quantity\": 10\n}"
            },
            "url": "{{baseUrl}}/api/v1/inventory/{{characterId}}/items/{{itemId}}"
          }
        },
        {
          "name": "Remove Item",
          "request": {
            "method": "DELETE",
            "header": [],
            "url": "{{baseUrl}}/api/v1/inventory/{{characterId}}/items/{{itemId}}?quantity=2"
          }
        },
        {
          "name": "Move Item",
          "request": {
            "method": "POST",
            "header": [
              {
                "key": "Content-Type",
                "value": "application/json"
              }
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"from_slot\": 0,\n  \"to_slot\": 5\n}"
            },
            "url": "{{baseUrl}}/api/v1/inventory/{{characterId}}/move"
          }
        },
        {
          "name": "Split Stack",
          "request": {
            "method": "POST",
            "header": [
              {
                "key": "Content-Type",
                "value": "application/json"
              }
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"from_slot\": 0,\n  \"to_slot\": 1,\n  \"quantity\": 3\n}"
            },
            "url": "{{baseUrl}}/api/v1/inventory/{{characterId}}/split"
          }
        },
        {
          "name": "Add Bulk Items",
          "request": {
            "method": "POST",
            "header": [
              {
                "key": "Content-Type",
                "value": "application/json"
              }
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"items\": [\n    {\n      \"item_id\": \"{{itemId}}\",\n      \"quantity\": 3\n    },\n    {\n      \"item_id\": \"456e7890-e89b-12d3-a456-426614174001\",\n      \"quantity\": 1\n    }\n  ]\n}"
            },
            "url": "{{baseUrl}}/api/v1/inventory/{{characterId}}/items/bulk/add"
          }
        },
        {
          "name": "Remove Bulk Items",
          "request": {
            "method": "POST",
            "header": [
              {
                "key": "Content-Type",
                "value": "application/json"
              }
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"items\": [\n    {\n      \"item_id\": \"{{itemId}}\",\n      \"quantity\": 1\n    }\n  ]\n}"
            },
            "url": "{{baseUrl}}/api/v1/inventory/{{characterId}}/items/bulk/remove"
          }
        }
      ]
    }
  ]
}
```

## Instructions de test

### Étape 1 : Vérifier que le service fonctionne
1. Testez d'abord les endpoints de santé
2. Vous devriez recevoir des réponses JSON avec le statut du service

### Étape 2 : Préparer les données de test
Avant de tester les fonctionnalités d'inventaire, vous devrez :

1. **Avoir une base de données configurée** (PostgreSQL)
2. **Créer des items de test** dans la base
3. **Créer un inventaire pour le personnage de test**

### Étape 3 : Tester les fonctionnalités
1. Commencez par `GET /inventory/{characterId}` pour voir si l'inventaire existe
2. Testez `Add Item` avec un item valide
3. Testez les autres opérations une fois que vous avez des items

## Réponses d'erreur courantes

### Inventaire non trouvé
```json
{
  "success": false,
  "error": {
    "code": "inventory_not_found",
    "message": "Inventory not found"
  }
}
```

### Item non trouvé
```json
{
  "success": false,
  "error": {
    "code": "item_not_found",
    "message": "Item not found"
  }
}
```

### ID de personnage invalide
```json
{
  "success": false,
  "error": {
    "code": "invalid_character_id",
    "message": "Invalid character ID format"
  }
}
```

## Notes importantes

- Le service d'inventaire utilise le port **8084** par défaut
- Tous les IDs doivent être des UUIDs valides
- Les quantités doivent être des entiers positifs
- Le service retourne toujours du JSON avec un format standardisé 