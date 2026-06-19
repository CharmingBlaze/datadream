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
       │  Phase[] (typecheck — default in check)
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
| `datadream check file.dd` | lex + parse + typecheck |
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
| `expr.go` | Expressions, builtins, namespaces, **method calls** |
| `stmts.go` | Statements, vec2 `+=`, loop/match/defer, const vs runtime global `let` |
| `array.go` | **`DD_Array` runtime**, `for x in array`, `for ch in string`, array methods |
| `arena.go` | Frame arena (`dd_frame_arena_reset`) + level arena on scene init |
| `attrs.go` | `@packed` / `@save` attribute codegen stubs |
| `decls.go` | Structs, entities, scenes, **methods** |
| `interop.go` | `use raylib`, extern calls |
| `types.go` | DataDream → C type mapping, **methodParamsC** |

### `internal/compiler`

```go
compiler.Compile(opts)   // full pipeline → C + link flags
compiler.Check(opts)     // lex + parse + typecheck; opts.Codegen = true for codegen
compiler.ReadSource(path) // include expansion
```

Default pipeline includes `typecheckPhase` in `Pipeline.Phases`.

### `internal/typecheck`

- Runs as a compiler `Phase` before codegen
- Validates builtins, namespace methods, struct literals, entity fields
- **`forin.go`** — `resolveForIn()` sets `ForInStmt.Kind` (`IterEntity`, `IterArray`, `IterString`); warns on `.remove()` during array iteration
- **`Error.Warning`** — non-blocking diagnostics; pipeline continues to codegen
- `hints.go` — suggested fixes attached to diagnostics

### `internal/errors`

- `Reporter` — Rust/Elm-style formatted output (snippet, caret, hint)
- Supports **errors** and **warnings** (`WarningHint`, yellow prefix)
- Used by `datadream check` and compile error paths via `internal/cli/diagnostics.go`

### `cmd/datadream`

- Thin `main` → `internal/cli.Run(os.Args[1:])`
- Must stay tracked in git (`.gitignore` uses `/datadream`, not `datadream`)

### Extension point

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

## `for x in …` disambiguation

The parser emits one node type for all collection iteration:

```go
type ForInStmt struct {
    Value    string   // loop binding
    Iter     Node     // expression after "in"
    Kind     IterKind // filled by type checker
    ElemType string   // for IterArray
    Entity   string   // for IterEntity
}
```

| User syntax | `IterKind` | Codegen |
|-------------|------------|---------|
| `for e in Enemy` | `IterEntity` | ECS registry loop (`ecs.go`) |
| `for x in arr` / `[1,2,3]` | `IterArray` | `DD_Array` index loop (`array.go`) |
| `for ch in "hi"` / `str` | `IterString` | C string byte loop (`array.go`) |
| `for i in 0..10` | *(separate `ForRangeStmt`)* | numeric for-loop |

Range loops use `ForRangeStmt`, not `ForInStmt`.

---

## Memory arenas (app programs)

| Arena | Reset | Emitted in |
|-------|-------|------------|
| Frame | Start of each frame | `arena.go` → `dd_frame_arena_reset()` |
| Level | Scene init | `arena.go` → `dd_level_arena_reset()` |

Used for scratch strings and level-scoped spawn data. See [LANGUAGE.md](LANGUAGE.md).

---

## Module system

- `use graphics;` resolves via `internal/pkg/resolver.go` + `compiler/modules.go`
- **`export fn` / `export let`** — `compiler/modules_export.go` filters visible symbols
- `include "path.dd"` remains textual preprocessing in `compiler/source.go`

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
3. `examples/raylib/array_demo.dd` — `Array<T>` + for-in
4. `internal/codegen/app_emit.go` — generated main
5. `internal/typecheck/forin.go` — iterable disambiguation
6. `internal/codegen/array.go` — `DD_Array` + for-in codegen
7. `internal/compiler/pipeline.go` — wiring
8. `docs/HANDOFF.md` — full state for next programmer
