# Build a DataDream distribution (maintainers only — end users never need Go).
# Output: dist/datadream-<platform>.zip with bin/, sdk/, examples/, libs/
#
# Usage:
#   .\scripts\build-dist.ps1
#   .\scripts\build-dist.ps1 -SkipClang          # smaller zip (user runs sdk install clang)
#   .\scripts\build-dist.ps1 -SkipVerify       # pack only, no doctor/build smoke test

param(
    [switch]$SkipClang,
    [switch]$SkipVerify,
    [switch]$SkipStudio
)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

$env:CGO_ENABLED = "0"

Write-Host "Building datadream compiler..."
go build -mod=mod -o datadream.exe ./cmd/datadream
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "Fetching offline Monaco editor..."
.\scripts\fetch-monaco.ps1
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "Installing raylib SDK..."
.\datadream.exe sdk install headers
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
.\datadream.exe sdk install raylib
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

if (-not $SkipClang) {
    Write-Host "Installing bundled Clang (large download)..."
    .\datadream.exe sdk install clang
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
} else {
    Write-Host "Skipping Clang install (-SkipClang)"
}

Write-Host "SDK doctor..."
.\datadream.exe doctor
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "Smoke build hello_friendly..."
.\datadream.exe build examples/raylib/hello_friendly.dd -o hello_dist_smoke
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "Smoke build hello_raw..."
.\datadream.exe build examples/raylib/hello_raw.dd -o hello_raw_dist_smoke
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "Smoke build coin-runner..."
Push-Location examples/coin-runner
..\..\datadream.exe build game.dd -o coin_runner_smoke
if ($LASTEXITCODE -ne 0) { Pop-Location; exit $LASTEXITCODE }
Pop-Location

if (-not $SkipStudio) {
    Write-Host "Building DataDream Studio (Wails)..."
    .\scripts\build-studio.ps1
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
} else {
    Write-Host "Skipping Studio build (-SkipStudio)"
}

Write-Host "Building packdist tool..."
go build -o packdist.exe ./tools/packdist
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

$plat = "windows-amd64"
if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { $plat = "windows-arm64" }
$Out = Join-Path $Root "dist\datadream-$plat.zip"
New-Item -ItemType Directory -Force -Path (Split-Path $Out) | Out-Null

Write-Host "Packing distribution..."
$packArgs = @("--out", $Out)
if (-not $SkipVerify) { $packArgs += "--verify" }
if ($SkipStudio) { $packArgs += "--skip-studio" }
& .\packdist.exe @packArgs
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "Done: $Out"
if ($SkipClang) {
    Write-Host "Note: zip omits bundled Clang - users must run: datadream sdk install clang"
}
