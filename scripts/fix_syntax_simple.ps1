#!/usr/bin/env pwsh

Write-Host "🔧 Correction manuelle des erreurs de syntaxe..." -ForegroundColor Cyan

# Fichiers avec les erreurs identifiées
$FilesToFix = @(
    "services/player/internal/repository/character.go",
    "services/player/internal/repository/player.go", 
    "services/world/internal/repository/stubs.go",
    "services/world/internal/repository/zone.go",
    "services/combat/internal/repository/pvp_repository.go"
)

foreach ($file in $FilesToFix) {
    Write-Host "Correction de: $file" -ForegroundColor Yellow
    
    $content = Get-Content -Path $file -Raw
    
    # Pattern simple: remplacer "if err := if err :=" par "if err :="
    $content = $content -replace 'if err := if err :=', 'if err :='
    
    # Nettoyer les structures cassées comme "}; err != nil {"
    $content = $content -replace '\}; err != nil \{\s*([^}]+)\s*\}', "`n`t`tlogrus.WithError(err).Warn(`"Erreur lors du unmarshaling JSON`")`n`t}"
    
    Set-Content -Path $file -Value $content -NoNewline
    Write-Host "✅ Corrigé: $file" -ForegroundColor Green
}

# Corriger aussi les imports manquants logrus
$logrusFiles = @(
    "services/combat/internal/repository/combat_repository.go",
    "services/combat/internal/repository/effect_repository.go"
)

foreach ($file in $logrusFiles) {
    Write-Host "Ajout de l'import logrus dans: $file" -ForegroundColor Yellow
    
    $content = Get-Content -Path $file -Raw
    
    # Ajouter l'import logrus s'il n'existe pas
    if ($content -notmatch 'github.com/sirupsen/logrus') {
        $content = $content -replace '(import \([\s\S]*?)"fmt"', '$1"fmt"
	"github.com/sirupsen/logrus"'
    }
    
    Set-Content -Path $file -Value $content -NoNewline
    Write-Host "✅ Import ajouté: $file" -ForegroundColor Green
}

Write-Host ""
Write-Host "🎉 Correction manuelle terminée !" -ForegroundColor Green 