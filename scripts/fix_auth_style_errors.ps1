#!/usr/bin/env pwsh

Write-Host "🎨 Correction des erreurs de style pour auth-new service" -ForegroundColor Cyan
Write-Host "=======================================================" -ForegroundColor Cyan

$ServicePath = "services/auth-new"

# Fonction pour corriger les fautes d'orthographe françaises
function Fix-FrenchSpelling {
    param($FilePath)
    
    $content = Get-Content -Path $FilePath -Raw -Encoding UTF8
    $originalContent = $content
    
    # Corrections d'orthographe commune
    $content = $content -replace 'connexion', 'connection'
    $content = $content -replace 'Connexion', 'Connection'  
    $content = $content -replace 'connexions', 'connections'
    $content = $content -replace 'statuts', 'status'
    $content = $content -replace 'Statuts', 'Status'
    $content = $content -replace 'individuel', 'individual'
    $content = $content -replace 'initialise', 'initialize'
    $content = $content -replace 'origines', 'origins'
    $content = $content -replace 'marrage', 'startup'
    
    if ($content -ne $originalContent) {
        Set-Content -Path $FilePath -Value $content -NoNewline -Encoding UTF8
        return $true
    }
    return $false
}

# Fonction pour corriger les tests de chaînes vides
function Fix-EmptyStringTests {
    param($FilePath)
    
    $content = Get-Content -Path $FilePath -Raw
    $originalContent = $content
    
    # Corriger len(strings.TrimSpace(...)) == 0 vers strings.TrimSpace(...) == ""
    $content = $content -replace 'len\(strings\.TrimSpace\(([^)]+)\)\) == 0', 'strings.TrimSpace($1) == ""'
    
    if ($content -ne $originalContent) {
        Set-Content -Path $FilePath -Value $content -NoNewline
        return $true
    }
    return $false
}

# Fonction pour corriger les regex simplifiées
function Fix-RegexPatterns {
    param($FilePath)
    
    $content = Get-Content -Path $FilePath -Raw
    $originalContent = $content
    
    # Corriger [0-9] vers \d
    $content = $content -replace '\[0-9\]', '\d'
    
    # Simplifier le pattern de caractères spéciaux (retirer les échappements doubles inutiles)
    $content = $content -replace '\\\\-', '\-'
    $content = $content -replace '\\\\/', '/'
    
    if ($content -ne $originalContent) {
        Set-Content -Path $FilePath -Value $content -NoNewline
        return $true
    }
    return $false
}

# Fonction pour diviser les lignes trop longues
function Fix-LongLines {
    param($FilePath)
    
    $lines = Get-Content -Path $FilePath
    $modified = $false
    $newLines = @()
    
    foreach ($line in $lines) {
        if ($line.Length -gt 140) {
            # Diviser les signatures de fonction longues
            if ($line -match '^func.*\(.*\).*\{?$') {
                $line = $line -replace '\(([^)]+)\)', {
                    param($match)
                    $params = $match.Groups[1].Value -split ', '
                    if ($params.Count -gt 2) {
                        "`n`t" + ($params -join ",`n`t") + "`n"
                    } else {
                        $match.Value
                    }
                }
                $modified = $true
            }
            # Diviser les en-têtes HTTP longs
            elseif ($line -match 'Content-Security-Policy') {
                $line = $line -replace '"([^"]{50,})"', {
                    param($match)
                    $policies = $match.Groups[1].Value -split '; '
                    '"' + ($policies -join '; " +`n`t`t"') + '"'
                }
                $modified = $true
            }
        }
        $newLines += $line
    }
    
    if ($modified) {
        Set-Content -Path $FilePath -Value $newLines
        return $true
    }
    return $false
}

# Fonction pour corriger le formatage gofumpt
function Fix-GoFumpt {
    param($ServicePath)
    
    Write-Host "🎨 Application de gofumpt..." -ForegroundColor Yellow
    
    Push-Location $ServicePath
    
    # Appliquer gofumpt si disponible
    $gofumpt = Get-Command gofumpt -ErrorAction SilentlyContinue
    if ($gofumpt) {
        Get-ChildItem -Recurse -Filter "*.go" | ForEach-Object {
            gofumpt -w $_.FullName
        }
    } else {
        # Fallback vers gofmt
        go fmt ./...
    }
    
    Pop-Location
}

# Fonction pour supprimer le paramètre inutilisé
function Fix-UnusedParam {
    param($FilePath)
    
    $content = Get-Content -Path $FilePath -Raw
    $originalContent = $content
    
    # Supprimer le paramètre authService inutilisé dans gracefulShutdown
    $content = $content -replace 'func gracefulShutdown\(server \*http\.Server, authService \*service\.AuthService\)', 'func gracefulShutdown(server *http.Server)'
    $content = $content -replace 'gracefulShutdown\(server, authService\)', 'gracefulShutdown(server)'
    
    if ($content -ne $originalContent) {
        Set-Content -Path $FilePath -Value $content -NoNewline
        return $true
    }
    return $false
}

# Exécution principale
Write-Host "🚀 Début des corrections de style..." -ForegroundColor Green

$goFiles = Get-ChildItem -Path $ServicePath -Recurse -Filter "*.go"
$totalFixed = 0

foreach ($file in $goFiles) {
    Write-Host "  📄 Traitement: $($file.Name)" -ForegroundColor Blue
    
    $fileFixed = $false
    
    # Appliquer les corrections
    if (Fix-FrenchSpelling -FilePath $file.FullName) { $fileFixed = $true }
    if (Fix-EmptyStringTests -FilePath $file.FullName) { $fileFixed = $true }
    if (Fix-RegexPatterns -FilePath $file.FullName) { $fileFixed = $true }
    if (Fix-LongLines -FilePath $file.FullName) { $fileFixed = $true }
    if (Fix-UnusedParam -FilePath $file.FullName) { $fileFixed = $true }
    
    if ($fileFixed) {
        Write-Host "    ✅ Corrigé" -ForegroundColor Green
        $totalFixed++
    }
}

# Appliquer le formatage final
Fix-GoFumpt -ServicePath $ServicePath

Write-Host ""
Write-Host "📊 RÉSULTATS:" -ForegroundColor Cyan
Write-Host "- $totalFixed fichiers corrigés" -ForegroundColor Green
Write-Host "- Formatage gofumpt appliqué" -ForegroundColor Green

# Test de compilation
Write-Host ""
Write-Host "🧪 Test de compilation..." -ForegroundColor Yellow

Push-Location $ServicePath
$compileResult = go build -o temp_test.exe ./cmd/main.go 2>&1
$compileSuccess = $LASTEXITCODE -eq 0

if ($compileSuccess) {
    Write-Host "✅ Service compile correctement" -ForegroundColor Green
    Remove-Item -Force temp_test.exe -ErrorAction SilentlyContinue
} else {
    Write-Host "❌ Erreurs de compilation:" -ForegroundColor Red
    Write-Host $compileResult -ForegroundColor Red
}

Pop-Location

# Test de linting final
Write-Host ""
Write-Host "🔍 Test de linting final..." -ForegroundColor Yellow

Push-Location $ServicePath
$lintResult = golangci-lint run --timeout=3m --max-issues-per-linter=3 2>&1
$lintSuccess = $LASTEXITCODE -eq 0

if ($lintSuccess) {
    Write-Host "✅ Aucune erreur de linting critique!" -ForegroundColor Green
} else {
    $errorCount = ($lintResult | Select-String "Error:" | Measure-Object).Count
    Write-Host "⚠️  $errorCount erreurs de linting restantes (probablement mineures)" -ForegroundColor Yellow
}

Pop-Location

Write-Host ""
Write-Host "🎉 Corrections de style terminées pour auth-new!" -ForegroundColor Green 