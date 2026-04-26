$results = Get-ChildItem -Path "C:\Users\aqi\Desktop\aurora-master" -Filter "go.mod" -Recurse -ErrorAction SilentlyContinue
if ($results) {
    foreach ($r in $results) { Write-Output $r.FullName }
} else {
    Write-Output "No go.mod found!"
}

Write-Output "--- Directory listing of aurora-master ---"
Get-ChildItem "C:\Users\aqi\Desktop\aurora-master" | Select-Object Name, PSIsContainer
