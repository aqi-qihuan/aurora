$env:JAVA_HOME = "D:\Java\jdk-17.0.12"
$env:PATH = "$env:JAVA_HOME\bin;$env:PATH"
Write-Host "Using Java version:"
java -version
Write-Host "Starting Spring Boot application..."
mvn spring-boot:run
