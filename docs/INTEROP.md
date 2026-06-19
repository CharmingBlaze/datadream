# C / raylib Interop

How DataDream talks to C libraries, with **raylib 6.0** as the reference integration.

---

## Two layers

```
┌─────────────────────────────────────┐
│  Friendly API                       │
│  draw.text, sprite, colors.*, input │
└─────────────────┬───────────────────┘
                  │ compiles to
┌─────────────────▼───────────────────┐
│  Generated C + game runtime helpers   │
└─────────────────┬───────────────────┘
                  │ calls
┌─────────────────▼───────────────────┐
│  raylib.h (full API)                │
│  libs/raylib/raw.dd                 │
└─────────────────────────────────────┘
```

Advanced users skip the friendly layer and call raylib directly.

---

## Import styles

### Plain import (recommended for raw games)

```dd
use raylib;

fn main() {
    InitWindow(800, 600, "Game");
    DrawText("Hello", 100, 100, 20, WHITE);
    CloseWindow();
}
```

Names land in file scope — no `raylib.` prefix.

### Namespaced import

```dd
use raylib as rl;

fn main() {
    rl.InitWindow(800, 600, "Game");
}
```

### Using declaration

```dd
using raylib;

fn main() {
    InitWindow(800, 600, "Game");
}
```

### App mode (no import needed)

```dd
app "Game";
window { size: 800, 600; title: "Game"; }
draw { clear(colors.black); draw.text("Hi", { ... }); }
```

Raylib is auto-linked when `app` + `window` + `draw` are present.

---

## Raw bindings file

`libs/raylib/raw.dd` — ~548 functions from `raylib.h`.

Regenerate:

```bash
datadream bind sdk/raylib/6.0/include/raylib.h \
  --module raylib --raw --out libs/raylib/raw.dd

datadream check --codegen libs/raylib/raw.dd
```

Bindgen: `tools/bindgen/` (Go only).

---

## SDK: headers, libs, Clang

| Path | Contents |
|------|----------|
| `sdk/toolchain/clang/bin/clang[.exe]` | Bundled compiler |
| `sdk/raylib/6.0/include/` | `raylib.h`, `raymath.h`, `rlgl.h` |
| `sdk/raylib/6.0/lib/windows-amd64/` | `libraylib.a`, `raylib.dll` |
| `vendor/raylib/include/` | Dev fallback headers |

Install:

```bash
datadream sdk install clang
datadream sdk install headers
datadream sdk install raylib
datadream doctor
```

---

## Link metadata

`libs/raylib/package.json` — platform libs list. Runtime flags: `internal/sdk/link.go` → `RaylibLinkLibs()`.

Windows: passes **full path** to `libraylib.a` (MinGW format).

---

## Friendly API → raylib mapping

| DataDream | C (raylib) |
|-----------|------------|
| `clear(colors.black)` | `ClearBackground(...)` |
| `draw.text(...)` | `DrawText(...)` |
| `draw.rect({...})` | `DrawRectangle(...)` |
| `draw.circle({...})` | `DrawCircle(...)` |
| `draw.line({...})` | `DrawLine(...)` |
| `draw.sprite(s)` | `DrawTextureEx(...)` or fallback rect |
| `window { size: W, H }` | `InitWindow(W, H, title)` |
| App main loop | `while (!WindowShouldClose())` … |

Game helpers: `codegen/game_runtime.go`.

---

## Windows linking

raylib 6.0 official prebuilts = **MinGW** (`libraylib.a`).

| Setup | Result |
|-------|--------|
| `datadream sdk install clang` (llvm-mingw) | ✅ Recommended |
| System LLVM + MinGW `-target` (auto) | ✅ Fallback if doctor shows compatible |
| MSVC-targeting LLVM alone | ❌ `__mingw_*` unresolved symbols |

---

## Color interop

DataDream `Color` → raylib `Color` `{ r, g, b, a }` (u8).

- `colors.white` → compile-time struct literal
- `#FF0000` → `ColorLit`
- `rgba(255,0,0,0.5)` → `datadream_rgba(...)` helper

CSS names in Go: `internal/colors/namespace.go`.

---

## Binding a new C library

```bash
datadream bind path/to/SDL.h --lib sdl2 --raw --out libs/sdl2/raw.dd
```

Add `libs/sdl2/package.json`, then `use sdl2;` in source.

---

## Generated bindings (`raw.dd`)

`libs/raylib/raw.dd` is produced by bindgen. The parser accepts C-isms needed for bindings:

- Field and parameter names that match DataDream keywords (`shader: Shader`, `data: void?`)
- Array field types (`float[]`, `char[]`)
- C float literals (`180.0f`)

If bindgen output changes, run:

```bash
datadream bind sdk/raylib/6.0/include/raylib.h --raw --out libs/raylib/raw.dd
./scripts/check-bindgen.sh   # CI runs this as bindgen-check
go test ./internal/parser/... -run TestParseRaylibRawBindings
```

Parse should finish quickly. See `lexer.IsBindingName` and `parser.parseBindingIdent`.

---

## Version policy

- Ship **raylib 6.0**
- Single source: `internal/sdk/version.go` → `RaylibVersion = "6.0"`
- Do not track raylib master / 6.1-dev in production SDK

---

## Examples

| File | Style |
|------|-------|
| `examples/raylib/hello_friendly.dd` | App + friendly draw |
| `examples/raylib/hello_raw.dd` | `use raylib as rl` |
| `examples/raylib/hello_using.dd` | `using raylib` |
| `examples/colors/alpha.dd` | Raw fn main + hex/rgba |
| `examples/coin-runner/game.dd` | Friendly game (auto raylib) |
