# Distribution

How DataDream is packaged for **end users** who never install Go.

---

## Principles

1. **End users get a native `datadream` binary** — not Go source, not `go.mod`.
2. **Bundled SDK** — Clang + raylib 6.0 in `sdk/`.
3. **`.dd` programs compile to C** via bundled Clang — not `go run`.
4. **Maintainers use Go** to build the compiler and pack releases.

---

## Release layout

```
datadream/                      ← DATADREAM_ROOT
  bin/
    datadream.exe               ← Windows CLI
    datadream                   ← Linux/macOS CLI
    datadream-studio.exe        ← Windows IDE (Wails)
    datadream-studio            ← Linux IDE (raw binary, dev only)
    datadream-studio-x86_64.AppImage  ← Linux IDE (release — self-contained)
    datadream-studio.app/       ← macOS IDE bundle
  GETTING_STARTED.txt           ← one-page quick start (also at root)
  Start DataDream Studio.bat    ← Windows double-click launcher
  DataDream Studio.exe          ← Windows IDE (root copy)
  start-studio.sh               ← Linux launcher (runs AppImage)
  datadream-studio-x86_64.AppImage  ← Linux IDE (root copy)
  DataDream Studio.app/         ← macOS launcher copy (optional)
  sdk/
    manifest.json
    toolchain/clang/            ← populated by sdk install clang (~700 MB)
      bin/clang[.exe]
    raylib/6.0/
      include/
      lib/<platform>/
  examples/
  libs/
    raylib/raw.dd
    raylib/package.json
  docs/                         ← optional in zip
  README.md
```

**Not included:** `go.mod`, `internal/`, `cmd/`, Python, Go toolchain.

**Do not commit** `sdk/toolchain/clang/` or platform-specific `sdk/raylib/6.0/lib/` to git — install at release build time.

---

## Building a release (maintainers)

### 1. Build compiler

```bash
go build -o datadream ./cmd/datadream
# Windows: datadream.exe → bin/
```

### 2. Populate SDK (on each target platform)

```bash
datadream sdk install clang
datadream sdk install headers    # if headers not bundled
datadream sdk install raylib
datadream doctor                 # all ✓
```

### 3. Smoke test

```bash
datadream build examples/raylib/hello_friendly.dd -o hello
datadream build examples/coin-runner/game.dd -o coin-runner
```

### 4. Pack zip

```powershell
.\scripts\build-dist.ps1
.\scripts\verify-dist.ps1 dist\datadream-windows-amd64.zip
```

This builds **DataDream Studio** (`datadream-studio`) via Wails and packs it into `bin/` alongside the CLI.

To omit the IDE from a zip (smaller artifact):

```powershell
.\scripts\build-dist.ps1 -SkipStudio
# or: ./packdist --out dist/....zip --skip-studio
```

Or on Linux/macOS:

```bash
./scripts/build-dist.sh
./scripts/verify-dist.sh dist/datadream-linux-amd64.zip
```

Or manually:

```bash
./packdist --out dist/datadream.zip --verify
```

`--verify` runs `doctor`, builds `hello_friendly`, `hello_raw`, and `coin-runner` inside the packed tree (sets `DATADREAM_ROOT`).

`tools/packdist` copies: `bin/`, `sdk/`, `examples/`, `libs/`, README.

---

## Before shipping checklist

- [ ] `datadream` built with `CGO_ENABLED=0` (static Go binary)
- [ ] `datadream sdk install clang` on target platform
- [ ] `datadream sdk install raylib` on target platform
- [ ] `datadream doctor` → all ✓ including toolchain line
- [ ] `datadream build examples/raylib/hello_friendly.dd` succeeds
- [ ] `datadream build examples/raylib/hello_raw.dd` succeeds
- [ ] `datadream build examples/coin-runner/game.dd` succeeds
- [ ] Zip excludes `.git`, `internal/`, `cmd/`, test artifacts

---

## End-user install

1. Download / unzip `datadream-<platform>.zip`
2. Read **`GETTING_STARTED.txt`** in the folder
3. **Windows:** double-click **`Start DataDream Studio.bat`** or **`DataDream Studio.exe`**
4. **Linux:** `./start-studio.sh`
5. **macOS:** double-click **`DataDream Studio.app`**

No Go, PATH setup, or separate SDK install required. The zip includes Clang, raylib, the compiler, and the IDE. Works offline.

Optional CLI: add `bin/` to PATH and run `datadream doctor`.
5. Launch the IDE: `datadream studio` or run `bin/datadream-studio` directly

To validate a zip before shipping:

```bash
./scripts/verify-dist.sh dist/datadream-linux-amd64.zip
# or: ./packdist --verify-only /path/to/unzipped-dist
```

No Go. No pip. No vcpkg. No separate LLVM install if zip includes populated `sdk/`.

---

## Environment variables

| Variable | Purpose |
|----------|---------|
| `DATADREAM_ROOT` | Override auto-detected distribution root |

Auto-detection (`internal/sdk/sdk.go`):

1. `DATADREAM_ROOT` env
2. Parent of `bin/datadream` executable
3. Walk up from cwd looking for `sdk/manifest.json`

---

## SDK manifest

`sdk/manifest.json` — layout per platform:

```json
{
  "raylib": "6.0",
  "platforms": {
    "windows-amd64": {
      "raylibLib": "sdk/raylib/6.0/lib/windows-amd64",
      "clangBin": "sdk/toolchain/clang/bin/clang.exe"
    }
  }
}
```

---

## Populating SDK components

| Component | Command | Notes |
|-----------|---------|-------|
| Clang | `datadream sdk install clang` | llvm-mingw on Windows; LLVM tarball on Linux/macOS |
| raylib headers | `datadream sdk install headers` | Or ship in repo under `sdk/raylib/6.0/include/` |
| raylib libs | `datadream sdk install raylib` | Per-platform under `lib/<platform>/` |

Pinned versions: `internal/sdk/version.go` (`LLVMMingwVersion`, `LLVMOrgVersion`, `RaylibVersion`).

---

## Platform matrix

| Platform | Clang source | raylib lib | Status |
|----------|--------------|------------|--------|
| windows-amd64 | llvm-mingw via `sdk install clang` | libraylib.a | ✅ CI dist-verify + **local cold-install verified** (~188 MB zip) |
| linux-amd64 | system Clang or `sdk install clang` | libraylib.a | ✅ CI dist-verify |
| darwin-arm64 | system Clang or `sdk install clang` | libraylib.a | ✅ CI dist-verify |
| darwin-amd64 | LLVM via `sdk install clang` | libraylib.a | 🟡 untested in CI |

---

## CI (`.github/workflows/ci.yml`)

On every push/PR:

- `go test ./internal/...` and `./tools/packdist/...`
- Linux + macOS: all examples `check --codegen` and `build` (`scripts/build-all-examples.sh`)
- `bindgen-check`: `scripts/check-bindgen.sh` keeps `libs/raylib/raw.dd` in sync
- **dist-verify** (Windows/Linux/macOS): `build-dist.*` + `verify-dist.*` — doctor, `hello_friendly`, `hello_raw`, and `coin-runner` in the packed tree

Release zips: `.github/workflows/release.yml` (`workflow_dispatch` or push tag `v*`, verify enabled by default). The `publish` job uploads all three platform zips to a GitHub Release.

**To ship v1.0:** see [PUBLISH.md](PUBLISH.md).

---

## What users compile

```
hello.dd  →  [datadream compiler]  →  hello.c  →  [bundled clang]  →  hello.exe
```

The Go runtime is **only inside `datadream`** — not in user programs.

---

## Anti-patterns

| Don't | Do instead |
|-------|------------|
| Tell users to `go install` | Ship prebuilt binary + SDK |
| Ship only source | Ship bin + sdk |
| Rely on system raylib | Bundle `sdk/raylib/6.0` |
| Commit 700 MB toolchain to git | `sdk install clang` at release time |
| Use MSVC LLVM alone on Windows | Bundled llvm-mingw or doctor-verified MinGW target |
