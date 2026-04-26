Set-Location "C:\Users\aqi\Desktop\aurora-master\aurora-go"
Write-Output "Current dir: $(Get-Location)"
Write-Output "go.mod exists: $(Test-Path go.mod)"
$output = go build ./... 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Output "BUILD SUCCESS - No errors!"
} else {
    $output | Select-Object -First 100 | ForEach-Object { Write-Output $_ }
}
