#!/usr/bin/env pwsh

Write-Host "🔧 Correction des erreurs de syntaxe json.Unmarshal dans tous les services..." -ForegroundColor Cyan

# Définir le répertoire des services
$ServicesDir = "services"

# Fonction pour corriger les double if err :=
function Fix-DoubleIfErr {
    param($FilePath)
    
    $content = Get-Content -Path $FilePath -Raw
    $originalContent = $content
    
    # Corriger les patterns de double if err := pour json.Unmarshal
    $patterns = @(
        @{
            # Pattern: if err := if err := json.Unmarshal(...); err != nil { ... }; err != nil { ... }
            Old = 'if err := if err := json\.Unmarshal\(([^)]+)\); err != nil \{\s*([^}]+)\s*\}; err != nil \{'
            New = 'if err := json.Unmarshal($1); err != nil {'
        }
    )
    
    foreach ($pattern in $patterns) {
        $content = $content -replace $pattern.Old, $pattern.New
    }
    
    # Si le contenu a changé, sauvegarder
    if ($content -ne $originalContent) {
        Set-Content -Path $FilePath -Value $content -NoNewline
        Write-Host "✅ Corrigé: $FilePath" -ForegroundColor Green
        return $true
    }
    
    return $false
}

# Fonction pour nettoyer les lignes vides en trop
function Clean-ExtraLines {
    param($FilePath)
    
    $lines = Get-Content -Path $FilePath
    $cleanedLines = @()
    $previousEmpty = $false
    
    foreach ($line in $lines) {
        $isEmpty = [string]::IsNullOrWhiteSpace($line)
        
        # Éviter les lignes vides consécutives
        if (-not ($isEmpty -and $previousEmpty)) {
            $cleanedLines += $line
        }
        
        $previousEmpty = $isEmpty
    }
    
    Set-Content -Path $FilePath -Value $cleanedLines
}

# Parcourir tous les fichiers .go dans les services
$goFiles = Get-ChildItem -Path $ServicesDir -Recurse -Filter "*.go" | Where-Object { 
    $_.FullName -notmatch "vendor|node_modules|\.git" 
}

$totalFixed = 0

foreach ($file in $goFiles) {
    Write-Host "Traitement: $($file.FullName)" -ForegroundColor Yellow
    
    if (Fix-DoubleIfErr -FilePath $file.FullName) {
        Clean-ExtraLines -FilePath $file.FullName
        $totalFixed++
    }
}

Write-Host ""
Write-Host "🎉 Correction terminée ! $totalFixed fichiers corrigés." -ForegroundColor Green

# Vérifier la compilation après correction
Write-Host ""
Write-Host "🔍 Vérification de la compilation des services critiques..." -ForegroundColor Cyan

$criticalServices = @("combat", "player", "world")

foreach ($service in $criticalServices) {
    Write-Host "Compilation de $service..." -ForegroundColor Yellow
    
    Set-Location "services/$service"
    $result = go build -o temp_test.exe ./cmd/main.go 2>&1
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ $service compile correctement" -ForegroundColor Green
        if (Test-Path "temp_test.exe") {
            Remove-Item "temp_test.exe" -Force
        }
    } else {
        Write-Host "❌ $service a encore des erreurs:" -ForegroundColor Red
        Write-Host $result -ForegroundColor Red
    }
    
    Set-Location "../.."
}

Write-Host ""
Write-Host "🔧 Phase 1 de correction terminée !" -ForegroundColor Cyan 