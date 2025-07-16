# run_all.ps1
Set-Location $PSScriptRoot

$tests = Get-ChildItem -Filter 'test_*_health.ps1' | Sort-Object Name
foreach ($test in $tests) {
    Write-Host "--- $($test.Name) ---"
    & pwsh -File $test.FullName
    if ($LASTEXITCODE -ne 0) {
        Write-Host "[FAIL] $($test.Name)"
        exit 1
    }
    Write-Host ""
    Start-Sleep -Seconds 1
}
Write-Host "[OK] Tous les tests de santé PowerShell sont passés." 