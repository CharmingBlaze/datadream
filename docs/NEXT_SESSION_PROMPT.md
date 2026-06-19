# Prompt for Next Programmer

Copy everything below the line into a new chat or ticket.

---

You are continuing development on **DataDream** — a Go compiler that translates `.dd` source → C → Clang → native binaries, with raylib as the primary graphics layer.

## Repo

`c:\Users\Rain\Downloads\DataDream` (Go module: `datadream`)

## Read first

1. `docs/HANDOFF.md` — current state and prioritized tasks
2. `docs/VISION.md` — anti-goals (do not violate)
3. `docs/SYNTAX.md` — keywords/operators by layer
4. `docs/LANGUAGE.md` — user-facing API reference

## Verify before changing anything

```bash
go build -o datadream ./cmd/datadream
datadream sdk install clang
datadream sdk install raylib
datadream doctor
go test ./internal/parser/... ./internal/codegen/... ./internal/colors/...
datadream check --codegen examples/raylib/hello_3d.dd
datadream check --codegen examples/raylib/control_flow.dd
datadream build examples/raylib/hello_friendly.dd -o hello
```

All should pass on Windows with a populated SDK.

## What is already done (do not redo)

- **Layer 1:** `loop`, `defer`, `match`, `break`, `continue`, `for i in 0..=n`, struct literals, `use`/`extern c`
- **Layer 2:** `use raylib;`, `libs/raylib/raw.dd`, bindgen, `hello_3d.dd` (v1 target)
- **Layer 4 (mostly):** `draw.*`, `input.*`, `screen.*`, `random.*`, `math.*`, `time.*`, `collision.*`
- **Layer 5 (partial):** `app`/`window`/`start`/`update`/`draw` work; `entity`/`scene`/`spawn` are stubs

## Your task — pick one track and finish it

### Track A — Language (recommended if on Windows)

**P1.1 — Audio + assets demo**

Implement and verify `audio.*` and `assets.*`:
- Harden `internal/codegen/stdlib.go` and `stdlib_runtime.go`
- Create `examples/raylib/audio_demo.dd` (load texture + sound, play, draw, `defer` cleanup)
- Add tests in `stdlib_codegen_test.go`
- Update `docs/LANGUAGE.md`

Acceptance: `datadream build examples/raylib/audio_demo.dd` succeeds.

**P1.2 — `ui.*` raygui wrapper**

Add `ui.button(...)` → `GuiButton`:
- New `internal/codegen/ui.go`, wire in `expr.go`
- Create `examples/raylib/ui_demo.dd` + test
- Update docs

### Track B — Compiler correctness

**P1.3 — Type checker phase**

Add `internal/typecheck/` as a `compiler.Phase`:
- Catch unknown builtins, bad struct fields, wrong `draw.text` options
- Wire into `datadream check` (default or `--types`)
- Tests required

### Track C — Infrastructure (gates v1)

- Linux amd64 + macOS arm64: `datadream build hello_friendly.dd` and `coin-runner`
- CI: `go test`, `doctor`, build hello_friendly
- Release zip smoke test

## Rules

- **Go only** in tooling — no Python
- **Every language feature needs an example** that passes `datadream check --codegen`
- **Match existing code style** — read surrounding files before editing
- **Small focused diffs** — one feature per change
- **Do not break:** `libs/raylib/raw.dd`, `internal/compiler/pipeline.go`, `examples/coin-runner/game.dd`
- **Do not start Layer 5 ECS** until Layer 4 (`audio`, `assets`, `ui`) is solid
- **Update docs** when status changes: `HANDOFF.md`, `LANGUAGE.md`, `ROADMAP.md`
- **Do not commit** unless asked

## Key files

```
internal/codegen/draw.go, friendly.go, stdlib.go, game_runtime.go, defer.go, analyze.go
internal/parser/stmt.go, decls.go, expr.go
internal/compiler/pipeline.go
examples/raylib/commands.dd    ← full friendly API reference
examples/raylib/control_flow.dd ← loop, defer, match
```

## Gotchas

- Default `datadream check` is parse-only — use `check --codegen` or `build`
- Global `let x = sprite(...)` defers to `datadream_init_globals()` — never emit calls at C file scope
- Build `coin-runner` from `examples/coin-runner/` (asset paths are relative)
- Windows needs bundled llvm-mingw (`sdk install clang`), not MSVC LLVM alone
- `internal/pkg` must NOT import `internal/sdk`

## Definition of done for your session

- [ ] Chosen track task completed with acceptance criteria met
- [ ] `go test ./internal/...` passes
- [ ] New/changed example passes `check --codegen` (and `build` if applicable)
- [ ] `docs/LANGUAGE.md` and `docs/HANDOFF.md` updated if behavior changed

Start by reading `docs/HANDOFF.md`, running the verify commands, then implementing **P1.1 (audio_demo)** unless I specify a different track.
