# DataDream — Language Design Map

**Canonical design reference.** For day-to-day syntax, see [LANGUAGE.md](LANGUAGE.md). For every keyword and operator by category, see [SYNTAX.md](SYNTAX.md). For goals and anti-goals, see [VISION.md](VISION.md).

---

## Identity

**DataDream is what raylib would feel like if it had its own modern language.**

| DataDream is | DataDream is not |
|--------------|------------------|
| C-style structure, BASIC-level ease | Python-like indentation |
| raylib-first, C-library-wrapper-first | A web scripting language first |
| game/app focused, fast native output | A JVM language or Python replacement |
| beginner-friendly but not toy-like | A toy BASIC clone or full C++ clone |

**Best one-liner:** A native app and game language built around C library power, raylib-style simplicity, modern syntax, and beginner-friendly tooling.

**Good for:** 2D/3D games, desktop apps, tools, editors, creative coding, game engines, small utilities, visual-scripting backends, C library wrapping.

---

## Syntax style

```
C-style braces · semicolons · fn functions · let variables
struct literals · use modules · native C interop · raw library access · friendly wrappers
```

**Not:** Python indentation · BASIC line numbers · C++ `#include`

```dd
use raylib;

fn main() {
    let x = 100;
    let y = 200;
    DrawCircle(x, y, 40.0, RED);
}
```

---

## Files and entry points

| Item | Value |
|------|--------|
| Language name | **DataDream** |
| Extension | **`.dd`** (also noted: `.dream` — `.dd` is official) |
| Raw entry | `fn main() { }` — **essential**, C-library-first |
| Friendly entry | `app "Title";` + `window { }` + `start` / `update` / `draw` |

Build order: **raw raylib first → C interop → colors/vectors → raygui → friendly wrappers → editor later.**

---

## Imports (no `#include`)

Official style:

```dd
use raylib;                    // names in scope: InitWindow, DrawCube, …
use raylib as rl;              // rl.InitWindow(...)
using raylib;                  // same as plain use for name injection
```

Planned (not v1):

```dd
use raylib { InitWindow, DrawText, WHITE };   // selective
import raylib;                                 // module-only: raylib.InitWindow
```

`#include` is text pasting. DataDream uses real modules + bindgen.

---

## Separators (critical rule)

| Construct | Separator |
|-----------|-----------|
| Statements | `;` |
| Config blocks (`window { }`) | `;` between properties |
| Struct / object literal fields | `,` |
| Option objects (`draw.text` opts) | `,` |

```dd
let camera = Camera3D {
    position: vec3(4.0, 4.0, 4.0),
    target: vec3(0.0, 0.0, 0.0),
    projection: CAMERA_PERSPECTIVE
};   // semicolon ends the let statement
```

---

## Variables and types

```dd
let score = 0;              // inferred, mutable (games default)
let speed: float = 4.5;     // explicit type allowed
const gravity = 9.8;        // immutable (keyword exists; codegen partial)
```

**Core types:** `bool`, `int`, `float`, `double`, `string`, `char`, `byte`  
**Exact types:** `i8`…`i64`, `u8`…`u64`, `f32`, `f64`, `usize`, `isize`, `cstring`, `voidptr`, `ptr<T>`

**Aliases:** `int` = `i32`, `float` = `f32`, `double` = `f64`, `byte` = `u8`

---

## Control flow

```dd
if health <= 0 { die(); }           // parens optional
while !WindowShouldClose() { }     // ! required (C-style)
loop { break; continue; }          // infinite loop — cleaner than while true

for i in 0..10 { }                  // 0..9 (exclusive end)
for i in 0..=10 { }                 // 0..10 (inclusive)

for enemy in enemies { }
for i, enemy in enemies { }         // planned

match state {
    0 => { }
    _ => { }                        // or else =>
}

defer UnloadTexture(tex);           // runs on block exit (LIFO)
break; continue;
```

Full keyword/operator tables: [SYNTAX.md](SYNTAX.md).

**Operators:** `+ - * / %`, `== != < <= > >=`, `&& || !`, also `and` / `or` / `not`, `&&` / `||`

---

## Structs, enums, methods

```dd
struct Player {
    position: Vector3;
    health: int = 100;

    fn damage(amount: int) {
        health -= amount;
    }
}

let player = Player {
    position: vec3(0.0, 1.0, 0.0),
    health: 100
};
```

Raylib C types work as struct literals: `Camera3D { … }`, `Rectangle { x: 20, y: 20, width: 160, height: 40 }`.

Enums: DataDream enums + C bindings via `extern c { enum … }`. Raylib constants (`CAMERA_PERSPECTIVE`, `RAYWHITE`) import with `use raylib`.

---

## Strings and colors

```dd
let title = "DataDream 3D";
DrawText("Score: {score}", 10, 10, 20, WHITE);   // interpolation in friendly + print

ClearBackground(#101018);
DrawCube(vec3(0,1,0), 2,2,2, colors.rebeccaPurple);
DrawCube(..., RED);                               // raylib constants
```

Forms: `#RGB` `#RGBA` `#RRGGBB` `#RRGGBBAA`, `rgb()` `rgba()` `hsl()`, `css("rebeccapurple")`, `colors.*`

---

## Vectors and math

```dd
vec2(10, 20)    vec3(0, 1, 0)    vec4(1, 0, 0, 1)
distance(a, b)  length(v)  clamp  lerp  min  max  abs
```

Vec2 `+=`, `* scalar` work in friendly mode. Full vec3 ops + methods planned.

---

## C interop and wrapping

```dd
extern c {
    link "raylib";
    fn InitWindow(width: int, height: int, title: cstring);
}
```

```bash
datadream bind sdk/raylib/6.0/include/raylib.h --raw --out libs/raylib/raw.dd
```

Package metadata: `libs/raylib/package.json` (platform libs, headers). No `#include` in user code.

---

## Friendly layers (compile down to raylib)

### Lifecycle sugar

```dd
app "My Game";
window { size: 800, 600; title: "My Game"; fps: 60; }
start { }  update { }  draw { clear(#000); draw.text("Hi", { … }); }
```

Expands to `InitWindow` + `while (!WindowShouldClose())` + `BeginDrawing` / `EndDrawing`.

### Namespaces (today)

| Namespace | Examples |
|-----------|----------|
| `draw.*` | `text`, `rect`, `rectOutline`, `circle`, `circleOutline`, `line`, `ellipse`, `triangle`, `point`, `sprite`, `fps` |
| `input.*` | `move2d`, `move3d`, `axis`, `pressed`, `down`, `released`, `mouse`, `mousePressed`, `mouseDown`, `scroll` |
| `screen.*` | `width`, `height`, `center`, `size` |
| `keys.*` | `w`, `space`, `escape`, arrows, `f1`… — for use with `input.*` |
| `random.*` | `int`, `float`, `point`, `screenPosition` |
| `collision.*` | `overlap`, `contains`, `pointInRect` |
| `time.*` | `fps`, `now`, `frame` |
| `math.*` | `dot`, `cross`, `normalize`, `length`, `distance`, `clamp`, `lerp` |
| `audio.*` | `play`, `stop` |
| `assets.*` | `texture`/`image`, `sound` |
| `colors.*` | CSS + raylib palette |

Raw calls always available after `use raylib`: `DrawText`, `BeginMode3D`, `GuiButton`, …

### Planned (Layer 4–5)

- `ui.button()` (raygui wrapper) — **Layer 5**
- `scene` / `entity` / `component` / `system` (engine sugar — **Layer 5**, after friendly 2D solid)
- `project.dd` / `dream.toml`, `datadream add`, cross-target builds

Keyword and operator reference: [SYNTAX.md](SYNTAX.md).

---

## Memory model (target)

| Layer | Rule |
|-------|------|
| Default | Automatic for DataDream objects |
| C resources | Explicit load/unload or `assets.*` manager |
| Advanced | Arenas, pools for games |

```dd
let tex = LoadTexture("player.png");    // raw — user unloads
let tex = assets.texture("player.png"); // friendly — shutdown cleanup (planned)
```

`none` = optional empty · `null` = C pointer null (separate concepts).

---

## Compiler layers (build order)

| Layer | Contents | Status |
|-------|----------|--------|
| **1 — Raw language** | lexer, parser, AST, `fn main`, `let`, if/while/for/**loop**, **match**, **defer**, struct literals, **struct/entity methods**, modules, C calls | ✅ |
| **2 — raylib** | bindgen, `use raylib`, link flags, raw 2D/3D hello | ✅ CI all platforms |
| **3 — Quality of life** | colors, vec2/3, interpolation, `\|\|/&&`, **loop/defer/match**, **typecheck + error hints** | 🟡 parser/lexer hints partial |
| **4 — Friendly wrappers** | app/window/draw, input, screen, random, sprites, math, time, **audio/assets**, **ui.*** | ✅ |
| **5 — Engine sugar** | scenes, entities, ECS, spawn, systems, events | ✅ · editor ❌ |

---

## Version 1 target program

This **must** compile on all platforms. Canonical example: `examples/raylib/hello_3d.dd`

```dd
use raylib;

fn main() {
    InitWindow(800, 600, "DataDream 3D");
    SetTargetFPS(60);
    defer CloseWindow();

    let camera = Camera3D {
        position: vec3(4.0, 4.0, 4.0),
        target: vec3(0.0, 0.0, 0.0),
        up: vec3(0.0, 1.0, 0.0),
        fovy: 45.0,
        projection: CAMERA_PERSPECTIVE
    };

    while !WindowShouldClose() {
        BeginDrawing();
        ClearBackground(#101018);

        BeginMode3D(camera);
        DrawGrid(10, 1.0);
        DrawCube(vec3(0.0, 1.0, 0.0), 2.0, 2.0, 2.0, colors.rebeccaPurple);
        EndMode3D();

        DrawText("DataDream 3D — raylib 6.0", 10, 10, 20, RAYWHITE);
        EndDrawing();
    }
}
```

---

## Minimum v1 checklist

| Requirement | Status |
|-------------|--------|
| `use raylib;` — names in scope | ✅ |
| `fn main()` | ✅ |
| `let`, primitives, calls | ✅ |
| `if`, `while`, `loop`, `for i in 0..n`, `match`, `defer` | ✅ |
| Struct literals (`Camera3D { }`) | ✅ |
| `extern c` / bindgen / link | ✅ |
| Hex + `colors.*` in raw mode | ✅ |
| `vec2` / `vec3` | ✅ |
| Friendly `app` / `window` / `draw` | ✅ |
| Bundled build (no Go for users) | ✅ Windows |
| Linux / macOS builds | ❌ P0 |
| Selective `use { }` | ❌ post-v1 |
| `for i in 0..=n` inclusive | ❌ post-v1 |
| Type checker | ❌ P1 |
| raygui / `ui.*` | ❌ Layer 4 |
| `assets.*` / `audio.*` | ❌ Layer 4 |

---

## Naming conventions

| Kind | Style | Example |
|------|-------|---------|
| Language | DataDream | — |
| Files | `snake_case.dd` | `hello_3d.dd` |
| C imports | Original C names | `InitWindow`, `DrawCube` |
| Friendly API | dotted lower | `draw.text`, `input.down` |
| Types | PascalCase | `Camera3D`, `Color` |
| Constants | C / raylib | `CAMERA_PERSPECTIVE`, `RAYWHITE` |
| Colors namespace | camelCase | `colors.rebeccaPurple` |

---

## Keywords (full set)

**Core today (Layer 1–2):** `use`, `using`, `as`, `fn`, `return`, `let`, `struct`, `enum`, `if`, `else`, `while`, `loop`, `for`, `in`, `break`, `continue`, `match`, `defer`, `true`, `false`, `null`, `none`, `extern`, `c`, `include`, `link`, `module`, `const`, `try`

**Game/app sugar today (Layer 5):** `app`, `window`, `start`, `update`, `draw`, `scene`, `entity`, `system`, `spawn`, `destroy`, `on`, `self`, `ui`

**Later:** `package`, `namespace`, `test`, `assert`, `arena`, `async`, `await`, `component`, `import`, `preload`, `shader`, `network`, `rpc`, `sync`, `data`, `state`, `asset`, `pool`

Full tables by purpose: [SYNTAX.md](SYNTAX.md).

---

## Error message philosophy

Bad: `unexpected token 91`  
Good: `Expected ';' after this statement` with file, line, caret, and a fix hint.

Color errors should suggest valid formats (`#FF0000`, `#F00`, etc.).

---

## Related docs

| Doc | Purpose |
|-----|---------|
| [SYNTAX.md](SYNTAX.md) | Keywords, operators, literals by category and layer |
| [LANGUAGE.md](LANGUAGE.md) | What works **today** (user reference) |
| [VISION.md](VISION.md) | Anti-goals and success criteria |
| [ROADMAP.md](ROADMAP.md) | Prioritized tasks |
| [INTEROP.md](INTEROP.md) | raylib, bindgen, linking |
| [ARCHITECTURE.md](ARCHITECTURE.md) | Compiler pipeline |

When design and implementation disagree, **fix implementation or update LANGUAGE.md** — do not let examples drift from reality.
