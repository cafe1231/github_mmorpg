#!/usr/bin/env pwsh

Write-Host "🎯 CORRECTION COMPLÈTE - Service Chat" -ForegroundColor Cyan
Write-Host "====================================" -ForegroundColor Cyan

$ServicePath = "services/chat"
$BackupPath = "backup_chat_" + (Get-Date).ToString("yyyyMMdd_HHmmss")

# Créer une sauvegarde
Write-Host "📦 Création de la sauvegarde: $BackupPath" -ForegroundColor Yellow
Copy-Item -Path $ServicePath -Destination $BackupPath -Recurse

function Fix-DeprecatedImports {
    param($FilePath)
    
    $content = Get-Content -Path $FilePath -Raw
    $originalContent = $content
    
    # Remplacer io/ioutil par io et os
    $content = $content -replace '"io/ioutil"', '"io"'
    $content = $content -replace 'ioutil\.ReadFile', 'os.ReadFile'
    $content = $content -replace 'ioutil\.WriteFile', 'os.WriteFile'
    $content = $content -replace 'ioutil\.ReadDir', 'os.ReadDir'
    $content = $content -replace 'ioutil\.TempFile', 'os.CreateTemp'
    $content = $content -replace 'ioutil\.TempDir', 'os.MkdirTemp'
    
    if ($content -ne $originalContent) {
        Set-Content -Path $FilePath -Value $content -NoNewline
        return $true
    }
    return $false
}

function Fix-BuiltinShadowing {
    param($FilePath)
    
    $content = Get-Content -Path $FilePath -Raw
    $originalContent = $content
    
    # Renommer la fonction min qui shadow la builtin
    $content = $content -replace 'func min\(', 'func minInt('
    $content = $content -replace 'min\(50,', 'minInt(50,'
    
    if ($content -ne $originalContent) {
        Set-Content -Path $FilePath -Value $content -NoNewline
        return $true
    }
    return $false
}

function Fix-BoolComparisons {
    param($FilePath)
    
    $content = Get-Content -Path $FilePath -Raw
    $originalContent = $content
    
    # Simplifier les comparaisons booléennes
    $content = $content -replace '(\w+\.?\w*) == false', '!$1'
    $content = $content -replace '(\w+\.?\w*) == true', '$1'
    
    if ($content -ne $originalContent) {
        Set-Content -Path $FilePath -Value $content -NoNewline
        return $true
    }
    return $false
}

function Add-MagicNumberConstants {
    param($FilePath)
    
    $content = Get-Content -Path $FilePath -Raw
    $originalContent = $content
    
    # Ajouter les constantes au début du fichier pour config.go
    if ($FilePath -like "*config.go") {
        $constants = @"
// Configuration constants
const (
	DefaultServerPort     = 8087
	DefaultDBPort         = 5432
	DefaultTimeout        = 30
	DefaultMaxOpenConns   = 25
	DefaultMaxIdleConns   = 10
	DefaultDBMaxLifetime  = 5
	DefaultMaxMessageLen  = 500
	DefaultRetentionDays  = 30
	DefaultMaxChannels    = 10
	DefaultAntiSpamMax    = 10
	DefaultBufferSize     = 1024
	DefaultHandshakeTO    = 10
	DefaultPongWait       = 60
	DefaultPingPeriod     = 54
	DefaultWriteWait      = 10
	DefaultMaxMsgSize     = 512
	DefaultMsgsPerMin     = 60
	DefaultBurstSize      = 10
	DefaultCleanupInt     = 5
	DefaultPrometheusPort = 9087
	DefaultJWTMinLength   = 32
	DefaultDBTimeout      = 5
	DefaultMigrationParts = 2
	DefaultLogContentLen  = 50
	DefaultShutdownTO     = 30
	MemoryConversionFactor = 1024
)

"@
        
        # Insérer après les imports
        $content = $content -replace '(import \([^)]+\))', "`$1`n`n$constants"
        
        # Remplacer les magic numbers par les constantes
        $content = $content -replace ': 8087,', ': DefaultServerPort,'
        $content = $content -replace ': 5432,', ': DefaultDBPort,'
        $content = $content -replace ': 30 \*', ': DefaultTimeout *'
        $content = $content -replace ': 25,', ': DefaultMaxOpenConns,'
        $content = $content -replace ': 10,', ': DefaultMaxIdleConns,'
        $content = $content -replace ': 5 \*', ': DefaultDBMaxLifetime *'
        $content = $content -replace ': 500,', ': DefaultMaxMessageLen,'
        $content = $content -replace ': 30,', ': DefaultRetentionDays,'
        $content = $content -replace ': 10,', ': DefaultMaxChannels,'
        $content = $content -replace ': 1024,', ': DefaultBufferSize,'
        $content = $content -replace ': 10 \*', ': DefaultHandshakeTO *'
        $content = $content -replace ': 60 \*', ': DefaultPongWait *'
        $content = $content -replace ': 54 \*', ': DefaultPingPeriod *'
        $content = $content -replace ': 512,', ': DefaultMaxMsgSize,'
        $content = $content -replace ': 60,', ': DefaultMsgsPerMin,'
        $content = $content -replace ': 5 \*', ': DefaultCleanupInt *'
        $content = $content -replace ': 9087,', ': DefaultPrometheusPort,'
        $content = $content -replace '< 32', '< DefaultJWTMinLength'
    }
    
    # Pour les autres fichiers, remplacer les magic numbers courants
    $content = $content -replace '5\*time\.Second', 'DefaultDBTimeout*time.Second'
    $content = $content -replace 'len\(parts\) < 2', 'len(parts) < DefaultMigrationParts'
    $content = $content -replace '50,', 'DefaultLogContentLen,'
    $content = $content -replace '30\*time\.Second', 'DefaultShutdownTO*time.Second'
    
    if ($content -ne $originalContent) {
        Set-Content -Path $FilePath -Value $content -NoNewline
        return $true
    }
    return $false
}

function Fix-HugeParams {
    param($FilePath)
    
    $content = Get-Content -Path $FilePath -Raw
    $originalContent = $content
    
    # Changer les gros paramètres pour utiliser des pointeurs
    $content = $content -replace 'func \(d DatabaseConfig\)', 'func (d *DatabaseConfig)'
    $content = $content -replace 'func NewConnection\(cfg config\.DatabaseConfig\)', 'func NewConnection(cfg *config.DatabaseConfig)'
    
    if ($content -ne $originalContent) {
        Set-Content -Path $FilePath -Value $content -NoNewline
        return $true
    }
    return $false
}

function Fix-LongLines {
    param($FilePath)
    
    $lines = Get-Content -Path $FilePath
    $modified = $false
    $newLines = @()
    
    foreach ($line in $lines) {
        if ($line.Length -gt 140) {
            # Diviser les signatures de fonction longues
            if ($line -match '^func.*\(.*\).*\{?$') {
                # Extraire les parties de la fonction
                if ($line -match '^(func.*?)\((.*?)\)(.*)$') {
                    $funcStart = $matches[1]
                    $params = $matches[2]
                    $funcEnd = $matches[3]
                    
                    # Diviser les paramètres s'ils sont longs
                    if ($params.Length -gt 80) {
                        $paramArray = $params -split ', '
                        if ($paramArray.Count -gt 1) {
                            $newLines += $funcStart + '('
                            foreach ($param in $paramArray) {
                                $newLines += "`t" + $param.Trim() + ','
                            }
                            # Enlever la dernière virgule et fermer
                            $newLines[-1] = $newLines[-1].TrimEnd(',')
                            $newLines += ')' + $funcEnd
                            $modified = $true
                            continue
                        }
                    }
                }
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

function Fix-SpellingErrors {
    param($FilePath)
    
    $content = Get-Content -Path $FilePath -Raw
    $originalContent = $content
    
    # Corrections d'orthographe courantes
    $content = $content -replace 'marrage', 'startup'
    
    if ($content -ne $originalContent) {
        Set-Content -Path $FilePath -Value $content -NoNewline
        return $true
    }
    return $false
}

# Exécution principale
Write-Host "🚀 Début des corrections..." -ForegroundColor Green

$goFiles = Get-ChildItem -Path $ServicePath -Recurse -Filter "*.go"
$totalFixed = 0

foreach ($file in $goFiles) {
    Write-Host "  📄 Traitement: $($file.Name)" -ForegroundColor Blue
    
    $fileFixed = $false
    
    # Appliquer toutes les corrections
    if (Fix-DeprecatedImports -FilePath $file.FullName) { $fileFixed = $true }
    if (Fix-BuiltinShadowing -FilePath $file.FullName) { $fileFixed = $true }
    if (Fix-BoolComparisons -FilePath $file.FullName) { $fileFixed = $true }
    if (Add-MagicNumberConstants -FilePath $file.FullName) { $fileFixed = $true }
    if (Fix-HugeParams -FilePath $file.FullName) { $fileFixed = $true }
    if (Fix-LongLines -FilePath $file.FullName) { $fileFixed = $true }
    if (Fix-SpellingErrors -FilePath $file.FullName) { $fileFixed = $true }
    
    if ($fileFixed) {
        Write-Host "    ✅ Corrigé" -ForegroundColor Green
        $totalFixed++
    }
}

# Ajouter les imports manquants pour les constantes
Write-Host "📦 Ajout des imports manquants..." -ForegroundColor Yellow
$configFile = Join-Path $ServicePath "internal/config/config.go"
if (Test-Path $configFile) {
    $content = Get-Content -Path $configFile -Raw
    # S'assurer que time est importé
    if ($content -notmatch '"time"') {
        $content = $content -replace '(import \()', '$1`n`t"time"'
        Set-Content -Path $configFile -Value $content -NoNewline
    }
}

# Corriger les références après les changements de pointeurs
Write-Host "🔧 Correction des références..." -ForegroundColor Yellow
$connectionFile = Join-Path $ServicePath "internal/database/connection.go"
if (Test-Path $connectionFile) {
    $content = Get-Content -Path $connectionFile -Raw
    # Corriger les appels qui doivent maintenant utiliser des pointeurs
    $content = $content -replace 'cfg\.GetDatabaseURL\(\)', 'cfg.GetDatabaseURL()'
    Set-Content -Path $connectionFile -Value $content -NoNewline
}

# Appliquer le formatage final
Write-Host "🎨 Application du formatage Go..." -ForegroundColor Yellow

Push-Location $ServicePath

# Appliquer go fmt
go fmt ./...

# Appliquer goimports si disponible
$goimports = Get-Command goimports -ErrorAction SilentlyContinue
if ($goimports) {
    Get-ChildItem -Recurse -Filter "*.go" | ForEach-Object {
        goimports -w $_.FullName
    }
}

Pop-Location

Write-Host ""
Write-Host "📊 RÉSULTATS:" -ForegroundColor Cyan
Write-Host "- $totalFixed fichiers corrigés" -ForegroundColor Green
Write-Host "- Sauvegarde créée: $BackupPath" -ForegroundColor Green
Write-Host "- Formatage Go appliqué" -ForegroundColor Green

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
    
    Write-Host ""
    Write-Host "🔄 Restauration de la sauvegarde..." -ForegroundColor Yellow
    Pop-Location
    Remove-Item -Path $ServicePath -Recurse -Force
    Copy-Item -Path $BackupPath -Destination $ServicePath -Recurse
    Write-Host "❗ Sauvegarde restaurée en raison d'erreurs de compilation" -ForegroundColor Red
    exit 1
}

Pop-Location

# Test de linting final
Write-Host ""
Write-Host "🔍 Test de linting final..." -ForegroundColor Yellow

Push-Location $ServicePath
$lintResult = golangci-lint run --timeout=5m 2>&1
$lintSuccess = $LASTEXITCODE -eq 0

if ($lintSuccess) {
    Write-Host "🎉 AUCUNE ERREUR DE LINTING!" -ForegroundColor Green
    Write-Host "Service chat est maintenant parfaitement propre!" -ForegroundColor Green
} else {
    $errorCount = ($lintResult | Select-String "Error:" | Measure-Object).Count
    Write-Host "⚠️  $errorCount erreurs de linting restantes" -ForegroundColor Yellow
    Write-Host "Détails:" -ForegroundColor Yellow
    Write-Host $lintResult -ForegroundColor Yellow
}

Pop-Location

Write-Host ""
Write-Host "🎯 Corrections terminées pour le service Chat!" -ForegroundColor Green 