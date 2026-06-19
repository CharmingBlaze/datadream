# DataDream SDK — Bundled Toolchain

The SDK ships **Clang**, **raylib 6.0**, headers, and prebuilt libs so users build games without installing compilers or package managers.

**Full docs:** [../docs/README.md](../docs/README.md) · [../docs/SETUP.md](../docs/SETUP.md) · [../docs/DISTRIBUTION.md](../docs/DISTRIBUTION.md)

---

## What end users need

1. Unzip the DataDream distribution (or clone + run install commands).
2. Run `datadream sdk install clang` and `datadream sdk install raylib` (if not pre-populated in zip).
3. Run `datadream doctor` — all lines must show ✓.
4. Write `.dd` files; `datadream run` / `datadream build`.

**No Go install.** The compiler is a standalone native binary.

---

## Layout

```
datadream/                  ← DATADREAM_ROOT
  bin/datadream[.exe]
  sdk/
    manifest.json
    toolchain/clang/        ← datadream sdk install clang
      bin/clang[.exe]
    raylib/6.0/
      include/              ← raylib.h (may ship in repo)
      lib/
        windows-amd64/      ← datadream sdk install raylib
        linux-amd64/
        darwin-arm64/
  examples/
  libs/
```

---

## Install commands

```bash
datadream sdk install clang      # llvm-mingw on Windows; LLVM on Linux/macOS
datadream sdk install headers    # raylib headers only
datadream sdk install raylib     # prebuilt lib for current OS/arch
datadream doctor
```

### Windows Clang

Downloads **llvm-mingw** (pinned in `internal/sdk/version.go`). Required for linking MinGW `libraylib.a`.

System LLVM may work if `doctor` reports `MinGW target for raylib` ✓ — bundled llvm-mingw is still preferred for releases.

### Linux / macOS Clang

Downloads official **LLVM** tarball into the same `sdk/toolchain/clang/` layout. Requires `tar` on PATH for extraction.

---

## Verify

```bash
datadream doctor
```

Expected:

```
✓ SDK ready — build and run .dd programs with no Go install.
```

---

## Override root

```powershell
$env:DATADREAM_ROOT = "C:\path\to\DataDream"   # Windows
```

```bash
export DATADREAM_ROOT=/opt/datadream            # Linux/macOS
```

---

## Release builds (maintainers)

```bash
datadream sdk install clang
datadream sdk install raylib
datadream doctor
datadream build examples/raylib/hello_friendly.dd -o hello
.\scripts\build-dist.ps1
```

The zip contains `bin/`, `sdk/`, `examples/`, `libs/` — **not** Go source or `go.mod`.

**Do not commit** `sdk/toolchain/clang/` to git (~700 MB). Install during release assembly.

See [../docs/HANDOFF.md](../docs/HANDOFF.md) for what the next programmer should finish.
