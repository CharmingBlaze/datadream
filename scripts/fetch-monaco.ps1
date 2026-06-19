#!/usr/bin/env pwsh
# Download Monaco editor into internal/ide/web/vendor for offline IDE use.
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$root = Split-Path -Parent $PSScriptRoot
$dest = Join-Path $root "internal\ide\web\vendor\monaco\min"
$marker = Join-Path $dest "vs\loader.js"

if (Test-Path $marker) {
    Write-Host "Monaco already present: $marker"
    exit 0
}

Write-Host "Fetching monaco-editor@0.45.0..."
$tmp = Join-Path $env:TEMP "datadream-monaco-$(Get-Random)"
New-Item -ItemType Directory -Force -Path $tmp | Out-Null

try {
    Push-Location $tmp
    npm pack monaco-editor@0.45.0 2>$null | Out-Null
    $tgz = Get-ChildItem -Filter "monaco-editor-*.tgz" | Select-Object -First 1
    if (-not $tgz) { throw "npm pack failed" }
    tar -xf $tgz.FullName
    New-Item -ItemType Directory -Force -Path (Split-Path $dest) | Out-Null
    if (Test-Path $dest) { Remove-Item -Recurse -Force $dest }
    Copy-Item -Recurse "package/min" (Split-Path $dest)
    if (-not (Test-Path $marker)) { throw "Monaco install incomplete" }
    Write-Host "Installed Monaco to $dest"
} finally {
    Pop-Location
    Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}
