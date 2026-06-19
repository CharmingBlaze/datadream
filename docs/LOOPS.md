# Loops, Frame Budget, and Performance

DataDream runs a game loop at the target FPS (often 60). Every `update { }` and `draw { }` block shares the frame budget (~16 ms at 60 FPS). The compiler protects that budget so beginners do not accidentally freeze the window, while keeping generated C fast enough for large games.

**Status key:** ✅ implemented · 🟡 partial · ❌ planned

---

## The five loop forms

### 1. `for i in 0..n` — numeric range

**Status:** ✅ Optimal C (`for (int i = lo; i < hi; i++)`).

**Planned:** 🟡 Warning when the upper bound is not compile-time visible inside `update` / `draw`:

```dd
let count = assets.load_count();
for i in 0..count { }  // warning: loop bound unknown at compile time
```

Not an error — the bound may still be small and known to the author.

---

### 2. `for x in Array<T>` — dynamic array

**Status:** ✅ Contiguous `DD_Array` buffer; codegen uses index loop over `.len` (no per-iteration allocation).

**Rules:**
- Do not call `.remove()` while iterating — use `.remove_dead()` after the loop (compiler warns).
- Growing `.push()` inside a per-frame loop triggers a warning (heap realloc).

**Planned:** Frame arena routes ephemeral strings in `draw { }` through `dd_frame_arena_alloc` (arena runtime exists; string concat wiring 🟡).

---

### 3. `for x in Entity` — ECS iteration (hot path)

**Status:** 🟡 Packed **entity pool** (struct array + `active` flag), not heap pointer chasing.

Generated shape:

```c
#define ENEMY_MAX 1024
static Enemy_Entity Enemy_pool[ENEMY_MAX];

for (int _i = 0; _i < ENEMY_MAX; _i++) {
    if (!Enemy_pool[_i].active) continue;
    Enemy_Entity* e = &Enemy_pool[_i];
    /* body */
}
```

**`@max(n)` on entities** sets pool capacity (default **1024**):

```dd
@max(10000)
entity Enemy {
    health: float;
}
```

`spawn` allocates from the pool (no `malloc` per entity). `destroy` clears `active` for reuse.

---

### 4. `loop { }` — intentional infinite loop

**Status:** ✅ Error in per-frame blocks without a `break` that exits **this** loop.

```dd
update {
    loop {              // error: infinite loop in update will hang the frame
        if done { break; }
    }
}
```

Allowed in `start { }`, loading screens, and `fn main()` where blocking is intentional.

---

### 5. `while cond { }` — conditional loop

**Status:** 🟡 Debug iteration guard in generated C (`#ifndef NDEBUG`); stripped when building with `--release` (`-DNDEBUG`).

Inside `update` / `draw` / entity lifecycle / `system` blocks, debug builds cap unbounded loops at 10 000 iterations and log via `TraceLog`.

**Planned:** `@max_iterations(n)` on individual loops.

---

## Cross-cutting checks

| Issue | Status | Behavior |
|-------|--------|----------|
| Nested `for a in Entity` inside `for b in Entity` | ✅ | Warning: O(n²); hint spatial / `collision.*` |
| `.push()` in per-frame loop | ✅ | Warning |
| String `+` concat in per-frame loop | ✅ | Warning; frame arena will absorb cost later |
| Mutation in `draw { }` (non-draw calls) | 🟡 | Warning on `self`/entity field assigns outside `draw.*` |
| `collision.*` broad-phase | ❌ v2 | Spatial grid / sweep-and-prune |

---

## Build modes

| Mode | Command | Behavior |
|------|---------|----------|
| **Debug** | `datadream build` / `datadream run` | While-loop guards, warnings as errors optional, `-O0` default |
| **Release** | `datadream build --release` | `-O3 -DNDEBUG`, guards stripped, no guard overhead |

**Planned v2:** Frame time logging when update+draw exceeds 16 ms in debug builds.

---

## Implementation map

| Feature | Package |
|---------|---------|
| Loop / lifecycle analysis | `internal/typecheck/loops.go` |
| Entity pool + `@max` | `internal/codegen/ecs.go` |
| While debug guards | `internal/codegen/stmts.go` |
| Frame arena | `internal/codegen/arena.go` |
| Array runtime | `internal/codegen/array.go` |

---

## v2 (do not block v1 on these)

- Frame arena for all string interpolation in lifecycle blocks
- `@max_iterations(n)` attribute
- Spatial partitioning in `collision.*`
- Full `--debug` flag mirroring guard + allocation runtime traces

See also [DESIGN.md](DESIGN.md), [ARCHITECTURE.md](ARCHITECTURE.md), [ROADMAP.md](ROADMAP.md).
