# CI/CD MMORPG Microservices

## Objectifs
- Garantir la qualité du code (lint, tests)
- Automatiser le build et le packaging Docker
- Préparer le déploiement continu (CD)

## Pipeline GitHub Actions

- **Déclencheurs** : push/pull_request sur `master` ou `main`
- **Étapes principales** :
  1. Checkout du code
  2. Setup Go 1.21
  3. Lint global (`golangci-lint`)
  4. Build de tous les services (`make build`)
  5. Tests unitaires (`make test`)
  6. Build Docker Compose (vérification Dockerfile)

Voir `.github/workflows/ci.yml` pour le détail.

## Scripts CI

- `scripts/ci/build_and_test.sh` : build, test, lint tous les services (utilisable en local ou CI)

## Bonnes pratiques
- Garder les Dockerfile et Makefile à jour
- Ajouter des tests d’intégration dans `/tests` ou `/integration`
- Utiliser des secrets GitHub pour les variables sensibles
- Ajouter des badges de statut dans le README
- Étendre le pipeline pour le déploiement (CD) si besoin

## Exemple d’extension (CD)
- Push automatique d’images Docker sur GitHub Container Registry
- Déploiement sur un cluster Kubernetes ou VM

## Gestion des secrets GitHub (production)

- Utiliser l’interface GitHub (Settings > Secrets and variables > Actions) pour ajouter les secrets sensibles :
  - `POSTGRES_PASSWORD`
  - `JWT_SECRET`
  - `DOCKERHUB_USERNAME` / `DOCKERHUB_TOKEN` (si push d’images)
  - Autres clés API, tokens, etc.
- Ne jamais commiter de secrets dans le code ou les fichiers de config.
- Utiliser les secrets dans les workflows via `${{ secrets.NOM_DU_SECRET }}`.

Exemple dans un job GitHub Actions :
```yaml
      - name: Run migration
        run: |
          export POSTGRES_PASSWORD=${{ secrets.POSTGRES_PASSWORD }}
          # ...
```

---

**Pour toute question, voir le README ou contacter les mainteneurs du projet.** 