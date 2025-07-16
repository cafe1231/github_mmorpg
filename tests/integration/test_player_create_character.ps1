# test_player_create_character.ps1
$rand = Get-Random -Maximum 99999
$email = "testuser$rand@example.com"
$user = "testuser$rand"
$pass = "Test1234!"
$urlReg = "http://localhost:8081/api/v1/auth/register"
$urlLogin = "http://localhost:8081/api/v1/auth/login"
$urlCreate = "http://localhost:8082/api/v1/player/characters"
$bodyReg = @{ email = $email; password = $pass; username = $user } | ConvertTo-Json
$bodyLogin = @{ email = $email; password = $pass } | ConvertTo-Json
Write-Host "Inscription: $urlReg"
try {
    Invoke-RestMethod -Uri $urlReg -Method Post -Body $bodyReg -ContentType "application/json" -ErrorAction Stop | Out-Null
    Start-Sleep -Seconds 1
    Write-Host "Login: $urlLogin"
    $resp = Invoke-RestMethod -Uri $urlLogin -Method Post -Body $bodyLogin -ContentType "application/json" -ErrorAction Stop
    $json = $resp | ConvertTo-Json -Compress
    $token = $resp.token
    if (-not $token) { Write-Host "[FAIL] Pas de token"; exit 1 }
    Write-Host "Création personnage: $urlCreate"
    $charBody = @{ name = "Hero$rand"; class = "warrior"; race = "human" } | ConvertTo-Json
    $resp2 = Invoke-RestMethod -Uri $urlCreate -Method Post -Body $charBody -ContentType "application/json" -Headers @{ Authorization = "Bearer $token" } -ErrorAction Stop
    $json2 = $resp2 | ConvertTo-Json -Compress
    if ($json2 -match 'character|id|name') {
        Write-Host "[OK] /player/characters (create)"
        exit 0
    } else {
        Write-Host "[FAIL] Body: $json2"
        exit 1
    }
} catch {
    if ($_.Exception.Response -and $_.Exception.Response.StatusCode.Value__) {
        $code = $_.Exception.Response.StatusCode.Value__
        Write-Host "[FAIL] Code HTTP: $code"
    } else {
        Write-Host "[FAIL] $($_.Exception.Message)"
    }
    exit 1
} 