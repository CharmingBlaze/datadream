# Prompt for Next Programmer

Copy **everything below the horizontal rule** into a new chat, ticket, or agent session.

---

## Mission

You are continuing **DataDream** — a Go compiler that turns `.dd` source into **C**, links with **bundled Clang + raylib 6.0**, and ships as a **single zip per platform** (no Go required for end users).

**Current state:** ~**99% v1-complete**. All **25** examples build; compiler + **DataDream Studio** IDE on `main`. **Ship gate:** publish GitHub Release v1.0.0 ([docs/PUBLISH.md](docs/PUBLISH.md)) and smoke-test Studio launchers from the zip.

**End-user path:** unzip → double-click **`Start DataDream Studio.bat`** (Windows) or **`DataDream Studio.app`** (macOS) → press Ctrl+Enter to run `clicker.dd`. See `docs/GETTING_STARTED.txt`.

**Identity:** C/raylib structure, modern syntax — **not** Python, not a VM, not English-sentence commands.

**Repo:** https://github.com/CharmingBlaze/datadream · Go module: `datadream`

---

## Read first (20 minutes)

| # | File | Why |
|---|------|-----|
| 1 | `docs/HANDOFF.md` | Full state, file map, gotchas, examples list, **remaining work** |
| 2 | `docs/VISION.md` | Anti-goals — do not violate |
| 3 | `docs/ROADMAP.md` | Task status + Definition of done |
| 4 | `docs/STUDIO.md` | Desktop IDE layout and build |
| 5 | `docs/LANGUAGE.md` | User-facing syntax (arrays, for-in, memory model) |
| 6 | `docs/ARCHITECTURE.md` | Pipeline: lex → parse → typecheck → codegen → driver |

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
# Then: unzip and double-click Start DataDream Studio.bat
```

Expected: all tests pass; examples check clean; Windows zip cold-installs with bundled Clang; Studio opens and runs a sample.

---

## What is already done — do NOT redo

### Shipping & tooling
- CI on Win/Linux/macOS: test, build examples, bindgen-check, dist-verify (**`--skip-studio`** in CI)
- Release workflow builds **Studio + Monaco + launchers** (`.github/workflows/release.yml`)
- `datadream new` scaffold, VS Code grammar + minimal extension
- Lex/parse/typecheck/codegen diagnostics with hints; **warnings** are non-blocking
- **DataDream Studio** — Wails desktop (`cmd/studio/`), web IDE (`datadream ide`), embedded Monaco

### Language (Layers 1–5)
- Control flow: `loop`, `defer`, `match`, `for i in 0..=n`, struct destructuring in `match`
- **`Array<T>` / `list T`** → `DD_Array` runtime; `.push`, `.len`, `.pop`, `.remove`, `.remove_dead`
- **`for x in …` disambiguation in type checker** (`IterKind`): entity / array / string
- **`for ch in "hello"`** — byte iteration only; **`c` is a reserved keyword** — use `ch`
- ECS: `app`, `entity`, `scene`, `spawn`, `destroy`, `system`, `for x in Entity`
- **`export fn` / `export let`**, selective **`use raylib { … }`**
- Frame + level arenas, `@packed`, `@save`
- Full Layer 4 + raw raylib

### Examples
25 `.dd` files under `examples/` — all pass `datadream check --codegen`.

---

## Definition of done — v1.0

1. Fresh zip (Win/Linux/macOS) → `doctor` ✓ → `run hello_friendly` — **verified via local + CI dist-verify**
2. `coin-runner` builds and runs with assets
3. All examples pass `check --codegen` and `build`
4. Docs match reality
5. **GitHub Release v1.0 published** with three platform zips ← **YOU DO THIS FIRST**
6. **Studio launcher smoke-tested** from at least one release zip (Windows recommended)

---

## Your work — strict priority order

### Phase 1 — Publish v1.0 (P0) ← START HERE

| Task | How |
|------|-----|
| Run release workflow | `docs/PUBLISH.md` — `gh auth login`, Actions → Release → tag `v1.0.0` |
| Fix CI if Wails fails | Linux: GTK/WebKit deps in workflow; macOS: may need ImageMagick; gate `-webview2 embed` to Windows in `build-studio.sh` |
| Verify release assets | Three zips; each contains Studio + `GETTING_STARTED.txt` + root launcher |
| Studio smoke test | Unzip → double-click launcher → Ctrl+Enter on `clicker.dd` |
| Sign off DoD | Check every item in ROADMAP § Definition of done |

### Phase 2 — Turnkey UX (P0.5)

| Task | Notes |
|------|-------|
| README Studio-first quick start | Root README still CLI + PATH |
| Release notes Studio-first | `release.yml` publish body |
| `verify-dist` Studio check | Assert `bin/datadream-studio*` when not `--skip-studio` |

### Phase 3 — Language / tooling (P2 / P4)

| Task | Notes |
|------|-------|
| LSP (hover, go-to-def) | v1.1 |
| Linux AppImage | ✅ | `build-studio-appimage.sh` bundles GTK/WebKit |
| Parser hint edge cases | Low priority |

---

## Key files

```
cmd/datadream/main.go
cmd/studio/                      Wails IDE — wails build only
internal/ide/service.go          Shared IDE backend
internal/ide/web/app.js          Wails bridge + Monaco loader
internal/cli/studio.go           datadream studio launcher
tools/packdist/main.go           Zips + GETTING_STARTED + launchers
scripts/build-dist.ps1           Full release (Monaco + Studio + Clang)
.github/workflows/release.yml    Publishes zips with Studio
docs/PUBLISH.md                  Release instructions
docs/GETTING_STARTED.txt         End-user one-pager
```

---

## Rules

1. **Go only** in tooling — no Python
2. **Every language feature** → example passing `datadream check --codegen`
3. **Small focused diffs** — one feature or fix per change
4. **Update docs** when status changes: `HANDOFF.md`, `ROADMAP.md`
5. **Do not commit** unless the user explicitly asks

---

## Gotchas

| Issue | Fix |
|-------|-----|
| `go build ./cmd/studio` fails | Use `scripts/build-studio.*` / `wails build` (needs `desktop,production` tags) |
| CI has no Studio | Expected — `--skip-studio` in dist-verify; full IDE in release workflow only |
| Linux Studio won't start | Use the AppImage from the zip (`datadream-studio-x86_64.AppImage`); no GTK/WebKit install needed |
| `for c in "hello"` parse error | `c` is C-interop keyword — use `ch` |
| `.gitignore` | Must be `/datadream` not `datadream` |

---

## Start command

1. Read `docs/HANDOFF.md` § **Your mission**
2. Run verify commands above
3. Execute **Phase 1** (publish v1.0 Release + Studio smoke test) unless told otherwise

**North star:** A beginner downloads one zip, double-clicks **DataDream Studio**, sees `clicker.dd`, presses Ctrl+Enter, and gets a window — no Go, no PATH, no internet.
