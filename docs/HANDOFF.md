# Handoff — Continue DataDream Language Development

**For:** the next programmer or agent session picking up this repo  
**Last updated:** June 2026  
**Compiler:** 0.1.0 · **raylib:** 6.0 · **Progress:** ~60% toward v1

---

## Continue from here

DataDream compiles `.dd` → C → Clang → native binary. **Windows end-to-end works.** The language core and friendly Layer 4 API are largely in place. Your job is to close the remaining language gaps and ship multi-platform builds.

**Read first (15 min):**
1. [VISION.md](VISION.md) — anti-goals are non-negotiable
2. [SYNTAX.md](SYNTAX.md) — keywords/operators by layer (Layer 5 = app sugar)
3. [LANGUAGE.md](LANGUAGE.md) — what users can write today

**Verify your environment:**

```bash
go build -o datadream ./cmd/datadream

datadream sdk install clang
datadream sdk install raylib
datadream doctor

go test ./internal/parser/... ./internal/codegen/... ./internal/colors/...
datadream check --codegen examples/raylib/hello_3d.dd
datadream check --codegen examples/raylib/control_flow.dd
datadream check --codegen examples/raylib/commands.dd
datadream check --codegen examples/raylib/audio_demo.dd
datadream check --codegen examples/raylib/ui_demo.dd
datadream build examples/raylib/hello_friendly.dd -o hello
```

All of the above should pass on Windows with a populated SDK.

---

## What is done (do not redo)

### Language core (Layer 1)
- `let`, `fn`, `if`/`else`, `while`, **`loop`**, **`for i in 0..n` / `0..=n`**, **`match`**, **`defer`**, `break`, `continue`, `return`
- Struct literals, `use`/`using`/`extern c`, `include`
- `none` alias for `null`
- C binding parse: keyword field names (`shader`, `data`), `float[]`, `180.0f` suffix

### Raw raylib (Layer 2)
- `use raylib;` + `libs/raylib/raw.dd` (~555 functions)
- `datadream bind` from headers
- `examples/raylib/hello_raw.dd`, `hello_using.dd`, **`hello_3d.dd`** (v1 target, uses `defer CloseWindow()`)

### Friendly API (Layer 4) — mostly done
| Namespace | Status | Codegen |
|-----------|--------|---------|
| `draw.*` | ✅ | `codegen/draw.go`, `friendly.go` |
| `input.*` / `keys.*` / mouse | ✅ | `game_runtime.go`, `friendly.go` |
| `screen.*` | ✅ | `friendly.go` |
| `random.*` | ✅ | `game_runtime.go` |
| `math.*` / `time.*` | ✅ | `stdlib.go`, `friendly.go` |
| `collision.*` | ✅ | `stdlib.go` |
| `audio.*` / `assets.*` | ✅ | `stdlib.go`, `stdlib_runtime.go`, `audio_demo.dd` |
| `ui.*` (raygui) | ✅ | `ui.go`, `ui_runtime.go`, `ui_demo.dd` |

### App sugar (Layer 5) — ✅ core wired
- ✅ `app`, `window { }`, `start` / `update` / `draw` → raylib loop
- ✅ `scene`, `entity`, `spawn`, `destroy`, `system`, `on`, entity `draw` — registry + app loop
- ✅ `for x in Entity` iterates live instances
- 🟡 Generic `for x in array` still stubbed; scene-local `let` scoping minimal

### Examples (all pass `check --codegen` on Windows)
```
examples/beginner/clicker.dd       ← start here for beginners
examples/raylib/hello_friendly.dd
examples/raylib/hello_3d.dd
examples/raylib/control_flow.dd    ← loop, defer, match
examples/raylib/commands.dd        ← all friendly namespaces
examples/raylib/audio_demo.dd      ← audio + assets load/play/unload
examples/raylib/ui_demo.dd         ← raygui ui.button / ui.label
examples/raylib/graphics_module.dd ← use graphics module (no include)
examples/raylib/features.dd
examples/coin-runner/game.dd       ← build from examples/coin-runner/
examples/game-loop/main.dd
examples/colors/*.dd
```

---

## Your mission — prioritized language work

Work in this order. Do not skip P0. Do not start Layer 5 ECS before Layer 4 gaps are closed.

### P0 — Ship blockers (infra, not language, but gates v1)

| Task | Acceptance criteria | Files |
|------|---------------------|-------|
| Linux amd64 build | `datadream build hello_friendly.dd` + `coin-runner` | `internal/sdk/link.go`, CI |
| macOS arm64 build | Same | CI macos job |
| CI workflow | `go test`, `doctor`, `build hello_friendly` | `.github/workflows/ci.yml` |
| Release zip | Fresh unzip → `doctor` ✓ → `run hello_friendly` | `scripts/build-dist.ps1`, `build-dist.sh`, `packdist --verify` |

### P1 — Finish Layer 4 (language — do this next for pure language work)

These are the highest-value **language** tasks right now.

#### 1. Harden `audio.*` and `assets.*`

**Goal:** A working example that loads a sound and texture, plays sound, draws texture, cleans up with `defer`.

**Start here:**
- `internal/codegen/stdlib.go` — `audio.play`, `assets.texture`
- `internal/codegen/stdlib_runtime.go` — C helpers
- `internal/codegen/analyze.go` — conditional emission flags

**Deliverables:**
- `examples/raylib/audio_demo.dd` (new)
- Extend `stdlib_codegen_test.go`
- Update [LANGUAGE.md](LANGUAGE.md) § audio/assets

**Acceptance:** `datadream build examples/raylib/audio_demo.dd` on Windows.

#### 2. `ui.*` raygui wrapper

**Goal:** `ui.button("Click", { position: vec2(10,10), width: 120, height: 32 })` → `GuiButton`.

**Start here:**
- `internal/codegen/expr.go` — namespace dispatch (copy `draw.*` pattern)
- New `internal/codegen/ui.go`
- Raygui is in raylib; link already works via `use raylib`

**Deliverables:**
- `examples/raylib/ui_demo.dd`
- `ui_codegen_test.go`
- [LANGUAGE.md](LANGUAGE.md) + [SYNTAX.md](SYNTAX.md) Layer 4 table

**Acceptance:** Button renders in a window; `check --codegen` passes.

#### 3. Type checker phase

**Goal:** Catch obvious errors before codegen: unknown identifiers, wrong arg counts on builtins, bad struct fields.

**Start here:**
- `internal/compiler/phases.go` — `Phase` interface exists
- `internal/compiler/pipeline.go` — insert phase between parse and codegen
- New `internal/typecheck/` package

**Scope for v1 (minimal):**
- `let` assignments vs type hints
- Builtin namespace calls (`draw.text`, `input.pressed`, …)
- Struct literal field names against known structs
- Do **not** build full inference yet — extend `inferTypeFromExpr` heuristics first

**Deliverables:**
- `internal/typecheck/typecheck_test.go`
- `datadream check` runs typecheck by default (or `--types` flag)
- Update [ROADMAP.md](ROADMAP.md)

#### 4. Module resolution (`use graphics`) — ✅

**Goal:** `use graphics;` loads `libs/graphics/wrapper.dd` without textual `include`.

**Implemented:** `internal/pkg/resolver.go` (`ModuleSourcePath`), `internal/compiler/modules.go`, `libs/graphics/wrapper.dd`, parser allows `app` after `use`.

**Example:** `examples/raylib/graphics_module.dd`

### P2 — Layer 5 engine sugar (after P1) — ✅ core wired

| Task | Status | Notes |
|------|--------|-------|
| `entity` lifecycle in loop | ✅ | Registry, `update_all`, `draw_all`, `self->` |
| `spawn X at vec2(...)` | ✅ | `Entity_spawn`, `let x = spawn Entity` |
| `for x in Entity` | ✅ | Iterates entity registry |
| `scene` blocks | ✅ | init/start/update/draw in app loop |
| `on key "space"` | ✅ | Polls input, calls handler |
| `system` blocks | ✅ | `system_*_run(dt)` each frame |
| Entity `draw { }` | ✅ | Per-entity draw + `draw_all` |

**Examples:** `examples/coin-runner/game.dd` (ECS), `examples/raylib/entity_demo.dd`

### P3 — Polish (later)

- `match` destructuring / type patterns (today: equality only → if-else chain)
- `defer` on early `return` paths (verify LIFO on all exit paths)
- Selective `use raylib { InitWindow, DrawText }`
- LSP / VS Code extension

---

## How to add a language feature (recipe)

Example: add `ui.slider(...)`.

```
lexer (usually nothing)
  → parser/expr.go (namespace root)
  → codegen/ui.go (emit raylib GuiSlider)
  → analyze.go (flag if conditional emission needed)
  → example in examples/raylib/ui_demo.dd
  → test in internal/codegen/ui_codegen_test.go
  → docs: LANGUAGE.md, SYNTAX.md, ROADMAP.md, this file
```

**Rules:**
- Every feature needs an example that `check --codegen` passes
- Match existing style in the file you edit
- Go only in tooling — no Python
- Do not break [VISION.md](VISION.md) anti-goals (no Python syntax, no hidden raylib, no VM)

---

## Key files map

```
internal/lexer/lexer.go          Tokens, IsBindingName, 0..=, ||/&&
internal/parser/
  stmt.go                        loop, defer, match, break, continue
  decls.go                       parseBindingIdent, float[] types
  expr.go                        draw.*, input.* namespace calls
  interop.go                     use, extern c
internal/ast/ast.go              LoopStmt, DeferStmt, TypeExpr.Array
internal/codegen/
  generator.go                   genNode dispatch
  analyze.go                     what runtimes to emit
  defer.go                       defer stack, genStmts
  friendly.go                    screen, keys, mouse
  stdlib.go                      time, math, audio, assets
  draw.go                        draw.* → raylib
  ui.go                         ui.* → raygui
  game_runtime.go                sprite, input, collision
  ecs.go                         entity registry, spawn, for-in, events
  app_emit.go                    main loop, lifecycle_*
  stmts.go                       if/for/while/loop/match
internal/compiler/pipeline.go    Check, Compile, Phase hook
libs/raylib/raw.dd               Generated — run bindgen to refresh
examples/raylib/commands.dd      Full friendly API reference
docs/SYNTAX.md                   Keywords by layer
docs/LANGUAGE.md                 User-facing API reference
```

---

## Known gotchas

1. **`check` ≠ `build`** — default `check` is parse + typecheck; use `check --codegen` or `build` for C verification.
2. **Global `let` with calls** — deferred to `datadream_init_globals()`; never emit function calls at C file scope.
3. **`use raylib` scope** — C names land in file scope; `isExternAPICall` in `interop.go` must distinguish user fns from extern.
4. **`internal/pkg` must NOT import `internal/sdk`** — import cycle.
5. **coin-runner cwd** — build from `examples/coin-runner/` so `assets/*.png` resolve.
6. **Windows linking** — use bundled llvm-mingw (`sdk install clang`), not MSVC LLVM alone.
7. **Parser on `raw.dd`** — completes in under 1s; keyword binding names via `IsBindingName`.
8. **Object field `c:`** — `c` is `TOKEN_C`; use `p1`/`p2`/`p3` for triangle points (parser quirk).
9. **`let` inside `update`/`draw`** — must not be treated as global (`topLevel = false` in lifecycle blocks).

---

## Definition of done — next milestone

Call **v0.2** done when:

- [x] `audio_demo.dd` builds on Windows
- [x] `ui_demo.dd` builds on Windows
- [x] Type checker catches at least: unknown builtin, wrong `draw.text` options, bad struct field
- [x] `go test ./internal/...` green
- [x] All examples pass `check --codegen` (17 `.dd` files; CI verifies on Linux + macOS)
- [x] Linux OR macOS build verified (CI: `.github/workflows/ci.yml`)

**v0.2 is complete.** Next gate is **v1** (fresh zip + coin-runner on all three OSes, docs match reality).

Call **v1** done when:

- Fresh zip → `doctor` ✓ → `build coin-runner` on **Windows, Linux, macOS** (`packdist --verify` or `scripts/build-dist.*`)
- CI **dist-verify** jobs on **Windows, Linux, and macOS** re-pack and verify every push/PR
- CI **bindgen-check** keeps `libs/raylib/raw.dd` in sync with `raylib.h`
- All examples pass `check --codegen` and `build` (CI: `build-all-examples.sh` on Linux + macOS)
- Docs match reality

---

## Daily commands

```bash
go build -o datadream ./cmd/datadream
go test ./internal/parser/... ./internal/codegen/... ./internal/colors/...

datadream check --codegen examples/raylib/<example>.dd
datadream build examples/raylib/<example>.dd -o /tmp/test

cd examples/coin-runner && ../../datadream build game.dd -o coin-runner

datadream bind sdk/raylib/6.0/include/raylib.h --raw --out libs/raylib/raw.dd

./scripts/build-all-examples.sh   # all 17 examples
./scripts/check-bindgen.sh
```

---

## Context

| Item | Value |
|------|-------|
| Go module | `datadream` |
| File extension | `.dd` |
| Former name | Koda — do not use |
| End users | `datadream` binary + `sdk/` — never require Go |
| Design identity | C/raylib structure, modern syntax, beginner-friendly |

**North star:** A native game language where beginners write `app`/`draw`/`input` and advanced users drop to raw `use raylib` for 3D — same compiler, same binary.

---

## Related docs

| Doc | When to read |
|-----|--------------|
| [ROADMAP.md](ROADMAP.md) | Task status table |
| [DESIGN.md](DESIGN.md) | Full design map, v1 target program |
| [ARCHITECTURE.md](ARCHITECTURE.md) | Pipeline and package layout |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Feature addition checklist |
| [INTEROP.md](INTEROP.md) | raylib, bindgen, `raw.dd` quirks |
| [SETUP.md](SETUP.md) | SDK install, troubleshooting |
