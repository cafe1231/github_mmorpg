# Makefile principal pour le projet MMORPG microservices

# Liste des services
SERVICES = gateway auth-new player world combat inventory guild chat analytics

# Build tous les services
build:
	@for svc in $(SERVICES); do \
		cd services/$$svc && echo "==> Build $$svc" && go build -o main.exe ./cmd/main.go || exit 1; \
		cd ../..; \
	done

# Run un service spécifique (ex: make run SERVICE=player)
run:
	cd services/$(SERVICE) && go run ./cmd/main.go

# Tests unitaires sur tous les services
.PHONY: test
test:
	@for svc in $(SERVICES); do \
		cd services/$$svc && echo "==> Test $$svc" && go test ./... || exit 1; \
		cd ../..; \
	done

# Lint sur tout le code Go
.PHONY: lint
lint:
	@golangci-lint run ./...

# Appliquer les migrations SQL (ex: make migrate SERVICE=player)
migrate:
	@echo "==> Migration pour $(SERVICE) (manuel: psql ou outil de migration recommandé)"

# Docker Compose
.PHONY: docker-up docker-down
docker-up:
	docker-compose up --build

docker-down:
	docker-compose down

# Logs de tous les services Docker
.PHONY: logs
logs:
	docker-compose logs -f

# Status des conteneurs Docker
.PHONY: ps
ps:
	docker-compose ps

# Ouvrir un shell psql sur une base (ex: make db-shell SERVICE=player)
db-shell:
	docker exec -it $(SERVICE)_db psql -U auth_user -d $(SERVICE)_db

# Reset d'une base (drop/create) (ex: make reset-db SERVICE=player)
reset-db:
	docker exec -it $(SERVICE)_db psql -U auth_user -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"

# Rapport de couverture de tests Go
.PHONY: coverage
coverage:
	go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out

# Regénérer les fichiers proto (si utilisé)
.PHONY: proto
proto:
	@echo "==> Génération des fichiers proto (adapter la commande si besoin)"
	@protoc --go_out=shared/proto --go-grpc_out=shared/proto shared/proto/*.proto

# Nettoyer les binaires
.PHONY: clean
clean:
	@for svc in $(SERVICES); do \
		rm -f services/$$svc/main.exe; \
	done
	@rm -f coverage.out
