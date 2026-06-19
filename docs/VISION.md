# DataDream Vision

This document defines what we are building and what we are **not** building. Read it before adding features.

## One-line pitch

**C/raylib structure, BASIC-level ease** — a native language for games and apps that compiles to fast C, with full library access when you want it.

> DataDream is what raylib would feel like if it had its own modern language.

Not Python-like. Not JavaScript-like. **C-style braces, semicolons, `fn main()`, `use raylib;`, struct literals** — plus optional friendly `app` / `draw` sugar on top.

## Core goals

### 1. Beginner-friendly surface, serious engine underneath

Beginners write:

```dd
app "My Game";

window { size: 800, 600; title: "Hello"; }

draw {
    clear(colors.black);
    draw.text("Hello", { position: vec2(300, 280), size: 32, color: colors.white });
}
```

Advanced users drop to raw C interop when needed:

```dd
use raylib;

fn main() {
    InitWindow(800, 600, "Raw");
    // full raylib 6.0 API
}
```

Both paths must work. Friendly APIs compile down to raylib — they are not a separate runtime VM.

### 2. Native performance via C

- DataDream → C → Clang → native binary.
- No interpreter, no bytecode VM, no hidden JIT for game logic.
- Performance-sensitive code can call C libraries directly.

### 3. Frictionless C library access

- `datadream bind header.h` generates DataDream bindings.
- `use raylib;` without alias brings C names into scope (`InitWindow`, not `raylib.InitWindow`).
- Goal: wrap **entire** libraries (raylib first), not a curated subset.

### 4. Modular compiler

Each stage is swappable (`lexer`, `parser`, `codegen`, `driver`). New passes (typecheck, lint) plug into `internal/compiler.Pipeline` without rewriting everything.

### 5. Self-contained distribution

End users get:

```
bin/datadream[.exe]
sdk/toolchain/clang/
sdk/raylib/6.0/
examples/
libs/
```

No Go. No system package manager required. `datadream doctor` verifies the layout.

### 6. Modern ergonomics

- Type inference (`let x = 10;`)
- Structs, enums, modules
- `colors.*`, hex colors (`#RRGGBB`), `rgb()`, `rgba()`, `hsl()`, CSS names
- String interpolation (`"Score: {score}"`)
- Config blocks (`window { }`) with semicolons; option objects with commas
- Namespaced drawing (`draw.text`, not sentence commands like `text "Hi" at 10, 20`)

## Anti-goals (do not drift here)

| Do NOT | Why |
|--------|-----|
| Require Go for end users | Compiler ships as a native binary |
| Use Python in tooling | Bindgen, packager, SDK install are Go |
| Build a custom VM / bytecode | We compile to C |
| Hide raylib behind an opaque runtime | Users must reach `raylib.h` when they want |
| Invent English-sentence syntax | `draw.text("Hi", { position: vec2(...) })` not `text "Hi" at 10, 20 size 32` |
| Ship half-finished syntax in examples | Examples must `check` and ideally `build` |
| Rewrite everything in Rust/C++ “eventually” | Compiler stays Go until there is a concrete reason |
| Add an editor/IDE before the language works | Roadmap layer 5 — not now |

## Syntax principles

1. **Config blocks** use semicolons: `window { size: 800, 600; title: "Hi"; }`
2. **Option objects** use commas: `{ position: vec2(0, 0), size: 32, color: colors.white }`
3. **Drawing** uses namespaces: `draw.text`, `draw.rect`, `draw.sprite`
4. **Lifecycle** uses blocks: `start { }`, `update { }`, `draw { }`
5. **Imports**: `use raylib;`, `use raylib as rl;`, `using raylib;`

## Roadmap layers (order matters)

Build bottom-up per [DESIGN.md](DESIGN.md). Do not skip layers.

1. **Raw language + raylib** — `fn main`, `use raylib`, struct literals, 2D/3D raw demos ✅ mostly done
2. **C interop + bindgen** — `extern c`, `datadream bind`, package linking ✅ mostly done
3. **Quality of life** — colors, vec2/3, interpolation, operators, **loop/defer/match**, errors, fmt 🟡 in progress
4. **Friendly wrappers** — app/window/draw, input, screen, sprites, random, math, time, audio/assets partial 🟡 2D works; ui.* pending
5. **Engine sugar** — scenes, entities, raygui `ui.*`, editor ❌ after Layer 4 solid on all platforms

## Success criteria for “language complete” (v1)

- [x] `datadream build` works on Windows with bundled SDK
- [ ] `datadream build` works on Linux/macOS with bundled SDK (CI verifies Linux + macOS; local `sdk install` still required for offline use)
- [x] Friendly app examples build on Windows (`hello_friendly`, `coin-runner`)
- [x] Raw raylib examples build on Windows (`hello_raw`, `hello_using`)
- [x] Color system works in friendly and raw modes
- [ ] `datadream check` catches all codegen errors by default (today: `--codegen` flag)
- [x] `datadream bind raylib.h --raw` regenerates `libs/raylib/raw.dd`
- [x] Multi-file programs via `include "file.dd";`
- [x] Core control flow: `loop`, `defer`, `match` ([SYNTAX.md](SYNTAX.md))
- [x] Type checker pass (at least for builtins and struct fields)
- [ ] Entity/scene lifecycle generates correct C
- [x] `datadream sdk install clang` populates bundled toolchain

## Branding

| Item | Value |
|------|-------|
| Language name | DataDream |
| File extension | `.dd` |
| CLI binary | `datadream` |
| Go module | `datadream` |
| Bundled raylib | **6.0** (latest stable) |

Former name: **Koda** — do not use in new code or docs.
