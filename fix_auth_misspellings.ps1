# Script pour corriger les fautes d'orthographe dans le service auth-new

Write-Host "🔧 Correction des fautes d'orthographe dans le service auth-new..."

# Définir les corrections (mot incorrect -> mot correct)
$corrections = @{
    'connexion' = 'connection'
    'connexions' = 'connections'
    'Connexion' = 'Connection'
    'statuts' = 'statutes'
    'Statuts' = 'Statutes'
    'initialise' = 'initialize'
    'marrage' = 'startup'
    'individuel' = 'individual'
}

# Obtenir tous les fichiers .go dans le répertoire courant et ses sous-répertoires
$files = Get-ChildItem -Path "." -Recurse -Filter "*.go" -File

foreach ($file in $files) {
    Write-Host "Traitement: $($file.FullName)"
    
    # Lire le contenu du fichier
    $content = Get-Content $file.FullName -Encoding UTF8 -Raw
    $original = $content
    
    # Appliquer chaque correction
    foreach ($incorrect in $corrections.Keys) {
        $correct = $corrections[$incorrect]
        $content = $content -replace [regex]::Escape($incorrect), $correct
    }
    
    # Écrire le fichier seulement s'il y a eu des changements
    if ($content -ne $original) {
        Set-Content -Path $file.FullName -Value $content -Encoding UTF8 -NoNewline
        Write-Host "✅ Corrigé: $($file.Name)"
    }
}

Write-Host "🎉 Correction des fautes d'orthographe terminée!" 