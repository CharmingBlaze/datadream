# Verify libs/raylib/raw.dd matches bindgen output from raylib.h.
# Usage: .\scripts\check-bindgen.ps1

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
Set-Location $Root

if (-not (Test-Path ".\datadream.exe")) {
    Write-Host "Building datadream..."
    go build -o datadream.exe ./cmd/datadream
}

& .\datadream.exe sdk install headers

$Gen = [System.IO.Path]::GetTempFileName()
try {
    Write-Host "Regenerating raw bindings..."
    & .\datadream.exe bind sdk/raylib/6.0/include/raylib.h --raw --out $Gen

    $committed = Get-Content "libs\raylib\raw.dd" -Raw
    $generated = Get-Content $Gen -Raw
    if ($committed -ne $generated) {
        Write-Error @"
libs/raylib/raw.dd is out of date. Regenerate with:
  datadream bind sdk/raylib/6.0/include/raylib.h --raw --out libs/raylib/raw.dd
"@
    }

    & .\datadream.exe check --codegen libs\raylib\raw.dd

    Write-Host "Verifying infer return-type map..."
    $InfGen = [System.IO.Path]::GetTempFileName()
    try {
        go run ./tools/infergen/main.go -raw libs/raylib/raw.dd -out $InfGen
        $infCommitted = Get-Content "internal\infer\raylib_returns_gen.go" -Raw
        $infGenerated = Get-Content $InfGen -Raw
        if ($infCommitted -ne $infGenerated) {
            Write-Error @"
internal/infer/raylib_returns_gen.go is out of date. Regenerate with:
  cd internal/infer; go generate .
"@
        }
    }
    finally {
        Remove-Item -Force $InfGen -ErrorAction SilentlyContinue
    }

    Write-Host "Done: raw.dd matches bindgen output"
}
finally {
    Remove-Item -Force $Gen -ErrorAction SilentlyContinue
}
