# Script pour corriger le formatage du service combat
# Basé sur la même approche que pour auth et analytics

Write-Host "Correction du formatage pour le service combat..." -ForegroundColor Green

# Aller dans le répertoire du service combat
Set-Location "services/combat"

# Appliquer gofumpt sur tous les fichiers Go
Write-Host "Application de gofumpt..." -ForegroundColor Yellow
gofumpt -w .

# Vérifier si gofumpt a fonctionné
if ($LASTEXITCODE -eq 0) {
    Write-Host "gofumpt appliqué avec succès" -ForegroundColor Green
} else {
    Write-Host "Erreur avec gofumpt, tentative avec go fmt..." -ForegroundColor Yellow
    go fmt ./...
}

# Retourner au répertoire racine
Set-Location "../.."

Write-Host "Formatage terminé pour le service combat" -ForegroundColor Green 