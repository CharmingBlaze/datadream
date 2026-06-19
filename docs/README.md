# DataDream Documentation

Start here if you are joining the project.

## New programmer?

1. **[HANDOFF.md](HANDOFF.md)** — current state, what works, what's next, file map, daily commands  
2. **[NEXT_SESSION_PROMPT.md](NEXT_SESSION_PROMPT.md)** — copy-paste into a new agent chat or ticket  
3. **[VISION.md](VISION.md)** — goals and anti-goals (read before any feature)  
4. **[SETUP.md](SETUP.md)** — build the compiler and populate the SDK  

If you only read two files: **HANDOFF.md** and **NEXT_SESSION_PROMPT.md**.

---

## Full doc index

| Document | Read this when you need to… |
|----------|-----------------------------|
| [HANDOFF.md](HANDOFF.md) | **Pick up development** — state, blockers, examples, gotchas |
| [NEXT_SESSION_PROMPT.md](NEXT_SESSION_PROMPT.md) | **Paste into a new chat** — mission, priorities, verify commands |
| [ROADMAP.md](ROADMAP.md) | See **task status** and v1 Definition of done |
| [PUBLISH.md](PUBLISH.md) | **Ship v1.0** — GitHub Release workflow |
| [VISION.md](VISION.md) | Understand **goals and anti-goals** |
| [DESIGN.md](DESIGN.md) | Full **language design map** (identity, layers, v1 target) |
| [SYNTAX.md](SYNTAX.md) | **Keywords, operators, literals** by layer |
| [LANGUAGE.md](LANGUAGE.md) | Learn **syntax and APIs** — source of truth for users |
| [ARCHITECTURE.md](ARCHITECTURE.md) | Understand the **compiler pipeline** and file map |
| [SETUP.md](SETUP.md) | **Build, test, and run** the compiler locally |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Add a feature **the right way** |
| [INTEROP.md](INTEROP.md) | Work with **raylib, colors, and C bindings** |
| [DISTRIBUTION.md](DISTRIBUTION.md) | Ship **releases** (no Go for end users) |

Also see:

- [../sdk/README.md](../sdk/README.md) — bundled SDK (Clang + raylib 6.0)
- [../examples/](../examples/) — 24 runnable `.dd` samples
- [../README.md](../README.md) — project overview

---

## Quick orientation

**DataDream** (`.dd`) compiles to **C**, then links with **bundled Clang + raylib 6.0**.

- **Identity:** C/raylib-like syntax — `fn main()`, `use raylib;`, struct literals — not Python or JS.
- **Two layers:** raw raylib (3D, full API) and friendly `app`/`draw` sugar (2D games).
- **Iteration:** `for i in 0..10` (range), `for e in Enemy` (ECS), `for x in array` (`DD_Array`), `for ch in "hi"` (bytes).

---

## Current status (June 2026)

**Progress:** ~98% v1-complete — all examples build; ship gate is GitHub Release v1.0.0.

| ✅ Done | 🔄 Remaining |
|---------|--------------|
| Win/Linux/macOS builds, CI dist-verify, release workflow | Publish **v1.0 GitHub Release** |
| Full Layer 4 + ECS + type checker + `@packed`/`@save` | GitHub Release v1.0.0 publish |
| `Array<T>`, `list T`, all four `for-in` forms | LSP |
| `export fn` / `export let`, frame/level arenas | Selective `use raylib { … }` |
| `datadream new`, VS Code grammar | Codegen-stage error hints |

**Next action for maintainers:** [PUBLISH.md](PUBLISH.md)
