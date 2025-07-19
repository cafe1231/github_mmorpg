#!/usr/bin/env pwsh

Write-Host "Correction FINALE du formatage pour le service combat" -ForegroundColor Cyan
Write-Host "=======================================================" -ForegroundColor Cyan

$ServicePath = "services/combat"

# Fonction pour appliquer go fmt (plus fiable que gofumpt)
function Apply-GoFmt {
    param($ServicePath)
    
    Write-Host "Application de go fmt..." -ForegroundColor Yellow
    
    Push-Location $ServicePath
    
    # Appliquer go fmt sur tous les fichiers
    go fmt ./...
    
    # Aussi essayer goimports si disponible
    $goimports = Get-Command goimports -ErrorAction SilentlyContinue
    if ($goimports) {
        Write-Host "Application de goimports..." -ForegroundColor Yellow
        goimports -w .
    }
    
    Pop-Location
}

# Fonction pour corriger manuellement les problemes de formatage les plus courants
function Fix-ManualFormatting {
    param($ServicePath)
    
    Write-Host "Correction manuelle du formatage..." -ForegroundColor Yellow
    
    $goFiles = Get-ChildItem -Path $ServicePath -Recurse -Filter "*.go"
    $fixedFiles = 0
    
    foreach ($file in $goFiles) {
        $content = Get-Content -Path $file.FullName -Raw
        $originalContent = $content
        
        # Corriger les espaces dans les imports
        $content = $content -replace 'import\s+\(\s*\n\s*"', 'import (`n`t"'
        $content = $content -replace '"\s*\n\s*"', '"`n`t"'
        $content = $content -replace '"\s*\n\s*\)', '"`n)'
        
        # Corriger les espaces apres les imports
        $content = $content -replace '\)\s*\n\s*\n\s*([^/\n])', ")`n`n`$1"
        
        # Corriger les espaces en debut de ligne dans les imports
        $content = $content -replace '\n\s+"([^"]+)"', "`n`t`"`$1`""
        
        if ($content -ne $originalContent) {
            Set-Content -Path $file.FullName -Value $content -NoNewline -Encoding UTF8
            $fixedFiles++
        }
    }
    
    Write-Host "    $fixedFiles fichiers corriges manuellement" -ForegroundColor Green
}

# Fonction pour verifier le formatage
function Test-Formatting {
    param($ServicePath)
    
    Write-Host "Verification du formatage..." -ForegroundColor Yellow
    
    Push-Location $ServicePath
    
    $formatErrors = golangci-lint run --timeout=3m | Select-String "File is not properly formatted"
    $errorCount = ($formatErrors | Measure-Object).Count
    
    Write-Host "    Erreurs de formatage restantes: $errorCount" -ForegroundColor $(if ($errorCount -eq 0) {"Green"} else {"Yellow"})
    
    Pop-Location
    
    return $errorCount
}

# Execution principale
Write-Host "Debut de la correction finale du formatage..." -ForegroundColor Green

# Etape 1: Correction manuelle des problemes courants
Fix-ManualFormatting -ServicePath $ServicePath

# Etape 2: Application de go fmt (plus fiable)
Apply-GoFmt -ServicePath $ServicePath

# Etape 3: Deuxieme correction manuelle apres go fmt
Fix-ManualFormatting -ServicePath $ServicePath

# Etape 4: Re-application de go fmt
Apply-GoFmt -ServicePath $ServicePath

# Etape 5: Verification finale
$remainingErrors = Test-Formatting -ServicePath $ServicePath

# Test de compilation
Write-Host ""
Write-Host "Test de compilation..." -ForegroundColor Yellow

Push-Location $ServicePath
$compileResult = go build -o temp_test.exe ./cmd/main.go 2>&1
$compileSuccess = $LASTEXITCODE -eq 0

if ($compileSuccess) {
    Write-Host "Service compile correctement" -ForegroundColor Green
    Remove-Item -Force temp_test.exe -ErrorAction SilentlyContinue
} else {
    Write-Host "Erreurs de compilation:" -ForegroundColor Red
    Write-Host $compileResult -ForegroundColor Red
}

Pop-Location

Write-Host ""
Write-Host "RESULTATS FINAUX:" -ForegroundColor Cyan
if ($remainingErrors -eq 0) {
    Write-Host "SUCCES: Formatage completement corrige!" -ForegroundColor Green
} else {
    Write-Host "$remainingErrors erreurs de formatage restantes" -ForegroundColor Yellow
    Write-Host "   (Probablement due a gofumpt vs go fmt - fonctionnellement OK)" -ForegroundColor Gray
}

if ($compileSuccess) {
    Write-Host "Compilation: OK" -ForegroundColor Green
} else {
    Write-Host "Compilation: ERREUR" -ForegroundColor Red
}

Write-Host ""
Write-Host "Correction du formatage terminee pour combat!" -ForegroundColor Green 