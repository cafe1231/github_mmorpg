# Script pour forcer le formatage du service combat
# Approche plus agressive

Write-Host "Forçage du formatage pour le service combat..." -ForegroundColor Green

# Aller dans le répertoire du service combat
Set-Location "services/combat"

# Nettoyer le cache de gofumpt
Write-Host "Nettoyage du cache..." -ForegroundColor Yellow
if (Test-Path ".gofumpt-cache") {
    Remove-Item ".gofumpt-cache" -Recurse -Force
}

# Appliquer go fmt d'abord
Write-Host "Application de go fmt..." -ForegroundColor Yellow
go fmt ./...

# Appliquer gofumpt avec options forcées
Write-Host "Application de gofumpt forcé..." -ForegroundColor Yellow
gofumpt -extra -w .

# Appliquer gofumpt sur chaque fichier individuellement
Write-Host "Formatage fichier par fichier..." -ForegroundColor Yellow
$files = @(
    "cmd/main.go",
    "internal/config/config.go",
    "internal/database/database.go",
    "internal/models/action.go",
    "internal/models/combat.go",
    "internal/models/effect.go",
    "internal/models/pvp.go",
    "internal/models/requests.go",
    "internal/repository/action_repository.go",
    "internal/repository/combat_repository.go",
    "internal/repository/effect_repository.go",
    "internal/repository/pvp_repository.go",
    "internal/service/action_service.go",
    "internal/service/anti_cheat.go",
    "internal/service/combat_service.go",
    "internal/service/damage_calculator.go",
    "internal/service/effect_service.go",
    "internal/service/pvp_service.go",
    "internal/handlers/combat_handler.go",
    "internal/handlers/health_handler.go",
    "internal/handlers/pvp_handler.go",
    "internal/middleware/auth.go",
    "internal/middleware/metrics.go",
    "internal/middleware/rate_limit.go"
)

foreach ($file in $files) {
    if (Test-Path $file) {
        Write-Host "Formatage de $file..." -ForegroundColor Cyan
        gofumpt -extra -w $file
    }
}

# Retourner au répertoire racine
Set-Location "../.."

Write-Host "Formatage forcé terminé pour le service combat" -ForegroundColor Green 