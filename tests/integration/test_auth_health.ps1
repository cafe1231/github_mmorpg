$url = "http://localhost:8081/health"
Write-Host "Test: $url"
$response = Invoke-WebRequest -Uri $url -UseBasicParsing -ErrorAction Stop
if ($response.StatusCode -ne 200) {
    Write-Host "[FAIL] Code HTTP: $($response.StatusCode)"
    exit 1
}
if ($response.Content -match 'status') {
    Write-Host "[OK] /auth-new/health"
} else {
    Write-Host "[FAIL] Body: $($response.Content)"
    exit 1
} 