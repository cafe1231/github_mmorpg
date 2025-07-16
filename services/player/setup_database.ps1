# Script PowerShell pour configurer la base de données player_db
# Assure-toi que PostgreSQL est installé et que psql est dans le PATH

Write-Host "🔧 Configuration de la base de données player_db..." -ForegroundColor Cyan

# Vérifier si psql est disponible
try {
    $psqlVersion = psql --version
    Write-Host "✅ PostgreSQL trouvé: $psqlVersion" -ForegroundColor Green
}
catch {
    Write-Host "❌ PostgreSQL (psql) n'est pas trouvé dans le PATH" -ForegroundColor Red
    Write-Host "💡 Assure-toi que PostgreSQL est installé et psql est dans le PATH" -ForegroundColor Yellow
    exit 1
}

# Variables de configuration
$POSTGRES_USER = "postgres"
$DB_NAME = "player_db"
$AUTH_USER = "auth_user"
$AUTH_PASSWORD = "auth_password"

Write-Host "📝 Paramètres:" -ForegroundColor Yellow
Write-Host "   - Base de données: $DB_NAME"
Write-Host "   - Utilisateur: $AUTH_USER"
Write-Host "   - Mot de passe: $AUTH_PASSWORD"

# Demander le mot de passe postgres
Write-Host ""
Write-Host "🔑 Veuillez entrer le mot de passe pour l'utilisateur 'postgres':" -ForegroundColor Yellow

# Créer l'utilisateur auth_user s'il n'existe pas
Write-Host "👤 Création de l'utilisateur $AUTH_USER..." -ForegroundColor Cyan
$createUserSQL = @"
DO `$`$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = '$AUTH_USER') THEN
        CREATE USER $AUTH_USER WITH PASSWORD '$AUTH_PASSWORD';
        GRANT CREATEDB TO $AUTH_USER;
    END IF;
END
`$`$;
"@

psql -U $POSTGRES_USER -d postgres -c $createUserSQL

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Utilisateur $AUTH_USER créé/vérifié" -ForegroundColor Green
} else {
    Write-Host "❌ Erreur lors de la création de l'utilisateur" -ForegroundColor Red
    exit 1
}

# Créer la base de données
Write-Host "🗄️ Création de la base de données $DB_NAME..." -ForegroundColor Cyan
$createDbSQL = "CREATE DATABASE $DB_NAME OWNER $AUTH_USER;"
psql -U $POSTGRES_USER -d postgres -c $createDbSQL 2>$null

# Configurer les permissions
Write-Host "🔐 Configuration des permissions..." -ForegroundColor Cyan
$permissionsSQL = @"
-- Donner tous les privilèges sur le schéma public
GRANT ALL PRIVILEGES ON SCHEMA public TO $AUTH_USER;
GRANT CREATE ON SCHEMA public TO $AUTH_USER;

-- Privilèges sur les objets existants
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO $AUTH_USER;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO $AUTH_USER;
GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO $AUTH_USER;

-- Privilèges par défaut pour les futurs objets
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO $AUTH_USER;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO $AUTH_USER;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON FUNCTIONS TO $AUTH_USER;
"@

psql -U $POSTGRES_USER -d $DB_NAME -c $permissionsSQL

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Permissions configurées avec succès" -ForegroundColor Green
} else {
    Write-Host "❌ Erreur lors de la configuration des permissions" -ForegroundColor Red
    exit 1
}

# Tester la connexion avec auth_user
Write-Host "🧪 Test de connexion avec $AUTH_USER..." -ForegroundColor Cyan
$env:PGPASSWORD = $AUTH_PASSWORD
$testSQL = "SELECT version();"
$result = psql -U $AUTH_USER -d $DB_NAME -c $testSQL 2>$null

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Connexion réussie avec $AUTH_USER" -ForegroundColor Green
} else {
    Write-Host "❌ Erreur de connexion avec $AUTH_USER" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "🎉 Configuration terminée avec succès!" -ForegroundColor Green
Write-Host "📋 Résumé:"
Write-Host "   ✅ Utilisateur $AUTH_USER créé"
Write-Host "   ✅ Base de données $DB_NAME créée"
Write-Host "   ✅ Permissions configurées"
Write-Host "   ✅ Connexion testée"
Write-Host ""
Write-Host "🚀 Tu peux maintenant démarrer le service Player:" -ForegroundColor Cyan
Write-Host "   cd services/player"
Write-Host "   go run ./cmd/main.go" 