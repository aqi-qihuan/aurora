$files = Get-ChildItem -Path "C:\Users\aqi\Desktop\aurora-master\aurora-go" -Recurse -ErrorAction SilentlyContinue
Write-Output "Total items: $($files.Count)"
foreach ($f in $files) {
    Write-Output $f.FullName
}
