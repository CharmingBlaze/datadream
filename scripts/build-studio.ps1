#!/usr/bin/env pwsh
# Build DataDream Studio (Wails) for the current or selected platform.
# Usage:
#   .\scripts\build-studio.ps1
#   .\scripts\build-studio.ps1 -Platform windows/amd64
param(
    [string]$Platform = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

function Ensure-AppIcon {
    param([string]$StudioDir)
    $buildDir = Join-Path $StudioDir "build"
    $windowsDir = Join-Path $buildDir "windows"
    New-Item -ItemType Directory -Force -Path $windowsDir | Out-Null
    $iconPath = Join-Path $buildDir "appicon.png"
    if (Test-Path $iconPath) { return }
    Add-Type -AssemblyName System.Drawing
    $bmp = New-Object System.Drawing.Bitmap 256,256
    $g = [System.Drawing.Graphics]::FromImage($bmp)
    $g.Clear([System.Drawing.Color]::FromArgb(255, 29, 58, 95))
    $accent = New-Object System.Drawing.SolidBrush ([System.Drawing.Color]::FromArgb(255, 61, 184, 201))
    $mark = New-Object System.Drawing.SolidBrush ([System.Drawing.Color]::FromArgb(255, 232, 236, 239))
    $g.FillRectangle($accent, 0, 0, 32, 256)
    $g.FillRectangle($mark, 48, 64, 160, 16)
    $g.FillRectangle($mark, 48, 96, 120, 16)
    $g.FillRectangle($mark, 48, 128, 140, 16)
    $bmp.Save($iconPath, [System.Drawing.Imaging.ImageFormat]::Png)
    $g.Dispose(); $bmp.Dispose()
}

$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

if (-not $Platform) {
    $arch = if ($env:PROCESSOR_ARCHITECTURE -eq 'ARM64') { 'arm64' } else { 'amd64' }
    $Platform = "windows/$arch"
}

Write-Host "Building DataDream Studio (Wails) for $Platform..." -ForegroundColor Cyan

$env:GOFLAGS = "-mod=mod"
$env:Path = "$env:USERPROFILE\go\bin;$env:Path"

if (-not (Get-Command wails -ErrorAction SilentlyContinue)) {
    Write-Host "Installing Wails CLI..." -ForegroundColor Yellow
    go install github.com/wailsapp/wails/v2/cmd/wails@latest
}

$studio = Join-Path $root "cmd\studio"
Ensure-AppIcon -StudioDir $studio

Push-Location $studio
try {
    wails build -clean -platform $Platform -webview2 embed
    $built = Join-Path $studio "build\bin\datadream-studio.exe"
    if ($Platform -notmatch '^windows/') {
        $built = Join-Path $studio "build\bin\datadream-studio"
        if ($Platform -match '^darwin/') {
            $built = Join-Path $studio "build\bin\datadream-studio.app"
        }
    }
    if (-not (Test-Path $built)) {
        throw "Build failed: expected output not found ($built)"
    }
    Write-Host ""
    Write-Host "Built: $built" -ForegroundColor Green
} finally {
    Pop-Location
}
