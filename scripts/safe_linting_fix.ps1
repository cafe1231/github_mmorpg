#!/usr/bin/env pwsh

Write-Host "🛡️ Script de correction de linting sécurisé et intelligent" -ForegroundColor Cyan
Write-Host "=========================================================" -ForegroundColor Cyan

# Configuration
$ServicesDir = "services"
$BackupDir = "backup_linting_$(Get-Date -Format 'yyyyMMdd_HHmmss')"

# Services à corriger (excluant ceux avec erreurs critiques déjà corrigées)
$ServicesToFix = @("auth-new", "chat", "guild", "inventory")

# Fonction pour créer une sauvegarde
function Create-Backup {
    param($ServiceName)
    
    $backupPath = "$BackupDir/$ServiceName"
    Write-Host "📦 Création sauvegarde: $backupPath" -ForegroundColor Blue
    
    New-Item -ItemType Directory -Path $backupPath -Force | Out-Null
    Copy-Item -Path "$ServicesDir/$ServiceName" -Destination $backupPath -Recurse -Force
    
    return $backupPath
}

# Fonction pour vérifier la compilation avant/après
function Test-Compilation {
    param($ServiceName)
    
    Push-Location "$ServicesDir/$ServiceName"
    $result = go build -o temp_test.exe ./cmd/main.go 2>&1
    $success = $LASTEXITCODE -eq 0
    
    if ($success) {
        Remove-Item -Force temp_test.exe -ErrorAction SilentlyContinue
    }
    
    Pop-Location
    return $success
}

# Fonction pour corriger les erreurs errcheck (retours d'erreur non vérifiés)
function Fix-ErrCheck {
    param($FilePath, $Content)
    
    # Corriger viper.BindEnv non vérifiés
    $Content = $Content -replace 'viper\.BindEnv\(([^)]+)\)', '_ = viper.BindEnv($1)'
    
    # Corriger w.Write non vérifiés
    $Content = $Content -replace '(\s+)w\.Write\(([^)]+)\)', '$1_, _ = w.Write($2)'
    
    # Corriger s.userRepo.Update non vérifiés
    $Content = $Content -replace '(\s+)s\.userRepo\.Update\(([^)]+)\)', '$1if err := s.userRepo.Update($2); err != nil {
$1	logrus.WithError(err).Error("Erreur lors de la mise à jour utilisateur")
$1}'
    
    return $Content
}

# Fonction pour corriger les hugeParam (paramètres lourds)
function Fix-HugeParam {
    param($FilePath, $Content)
    
    # Corriger les fonctions avec paramètres DatabaseConfig lourds
    $Content = $Content -replace '(\w+)\((\w+) config\.DatabaseConfig\)', '$1($2 *config.DatabaseConfig)'
    $Content = $Content -replace '(\w+)\((\w+) (\w+\.)?DatabaseConfig\)', '$1($2 *$3DatabaseConfig)'
    
    # Corriger les structures de requête lourdes
    $patterns = @(
        'func \((\w+) ([^)]+Request)\)',
        'func \((\w+) ([^)]+Config)\)'
    )
    
    foreach ($pattern in $patterns) {
        $Content = $Content -replace $pattern, 'func ($1 *$2)'
    }
    
    return $Content
}

# Fonction pour corriger goconst (chaînes dupliquées)
function Fix-GoConst {
    param($FilePath, $Content)
    
    # Définir les constantes communes en haut du fichier
    $constants = @{
        'healthy' = 'StatusHealthy'
        'unhealthy' = 'StatusUnhealthy'
        'leader' = 'RoleLeader'
        'officer' = 'RoleOfficer'
        'magical' = 'DamageTypeMagical'
        'POST' = 'MethodPost'
        'unknown' = 'StatusUnknown'
    }
    
    foreach ($string in $constants.Keys) {
        $constantName = $constants[$string]
        $Content = $Content -replace "`"$string`"", $constantName
    }
    
    return $Content
}

# Fonction pour corriger mnd (nombres magiques)
function Fix-MagicNumbers {
    param($FilePath, $Content)
    
    # Définir les constantes pour les nombres magiques communs
    $magicNumbers = @{
        '8081' = 'DefaultAuthPort'
        '8082' = 'DefaultPlayerPort'  
        '8083' = 'DefaultCombatPort'
        '8084' = 'DefaultInventoryPort'
        '8087' = 'DefaultChatPort'
        '8088' = 'DefaultAnalyticsPort'
        '5432' = 'DefaultPostgresPort'
        '30' = 'DefaultTimeoutSeconds'
        '32' = 'MinJWTSecretLength'
        '25' = 'DefaultMaxConnections'
        '1024' = 'BytesToMB'
    }
    
    foreach ($number in $magicNumbers.Keys) {
        $constantName = $magicNumbers[$number]
        # Remplacer seulement dans les assignations, pas dans les chaînes
        $Content = $Content -replace "(\W)$number(\W)", "`$1$constantName`$2"
    }
    
    return $Content
}

# Fonction pour corriger misspell (fautes d'orthographe)
function Fix-Misspell {
    param($FilePath, $Content)
    
    $corrections = @{
        'connexion' = 'connection'
        'connexions' = 'connections'
        'statuts' = 'status'
        'individuel' = 'individual'
        'initialise' = 'initialize'
        'origines' = 'origins'
        'marrage' = 'startup'
        'compense' = 'reward'
        'Attributs' = 'Attributes'
    }
    
    foreach ($mistake in $corrections.Keys) {
        $correction = $corrections[$mistake]
        $Content = $Content -replace $mistake, $correction
    }
    
    return $Content
}

# Fonction pour corriger gocritic (suggestions de style)
function Fix-GoCritic {
    param($FilePath, $Content)
    
    # Corriger emptyStringTest
    $Content = $Content -replace 'len\(strings\.TrimSpace\(([^)]+)\)\) == 0', 'strings.TrimSpace($1) == ""'
    
    # Corriger builtinShadowDecl (shadowing d'identifiants prédéfinis)
    $Content = $Content -replace 'func min\(', 'func minInt('
    $Content = $Content -replace 'func ([^(]*)\(error,', 'func $1(err,'
    
    # Corriger assignOp
    $Content = $Content -replace '(\w+) = \1 \*', '$1 *='
    $Content = $Content -replace '(\w+) = \1 \+', '$1 +='
    
    return $Content
}

# Fonction pour corriger gofmt/gofumpt (formatage)
function Fix-Formatting {
    param($ServiceName)
    
    Write-Host "🎨 Application du formatage Go sur $ServiceName..." -ForegroundColor Yellow
    
    Push-Location "$ServicesDir/$ServiceName"
    
    # Appliquer gofmt
    go fmt ./...
    
    # Appliquer goimports si disponible
    $goimports = Get-Command goimports -ErrorAction SilentlyContinue
    if ($goimports) {
        Get-ChildItem -Recurse -Filter "*.go" | ForEach-Object {
            goimports -w $_.FullName
        }
    }
    
    Pop-Location
}

# Fonction principale de correction
function Fix-Service {
    param($ServiceName)
    
    Write-Host ""
    Write-Host "🔧 Traitement du service: $ServiceName" -ForegroundColor Green
    Write-Host "=====================================" -ForegroundColor Green
    
    # Créer sauvegarde
    $backupPath = Create-Backup -ServiceName $ServiceName
    
    # Test compilation avant
    Write-Host "🧪 Test compilation avant correction..." -ForegroundColor Yellow
    $compileBefore = Test-Compilation -ServiceName $ServiceName
    
    if (-not $compileBefore) {
        Write-Host "⚠️  Service $ServiceName ne compile pas avant correction!" -ForegroundColor Red
        return $false
    }
    
    # Appliquer les corrections
    $goFiles = Get-ChildItem -Path "$ServicesDir/$ServiceName" -Recurse -Filter "*.go"
    $filesFixed = 0
    
    foreach ($file in $goFiles) {
        $content = Get-Content -Path $file.FullName -Raw
        $originalContent = $content
        
        # Appliquer toutes les corrections
        $content = Fix-ErrCheck -FilePath $file.FullName -Content $content
        $content = Fix-HugeParam -FilePath $file.FullName -Content $content
        $content = Fix-GoConst -FilePath $file.FullName -Content $content
        $content = Fix-MagicNumbers -FilePath $file.FullName -Content $content
        $content = Fix-Misspell -FilePath $file.FullName -Content $content
        $content = Fix-GoCritic -FilePath $file.FullName -Content $content
        
        # Sauvegarder si modifié
        if ($content -ne $originalContent) {
            Set-Content -Path $file.FullName -Value $content -NoNewline
            $filesFixed++
            Write-Host "  ✅ Corrigé: $($file.Name)" -ForegroundColor Green
        }
    }
    
    # Appliquer le formatage
    Fix-Formatting -ServiceName $ServiceName
    
    # Test compilation après
    Write-Host "🧪 Test compilation après correction..." -ForegroundColor Yellow
    $compileAfter = Test-Compilation -ServiceName $ServiceName
    
    if ($compileAfter) {
        Write-Host "✅ Service $ServiceName corrigé avec succès! ($filesFixed fichiers modifiés)" -ForegroundColor Green
        return $true
    } else {
        Write-Host "❌ Erreur lors de la correction de $ServiceName, restauration sauvegarde..." -ForegroundColor Red
        
        # Restaurer la sauvegarde
        Remove-Item -Path "$ServicesDir/$ServiceName" -Recurse -Force
        Copy-Item -Path "$backupPath/$ServiceName" -Destination "$ServicesDir/" -Recurse -Force
        
        Write-Host "🔄 Sauvegarde restaurée pour $ServiceName" -ForegroundColor Yellow
        return $false
    }
}

# Exécution principale
Write-Host "🚀 Début de la correction automatique des services..." -ForegroundColor Cyan

# Créer le répertoire de sauvegarde
New-Item -ItemType Directory -Path $BackupDir -Force | Out-Null

$results = @{}
$totalSuccess = 0

foreach ($service in $ServicesToFix) {
    $success = Fix-Service -ServiceName $service
    $results[$service] = $success
    
    if ($success) {
        $totalSuccess++
    }
}

# Rapport final
Write-Host ""
Write-Host "📊 RAPPORT FINAL" -ForegroundColor Cyan
Write-Host "================" -ForegroundColor Cyan

foreach ($service in $ServicesToFix) {
    $status = if ($results[$service]) { "✅ SUCCÈS" } else { "❌ ÉCHEC" }
    $color = if ($results[$service]) { "Green" } else { "Red" }
    Write-Host "$service : $status" -ForegroundColor $color
}

Write-Host ""
Write-Host "🎯 Résumé: $totalSuccess/$($ServicesToFix.Count) services corrigés avec succès" -ForegroundColor Cyan
Write-Host "📦 Sauvegardes disponibles dans: $BackupDir" -ForegroundColor Blue

if ($totalSuccess -eq $ServicesToFix.Count) {
    Write-Host "🎉 Toutes les corrections appliquées avec succès!" -ForegroundColor Green
} else {
    Write-Host "⚠️  Certains services nécessitent une correction manuelle" -ForegroundColor Yellow
} 