# FSM Backend Auto-Start Script

Write-Host "Loading environment configurations from .env..." -ForegroundColor Cyan
if (Test-Path .env) {
    Get-Content .env | ForEach-Object {
        if ($_ -and -not $_.Trim().StartsWith("#") -and $_.Contains("=")) {
            $name, $value = $_.Split('=', 2)
            [System.Environment]::SetEnvironmentVariable($name.Trim(), $value.Trim(), "Process")
        }
    }
    Write-Host "Environment variables loaded successfully." -ForegroundColor Green
} else {
    Write-Host "Warning: .env file not found." -ForegroundColor Yellow
}

Write-Host "Starting Air (Hot Reloading)..." -ForegroundColor Cyan
if (Get-Command air -ErrorAction SilentlyContinue) {
    air
} else {
    $userAirPath = "$env:USERPROFILE\go\bin\air.exe"
    if (Test-Path $userAirPath) {
        & $userAirPath
    } else {
        Write-Host "Error: 'air' is not installed. Please install it first by running: go install github.com/air-verse/air@latest" -ForegroundColor Red
    }
}
