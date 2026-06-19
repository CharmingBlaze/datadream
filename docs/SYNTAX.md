# DataDream — Syntax Reference

Every keyword, operator, and literal the language needs — organized by what they do.

**Status legend:** ✅ implemented · 🟡 partial · ❌ planned (lexer only or design)

For tutorials and friendly APIs, see [LANGUAGE.md](LANGUAGE.md). For design rationale and layers, see [DESIGN.md](DESIGN.md).

**Build order (layers):** core language (1) → C interop (2) → QoL (3) → friendly wrappers (4) → **game/app sugar (5)**. Layer 5 keywords compile down to Layer 1–4; they are marked below so it is clear they come later in the build order.

**A few things worth noting:**

- **`loop`** — explicit infinite-loop keyword; cleaner than `while true` for frame pumps and event loops.
- **`defer`** — scoped C cleanup so you do not forget `UnloadTexture`, `CloseAudioDevice`, or similar teardown.
- **`match`** — pattern-style switch for values that would otherwise need long `if` / `else if` chains.
- **Layer 5** — game/app sugar (`app`, `window`, `start`, `update`, `draw`, …) is marked **5** so it is clear those keywords come later in the build order.

---

## Literals

| Form | Example | Layer | Status |
|------|---------|-------|--------|
| Integer | `0`, `42`, `-7` | 1 | ✅ |
| Float | `3.14`, `0.5` | 1 | ✅ |
| Boolean | `true`, `false` | 1 | ✅ |
| String | `"hello"`, `"Score: {score}"` | 1 | ✅ interpolation in `print` / friendly draw |
| Null / none | `null`, `none` | 1 | ✅ both map to null |
| Hex color | `#RGB`, `#RRGGBB`, `#RRGGBBAA` | 3 | ✅ |
| Color constructors | `rgb()`, `rgba()`, `hsl()`, `css("name")` | 3 | ✅ |
| Identifier | `player`, `InitWindow` | 1 | ✅ |

---

## Operators

### Arithmetic

| Op | Meaning | Layer | Status |
|----|---------|-------|--------|
| `+` `-` `*` `/` `%` | add, subtract, multiply, divide, modulo | 1 | ✅ |
| `+=` `-=` `*=` `/=` | compound assignment | 1 | ✅ |

### Comparison

| Op | Meaning | Layer | Status |
|----|---------|-------|--------|
| `==` `!=` | equal, not equal | 1 | ✅ |
| `<` `>` `<=` `>=` | ordering | 1 | ✅ |

### Logical

| Op | Meaning | Layer | Status |
|----|---------|-------|--------|
| `!` | not (prefix) | 1 | ✅ required for negation (`!WindowShouldClose()`) |
| `&&` `\|\|` | and, or | 1 | ✅ |
| `and` `or` `not` | word forms | 1 | ✅ |

### Assignment & range

| Op | Meaning | Layer | Status |
|----|---------|-------|--------|
| `=` | assign | 1 | ✅ |
| `..` | half-open range end (`0..10` → 0..9) | 1 | ✅ |
| `..=` | inclusive range end (`0..=10` → 0..10) | 1 | ✅ |

### Other punctuation used in expressions

| Token | Meaning | Layer | Status |
|-------|---------|-------|--------|
| `?` | optional / ternary (planned) | 3 | 🟡 |
| `?.` | optional chain (planned) | 3 | ❌ |
| `->` | function return type | 1 | ✅ |
| `=>` | match arm separator | 1 | ✅ |
| `@` | attribute / decorator (planned) | 5 | ❌ |

---

## Delimiters & separators

| Token | Role | Layer | Status |
|-------|------|-------|--------|
| `{` `}` | blocks, struct literals | 1 | ✅ |
| `(` `)` | calls, grouping, `if`/`while` conditions optional | 1 | ✅ |
| `[` `]` | arrays (planned) | 3 | ❌ |
| `;` | statement separator; **config block** fields (`window { … }`) | 1 | ✅ |
| `,` | argument lists; **option object** fields (`draw.text` opts) | 1 | ✅ |
| `:` | type annotations; struct field names | 1 | ✅ |
| `.` | field access (`colors.red`, `input.down`) | 1 | ✅ |

---

## Control flow

| Keyword | Purpose | Layer | Status |
|---------|---------|-------|--------|
| `if` | conditional | 1 | ✅ |
| `else` | alternative branch; also default arm in `match` (`else =>`) | 1 | ✅ |
| `while` | loop while condition is true | 1 | ✅ |
| **`loop`** | **explicit infinite loop** — cleaner than `while true` | 1 | ✅ emits `while (1)` |
| `for` | numeric range or collection iteration | 1 | ✅ range; 🟡 `for x in arr` |
| `in` | range / iteration binding | 1 | ✅ |
| **`break`** | exit innermost loop | 1 | ✅ |
| **`continue`** | next loop iteration | 1 | ✅ |
| **`match`** | **pattern-style switch** — replaces long `if` / `else if` chains | 1 | ✅ emits if-else chain |
| **`defer`** | **scoped cleanup** — run call on block exit (LIFO) | 1 | ✅ |
| `return` | exit function with optional value | 1 | ✅ |
| `try` | fallible call + `else` block | 3 | 🟡 |

### `loop`

Prefer `loop { }` over `while true { }` for game loops and event pumps:

```dd
loop {
    if input.pressed("escape") { break; }
    update();
}
```

### `defer`

Register C cleanup so you do not forget `UnloadTexture`, `CloseAudioDevice`, `UnmapBuffer`, etc. Defers run in **reverse order** when the enclosing block or function returns:

```dd
fn load() {
    let tex = LoadTexture("coin.png");
    defer UnloadTexture(tex);

    let music = LoadMusicStream("theme.ogg");
    defer CloseAudioDevice();

    loop { /* use tex + music */ }
}
```

### `match`

Switch on a value with `=>` arms. Use `_ =>` or `else =>` for the default:

```dd
match state {
    0 => { draw.title(); }
    1 => { draw.game(); }
    2 => { draw.pause(); }
    _ => { draw.unknown(); }
}
```

Today: equality patterns only (compiles to C `if` / `else if`). Destructuring and type patterns are planned.

---

## Declarations & functions

| Keyword | Purpose | Layer | Status |
|---------|---------|-------|--------|
| `let` | mutable binding (inferred or typed) | 1 | ✅ |
| `const` | immutable binding | 1 | 🟡 |
| `fn` | function | 1 | ✅ |
| `struct` | user struct | 1 | ✅ |
| `enum` | enumeration | 1 | 🟡 |
| `return` | return from function | 1 | ✅ |

```dd
let score = 0;
let speed: float = 4.5;
const MAX = 100;

fn add(a: int, b: int) -> int {
    return a + b;
}
```

---

## Types (keywords & names)

| Name | Role | Layer | Status |
|------|------|-------|--------|
| `bool` | boolean | 1 | ✅ |
| `int` `float` `double` | aliases (`i32`, `f32`, `f64`) | 1 | ✅ |
| `string` `char` `byte` | strings and scalars | 1 | ✅ |
| `i8`…`i64`, `u8`…`u64` | exact integers | 1 | ✅ |
| `f32` `f64` | exact floats | 1 | ✅ |
| `usize` `isize` | pointer-sized integers | 1 | ✅ |
| `cstring` | C string pointer | 2 | ✅ |
| `voidptr` `ptr<T>` | raw pointers | 2 | 🟡 |
| `void` | no return | 1 | ✅ |

Raylib C types (`Vector2`, `Camera3D`, `Color`, …) are used via `use raylib;` or struct literals.

---

## Modules & C interop (Layer 2)

| Keyword | Purpose | Status |
|---------|---------|--------|
| `use` | import module / library into scope | ✅ |
| `using` | same as `use` for name injection | ✅ |
| `as` | import alias (`use raylib as rl`) | ✅ |
| `module` | declare module | 🟡 |
| `include` | textual include of another `.dd` file | ✅ |
| `import` | module-only access (planned) | ❌ |
| `extern` | foreign declarations | ✅ |
| `c` | C ABI marker in `extern c { }` | ✅ |
| `link` | link library name in extern block | ✅ |

```dd
use raylib;
include "utils.dd";

extern c {
    link "raylib";
    fn InitWindow(width: int, height: int, title: cstring);
}
```

---

## Async & concurrency (planned)

| Keyword | Purpose | Layer | Status |
|---------|---------|-------|--------|
| `async` | async function | 5 | ❌ |
| `await` | wait on async | 5 | ❌ |
| `sync` | synchronized region | 5 | ❌ |
| `rpc` | remote call stub | 5 | ❌ |
| `network` | networking module marker | 5 | ❌ |

---

## Memory & data (planned)

| Keyword | Purpose | Layer | Status |
|---------|---------|-------|--------|
| `arena` | arena allocator | 3 | ❌ |
| `pool` | object pool | 3 | ❌ |
| `data` | data-oriented block | 5 | ❌ |
| `preload` | asset preload hook | 4 | ❌ |
| `asset` | asset declaration | 4 | 🟡 |
| `state` | persistent state block | 5 | 🟡 |
| `shader` | shader program | 4 | ❌ |

---

## Game / app sugar — **Layer 5**

These keywords expand to raylib init, main loop, and draw calls. They are **not** required for raw C-style programs (`fn main()` + `use raylib;`).

| Keyword | Purpose | Layer | Status |
|---------|---------|-------|--------|
| **`app`** | program title / entry metadata | **5** | ✅ |
| **`window`** | window config block (`size`, `title`, `fps`) | **5** | ✅ |
| **`start`** | runs once at startup | **5** | ✅ |
| **`update`** | per-frame logic (`dt` available) | **5** | ✅ |
| **`draw`** | per-frame rendering (after update) | **5** | ✅ |
| **`scene`** | named scene with optional lifecycle | **5** | 🟡 |
| **`entity`** | component-style game object | **5** | 🟡 |
| **`spawn`** | create entity instance | **5** | 🟡 |
| **`destroy`** | free entity | **5** | 🟡 |
| **`self`** | current entity in entity methods | **5** | 🟡 |
| **`system`** | ECS system block | **5** | 🟡 |
| **`on`** | event handler (`on key "space" { }`) | **5** | 🟡 |
| **`ui`** | raygui wrapper namespace | **5** | ❌ |

```dd
app "Coin Runner";

window {
    size: 800, 600;
    title: "Coin Runner";
    fps: 60;
}

start { /* once */ }
update { /* each frame */ }
draw { clear(#101018); draw.text("Go!", { position: screen.center }); }
```

Layer 5 **does not** replace raw raylib — it sits on top. Use `fn main()` when you need full control (3D, custom loops, minimal runtime).

---

## Friendly runtime namespaces — **Layer 4**

Not keywords, but core vocabulary once `app` / `draw` are enabled:

| Namespace | Role | Status |
|-----------|------|--------|
| `draw.*` | shapes, text, sprites, clear | ✅ |
| `input.*` | keys, mouse, pressed/down/released | ✅ |
| `keys.*` | named key constants | ✅ |
| `screen.*` | width, height, center, size | ✅ |
| `colors.*` | named + raylib palette | ✅ |
| `random.*` | int, float, point | ✅ |
| `time.*` | fps, now, frame | ✅ |
| `math.*` | sqrt, pow, sign, … | ✅ |
| `audio.*` | play, stop, unload, shutdown | ✅ |
| `assets.*` | texture, sound, unload | ✅ |
| `ui.*` | button, label, labelButton | ✅ |
| `collision.*` | contains, pointInRect | ✅ |
| `quit()` | request exit | ✅ |

See [LANGUAGE.md](LANGUAGE.md) for the full command list.

---

## Reserved / future keywords

Present in the lexer for forward compatibility; no codegen yet unless noted:

| Keyword | Intended use |
|---------|----------------|
| `import` | module path imports |
| `async` / `await` | async I/O |
| `shader` | GPU programs |
| `network` / `rpc` / `sync` | multiplayer |
| `arena` / `pool` | memory patterns |
| `data` | SOA / buffers |
| `preload` | loading screens |

---

## Quick comparison: three loop styles

| Style | When to use |
|-------|-------------|
| `loop { }` | Game frame pump, server accept loop, “run forever” |
| `while cond { }` | Condition may become false (`!WindowShouldClose()`) |
| `for i in 0..n { }` | Indexed iteration |

---

## Related docs

| Doc | Contents |
|-----|----------|
| [LANGUAGE.md](LANGUAGE.md) | Friendly API reference, examples |
| [SYNTAX.md](SYNTAX.md) | Keywords, operators, literals by category and layer |
| [DESIGN.md](DESIGN.md) | Identity, v1 target, namespace table |
| [INTEROP.md](INTEROP.md) | raylib, bindgen, `extern c` |
| [VISION.md](VISION.md) | Layers 1–5 roadmap order |
| `examples/raylib/control_flow.dd` | Runnable `loop` / `defer` / `match` demo |
