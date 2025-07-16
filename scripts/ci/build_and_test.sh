#!/bin/bash
set -e
SERVICES="gateway auth-new player world combat inventory guild chat analytics"
echo "==> Build de tous les services"
for svc in $SERVICES; do
  echo "==> Build $svc"
  (cd services/$svc && go build -o main.exe ./cmd/main.go)
done
echo "==> Tests unitaires de tous les services"
for svc in $SERVICES; do
  echo "==> Test $svc"
  (cd services/$svc && go test ./...)
done
echo "==> Lint global"
golangci-lint run ./...
echo "==> OK" 