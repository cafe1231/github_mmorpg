# Script simple pour corriger les erreurs de linting
# Usage: .\scripts\fix_linting_simple.ps1

Write-Host "🔧 Correction automatique des erreurs de linting..." -ForegroundColor Green

# Fonction pour corriger un service
function Fix-Service {
    param([string]$ServiceName)
    
    $servicePath = Join-Path $PSScriptRoot "..\services\$ServiceName"
    if (-not (Test-Path $servicePath)) {
        Write-Host "❌ Service non trouvé: $servicePath" -ForegroundColor Red
        return
    }
    
    Write-Host "📁 Traitement: $ServiceName" -ForegroundColor Yellow
    Set-Location $servicePath
    
    # 1. Formater le code
    Write-Host "  🔄 Formatage..." -ForegroundColor Cyan
    go fmt ./... 2>$null
    
    # 2. Corriger les erreurs dans les fichiers Go
    $goFiles = Get-ChildItem -Recurse -Filter "*.go" | Where-Object { $_.FullName -notlike "*_test.go" }
    
    foreach ($file in $goFiles) {
        $content = Get-Content $file.FullName -Raw
        $originalContent = $content
        
        # Corriger les fautes d'orthographe françaises
        $content = $content -replace '\bconnexions\b', 'connections'
        $content = $content -replace '\bconnexion\b', 'connection'
        $content = $content -replace '\butilise\b', 'utilize'
        $content = $content -replace '\bstatuts\b', 'status'
        $content = $content -replace '\bcancelled\b', 'canceled'
        $content = $content -replace '\bcompense\b', 'compensate'
        $content = $content -replace '\bindividuel\b', 'individual'
        $content = $content -replace '\banalyse\b', 'analyze'
        $content = $content -replace '\bsuspectes\b', 'suspicious'
        $content = $content -replace '\bvictoires\b', 'victories'
        $content = $content -replace '\babsolue\b', 'absolute'
        $content = $content -replace '\bcandidats\b', 'candidates'
        $content = $content -replace '\bactivit\b', 'activity'
        $content = $content -replace '\bcalculs\b', 'calculations'
        $content = $content -replace '\bActivit\b', 'Activity'
        $content = $content -replace '\bConsistance\b', 'Consistency'
        $content = $content -replace '\bintervalles\b', 'intervals'
        $content = $content -replace '\bexistantes\b', 'existing'
        $content = $content -replace '\bparticipe\b', 'participate'
        $content = $content -replace '\bCalculs\b', 'Calculations'
        $content = $content -replace '\bexistant\b', 'existing'
        $content = $content -replace '\btrafic\b', 'traffic'
        $content = $content -replace '\bexemple\b', 'example'
        $content = $content -replace '\binitialise\b', 'initialize'
        
        # Corriger les erreurs errcheck pour tx.Rollback
        $content = $content -replace 'defer tx\.Rollback\(\)', 'defer func() {
		if err := tx.Rollback(); err != nil {
			logrus.WithError(err).Warn("Erreur lors du rollback")
		}
	}()'
        
        # Corriger les erreurs errcheck pour json.Unmarshal
        $content = $content -replace 'json\.Unmarshal\(([^,]+), ([^)]+)\)', 'if err := json.Unmarshal($1, $2); err != nil {
		logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")
	}'
        
        if ($content -ne $originalContent) {
            Set-Content $file.FullName $content -Encoding UTF8
            Write-Host "  ✅ Corrigé: $($file.Name)" -ForegroundColor Green
        }
    }
    
    Write-Host "  ✅ Terminé: $ServiceName" -ForegroundColor Green
}

# Traiter les services principaux
$services = @("analytics", "chat", "combat", "guild", "inventory", "player", "world")

foreach ($service in $services) {
    Fix-Service -ServiceName $service
}

# Retourner au répertoire racine
Set-Location $PSScriptRoot/..

Write-Host "🎉 Correction terminée !" -ForegroundColor Green 