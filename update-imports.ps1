# PowerShell script to replace javax.* imports with jakarta.*
$basePath = "c:\Users\aqi\Desktop\aurora-master\aurora-springboot\src\main\java"

Write-Host "Starting import replacement..."

# Get all Java files
$javaFiles = Get-ChildItem -Path $basePath -Filter "*.java" -Recurse

foreach ($file in $javaFiles) {
    $content = Get-Content $file.FullName -Raw
    $updated = $false
    
    # Replace common imports
    $patterns = @(
        @{Old = 'import javax.servlet.'; New = 'import jakarta.servlet.'},
        @{Old = 'import javax.validation.'; New = 'import jakarta.validation.'},
        @{Old = 'import javax.mail.'; New = 'import jakarta.mail.'},
        @{Old = 'import javax.annotation.'; New = 'import jakarta.annotation.'},
        @{Old = 'import javax.persistence.'; New = 'import jakarta.persistence.'},
        @{Old = 'import javax.transaction.'; New = 'import jakarta.transaction.'}
    )
    
    foreach ($pattern in $patterns) {
        if ($content -match $pattern.Old) {
            $content = $content -replace $pattern.Old, $pattern.New
            $updated = $true
        }
    }
    
    # Also update fastjson to fastjson2
    if ($content -match 'import com.alibaba.fastjson\.') {
        $content = $content -replace 'import com.alibaba.fastjson\.', 'import com.alibaba.fastjson2.'
        $updated = $true
    }
    
    if ($updated) {
        Set-Content -Path $file.FullName -Value $content -NoNewline
        Write-Host "Updated: $($file.FullName)"
    }
}

Write-Host "Import replacement completed!"
