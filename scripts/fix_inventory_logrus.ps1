#!/usr/bin/env pwsh

Write-Host "🔧 Ajout des imports logrus manquants dans inventory..." -ForegroundColor Cyan

$files = @(
    "services/inventory/internal/repository/inventory_repository.go",
    "services/inventory/internal/repository/item_repository.go", 
    "services/inventory/internal/repository/trade_repository.go"
)

foreach ($file in $files) {
    Write-Host "Traitement: $file" -ForegroundColor Yellow
    
    $content = Get-Content -Path $file -Raw
    
    # Ajouter logrus s'il n'existe pas
    if ($content -notmatch 'github.com/sirupsen/logrus') {
        $content = $content -replace '(\s+"github\.com/jmoiron/sqlx")', '$1
	"github.com/sirupsen/logrus"'
        
        Set-Content -Path $file -Value $content -NoNewline
        Write-Host "✅ Import logrus ajouté: $file" -ForegroundColor Green
    } else {
        Write-Host "ℹ️  Import logrus déjà présent: $file" -ForegroundColor Blue
    }
}

Write-Host ""
Write-Host "🎉 Correction des imports terminée !" -ForegroundColor Green 