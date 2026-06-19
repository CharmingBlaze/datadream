# Developer Setup

How to build and hack on the **DataDream compiler**. End users who only run `.dd` programs do **not** need this — see [DISTRIBUTION.md](DISTRIBUTION.md).

---

## Requirements

| Role | Needs |
|------|-------|
| **End user** | `datadream` binary + `sdk/` folder only |
| **Contributor** | Go 1.22+, Git, C compiler for testing builds |

**Do not install Python** for this project. All tooling is Go.

---

## Clone and build compiler

```bash
git clone <repo-url> DataDream
cd DataDream

go build -o datadream ./cmd/datadream
# Windows: datadream.exe
```

Verify:

```bash
./datadream version
# DataDream compiler version 0.1.0
```

Entry point: `cmd/datadream/main.go` → `internal/cli.Run`. The `.gitignore` pattern must be `/datadream` (root only), not bare `datadream`, or this directory is ignored.

---

## SDK setup

The SDK provides Clang and raylib 6.0 so builds are reproducible.

### Layout

```
sdk/
  manifest.json
  toolchain/clang/          ← put portable Clang here
  raylib/6.0/
    include/                ← raylib.h (may be pre-populated)
    lib/windows-amd64/      ← libraylib.a (from sdk install)
```

### Populate raylib

```bash
datadream sdk install headers
datadream sdk install raylib
```

Downloads official raylib **6.0** prebuilts from GitHub into `sdk/raylib/6.0/`.

### Populate Clang

```bash
datadream sdk install clang
```

Downloads **llvm-mingw** on Windows (MinGW, raylib-compatible) or official **LLVM** on Linux/macOS into `sdk/toolchain/clang/`.

Manual alternative: extract [llvm-mingw](https://github.com/mstorsjo/llvm-mingw/releases) so `sdk/toolchain/clang/bin/clang.exe` exists. System LLVM may work if `datadream doctor` reports `MinGW target for raylib` ✓.

### Verify

```bash
datadream doctor
```

Expected when healthy:

```
✓ SDK ready — build and run .dd programs with no Go install.
```

Set `DATADREAM_ROOT` if auto-detection fails:

```powershell
# Windows
$env:DATADREAM_ROOT = "C:\path\to\DataDream"

# Linux/macOS
export DATADREAM_ROOT=/path/to/DataDream
```

---

## Windows linking note

If you see:

```
unresolved external symbol __mingw_sscanf
cannot open file 'raylib.lib'
```

You are using the **wrong Clang**. Run `datadream sdk install clang` (llvm-mingw). See [HANDOFF.md](HANDOFF.md).

---

## Daily workflow

```bash
# 1. Edit language / codegen
# 2. Rebuild compiler
go build -o datadream ./cmd/datadream

# 3. Fast check (parse + typecheck)
datadream check examples/coin-runner/game.dd
datadream check examples/coin-runner/game.dd --codegen

# 4. Full compile (catches link errors)
datadream build examples/raylib/hello_friendly.dd -o hello

# 5. Run
datadream run examples/raylib/hello_friendly.dd

# 6. Unit tests
go test ./internal/...
```

---

## Project entry points

| Task | Command / file |
|------|----------------|
| CLI entry | `cmd/datadream/main.go` |
| Add CLI command | `internal/cli/cli.go` + new `cmd_*.go` |
| Add typecheck rule | `internal/typecheck/typecheck.go`, `builtins.go`, `hints.go` |
| Formatted errors | `internal/errors/`, `internal/cli/diagnostics.go` |
| Add keyword | `internal/lexer/lexer.go` → parser → ast → codegen |
| Add friendly builtin | `internal/codegen/expr.go` or `draw.go` |
| Add C runtime helper | `internal/codegen/game_runtime.go` or `runtime.go` |
| Change link flags | `internal/sdk/link.go` |
| Add SDK install target | `internal/sdk/fetch_*.go` + `internal/cli/sdk.go` |

---

## IDE / editor

- **VS Code / Cursor:** Open repo root. Go extension for compiler work.
- **`.dd` files:** No official LSP yet. Use `datadream check` in terminal.

---

## Build release zip (maintainers)

```powershell
.\scripts\build-dist.ps1              # full zip + packdist --verify (doctor, hello, coin-runner)
.\scripts\build-dist.ps1 -SkipVerify  # pack only
```

Produces `dist/datadream-<platform>.zip` with `bin/`, `sdk/`, `examples/`, `libs/` — **no Go source, no go.mod**.

Manual:

```bash
go build -o datadream ./cmd/datadream
go build -o packdist ./tools/packdist
./packdist --out dist/datadream.zip --verify
```

---

## Troubleshooting

| Problem | Solution |
|---------|----------|
| `doctor` finds no root | Set `DATADREAM_ROOT` to repo/distribution root |
| Headers not found | `datadream sdk install headers` |
| Link fails Windows | `datadream sdk install clang` then `sdk install raylib`; verify doctor toolchain ✓ |
| `check` passes, `build` fails | Use `check --codegen`; on Windows verify doctor toolchain line |
| Import cycle in Go | Never import `sdk` from `pkg` |
| Slow parser test | `TestParseRaylibRawBindings` parses full `raw.dd` — should complete in under 1 second; if it hangs, check `parseBindingIdent` / `float[]` handling |

---

## What not to set up

- Python virtualenvs
- Node.js / npm (unless you add a separate frontend later)
- CMake for raylib (use prebuilts via `sdk install`)
- Go for end-user distribution packages
