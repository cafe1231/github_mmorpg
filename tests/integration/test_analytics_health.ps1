# test_analytics_health.ps1
$url = "http://localhost:8088/health"
Write-Host "Test: $url"

try {
    $response = Invoke-WebRequest -Uri $url -UseBasicParsing -ErrorAction Stop
    if ($response.StatusCode -eq 200 -and $response.Content -match 'status') {
        Write-Host "[OK] /analytics/health"
        exit 0
    } else {
        Write-Host "[FAIL] Code HTTP: $($response.StatusCode)"
        exit 1
    }
} catch {
    Write-Host "[FAIL] $($_.Exception.Message)"
    exit 1
} 