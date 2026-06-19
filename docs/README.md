# DataDream Documentation

Start here if you are joining the project.

## New programmer?

1. **[HANDOFF.md](HANDOFF.md)** — current state, what works, what's next, daily commands  
2. **[VISION.md](VISION.md)** — goals and anti-goals (read before any feature)  
3. **[SETUP.md](SETUP.md)** — build the compiler and populate the SDK  

## Full doc index

| Document | Read this when you need to… |
|----------|-----------------------------|
| [HANDOFF.md](HANDOFF.md) | **Pick up development** — state, blockers, sprint plan |
| [VISION.md](VISION.md) | Understand **goals and anti-goals** |
| [DESIGN.md](DESIGN.md) | Full **language design map** (identity, layers, v1 target) |
| [SYNTAX.md](SYNTAX.md) | **Keywords, operators, literals** — organized by purpose and layer |
| [LANGUAGE.md](LANGUAGE.md) | Learn **syntax and APIs** — source of truth for users |
| [SETUP.md](SETUP.md) | **Build, test, and run** the compiler locally |
| [ARCHITECTURE.md](ARCHITECTURE.md) | Understand the **compiler pipeline** and file map |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Add a feature **the right way** |
| [ROADMAP.md](ROADMAP.md) | See **prioritized remaining work** |
| [INTEROP.md](INTEROP.md) | Work with **raylib, colors, and C bindings** |
| [DISTRIBUTION.md](DISTRIBUTION.md) | Ship **releases** (no Go for end users) |

Also see:

- [../sdk/README.md](../sdk/README.md) — bundled SDK (Clang + raylib 6.0)
- [../examples/](../examples/) — runnable `.dd` samples
- [../README.md](../README.md) — project overview

## Quick orientation

**DataDream** (`.dd`) compiles to **C**, then links with **bundled Clang + raylib 6.0**.

- **Identity:** C/raylib-like syntax — `fn main()`, `use raylib;`, struct literals — not Python or JS.
- **Two layers:** raw raylib (3D, full API) and friendly `app`/`draw` sugar (2D games).

## Current status (June 2026)

| Works | Not yet |
|-------|---------|
| Windows/Linux/macOS builds with bundled or system Clang | `darwin-amd64` CI coverage |
| `sdk install clang` / `raylib` | CI + release zips on all platforms |
| Friendly app + game API (Layer 4) | Type checker (partial) |
| Entity/scene ECS wired to loop | Module exports (`include` → modules) |
| Raw raylib + bindgen (`raw.dd` parses fast) | LSP / VS Code extension |
| [SYNTAX.md](SYNTAX.md) keyword reference | Selective `use raylib { }` |
| `check --codegen` on all examples | Default `check` without `--codegen` |

If you only read two files: **HANDOFF.md** and **VISION.md**.
