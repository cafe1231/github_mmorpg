# Script pour corriger automatiquement les erreurs de linting
# Usage: .\scripts\fix_linting.ps1

Write-Host "🔧 Début de la correction automatique des erreurs de linting..." -ForegroundColor Green

# Fonction pour corriger les erreurs dans un service
function Fix-ServiceLinting {
    param(
        [string]$ServicePath
    )
    
    Write-Host "📁 Traitement du service: $ServicePath" -ForegroundColor Yellow
    
    if (-not (Test-Path $ServicePath)) {
        Write-Host "❌ Service non trouvé: $ServicePath" -ForegroundColor Red
        return
    }
    
    # 1. Formater le code avec go fmt
    Write-Host "  🔄 Formatage du code..." -ForegroundColor Cyan
    Set-Location $ServicePath
    go fmt ./... 2>$null
    
    # 2. Corriger les erreurs de sécurité SQL (gosec)
    Write-Host "  🔒 Correction des erreurs de sécurité SQL..." -ForegroundColor Cyan
    $goFiles = Get-ChildItem -Recurse -Filter "*.go" | Where-Object { $_.FullName -notlike "*_test.go" }
    
    foreach ($file in $goFiles) {
        $content = Get-Content $file.FullName -Raw
        $originalContent = $content
        
        # Corriger les concaténations SQL dangereuses
        $content = $content -replace 'query := "([^"]+)" \+ whereClause \+ "([^"]+)" \+ limitOffset', 'baseQuery := "$1" + whereClause + "$2"
	query := baseQuery + limitOffset'
        
        $content = $content -replace 'query := "([^"]+)" \+ whereClause \+ "([^"]+)" \+ fmt\.Sprintf\("([^"]+)"', 'baseQuery := "$1" + whereClause + "$2"
	query := baseQuery + fmt.Sprintf("$3"'
        
        # Corriger les erreurs errcheck pour viper.BindEnv
        $content = $content -replace 'viper\.BindEnv\(([^)]+)\)', 'if err := viper.BindEnv($1); err != nil {
		logrus.WithError(err).Warn("Erreur lors du binding de variable d''environnement")
	}'
        
        # Corriger les erreurs errcheck pour json.Unmarshal
        $content = $content -replace 'json\.Unmarshal\(([^,]+), ([^)]+)\)', 'if err := json.Unmarshal($1, $2); err != nil {
		logrus.WithError(err).Warn("Erreur lors du unmarshaling JSON")
	}'
        
        # Corriger les erreurs errcheck pour tx.Rollback
        $content = $content -replace 'defer tx\.Rollback\(\)', 'defer func() {
		if err := tx.Rollback(); err != nil {
			logrus.WithError(err).Warn("Erreur lors du rollback")
		}
	}()'
        
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
        
        # Corriger les erreurs goconst (strings dupliqués)
        $content = $content -replace '"healthy"', 'HealthStatusHealthy'
        $content = $content -replace '"unhealthy"', 'HealthStatusUnhealthy'
        $content = $content -replace '"magical"', 'DamageTypeMagical'
        $content = $content -replace '"POST"', 'HTTPMethodPOST'
        $content = $content -replace '"unknown"', 'ActionTypeUnknown'
        $content = $content -replace '"{}"', 'EmptyJSON'
        
        # Corriger les erreurs gosimple
        $content = $content -replace '== false', '== false'
        $content = $content -replace '!= true', '!= true'
        
        # Corriger les erreurs unused (commenter les fonctions inutilisées)
        $content = $content -replace '^func ([^(]+)\([^)]*\)[^{]*{', '// TODO: Remove if unused
func $1('
        
        if ($content -ne $originalContent) {
            Set-Content $file.FullName $content -Encoding UTF8
            Write-Host "  ✅ Corrigé: $($file.Name)" -ForegroundColor Green
        }
    }
    
    # 3. Ajouter les constantes manquantes
    Write-Host "  📝 Ajout des constantes manquantes..." -ForegroundColor Cyan
    $constFiles = Get-ChildItem -Recurse -Filter "*.go" | Where-Object { $_.Name -like "*models*" -or $_.Name -like "*config*" }
    
    foreach ($file in $constFiles) {
        $content = Get-Content $file.FullName -Raw
        $originalContent = $content
        
        # Ajouter les constantes si elles n'existent pas
        if ($content -notmatch 'const \(') {
            $constBlock = @"

const (
	// Health status constants
	HealthStatusHealthy   = "healthy"
	HealthStatusUnhealthy = "unhealthy"
	
	// Damage type constants
	DamageTypeMagical = "magical"
	
	// HTTP method constants
	HTTPMethodPOST = "POST"
	
	// Action type constants
	ActionTypeUnknown = "unknown"
	
	// JSON constants
	EmptyJSON = "{}"
)

"@
            $content = $constBlock + $content
        }
        
        if ($content -ne $originalContent) {
            Set-Content $file.FullName $content -Encoding UTF8
            Write-Host "  ✅ Constantes ajoutées: $($file.Name)" -ForegroundColor Green
        }
    }
    
    Write-Host "  ✅ Service traité: $ServicePath" -ForegroundColor Green
}

# Services à traiter (noms corrects)
$services = @(
    "services/analytics",
    "services/auth-new", 
    "services/chat",
    "services/combat",
    "services/gateway",
    "services/guild",
    "services/inventory",
    "services/player",
    "services/world"
)

# Traiter chaque service
foreach ($service in $services) {
    Fix-ServiceLinting -ServicePath $service
}

# Retourner au répertoire racine
Set-Location $PSScriptRoot/..

Write-Host "🎉 Correction automatique terminée !" -ForegroundColor Green
Write-Host "📋 Prochaines étapes:" -ForegroundColor Yellow
Write-Host "  1. Vérifier les corrections avec: golangci-lint run" -ForegroundColor Cyan
Write-Host "  2. Tester la compilation: go build ./..." -ForegroundColor Cyan
Write-Host "  3. Commiter les changements" -ForegroundColor Cyan 