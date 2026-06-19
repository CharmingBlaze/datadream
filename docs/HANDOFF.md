# Handoff — DataDream Language Development

**For:** the next programmer or agent session picking up this repo  
**Last updated:** June 2026 (post-Studio + turnkey packaging push to `main`)  
**Compiler:** 0.1.0 · **raylib:** 6.0 · **Progress:** ~**99% v1-complete** — language, tooling, Studio IDE, and all **25** examples build; **ship gate:** publish GitHub Release v1.0.0 ([PUBLISH.md](PUBLISH.md))

---

## Executive summary

DataDream is a **Go compiler** that turns `.dd` source into **C**, then links with **bundled Clang + raylib 6.0** to produce a native binary. End users never need Go.

**What works today:** full friendly Layer 4 API, ECS (Layer 5), raw raylib interop (~548 functions), type checker with hints, growable `Array<T>` with `for x in array`, entity iteration, module exports, selective `use raylib { … }`, frame/level arenas, **DataDream Studio** (Wails desktop IDE + `datadream ide` web server), offline Monaco editor, turnkey release zips with root launchers, `datadream new`, VS Code grammar.

**Primary blocker for v1.0 ship:** publish the GitHub Release — workflow is ready; maintainer runs [PUBLISH.md](PUBLISH.md) and verifies Studio launchers in each zip.

**Primary gaps (post-ship):** LSP, automated Studio smoke test in `verify-dist`, README quick start still CLI-first. Language v1 surface is complete (const, enum, import, enum match, indexed for-in, ptr types).

---

## Continue from here

### Read first (~20 min)

| Order | Doc | Why |
|-------|-----|-----|
| 1 | [VISION.md](VISION.md) | Anti-goals are non-negotiable |
| 2 | [ARCHITECTURE.md](ARCHITECTURE.md) | Pipeline, package map |
| 3 | [LANGUAGE.md](LANGUAGE.md) | What users can write today |
| 4 | [SYNTAX.md](SYNTAX.md) | Keywords/operators by layer |
| 5 | [ROADMAP.md](ROADMAP.md) | Task status table |

### Verify your environment

```bash
go build -o datadream ./cmd/datadream

datadream sdk install clang
datadream sdk install raylib
datadream doctor

go test ./internal/...
datadream check --codegen examples/raylib/hello_3d.dd
datadream check --codegen examples/raylib/array_demo.dd
datadream check --codegen examples/coin-runner/game.dd
datadream build examples/raylib/hello_friendly.dd -o hello
```

**Release smoke test (maintainers, Windows):**

```powershell
.\scripts\build-dist.ps1
.\scripts\verify-dist.ps1 dist\datadream-windows-amd64.zip
```

Cold-install on Windows with bundled Clang (~188 MB zip) is verified: extract → `doctor` ✓ → `run hello_friendly` with no system compiler on PATH. Linux/macOS cold-install verified via CI `dist-verify` (CI skips Studio build; release workflow builds Studio).

**Studio smoke test (maintainers, after `build-dist`):**

```powershell
# Windows — after build-dist.ps1
.\dist\...\Start DataDream Studio.bat   # or root DataDream Studio.exe in zip tree
# Expect: IDE opens, clicker.dd loads, Ctrl+Enter runs
```

```bash
./scripts/build-studio.sh              # dev build only
datadream studio examples/beginner     # from repo root with built binary in bin/
```

---

## What is done (do not redo)

### Compiler & distribution
- ✅ `cmd/datadream/main.go` → `internal/cli.Run`
- ✅ CI: `go test`, example builds, bindgen-check, dist-verify (Win/Linux/macOS; **`--skip-studio`** in CI)
- ✅ `scripts/build-dist.ps1` / `build-dist.sh` + `packdist --verify`
- ✅ `.github/workflows/release.yml` — publish zips to GitHub Release **with Studio** (GTK/WebKit on Linux runner)
- ✅ Windows cold-install with bundled llvm-mingw
- ✅ `datadream new` — scaffold `game.dd`, `assets/`, README
- ✅ VS Code extension + TextMate grammar (`syntaxes/`, `editor/datadream/`)

### DataDream Studio (IDE)
- ✅ **`datadream studio`** — Wails desktop app (`cmd/studio/`, `internal/cli/studio.go`)
- ✅ **`datadream ide`** — embedded HTTP IDE (`internal/ide/server.go`, port 3847)
- ✅ Shared backend: `internal/ide/service.go` (tree, read/write, check, build, run, doctor)
- ✅ Frontend: `internal/ide/web/` — Monaco, file tree, corporate dark theme, Wails bridge in `app.js`
- ✅ **Offline Monaco** — vendored under `internal/ide/web/vendor/monaco/` via `scripts/fetch-monaco.*`
- ✅ **`DATADREAM_ROOT` auto-detect** — `internal/ide/root.go` from install layout (`sdk/manifest.json`)
- ✅ **Turnkey launchers** — `packdist` writes `GETTING_STARTED.txt`, `Start DataDream Studio.bat`, `start-studio.sh`, root IDE copies ([DISTRIBUTION.md](DISTRIBUTION.md), [STUDIO.md](STUDIO.md))
- ✅ Windows WebView2 embedded (`wails build -webview2 embed`)
- 🟡 **Release CI not yet run end-to-end** — first `v1.0.0` workflow may surface Wails/macOS deps issues
- ✅ **Linux AppImage** — GTK + WebKit bundled via `scripts/build-studio-appimage.sh` (no system install)
- ❌ **`verify-dist` does not launch Studio** — only doctor + hello + coin-runner builds

### Language core (Layer 1)
- ✅ `let`, `fn`, `if`/`else`, `while`, **`loop`**, **`for i in 0..n` / `0..=n`**, **`match`**, **`defer`**, `break`, `continue`, `return`
- ✅ **`let name: Type;`** — type-only declaration (no `=`) for empty arrays etc.
- ✅ Struct literals, **`struct` / `entity` methods** (`obj.method()` → `Type_method(&obj, …)`)
- ✅ **`match` struct destructuring** (`Point { x, y }`)
- ✅ `use`/`using`/`extern c`, `include`
- ✅ `none` alias for `null`
- ✅ C binding parse: keyword field names, `float[]`, `180.0f` suffix

### Arrays & iteration (Layer 1 + 5)
- ✅ **`Array<T>`** / **`list T`** sugar → growable **`DD_Array`** runtime (`data`, `len`, `cap`, `elem_size`)
- ✅ Methods: `.push()`, `.len` / `.len()`, `.pop()`, `.remove(i)`, `.remove_dead()`
- ✅ **`for x in array`** — disambiguated in **type checker** via `IterKind` (not parser)
- ✅ **`for e in Entity`** — ECS registry iteration
- ✅ **`for ch in "hello"`** / string variable — UTF-8 **byte** iteration (`byte` binding)
- ✅ Compile-time **warning** (non-blocking) for `.remove()` on same array during `for x in`
- ✅ Examples: `array_for_in.dd`, `array_demo.dd`, `string_for_in.dd`, `array_remove_warning.dd`

**Four `for x in …` forms** (see [LANGUAGE.md](LANGUAGE.md)):

| Syntax | AST / kind | Where resolved |
|--------|------------|----------------|
| `for i in 0..10` | `ForRangeStmt` | parser |
| `for e in Enemy` | `ForInStmt`, `IterEntity` | typecheck |
| `for x in arr` / `[1,2,3]` | `ForInStmt`, `IterArray` | typecheck |
| `for ch in "hi"` / `msg` | `ForInStmt`, `IterString` | typecheck |

Key files: `internal/ast/ast.go` (`IterKind`), `internal/typecheck/forin.go`, `internal/codegen/array.go`, `internal/codegen/stmts.go`.

### Raw raylib (Layer 2)
- ✅ `use raylib;` + `libs/raylib/raw.dd` (~548 functions)
- ✅ Selective **`use raylib { InitWindow, DrawText };`** (`hello_selective.dd`)
- ✅ **`import raylib { … }`** — qualified-only access (`import_demo.dd`)
- ✅ `datadream bind` from headers
- ✅ `hello_raw.dd`, `hello_using.dd`, **`hello_3d.dd`**, **`hello_selective.dd`**

### Language completeness (Layer 1)
- ✅ **`const` bindings** — top-level and block scope; reassignment rejected
- ✅ **`enum` + `Type.Variant` + enum `match` arms** (`language_features.dd`)
- ✅ **`for i, x in array`** — index + value iteration
- ✅ **`ptr<T>` / `voidptr`** — C pointer types in codegen

### Friendly API (Layer 4)
| Namespace | Codegen |
|-----------|---------|
| `draw.*` | `codegen/draw.go`, `friendly.go` |
| `input.*` / `keys.*` / mouse | `game_runtime.go`, `friendly.go` |
| `screen.*` | `friendly.go` |
| `random.*` | `game_runtime.go` |
| `math.*` / `time.*` | `stdlib.go`, `friendly.go` |
| `collision.*` | `stdlib.go` |
| `audio.*` / `assets.*` | `stdlib.go`, `stdlib_runtime.go` |
| `ui.*` (raygui) | `ui.go`, `ui_runtime.go` |

Living reference: `examples/raylib/commands.dd`.

### App sugar (Layer 5)
- ✅ `app`, `window { }`, `start` / `update` / `draw` → raylib loop
- ✅ `scene`, `entity`, `spawn`, `destroy`, `system`, `on`, entity `draw`
- ✅ Entity `fn` methods + `self.method()` calls
- ✅ **`export fn` / `export let`** — module boundaries (`exports_module.dd`, `libs/demoexports/`)

### Memory model
- ✅ **Frame arena** — `dd_frame_arena_reset()` each app frame (`codegen/arena.go`)
- ✅ **Level arena** — `dd_level_arena_reset()` on scene init
- ✅ Documented in [LANGUAGE.md](LANGUAGE.md) § Memory

### Type checker & diagnostics
- ✅ `internal/typecheck/` — default phase in `datadream check`
- ✅ Hints: unknown identifiers, bad namespace methods, arg counts, struct/entity fields
- ✅ **`internal/errors/`** — Rust-style snippet + caret + hint
- ✅ Lex/parse diagnostics with hints (`compiler/diagnostics.go`)
- ✅ **Warnings** — typecheck warnings do not fail `check`/`compile` (`Diagnostic.Warning`, yellow output)
- ✅ Codegen-stage errors use hint pipeline (`codegen/diagnostic.go`)
- 🟡 Parser/lexer hint coverage — edge cases remain

### Attributes
- ✅ **`@packed` entity** — SoA pool (`EntityPool` + `idx` handle), field access via pool arrays
- ✅ **`@save` struct** — binary `fwrite`/`fread` serialize/deserialize (`attributes_demo.dd`)

---

## Examples (27 `.dd` files)

All core examples pass `datadream check --codegen`. CI builds on Linux + macOS.

```
examples/beginner/clicker.dd
examples/coin-runner/game.dd
examples/colors/alpha.dd
examples/colors/css_demo.dd
examples/game-loop/main.dd
examples/hello/hello.dd
examples/raylib/array_demo.dd
examples/raylib/array_for_in.dd
examples/raylib/array_remove_warning.dd   ← warning demo (check still exits 0)
examples/raylib/attributes_demo.dd
examples/raylib/audio_demo.dd
examples/raylib/commands.dd
examples/raylib/control_flow.dd
examples/raylib/entity_demo.dd
examples/raylib/exports_module.dd
examples/raylib/features.dd
examples/raylib/graphics_module.dd
examples/raylib/hello_3d.dd
examples/raylib/hello_friendly.dd
examples/raylib/hello_raw.dd
examples/raylib/hello_selective.dd
examples/raylib/hello_using.dd
examples/raylib/import_demo.dd
examples/raylib/language_features.dd
examples/raylib/mario64_demo.dd
examples/raylib/match_destruct.dd
examples/raylib/string_for_in.dd
examples/raylib/ui_demo.dd
```

---

## Your mission — prioritized work

### P0 — Ship v1.0 (do this first)

| Task | Status | Acceptance |
|------|--------|------------|
| Windows bundled-Clang zip | ✅ | `build-dist.ps1` + cold PATH test |
| Linux/macOS dist-verify | ✅ | CI `dist-verify` job (no Studio) |
| CI on every push/PR | ✅ | `.github/workflows/ci.yml` |
| GitHub release workflow + Studio | ✅ | `.github/workflows/release.yml` |
| Code on `main` | ✅ | Pushed June 2026 |
| **Run release workflow** | 🔄 | [PUBLISH.md](PUBLISH.md) — tag `v1.0.0`, all three zips |
| **Verify Studio in release zips** | 🔄 | Unzip → double-click launcher → run `clicker.dd` (Win first) |
| **Fix release build if CI fails** | ❓ | Wails on Linux/macOS; macOS may need ImageMagick/npm in workflow |

Do not start large new language features until P0 is signed off or explicitly deprioritized.

### P0.5 — Turnkey UX polish (right after first release)

| Task | Status | Notes |
|------|--------|-------|
| README quick start → Studio-first | 🟡 | Root README still leads with CLI + PATH |
| Release notes body → Studio-first | 🟡 | `release.yml` publish body still CLI-first |
| `verify-dist` includes Studio binary check | ❌ | At least assert `bin/datadream-studio*` exists when not `--skip-studio` |
| `build-studio.sh` platform flags | ✅ | `-webview2 embed` Windows only; `-tags webkit2_41` on Linux |
| Linux AppImage (optional) | ✅ | `build-studio-appimage.sh` + packdist launchers |

### P1 — Friendly game layer

**Complete.** See [ROADMAP.md](ROADMAP.md) P1 table.

### P2 — Language correctness (remaining)

| Task | Status | Files |
|------|--------|-------|
| Type checker | ✅ | `internal/typecheck/` |
| Module exports | ✅ | `modules_export.go`, `exports_module.dd` |
| `for x in array` / `DD_Array` | ✅ | `array.go`, `forin.go` |
| `match` destructuring | ✅ | `match_destruct.dd` |
| `defer` on return/break/continue | ✅ | `defer.go` |
| **`@packed` full SoA codegen** | ✅ | `ecs_packed.go`, `attrs.go` |
| **`@save` real serialize** | ✅ | `attrs.go` — int/float/bool/Vec/string |
| Codegen error hints | ✅ | `codegen/diagnostic.go` → pipeline file/line/hint |
| Parser/lexer hint coverage | 🟡 | Most paths done; edge cases remain |

### P3 — ECS

**Complete.** See [ROADMAP.md](ROADMAP.md) P3 table.

### P4 — Tooling & adoption

| Task | Status | Notes |
|------|--------|-------|
| `datadream new` | ✅ | `internal/cli/new.go` |
| VS Code / TextMate grammar | ✅ | `syntaxes/`, `editor/datadream/` |
| **DataDream Studio (desktop IDE)** | ✅ | `cmd/studio/`, Wails, embedded Monaco |
| **`datadream ide` (web IDE)** | ✅ | `internal/ide/`, for dev / fallback |
| Turnkey zip launchers | ✅ | `packdist` + `GETTING_STARTED.txt` |
| **LSP** (hover, go-to-def) | ❌ | After grammar; v1.1 |

### P5 — Polish (later)

- `match` type patterns (beyond struct destructuring)
- Entity array demo with `.dead` + `.remove_dead()` in a real game loop
- String iteration: codepoint/grapheme (v1 is byte-only by design)
- Spatial partitioning for `collision.*` (see [LOOPS.md](LOOPS.md) v2)
- Full debug/release build mode (frame-time logging, pool overflow at runtime)

### Frame-budget loop protection (June 2026)

Design: [LOOPS.md](LOOPS.md). Implementation status:

| Item | Status | Location |
|------|--------|----------|
| Refuse infinite `loop` in per-frame blocks | ✅ | `typecheck/loops.go` |
| Runtime range bound warning | ✅ | `typecheck/loops.go` |
| Allocation / `.push()` in loop warnings | ✅ | `typecheck/loops.go` |
| Nested entity loop O(n²) warning | ✅ | `typecheck/loops.go` |
| Draw-block mutation warning | ✅ | `typecheck/loops.go` |
| Entity packed pool codegen (default 1024) | ✅ | `codegen/ecs.go` |
| `@max(n)` entity pool cap | ✅ | `codegen/ecs.go` |
| Debug `while` iteration guard (`#ifndef NDEBUG`) | ✅ | `codegen/stmts.go`, `lifecycle.go` |
| Spatial grid collision | ❌ v2 | — |
| `@max_iterations(n)` on `while` | ❌ v2 | — |


## How to add a language feature (recipe)

Example: add `ui.slider(...)`.

```
lexer (usually nothing)
  → parser/expr.go (namespace root)
  → ast/*.go (if new node or fields)
  → typecheck/forin.go or builtins.go (if disambiguation / namespace)
  → codegen/ui.go (emit raylib GuiSlider)
  → analyze.go (flag conditional runtime emission)
  → example in examples/raylib/ui_demo.dd
  → test in internal/codegen/ui_codegen_test.go
  → docs: LANGUAGE.md, SYNTAX.md, ROADMAP.md, this file
```

**For `for x in …` iterable types:** keep parser dumb (`ForInStmt` only); add `IterKind` case in typecheck + codegen switch — do not special-case in parser.

**Rules:**
- Every feature needs an example passing `datadream check --codegen`
- Match existing style; Go only in tooling
- Do not break [VISION.md](VISION.md) anti-goals
- Update this file + ROADMAP when status changes
- **Do not commit** unless explicitly asked

---

## Key files map

```
cmd/datadream/main.go              CLI entry
cmd/studio/                        Wails desktop IDE (build with wails build, not go build)
internal/cli/                      run, build, check, bind, doctor, sdk, new, ide, studio
internal/cli/diagnostics.go        Errors + warnings formatting
internal/ide/
  service.go                       Shared IDE API (tree, read, write, check, build, run)
  server.go                        HTTP server for datadream ide
  root.go                          DATADREAM_ROOT / distribution layout detection
  web/                             Embedded frontend + vendor/monaco/
internal/lexer/lexer.go            Tokens, IsBindingName, 0..=, ||/&&
internal/parser/                   loop, defer, match, entity methods, list T sugar
internal/ast/ast.go                ForInStmt, IterKind, AST nodes
internal/typecheck/
  typecheck.go                     Main checker, scopes, warnings, loop depth
  loops.go                         Per-frame loop static analysis + warnings
  forin.go                         resolveForIn, IterKind, remove-during-iter warning
  hints.go                         Suggested fixes
internal/errors/errors.go          Diagnostic reporter (error/warning/hint)
internal/codegen/
  array.go                         DD_Array runtime, for-in array/string codegen
  arena.go                         Frame + level arenas
  defer.go                         defer LIFO on all exit paths
  ecs.go                           Entity pool, spawn, for-in entity, ECS hooks
  lifecycle.go                     Per-frame lifecycle depth for debug while guards
  decls.go                         struct/entity + methods
  expr.go                          Method calls, namespaces, array methods
  stmts.go                         genForIn → genForInByKind
  analyze.go                       needsArrayRuntime, usage analysis
internal/compiler/
  pipeline.go                      Check, Compile; warnings don't block
  typecheck_phase.go               Maps typecheck errors/warnings
  modules_export.go                export fn / export let filter
tools/packdist/                    Release zip + --verify + launchers
scripts/build-dist.ps1             Windows dist (bundled Clang + Studio)
scripts/build-studio.sh           Wails Studio only
scripts/build-studio-appimage.sh  Linux AppImage (bundles GTK/WebKit)
scripts/fetch-monaco.ps1           Vendor Monaco for offline IDE
libs/raylib/raw.dd                 Generated — run check-bindgen to refresh
```

---

## Known gotchas

| # | Issue | Fix |
|---|-------|-----|
| 1 | `check` vs `build` | `check` = lex + parse + typecheck; `--codegen` adds C without link |
| 2 | Global `let` with calls | Deferred to `datadream_init_globals()` — no calls at C file scope |
| 3 | `.gitignore` | Use `/datadream` not `datadream` or `cmd/datadream/` is ignored |
| 4 | `internal/pkg` → `internal/sdk` | Import cycle — forbidden |
| 5 | coin-runner cwd | Build from `examples/coin-runner/` for asset paths |
| 6 | Windows linking | Release zips bundle llvm-mingw; MSVC-only LLVM fails raylib link |
| 7 | Parser on `raw.dd` | Completes in <1s; keyword binding names via `IsBindingName` |
| 8 | Entity method signatures | Empty param list emits `(self)` not `(self, void)` |
| 9 | `let` in lifecycle blocks | Not global (`topLevel = false` in update/draw) |
| 10 | **`c` is a keyword** | C interop — do not use `for c in "hello"`; use `ch` or `byte` |
| 11 | Array remove during loop | Warning only — use `.dead` + `.remove_dead()` after loop |
| 12 | Struct elements in `for x in arr` | Binding is **pointer** — mutations apply in place |
| 13 | **Studio build** | Use `wails build` via `scripts/build-studio.*` — plain `go build ./cmd/studio` fails (build tags) |
| 14 | **Wails vendor** | Studio builds use `GOFLAGS=-mod=mod` (root `vendor/` is raylib headers only) |
| 15 | **CI vs release** | CI `dist-verify` uses `--skip-studio`; full IDE only in release workflow |
| 16 | **Monaco in repo** | Large vendor tree under `internal/ide/web/vendor/` — `fetch-monaco` runs in `build-dist` |

---

## Milestones

### v0.2 — ✅ complete

Audio/ui demos, type checker, all examples, CI on Linux + macOS.

### v1.0 — gate (two steps left)

- [x] Fresh zip → `doctor` ✓ → `run hello_friendly` (Win/Linux/macOS)
- [x] `coin-runner` builds with assets
- [x] `hello_raw` + `hello_3d` + `hello_selective` build
- [x] All 25 examples pass `check --codegen`; CI builds
- [x] bindgen-check keeps `raw.dd` in sync
- [x] Docs match reality (Studio, GETTING_STARTED, DISTRIBUTION)
- [x] Studio IDE + turnkey launchers in packdist
- [ ] **GitHub v1.0 Release published** — [PUBLISH.md](PUBLISH.md)
- [ ] **Studio smoke-tested from release zip** (double-click launcher)

### v1.1 — suggested next

- LSP basics
- `verify-dist` Studio checks
- README / release notes Studio-first
- Linux AppImage (optional)

---

## Daily commands

```bash
go build -o datadream ./cmd/datadream
go test ./internal/...

datadream check examples/raylib/<example>.dd
datadream check --codegen examples/raylib/<example>.dd
datadream build examples/raylib/<example>.dd -o /tmp/test

cd examples/coin-runner && ../../datadream build game.dd -o coin-runner

datadream bind sdk/raylib/6.0/include/raylib.h --raw --out libs/raylib/raw.dd
./scripts/check-bindgen.sh
./scripts/build-all-examples.sh

# Studio (maintainers)
./scripts/fetch-monaco.sh
./scripts/build-studio.sh
datadream studio examples/beginner
datadream ide --port 3847
```

---

## Context

| Item | Value |
|------|-------|
| Go module | `datadream` |
| File extension | `.dd` |
| Former name | Koda — do not use |
| End users | `datadream` binary + `sdk/` — never require Go |
| Repo | https://github.com/CharmingBlaze/datadream |

**North star:** A native game language where beginners write `app`/`draw`/`input` and advanced users drop to raw `use raylib` for 3D — same compiler, same binary.

---

## Related docs

| Doc | When to read |
|-----|--------------|
| [NEXT_SESSION_PROMPT.md](NEXT_SESSION_PROMPT.md) | **Paste into a new agent chat** |
| [ROADMAP.md](ROADMAP.md) | Task status table |
| [PUBLISH.md](PUBLISH.md) | Release v1.0 step-by-step |
| [STUDIO.md](STUDIO.md) | Desktop IDE — end users + maintainers |
| [GETTING_STARTED.txt](GETTING_STARTED.txt) | One-page quick start (copied into release zip) |
| [DESIGN.md](DESIGN.md) | Full design map, v1 target program |
| [ARCHITECTURE.md](ARCHITECTURE.md) | Pipeline and package layout |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Feature addition checklist |
| [INTEROP.md](INTEROP.md) | raylib, bindgen, `raw.dd` quirks |
| [SETUP.md](SETUP.md) | SDK install, troubleshooting |
| [DISTRIBUTION.md](DISTRIBUTION.md) | Release zips, packdist |
