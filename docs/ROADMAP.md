# Roadmap

Prioritized work to complete DataDream v1. **Do not skip open P0 items.**

Status key: ✅ done · 🔄 in progress · 🟡 partial · ❌ not started

---

## P0 — Must work before v1 ship

| # | Task | Status | Notes |
|---|------|--------|-------|
| 1 | Bundled MinGW Clang on Windows | ✅ | `datadream sdk install clang` → llvm-mingw |
| 2 | `datadream build hello_friendly.dd` on Windows | ✅ | Verified with bundled SDK |
| 3 | `datadream build coin-runner` on Windows | ✅ | Assets + sprite runtime fixes |
| 4 | `datadream doctor` toolchain reporting | ✅ | MinGW / MSVC + compatibility |
| 5 | `datadream check --codegen` | ✅ | Opt-in codegen without link |
| 6 | coin-runner placeholder assets | ✅ | `examples/coin-runner/assets/*.png` |
| 7 | **`build` on Linux amd64** | ✅ | CI + `-lGL` link fix, doctor fix |
| 8 | **`build` on macOS arm64/amd64** | ✅ | CI macos-latest job |
| 9 | **Release zip per platform** | ✅ | `build-dist.sh/ps1`, `packdist --verify`, `release.yml` |

---

## P1 — Complete the friendly game layer

| # | Task | Status | Files |
|---|------|--------|-------|
| 10 | `draw.rect` → raylib | ✅ | `codegen/draw.go` |
| 11 | `draw.circle` → raylib | ✅ | `codegen/draw.go` |
| 12 | `draw.line` → raylib | ✅ | `codegen/draw.go` |
| 13 | `examples/raylib/features.dd` living spec | ✅ | draw + input reference |
| 14 | Conditional runtime emission | ✅ | `analyze.go`, `game_runtime.go` |
| 15 | `input.pressed` / key map expansion | ✅ | `game_runtime.go` — down/released + more keys |
| 16 | Audio friendly API (`audio.play`) | ✅ | `stdlib_runtime.go`, `audio_demo.dd` |
| 17 | `assets.texture` / `assets.sound` | ✅ | `stdlib.go`, `audio_demo.dd` |
| 18 | **`ui.*` raygui wrapper** | ✅ | `ui.go`, `ui_demo.dd` |
| 19 | **`loop` / `defer` / `match`** | ✅ | `defer.go`, `stmts.go`, [SYNTAX.md](SYNTAX.md) |
| 20 | **`docs/SYNTAX.md`** keyword reference | ✅ | layers 1–5 documented |
| 21 | **`examples/raylib/commands.dd`** | ✅ | all friendly namespaces |
| 22 | Parser: `raw.dd` without hang | ✅ | `IsBindingName`, `float[]`, `180.0f` |

---

## P2 — Language correctness

| # | Task | Status | Files |
|---|------|--------|-------|
| 22 | Type checker `Phase` | ✅ | `internal/typecheck/`, default in `check` |
| 23 | Better error messages everywhere | 🟡 | `internal/errors/`, hints for typecheck ✅ + lex/parse ✅; codegen diagnostics have file/line/hint ✅ |
| 24 | Module system (`use graphics` → `libs/`) | ✅ | `compiler/modules.go`, `pkg/resolver.go` |
| 25 | `include` → proper modules with exports | ✅ | `export fn` / `export let`; `filterModuleExports` in `modules_export.go` |
| 26 | Struct method calls on game types | ✅ | `codegen/decls.go`, `codegen/expr.go` |
| 27 | **`Array<T>` / `list T` + `DD_Array`** | ✅ | `codegen/array.go`, `typecheck/forin.go` |
| 28 | **`for x in array` (IterKind disambiguation)** | ✅ | `array_for_in.dd`, `array_demo.dd` |
| 29 | **String `for ch in "hello"` (byte iter)** | ✅ | `string_for_in.dd` |
| 30 | **Warning: `.remove()` during for-in** | ✅ | `forin.go`, `array_remove_warning.dd` |
| 31 | **`match` struct destructuring** | ✅ | `match_destruct.dd` |
| 32 | **`defer` on return/break/continue** | ✅ | `defer.go` |
| 33 | **`@packed` / `@save` attributes** | ✅ | `attrs.go`, `ecs_packed.go`, `attributes_demo.dd` |
| 34 | Frame + level arena runtime | ✅ | `codegen/arena.go`, LANGUAGE.md |

---

## P3 — Entity / scene / ECS

| # | Task | Status | Files |
|---|------|--------|-------|
| 27 | `entity` codegen in game loop | ✅ | `codegen/ecs.go`, `decls.go` |
| 28 | `scene` init/start/update/draw | ✅ | `codegen/decls.go`, `app_emit.go` |
| 29 | `spawn` / `destroy` runtime | ✅ | `codegen/ecs.go`, `stmts.go` |
| 30 | `for x in Entity` iteration | ✅ | `codegen/ecs.go`, `stmts.go` |
| 31 | `system` blocks | ✅ | `codegen/decls.go`, `app_emit.go` |
| 32 | `on key` / event handlers | ✅ | `codegen/ecs.go`, `stmts.go` |
| 33 | Entity `draw { }` blocks | ✅ | `codegen/decls.go`, `app_emit.go` |

---

## P4 — Tooling & distribution

| # | Task | Status | Files |
|---|------|--------|-------|
| 34 | `datadream sdk install clang` | ✅ | `internal/sdk/fetch_clang.go` |
| 35 | CI: test + build + dist verify | ✅ | `.github/workflows/ci.yml` (dist-verify Linux/macOS/Windows) |
| 36 | Release zips per platform | ✅ | `scripts/build-dist.*`, `packdist --verify`, `.github/workflows/release.yml` |
| 37 | Regenerate `raw.dd` in CI | ✅ | `scripts/check-bindgen.*`, CI `bindgen-check` |
| 38 | LSP / VS Code extension | 🟡 | TextMate grammar + minimal extension in `editor/datadream/`; LSP later |
| 39 | `datadream new` project scaffold | ✅ | `internal/cli/new.go` — `game.dd`, `assets/`, README |
| 40 | **DataDream Studio (Wails IDE)** | ✅ | `cmd/studio/`, `internal/ide/`, embedded Monaco, Wails bridge |
| 41 | **`datadream ide` web IDE** | ✅ | `internal/ide/server.go`, port 3847 |
| 42 | **Turnkey zip launchers** | ✅ | `packdist` → `GETTING_STARTED.txt`, `Start DataDream Studio.bat`, etc. |
| 43 | **Offline Monaco vendor** | ✅ | `scripts/fetch-monaco.*`, `internal/ide/web/vendor/` |
| 44 | **Release workflow includes Studio** | ✅ | `release.yml`; CI dist-verify uses `--skip-studio` |
| 45 | **Studio smoke test in verify-dist** | ❌ | Manual only today |
| 46 | **README / release notes Studio-first** | 🟡 | Still CLI-first in README and release body |

---

## P5 — Future (do not start yet)

- Visual editor / inspector
- Live reload
- WebAssembly target
- UI declarative blocks (`ui { }`)
- Hot reload assets
- Package registry / `datadream install`

---

## Completed (June 2026)

- ✅ `datadream sdk install clang` (llvm-mingw Windows, LLVM Linux/macOS)
- ✅ Windows end-to-end builds (hello_friendly, coin-runner, hello_raw)
- ✅ Sprite runtime: global init, texture fallback rects, `s->path` fix
- ✅ Windows MinGW linking: bundled Clang + `-target` fallback for system LLVM
- ✅ `datadream check --codegen`
- ✅ Conditional game runtime emission (sprite/input/collision/random)
- ✅ `input.down`, `input.released`, expanded key names
- ✅ Scene `start`/`update`/`draw` wired into app main loop
- ✅ `draw.rect` accepts `width`/`height` options
- ✅ Rebrand Koda → DataDream
- ✅ Modular compiler pipeline
- ✅ Friendly `window { }`, `draw.text`, `clear`, colors
- ✅ Full color system (hex, rgb, hsl, CSS, methods)
- ✅ Game runtime (sprite, input, collision, random, dt)
- ✅ Auto-raylib for app programs
- ✅ `use raylib` plain import
- ✅ Bindgen + `libs/raylib/raw.dd`
- ✅ SDK layout + `doctor` + `sdk install raylib`
- ✅ No Go in end-user distribution model
- ✅ `loop`, `defer`, `match`, `break`, `continue` — core control flow
- ✅ `docs/SYNTAX.md` — keywords/operators/literals by layer (Layer 5 = app sugar)
- ✅ `examples/raylib/commands.dd` — all friendly namespace commands
- ✅ Parser: `raw.dd` binding names (`shader`, `data`), `float[]`, C float suffix
- ✅ Beginner APIs — `screen.*`, mouse, `random.*`, `math.*`, `time.*`, `quit()`
- ✅ `examples/beginner/clicker.dd` tutorial
- ✅ Bindgen drift check in CI (`scripts/check-bindgen.*`, `bindgen-check` job)
- ✅ CI dist-verify on Windows, Linux, macOS (`build-dist.*`, `packdist --verify` includes `hello_raw`)
- ✅ All 17 examples build; CI runs `scripts/build-all-examples.sh` on Linux + macOS
- ✅ Fixed `let bg = colors.*` type inference (`inferTypeFromExpr` → `Color`)
- ✅ `hello_3d.dd` v1 target with `defer CloseWindow()`
- ✅ `cmd/datadream/main.go` — CLI entrypoint (was missing from initial commit)
- ✅ `.gitignore` fix — `/datadream` so `cmd/datadream/` is tracked
- ✅ Windows cold-install verified: bundled Clang zip → doctor → run hello_friendly (no system compiler on PATH)
- ✅ Type-check error hints (unknown id, bad draw.*, struct fields, arg counts)
- ✅ Entity/struct method calls (`self.boost()` → `Type_boost(self)`)
- ✅ Removed legacy `koda-linux-amd64` from repo
- ✅ `datadream new` scaffold (`internal/cli/new.go`)
- ✅ TextMate grammar + minimal VS Code extension (`syntaxes/`, `editor/datadream/`)
- ✅ Release workflow publishes GitHub Release with Win/Linux/macOS zips
- ✅ Lex/parse diagnostics with file, line, column, and hints (`compiler/diagnostics.go`)
- ✅ Frame arena runtime for app programs (`codegen/arena.go`)
- ✅ `export fn` / `export let` module boundaries (`exports_module.dd`, `libs/demoexports/`)
- ✅ `match` struct destructuring (`match_destruct.dd`)
- ✅ `@save` / `@packed` attribute syntax + codegen stubs (`attributes_demo.dd`)
- ✅ Generic `for x in array` — `DD_Array` runtime, `IterKind` disambiguation (`array_for_in.dd`, `array_demo.dd`)
- ✅ Level arena runtime (`dd_level_arena_reset` on scene init)
- ✅ Publish guide ([PUBLISH.md](PUBLISH.md))
- ✅ String `for ch in "hello"` byte iteration (`string_for_in.dd`)
- ✅ `list T` sugar for `Array<T>`
- ✅ Compile-time warning for `.remove()` during array iteration
- ✅ Frame-budget loop protection — static analysis, entity pools, `@max`, debug while guards ([LOOPS.md](LOOPS.md))
- ✅ Typecheck warnings (non-blocking) wired through CLI
- ✅ `@packed` SoA entity layout (`ecs_packed.go`, `BulletPool` parallel arrays)
- ✅ `@save` binary serialize/deserialize (`fwrite`/`fread` for int/float/bool/Vec/string)
- ✅ Codegen diagnostics with file, line, column, and hints (`codegen/diagnostic.go`)
- ✅ Array `.push()` portable C (no GNU statement expressions)
- ✅ Match struct pattern literals (`v.x == 0.0f`)
- ✅ Selective `use raylib { InitWindow, … }` with typecheck whitelist
- ✅ Fixed `isRaylibSymbol` for PascalCase API names (`InitWindow`, etc.)
- ✅ Windows dist zip built + cold-install verified locally
- ✅ **DataDream Studio** — Wails desktop IDE, web IDE, embedded Monaco, turnkey launchers
- ✅ **Linux Studio AppImage** — self-contained GTK/WebKit via `build-studio-appimage.sh`
- ✅ Selective `use raylib { … }` (`hello_selective.dd`)
- ✅ Code pushed to `main` (June 2026)

---

## Definition of done — v1.0

Ship when **all** are true:

1. Fresh install from zip → `doctor` ✓ → `run hello_friendly` works (Win/Linux/macOS) — **Windows verified locally; Linux/macOS verified via CI dist-verify (June 2026)**
2. `coin-runner` builds and runs with assets
3. `hello_raw` builds with full raylib API
4. All `examples/` pass `check --codegen` and `build`
5. Docs in `docs/` match reality
6. Bindgen regenerates raylib without manual edits
7. GitHub release zips published for all three platforms — **workflow ready**; run [PUBLISH.md](PUBLISH.md) to attach assets to v1.0.0
8. **Studio launcher works from release zip** — double-click → run sample (manual smoke test until verify-dist extended)

---

## Remaining after v1.0 ship

| Priority | Task | Status |
|----------|------|--------|
| P0 | Publish v1.0 GitHub Release | 🔄 |
| P0 | Studio smoke test from release zip | 🔄 |
| P0.5 | README + release notes Studio-first | 🟡 |
| P0.5 | `verify-dist` Studio binary check | ❌ |
| P2 | Parser/lexer hint edge cases | 🟡 |
| P4 | LSP (hover, go-to-def) | ❌ |
| P5 | `match` type patterns (beyond struct fields) | ❌ |

---

## How to update this file

When you finish a task:

1. Change status in the table
2. Add a line under **Completed** with date
3. If scope changed, update [VISION.md](VISION.md) first
