# Build every .dd example under examples/ (smoke test for releases and CI).
# Usage: .\scripts\build-all-examples.ps1

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
Set-Location $Root

if (-not (Test-Path ".\datadream.exe")) {
    Write-Host "Building datadream..."
    go build -o datadream.exe ./cmd/datadream
}

$OutDir = Join-Path $env:TEMP ("datadream-examples-" + [guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $OutDir | Out-Null
try {
    $files = Get-ChildItem -Recurse examples -Filter *.dd | Sort-Object FullName
    foreach ($file in $files) {
        $rel = $file.FullName.Substring($Root.Length + 1)
        $out = Join-Path $OutDir ($file.BaseName)
        Write-Host "build $rel"
        if ($file.Directory.Name -eq "coin-runner") {
            Push-Location $file.DirectoryName
            & ..\..\datadream.exe build $file.Name -o $out
            if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
            Pop-Location
        } else {
            & .\datadream.exe build $rel -o $out
            if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
        }
    }
    Write-Host "Done: built $($files.Count) examples"
}
finally {
    Remove-Item -Recurse -Force $OutDir -ErrorAction SilentlyContinue
}
