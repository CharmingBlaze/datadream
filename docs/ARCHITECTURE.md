# Compiler Architecture

How DataDream turns `.dd` source into a native binary.

---

## Pipeline overview

```
┌──────────────┐
│  source.dd   │
└──────┬───────┘
       │  compiler.ReadSource (include expansion)
       ▼
┌──────────────┐
│    lexer     │  internal/lexer/
└──────┬───────┘
       ▼
┌──────────────┐
│    parser    │  internal/parser/
└──────┬───────┘
       ▼
┌──────────────┐
│     AST      │  internal/ast/
└──────┬───────┘
       │  optional Phase[] (typecheck — not yet)
       ▼
┌──────────────┐
│   codegen    │  internal/codegen/
└──────┬───────┘
       ▼
┌──────────────┐
│   driver     │  internal/driver/  → bundled or system Clang
└──────┬───────┘
       ▼
   native binary
```

Orchestration: `internal/compiler/pipeline.go`

**Check modes:**

| Command | Stages |
|---------|--------|
| `datadream check file.dd` | lex + parse (+ optional Phases) |
| `datadream check --codegen file.dd` | above + codegen (no link) |
| `datadream build file.dd` | full pipeline + driver |

---

## Package responsibilities

### `internal/lexer`

- Tokenizes `.dd` source
- Special tokens: `TOKEN_HEX_COLOR`, lifecycle keywords
- File: `lexer.go`

### `internal/parser`

| File | Parses |
|------|--------|
| `parser.go` | Top-level dispatch, `app` header |
| `decls.go` | `fn`, `struct`, `entity`, `scene`, `window`, `enum` |
| `stmt.go` | `let`, `if`, `for`, `while`, **`loop`**, **`match`**, **`defer`**, assign |
| `expr.go` | Expressions, calls, `draw.*` namespaces |
| `colors.go` | Fold `rgb()`, `colors.*` at parse time |
| `interop.go` | `use`, `using`, `extern c` |
| `config.go` | `window { }`, lifecycle blocks, object literals |

**Binding parse:** C struct fields and params may use names that match keywords (`shader`, `data`). `lexer.IsBindingName()` + `parseBindingIdent()` handle these. Array field types (`float[]`) and C float suffix (`180.0f`) are supported in extern blocks.

### `internal/codegen`

| File | Emits |
|------|-------|
| `generator.go` | `Generate()`, top-level dispatch, deferred global init |
| `analyze.go` | Usage analysis → conditional runtime emission |
| `defer.go` | `defer` stack, `genStmts()` scoped cleanup |
| `friendly.go` | `screen.*`, mouse, keys, math builtins |
| `stdlib.go` / `stdlib_runtime.go` | `time.*`, `audio.*`, `assets.*` |
| `app.go` | Program analysis, auto-raylib, link libs |
| `app_emit.go` | raylib game loop `main()`, calls `datadream_init_globals()` |
| `runtime.go` | C header, types, color runtime |
| `game_runtime.go` | Sprite, input, collision helpers |
| `draw.go` | `draw.text/rect/circle/line/sprite` → raylib |
| `expr.go` | Expressions, builtins, namespaces |
| `stmts.go` | Statements, vec2 `+=`, loop/match/defer, const vs runtime global `let` |
| `decls.go` | Structs, entities, scenes |
| `interop.go` | `use raylib`, extern calls |
| `types.go` | DataDream → C type mapping |

### `internal/compiler`

```go
compiler.Compile(opts)   // full pipeline → C + link flags
compiler.Check(opts)     // lex + parse; opts.Codegen = true for codegen
compiler.ReadSource(path) // include expansion
```

Extension point:

```go
type Phase interface {
    Run(prog *ast.Program) ([]Diagnostic, bool)
}
```

Insert between parse and codegen via `Pipeline.Phases`.

### `internal/driver`

- Writes C to temp file
- Invokes Clang: `sdk.CompileFlags()` + `sdk.ToolchainFlags()` + codegen link flags
- Skips `-lm` on Windows; `-Wno-unused-function` for generated helpers

### `internal/sdk`

| File | Role |
|------|------|
| `sdk.go` | `Root()`, `ClangPath()`, `Doctor()`, raylib paths |
| `fetch_clang.go` | `InstallClang()` — llvm-mingw / LLVM download |
| `fetch_raylib.go` | `InstallRaylib()`, `InstallRaylibHeaders()` |
| `link.go` | `RaylibLinkLibs()` — platform link line |
| `toolchain.go` | MinGW vs MSVC detection, `-target x86_64-w64-mingw32` |
| `version.go` | `RaylibVersion`, `LLVMMingwVersion`, `LLVMOrgVersion` |

**Clang resolution order** (`ClangPath()`):

1. Bundled `sdk/toolchain/clang/bin/clang[.exe]` (from manifest)
2. `clang` / `clang-cl` / `gcc` on PATH
3. Fallback string `"clang"`

### `internal/colors`

Pure Go: `ParseHex`, CSS table, HSL, `colors.*` namespace. Used by parser and codegen.

### `tools/bindgen`

C header → `extern c { }` or raw `.dd`. Invoked via `datadream bind`.

---

## Two compilation modes

### App mode

When: `app` + `window` + `draw` (+ optional `start` / `update`)

- `analyzeProgram()` / `finalizeAnalysis()` enable raylib without explicit `use`
- `emitAppMain()` generates game loop
- Lifecycle: `lifecycle_start`, `lifecycle_update`, `lifecycle_draw`
- Non-const global `let` (e.g. `sprite(...)`) → `datadream_init_globals()` before `lifecycle_start`

### Script mode

`fn main()` or no app/window — emits `user_main` or simple `main`.

---

## Link flag flow

```
codegen → generator.linkLibs += sdk.RaylibLinkLibs()
pipeline → result.LinkFlags += sdk.CompileFlags() + pkg includes
driver   → clang file.c -o out [ToolchainFlags] [flags...]
```

Windows: full path to `libraylib.a`; MinGW Clang required (or `-target` fallback).

---

## Include system

`include "path.dd";` is **textual** preprocessing in `compiler/source.go` — not a module system. Future: exports via `use`.

---

## Testing strategy

| Layer | Command |
|-------|---------|
| Colors | `go test ./internal/colors/...` |
| Codegen | `go test ./internal/codegen/...` |
| Parser | `go test ./internal/parser/...` (slow on raw.dd) |
| E2E | `datadream build examples/...` |
| Bindgen | `datadream bind ...` + `datadream check --codegen` |

Prefer codegen unit tests before full builds.

---

## Files to read first

1. `examples/raylib/hello_friendly.dd` — target UX
2. `examples/coin-runner/game.dd` — game loop + sprites
3. `internal/codegen/app_emit.go` — generated main
4. `internal/codegen/game_runtime.go` — sprite/input runtime
5. `internal/compiler/pipeline.go` — wiring
6. `internal/sdk/fetch_clang.go` — SDK install
