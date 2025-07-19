#!/usr/bin/env pwsh

Write-Host "🧹 Nettoyage des accolades et erreurs de syntaxe..." -ForegroundColor Cyan

$file = "services/world/internal/repository/stubs.go"

Write-Host "Nettoyage de: $file" -ForegroundColor Yellow

$content = Get-Content -Path $file -Raw

# Corriger les accolades doubles
$content = $content -replace '\}\s*\}', '}'

# Corriger les lignes avec seulement "logrus.WithError..." orphelines
$content = $content -replace '\s*logrus\.WithError\(err\)\.Warn\("Erreur lors du unmarshaling JSON"\)\s*\}\s*\}', '
		logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")
	}'

# Nettoyer les espaces en trop
$content = $content -replace '\n\s*\n\s*\n', "`n`n"

Set-Content -Path $file -Value $content -NoNewline
Write-Host "✅ Nettoyé: $file" -ForegroundColor Green

# Ajouter l'import logrus s'il manque
$content = Get-Content -Path $file -Raw
if ($content -notmatch 'github.com/sirupsen/logrus') {
    $content = $content -replace '(import \([\s\S]*?)"fmt"', '$1"fmt"
	"github.com/sirupsen/logrus"'
    Set-Content -Path $file -Value $content -NoNewline
    Write-Host "✅ Import logrus ajouté" -ForegroundColor Green
}

Write-Host ""
Write-Host "🎉 Nettoyage terminé !" -ForegroundColor Green 