# clerk uninstaller for Windows
# Usage: irm https://raw.githubusercontent.com/vulcanshen/clerk/main/uninstall.ps1 | iex

$ErrorActionPreference = "Stop"

$installDir = "$env:LOCALAPPDATA\clerk"

if (-not (Test-Path "$installDir\clerk.exe")) {
    Write-Host "clerk not found in $installDir" -ForegroundColor Red
    exit 1
}

Remove-Item $installDir -Recurse -Force
Write-Host "removed $installDir" -ForegroundColor Green

# Remove from PATH
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -like "*$installDir*") {
    $newPath = ($userPath -split ";" | Where-Object { $_ -ne $installDir }) -join ";"
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    Write-Host "removed $installDir from PATH" -ForegroundColor Green
}

# Remove summaries
$saveDir = Join-Path $env:USERPROFILE ".clerk"
if (Test-Path $saveDir) {
    $answer = Read-Host "Remove saved summaries in $saveDir? [y/N]"
    if ($answer -eq "y" -or $answer -eq "Y") {
        Remove-Item $saveDir -Recurse -Force
        Write-Host "removed $saveDir" -ForegroundColor Green
    } else {
        Write-Host "kept $saveDir" -ForegroundColor Yellow
    }
}

# Remove config
$configDir = Join-Path $env:USERPROFILE ".config\clerk"
if (Test-Path $configDir) {
    $answer = Read-Host "Remove config in $configDir? [y/N]"
    if ($answer -eq "y" -or $answer -eq "Y") {
        Remove-Item $configDir -Recurse -Force
        Write-Host "removed $configDir" -ForegroundColor Green
    } else {
        Write-Host "kept $configDir" -ForegroundColor Yellow
    }
}

Write-Host ""
Write-Host "clerk uninstalled. Restart your terminal for PATH changes to take effect." -ForegroundColor Green
