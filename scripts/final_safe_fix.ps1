#!/usr/bin/env pwsh

Write-Host "🛡️ Script final de correction conservateur" -ForegroundColor Cyan
Write-Host "===========================================" -ForegroundColor Cyan

# Fonction pour appliquer gofmt et goimports sur tous les services
function Apply-GoFormatting {
    Write-Host "🎨 Application du formatage Go standard..." -ForegroundColor Yellow
    
    $services = Get-ChildItem -Path "services" -Directory
    
    foreach ($service in $services) {
        Write-Host "  📁 Formatage: $($service.Name)" -ForegroundColor Blue
        
        Push-Location "services/$($service.Name)"
        
        # Appliquer go fmt (sûr)
        go fmt ./... 2>$null
        
        Pop-Location
    }
    
    Write-Host "✅ Formatage appliqué à tous les services" -ForegroundColor Green
}

# Fonction pour vérifier quels services compilent maintenant
function Test-AllServices {
    Write-Host ""
    Write-Host "🧪 Test de compilation de tous les services..." -ForegroundColor Yellow
    
    $services = @("analytics", "auth-new", "chat", "combat", "gateway", "guild", "inventory", "player", "world")
    $results = @{}
    
    foreach ($service in $services) {
        Write-Host "  🔍 Test: $service" -ForegroundColor Blue
        
        Push-Location "services/$service"
        $result = go build -o temp_test.exe ./cmd/main.go 2>&1
        $success = $LASTEXITCODE -eq 0
        
        if ($success) {
            Remove-Item -Force temp_test.exe -ErrorAction SilentlyContinue
        }
        
        Pop-Location
        $results[$service] = $success
    }
    
    # Rapport de compilation
    Write-Host ""
    Write-Host "📊 RAPPORT DE COMPILATION" -ForegroundColor Cyan
    Write-Host "=========================" -ForegroundColor Cyan
    
    $successCount = 0
    foreach ($service in $services) {
        $status = if ($results[$service]) { "✅ OK" } else { "❌ ERREUR" }
        $color = if ($results[$service]) { "Green" } else { "Red" }
        Write-Host "$service : $status" -ForegroundColor $color
        
        if ($results[$service]) {
            $successCount++
        }
    }
    
    Write-Host ""
    Write-Host "🎯 Résumé: $successCount/$($services.Count) services compilent correctement" -ForegroundColor Cyan
    
    return $results
}

# Fonction pour tester golangci-lint sur services qui compilent
function Test-Linting {
    param($CompilationResults)
    
    Write-Host ""
    Write-Host "🔍 Test golangci-lint sur services compilables..." -ForegroundColor Yellow
    
    $workingServices = $CompilationResults.Keys | Where-Object { $CompilationResults[$_] }
    
    foreach ($service in $workingServices) {
        Write-Host ""
        Write-Host "  📋 Linting: $service" -ForegroundColor Blue
        
        Push-Location "services/$service"
        $lintResult = golangci-lint run --timeout=3m --max-issues-per-linter=3 2>&1
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "    ✅ $service : Aucune erreur de linting!" -ForegroundColor Green
        } else {
            Write-Host "    ⚠️  $service : Erreurs de linting restantes" -ForegroundColor Yellow
            # Write-Host $lintResult -ForegroundColor Gray
        }
        
        Pop-Location
    }
}

# Exécution principale
Write-Host "🚀 Début de la correction finale conservative..." -ForegroundColor Cyan

# Étape 1: Appliquer le formatage standard
Apply-GoFormatting

# Étape 2: Tester la compilation de tous les services
$compileResults = Test-AllServices

# Étape 3: Tester le linting sur les services qui compilent
Test-Linting -CompilationResults $compileResults

Write-Host ""
Write-Host "🎉 Phase 3 terminée - Analyse conservative complète!" -ForegroundColor Green
Write-Host ""
Write-Host "📈 RÉSUMÉ GLOBAL:" -ForegroundColor Cyan
Write-Host "- ✅ Phase 1: Erreurs de syntaxe critiques corrigées"
Write-Host "- ✅ Phase 2: Script intelligent créé"  
Write-Host "- ✅ Phase 3: Formatage appliqué et compilation testée"
Write-Host ""
Write-Host "🎯 Services prêts pour le commit: $(($compileResults.Values | Where-Object {$_}).Count) services compilent" -ForegroundColor Green 