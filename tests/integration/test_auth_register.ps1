# test_auth_register.ps1
$rand = Get-Random -Maximum 99999
$email = "testuser$rand@example.com"
$username = "testuser$rand"
$url = "http://localhost:8081/api/v1/auth/register"
$body = @{ email = $email; password = "Test1234!"; username = $username } | ConvertTo-Json
Write-Host "Test: $url"

try {
    $response = Invoke-RestMethod -Uri $url -Method Post -Body $body -ContentType "application/json" -ErrorAction Stop -SkipHttpErrorCheck
    $code = $LASTEXITCODE
    $json = $response | ConvertTo-Json -Compress
    if ($null -eq $response -or $json -notmatch 'user|id|success|token') {
        Write-Host "[FAIL] Body: $json"
        exit 1
    }
    Write-Host "[OK] /auth/register"
    exit 0
} catch {
    if ($_.Exception.Response -and $_.Exception.Response.StatusCode.Value__) {
        $code = $_.Exception.Response.StatusCode.Value__
        Write-Host "[FAIL] Code HTTP: $code"
    } else {
        Write-Host "[FAIL] $($_.Exception.Message)"
    }
    exit 1
} 