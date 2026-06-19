# Prompt for Next Programmer

Copy **everything below the horizontal rule** into a new chat, ticket, or agent session.

---

## Mission

You are continuing **DataDream** — a Go compiler that turns `.dd` source into **C**, links with **bundled Clang + raylib 6.0**, and ships as a **single zip per platform** (no Go required for end users).

**Current state:** ~**99% v1-complete**. All 25 examples build; dist zip verified locally. **Ship gate:** publish GitHub Release v1.0.0 ([docs/PUBLISH.md](docs/PUBLISH.md)).

**Identity:** C/raylib structure, modern syntax — **not** Python, not a VM, not English-sentence commands.

**Repo:** https://github.com/CharmingBlaze/datadream · Go module: `datadream`

---

## Read first (20 minutes)

| # | File | Why |
|---|------|-----|
| 1 | `docs/HANDOFF.md` | Full state, file map, gotchas, examples list |
| 2 | `docs/VISION.md` | Anti-goals — do not violate |
| 3 | `docs/ROADMAP.md` | Task status + Definition of done |
| 4 | `docs/LANGUAGE.md` | User-facing syntax (arrays, for-in, memory model) |
| 5 | `docs/ARCHITECTURE.md` | Pipeline: lex → parse → typecheck → codegen → driver |

---

## Verify before changing anything

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

**Windows release smoke (maintainers):**

```powershell
.\scripts\build-dist.ps1
.\scripts\verify-dist.ps1 dist\datadream-windows-amd64.zip
```

Expected: all tests pass; examples check clean; Windows zip cold-installs with bundled Clang.

---

## What is already done — do NOT redo

### Shipping & tooling
- CI on Win/Linux/macOS: test, build examples, bindgen-check, dist-verify
- Release workflow (`.github/workflows/release.yml`) — publishes zips to GitHub Release
- `datadream new` scaffold, VS Code grammar + minimal extension
- Lex/parse/typecheck diagnostics with hints; **warnings** are non-blocking

### Language (Layers 1–5)
- Control flow: `loop`, `defer`, `match`, `for i in 0..=n`, struct destructuring in `match`
- **`Array<T>` / `list T`** → `DD_Array` runtime; `.push`, `.len`, `.pop`, `.remove`, `.remove_dead`
- **`for x in …` disambiguation in type checker** (`IterKind`): entity / array / string (range uses `ForRangeStmt`)
- **`for ch in "hello"`** — byte iteration only (not graphemes); **`c` is a reserved keyword** — use `ch`
- ECS: `app`, `entity`, `scene`, `spawn`, `destroy`, `system`, `for x in Entity`
- **`export fn` / `export let`** for module boundaries
- Frame + level arenas documented and emitted
- Full Layer 4: `draw`, `input`, `audio`, `assets`, `ui`, … — see `examples/raylib/commands.dd`
- Raw raylib: `use raylib;`, `libs/raylib/raw.dd` (~548 fns)

### Examples
24 `.dd` files under `examples/` — all pass `datadream check --codegen`. Start with:
- `examples/beginner/clicker.dd`
- `examples/raylib/hello_friendly.dd`
- `examples/raylib/array_demo.dd`
- `examples/coin-runner/game.dd`

---

## Definition of done — v1.0

Ship when **all** are true (see `docs/ROADMAP.md`):

1. Fresh zip (Win/Linux/macOS) → `doctor` ✓ → `run hello_friendly` — **verified via local + CI dist-verify**
2. `coin-runner` builds and runs with assets
3. All examples pass `check --codegen` and `build`
4. Docs match reality
5. **GitHub Release v1.0 published** with three platform zips ← **YOU DO THIS FIRST**

---

## Your work — strict priority order

### Phase 1 — Publish v1.0 (P0) ← START HERE

| Task | How |
|------|-----|
| Run release workflow | `docs/PUBLISH.md` — `gh auth login` if needed, tag `v1.0.0` or workflow_dispatch |
| Verify release assets | Three zips attached; README links to Releases |
| Sign off DoD | Check every item in ROADMAP § Definition of done |

**Do not** start large language features until Phase 1 is done or user explicitly deprioritizes ship.

### Phase 2 — Language hardening (P2)

Pick **one** per session; each needs example + test + doc update:

| Task | Files | Acceptance |
|------|-------|------------|
| `@packed` full SoA entity layout | `codegen/attrs.go`, `ecs.go` | Entity fields in packed C struct; example builds |
| `@save` serialize/deserialize | TBD | Round-trip struct in example |
| Codegen error hints | `codegen/*.go`, `errors/` | Bad field/method shows caret + hint |
| Selective `use raylib { … }` | `parser/interop.go`, `codegen/interop.go` | Only listed symbols in scope |

**Done (June 2026):** frame-budget loop protection per [LOOPS.md](LOOPS.md) — `typecheck/loops.go`, entity pool + `@max` in `ecs.go`, debug while guards in `stmts.go`. v2: spatial collision grid, `@max_iterations`, full `--debug`/`--release` telemetry.

### Phase 3 — Tooling (P4)

| Task | Notes |
|------|-------|
| LSP (hover, go-to-def) | Build on `syntaxes/datadream.tmLanguage.json` + AST |
| Improve `datadream doctor` messages | Platform-specific install hints |

### Phase 4 — Polish (P5)

- Entity + `Array<Entity>` gameplay demo using `.dead` / `.remove_dead()`
- `match` type patterns
- String codepoint iteration (only if user requests — v1 is byte-by design)

---

## Architecture cheat sheet

```
Parser:     ForInStmt { binding, iterable, body }     — same node for all iterables
Typecheck:  resolveForIn() → sets IterKind on ForInStmt
Codegen:    genForInByKind() switch on IterKind
Arrays:     DD_Array in codegen/array.go
Warnings:   typecheck.Error{Warning:true} → compiler.Diagnostic{Warning:true} → yellow CLI output
```

**Adding a new iterable type:** extend `IterKind` in `ast.go`, `resolveForIn()` in `forin.go`, `genForInByKind()` in `array.go` — **never** fork the parser for each kind.

---

## Key files

```
cmd/datadream/main.go
internal/cli/                    check, build, run, new, doctor
internal/compiler/pipeline.go      Compile, Check (warnings don't block)
internal/typecheck/forin.go        IterKind, array/string/entity resolution
internal/codegen/array.go          DD_Array + for-in codegen
internal/codegen/arena.go          Frame/level memory
internal/codegen/ecs.go            Entities, spawn, for-in entity
internal/codegen/defer.go          defer on all exit paths
internal/compiler/modules_export.go export fn/let
examples/raylib/commands.dd        Layer 4 reference
examples/coin-runner/game.dd       Full game sample
.github/workflows/release.yml      Publish zips
docs/PUBLISH.md                    Release instructions
```

---

## Rules

1. **Go only** in tooling — no Python
2. **Every language feature** → example passing `datadream check --codegen` (+ `build` if runnable)
3. **Small focused diffs** — one feature or fix per change
4. **Match existing style** — read the file before editing
5. **Update docs** when status changes: `HANDOFF.md`, `ROADMAP.md`, `LANGUAGE.md`
6. **Do not break** `coin-runner`, bindgen drift (`scripts/check-bindgen.sh`), or `internal/pkg` → `internal/sdk` cycle
7. **Do not commit** unless the user explicitly asks

---

## Gotchas (read before debugging)

| Issue | Fix |
|-------|-----|
| `for c in "hello"` parse error | `c` is C-interop keyword — use `ch` |
| `.remove()` in `for x in arr` | Warning only; use `.dead` + `.remove_dead()` after loop |
| Struct in `for x in arr` | Loop var is **pointer** — `p.x += 1` mutates in place |
| `let bullets: Array<int>;` | Type-only `let` OK — no `=` required |
| Global `let x = sprite(...)` | Emitted via `datadream_init_globals()` |
| `.gitignore` | Must be `/datadream` not `datadream` |
| Former name **Koda** | Never use |

---

## Session definition of done

- [ ] Picked a phase and completed at least one task with acceptance criteria met
- [ ] `go test ./internal/...` passes
- [ ] Affected examples pass `check --codegen` (and `build` where relevant)
- [ ] If ship-related: release published or blocker documented
- [ ] Updated `docs/HANDOFF.md` and `docs/ROADMAP.md` if status changed

---

## Start command

1. Read `docs/HANDOFF.md`
2. Run verify commands above
3. Execute **Phase 1** (publish v1.0 Release) unless told otherwise
4. Then pick **one** Phase 2 item with tests + example

**North star:** A beginner downloads one zip, runs `datadream doctor`, runs `datadream new my-game`, sees a window with `draw.text("Hello")`, and gets a helpful error (with hint) when they typo a name — on Windows, Linux, or macOS.
