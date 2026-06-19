# Verify a release zip (Windows).
# Usage: .\scripts\verify-dist.ps1 dist\datadream-windows-amd64.zip

param(
    [Parameter(Mandatory = $true)]
    [string]$Zip
)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

if (-not (Test-Path $Zip)) {
    Write-Error "zip not found: $Zip"
}

if (-not (Test-Path packdist.exe)) {
    Write-Host "Building packdist..."
    go build -o packdist.exe ./tools/packdist
}

$Tmp = Join-Path $env:TEMP ("datadream-verify-" + [guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $Tmp | Out-Null
try {
    Write-Host "Extracting $Zip..."
    Expand-Archive -Path $Zip -DestinationPath $Tmp -Force

    Write-Host "Verifying unpacked tree..."
    & .\packdist.exe --verify-only $Tmp
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    Write-Host "Done: $Zip"
} finally {
    Remove-Item -Recurse -Force $Tmp -ErrorAction SilentlyContinue
}
