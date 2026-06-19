# DataDream Language Reference

What the language supports **today**. For the full design map, see [DESIGN.md](DESIGN.md). For every keyword, operator, and literal by category, see [SYNTAX.md](SYNTAX.md). For goals and future syntax, see [VISION.md](VISION.md).

File extension: `.dd`  
Entry styles: **app program** (lifecycle) or **script** (`fn main()`)

---

## Program shapes

### App / game (friendly)

```dd
app "Title";

window {
    size: 800, 600;
    title: "Hello";
    fps: 60;
}

let score = 0;

start { /* runs once */ }
update { /* runs each frame; `dt` is available */ }
draw { /* runs each frame after update */ }
```

App programs **auto-enable raylib** — you do not need `use raylib;` for `window` + `draw`.

### Script (raw / mixed)

```dd
use raylib;

fn main() {
    InitWindow(800, 600, "Hello");
    while !WindowShouldClose() {
        BeginDrawing();
        ClearBackground(BLACK);
        EndDrawing();
    }
    CloseWindow();
}
```

---

## Syntax rules

| Construct | Separator | Example |
|-----------|-----------|---------|
| Config block | `;` | `window { size: 800, 600; title: "Hi"; }` |
| Option object | `,` | `{ position: vec2(0,0), size: 32 }` |
| Statements | `;` | `let x = 10;` |

### Invalid (rejected by parser)

```dd
// OLD — do not use
window 800, 600, "Hello";
text "Hello" at 300, 280 size 32 color white;
```

Use structured namespaces and option objects instead.

---

## Imports and modules

```dd
use raylib;              // C API in scope: InitWindow, DrawText, …
use raylib as rl;        // rl.InitWindow(...)
use graphics;            // loads libs/graphics/wrapper.dd (friendly Layer 4 + raylib)
using raylib;            // same as plain use for name injection
include "utils.dd";      // textual include (legacy — prefer use for bundled libs)
extern c { }             // inline C bindings block
```

| Form | Effect |
|------|--------|
| `use raylib;` | Links raylib; names available without prefix |
| `use raylib as rl;` | Namespaced access only |
| `use graphics;` | Merges `libs/graphics/wrapper.dd`; enables friendly `draw.*` / `input.*` with raylib link |
| `using raylib;` | Brings names into file scope |

---

## Types (codegen mapping)

| DataDream | C (raylib mode) |
|-----------|-----------------|
| `int` | `int` |
| `float` | `float` |
| `bool` | `bool` |
| `string` | `char*` / string literal |
| `Vec2` | `Vector2` |
| `Vec3` | `Vector3` |
| `Color` | `Color` |
| `Sprite` | `Sprite` (game runtime struct) |

Type inference works for literals and some builtins. **`datadream check` runs a type checker** (unknown identifiers, builtin arg counts, bad option/struct fields). Use `check --codegen` to also verify C emission.

---

## Variables and control flow

```dd
let x = 10;
let name: string = "Ada";
let pos = vec2(100, 200);

if x > 0 {
    print("positive");
}

if input.down("space") or input.pressed("enter") {
    // `or` / `||` and `and` / `&&` both work
}

for i in 0..10 { }      // 0 through 9 (exclusive end)
for i in 0..=10 { }     // 0 through 10 (inclusive end)
while condition { }
loop { }                // infinite loop — cleaner than while true
break; continue;

match value {
    1 => { }
    _ => { }
}

defer CloseWindow();    // runs on block exit (LIFO) — UnloadTexture, CloseAudioDevice, …
```

Full keyword, operator, and literal tables (with layer numbers): [SYNTAX.md](SYNTAX.md).

---

## Colors

### Namespace

```dd
colors.black
colors.white
colors.sky        // alias for sky blue
colors.raywhite
```

~140 CSS named colors + raylib palette. Resolved at compile time when possible.

### Literals and constructors

```dd
#RRGGBB
#RRGGBBAA
rgb(255, 0, 0)
rgba(255, 0, 0, 0.5)
hsl(200, 0.5, 0.5)
css("cornflowerblue")
```

### Methods (compile to C helpers)

```dd
colors.red.withAlpha(0.5)
someColor.lighten(0.2)
someColor.mix(colors.blue, 0.5)
```

---

## Screen (window size)

Use `screen` to position things without hard-coding pixel values:

```dd
draw.text("Centered title", {
    position: vec2(screen.width / 2, 40),
    size: 24,
    color: colors.white,
    align: "center"    // "left" (default), "center", "right"
});

let spot = random.point(screen.size);   // random x/y inside the window
let middle = screen.center;             // vec2 at window center
```

| Property | Type | Meaning |
|----------|------|---------|
| `screen.width` | float | Window width in pixels |
| `screen.height` | float | Window height in pixels |
| `screen.center` | Vec2 | Middle of the window |
| `screen.size` | Vec2 | `vec2(width, height)` |

---

## Mouse input

```dd
let mouse = input.mouse();                 // current cursor position (Vec2)

if input.mousePressed("left") { }          // clicked this frame
if input.mouseDown("right") { }            // button held
if input.mouseReleased("middle") { }       // released this frame
```

Buttons: `"left"`, `"right"`, `"middle"`.

---

## Random numbers

```dd
let n = random.int(1, 6);                  // dice roll
let x = random.float(0.0, 1.0);            // 0..1
let pos = random.point(screen.size);       // random point in window
let anywhere = random.screenPosition();    // same idea, full screen
```

---

## Math helpers

```dd
let d = distance(player.position, coin.position);
let len = length(velocity);
let t = lerp(0.0, 100.0, 0.5);             // 50
let clamped = clamp(speed, 0.0, 400.0);
let lo = min(a, b);
let hi = max(a, b);
let a = abs(-3.5);
```

---

## Quitting the app

```dd
update {
    if input.pressed("escape") {
        quit();
    }
}
```

`quit()` closes the game at the end of the current frame (works in app programs).

---

## Key names (`keys.*`)

Use with `input.pressed(keys.w)` instead of string literals:

```dd
if input.down(keys.space) { }
if input.pressed(keys.escape) { }
```

Available: `keys.w`, `keys.a`, `keys.s`, `keys.d`, `keys.space`, `keys.enter`, `keys.escape`, `keys.shift`, `keys.ctrl`, `keys.alt`, arrows, `keys.f1`…`keys.f9`, etc.

---

## Time

```dd
let fps = time.fps();           // current FPS (int)
let t = time.now();             // seconds since InitWindow
let frame = time.frame();       // GetFrameTime() — same as dt in update
```

In `update { }`, prefer the built-in `dt` variable.

---

## Math namespace

Global builtins: `distance`, `length`, `clamp`, `lerp`, `min`, `max`, `abs`, `sqrt`, `pow`, `sign`, `floor`, `ceil`, `round`

Namespaced (raymath):

```dd
let d = math.distance(a, b);
let n = math.normalize(v);
let dot = math.dot(a, b);
let cross = math.cross(a, b);
```

---

## Audio and assets

```dd
let jump = sound("jump.wav");           // or assets.sound("jump.wav")
audio.play(jump);
audio.stop(jump);
audio.unload(jump);                     // UnloadSound — use with defer
audio.shutdown();                       // CloseAudioDevice — defer in fn main()

let tex = sprite("player.png");         // or assets.texture("player.png")
draw.sprite(tex);
assets.unload(tex);                     // UnloadTexture — use with defer
```

Load helpers store the path only; raylib load happens on first `play` or `draw.sprite`. In `fn main()` programs, defer cleanup in reverse order (textures/sounds first, then `audio.shutdown()`, then `CloseWindow()`). App programs call `datadream_audio_shutdown()` automatically before `CloseWindow()`.

Audio initializes on first use (`InitAudioDevice`).

Example: `examples/raylib/audio_demo.dd` — build from `examples/raylib/` so `assets/` resolves.

---

## UI (raygui)

Immediate-mode widgets via the bundled `raygui.h` header (included automatically when you use `ui.*`):

```dd
if ui.button("Click me", { position: vec2(20, 70), width: 140, height: 32 }) {
    score += 1;
}

ui.label("Status text", { position: vec2(20, 120), width: 260, height: 24 });
ui.labelButton("Toggle", { position: vec2(20, 160), width: 120, height: 32 });
```

| Call | Returns | Notes |
|------|---------|-------|
| `ui.button(text, opts)` | `bool` | → `GuiButton` |
| `ui.label(text, opts)` | — | → `GuiLabel` |
| `ui.labelButton(text, opts)` | `bool` | → `GuiLabelButton` |

Option object fields: `position` (`vec2`), `width`, `height`, or `size` (`vec2`). Defaults: button width 120, height 32; label height 24.

Example: `examples/raylib/ui_demo.dd`.

---

## Full command demo

See `examples/raylib/commands.dd` — living spec for all friendly namespaces.

---

## Drawing (friendly API)

| Call | Status | Notes |
|------|--------|-------|
| `clear(colors.black)` | ✅ | → `ClearBackground` |
| `draw.text("Hi", { ... })` | ✅ | → `DrawText` |
| `draw.sprite(player)` | ✅ | → `DrawTextureEx` |
| `draw.rect({ ... })` | ✅ | → `DrawRectangle` — `position`, `size` or `width`/`height`, `color` |
| `draw.circle({ ... })` | ✅ | → `DrawCircle` |
| `draw.line({ ... })` | ✅ | → `DrawLine` |
| `draw.rectOutline({ ... })` | ✅ | → `DrawRectangleLines` |
| `draw.circleOutline({ ... })` | ✅ | → `DrawCircleLines` |
| `draw.ellipse({ ... })` | ✅ | → `DrawEllipse` |
| `draw.triangle({ p1, p2, p3, color })` | ✅ | → `DrawTriangle` |
| `draw.point({ position, color })` | ✅ | → `DrawPixelV` |
| `draw.fps({ position: vec2(10, 10) })` | ✅ | → `DrawFPS` overlay |

### Text alignment

```dd
draw.text("Game Over", {
    position: vec2(screen.width / 2, 300),
    size: 48,
    color: colors.white,
    align: "center"   // or "right"
});
```

Dynamic positions (e.g. `screen.width / 2`) are supported in `draw.text` options.

---

## Game API

```dd
draw.text("Hello", {
    position: vec2(300, 280),
    size: 32,
    color: colors.white
});
```

### String interpolation

```dd
let score = 0;
draw.text("Score: {score}", { position: vec2(20, 20), size: 32, color: colors.white });
```

Integer variables use `%d` in generated `snprintf`.

---

## Game API

| API | Description |
|-----|-------------|
| `sprite("path.png")` | Create lazy-loaded sprite (runtime init if global) |
| `player.position` | `Vec2` field |
| `input.move2d()` | Normalized WASD / arrows |
| `input.axis("left","right","up","down")` | Axis from key names |
| `input.pressed("space")` | Key pressed this frame |
| `input.down("space")` | Key held down |
| `input.released("space")` | Key released this frame |
| `input.mouse()` | Cursor position (Vec2) |
| `input.mousePressed("left")` | Mouse button pressed |
| `input.scroll()` | Mouse wheel delta (float) |
| `input.move3d()` | WASD on XZ + Q/E vertical (Vec3) |
| `collision.overlap(a, b)` | Rectangle overlap |
| `collision.contains(sprite, point)` | Point inside sprite bounds |
| `collision.pointInRect(point, { ... })` | Point inside rect |
| `random.screenPosition()` | Random position on screen |
| `random.point(size)` | Random point within `vec2(width, height)` |
| `random.int(min, max)` | Random integer (inclusive) |
| `random.float(min, max)` | Random float |
| `dt` | Delta time in `update` block only |
| `quit()` | Exit app at end of frame |

### Example

See `examples/coin-runner/game.dd`. Build/run from `examples/coin-runner/` so `assets/*.png` resolve.

Missing textures draw colored placeholder rectangles at runtime.

---

## Lifecycle codegen

| Block | Generated |
|-------|-----------|
| `start { }` | `void lifecycle_start(void)` — after `datadream_init_globals()` if needed |
| `update { }` | `void lifecycle_update(float dt)` |
| `draw { }` | `void lifecycle_draw(void)` — inside `BeginDrawing` |

Global `let` with non-constant values (e.g. `sprite("x.png")`) emit as declarations plus assignments in `datadream_init_globals()`.

Main loop (simplified):

```c
InitWindow(...);
while (!WindowShouldClose()) {
    float dt = GetFrameTime();
    lifecycle_update(dt);
    BeginDrawing();
    lifecycle_draw();
    EndDrawing();
}
CloseWindow();
```

---

## Declarations (parsed; codegen varies)

| Syntax | Parse | Codegen |
|--------|-------|---------|
| `fn` / `async fn` | ✅ | ✅ |
| `struct` | ✅ | ✅ |
| `enum` | ✅ | ✅ |
| `entity` | ✅ | 🟡 Partial |
| `scene` | ✅ | 🟡 Partial |
| `system` | ✅ | 🟡 Partial |
| `spawn` / `destroy` | ✅ | 🟡 Partial |
| `on event` | ✅ | 🟡 Stub |
| `try` | ✅ | 🟡 Stub |
| `ui { }` | ✅ | ❌ Not implemented |

---

## Built-in functions

| Builtin | Status |
|---------|--------|
| `print(...)` | ✅ → `printf` (supports `"Score: {score}"` interpolation) |
| `vec2(x, y)` | ✅ |
| `vec3`, `vec4` | ✅ |
| `sound(path)` | ✅ → lazy-loaded `SoundAsset` |
| `clear(color)` | ✅ (raylib mode) |
| `sprite(path)` | ✅ |
| `distance(a, b)` | ✅ → `Vector2Distance` (Vec2) |
| `length(v)` | ✅ → `Vector2Length` |
| `clamp`, `lerp`, `min`, `max`, `abs` | ✅ |
| `sqrt`, `pow`, `sign`, `floor`, `ceil`, `round` | ✅ |
| `quit()` | ✅ (app programs) |

---

## C interop

Full raylib 6.0 API in `libs/raylib/raw.dd`. Use plain names after `use raylib;`:

```dd
use raylib;

fn main() {
    InitWindow(800, 600, "Game");
    DrawText("Hi", 100, 100, 20, WHITE);
    CloseWindow();
}
```

Generate bindings:

```bash
datadream bind path/to/header.h --raw --out libs/mylib/raw.dd
```

---

## CLI

```bash
datadream check file.dd [--codegen]
datadream build file.dd [-o name] [--release]
datadream run file.dd
datadream bind header.h [--raw] [--out file.dd]
datadream doctor
datadream sdk install clang|raylib|headers
datadream version
```

---

## Examples index

| File | Demonstrates |
|------|--------------|
| `examples/beginner/clicker.dd` | **Start here** — mouse, screen, random, quit |
| `examples/hello/hello.dd` | Minimal print |
| `examples/raylib/hello_friendly.dd` | App + draw.text |
| `examples/raylib/hello_raw.dd` | Raw raylib + `use` |
| `examples/raylib/hello_using.dd` | `using raylib` |
| `examples/raylib/hello_3d.dd` | **v1 target** — raw 3D raylib |
| `examples/raylib/control_flow.dd` | **`loop`, `defer`, `match`** — core control flow |
| `examples/raylib/features.dd` | Friendly draw + input living spec |
| `examples/raylib/commands.dd` | **All friendly namespaces** — full command reference |
| `examples/coin-runner/game.dd` | Full game loop |
| `examples/game-loop/main.dd` | Input + sprite movement |
| `examples/colors/alpha.dd` | Hex + rgba colors |
| `examples/colors/css_demo.dd` | CSS color names |
