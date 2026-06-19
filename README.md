# DataDream

**DataDream** is a native app and game language — C/raylib structure with modern syntax, fast enough for 3D games, and simple enough for beginners.

> What raylib would feel like if it had its own language — `use raylib;`, `fn main()`, struct literals, optional friendly `app`/`draw` sugar.

Compile to native C, link with bundled Clang + raylib 6.0. Wrap C libraries with one command.

## Quick start (end users)

Download the latest release for your platform from [GitHub Releases](https://github.com/CharmingBlaze/datadream/releases):

| Platform | Zip |
|----------|-----|
| Windows | `datadream-windows-amd64.zip` |
| Linux | `datadream-linux-amd64.zip` |
| macOS (Apple Silicon) | `datadream-darwin-arm64.zip` |

Extract the zip, add `bin/` to your PATH (or set `DATADREAM_ROOT` to the extract folder), then:

```bash
datadream doctor
datadream new my-game
cd my-game
datadream run game.dd
```

Try the bundled examples:

```bash
datadream run examples/raylib/hello_friendly.dd
datadream run examples/beginner/clicker.dd
datadream build examples/coin-runner/game.dd -o coin-runner
```

No Go install required. Windows release zips include bundled Clang (~188 MB). On Linux/macOS, run `datadream sdk install clang` once if the zip was built without bundled toolchain.

**Syntax highlighting:** install the VS Code extension from `editor/datadream/` (Open Folder → Run Extension), or point your editor at `syntaxes/datadream.tmLanguage.json`.

**Documentation:** [docs/README.md](docs/README.md) · [docs/SYNTAX.md](docs/SYNTAX.md) (keywords & operators) · **Next programmer:** [docs/HANDOFF.md](docs/HANDOFF.md)

## What DataDream feels like

```dd
app "Hello";

window {
    size: 800, 600;
    title: "Hello";
}

draw {
    clear(colors.black);

    draw.text("Hello World", {
        position: vec2(300, 280),
        size: 32,
        color: colors.white
    });
}
```

### Syntax rules

- **Config blocks** (`window { }`) use semicolons between properties
- **Option objects** use commas: `{ position: vec2(...), size: 32 }`
- **Drawing** uses namespaces: `draw.text`, `draw.rect`, `draw.sprite`
- **Not** English-sentence commands (`text "Hi" at 10, 20`)

## Documentation

| Doc | Purpose |
|-----|---------|
| [docs/HANDOFF.md](docs/HANDOFF.md) | **Start here** — state, blockers, sprint plan |
| [docs/NEXT_SESSION_PROMPT.md](docs/NEXT_SESSION_PROMPT.md) | Paste into a new agent chat |
| [docs/SYNTAX.md](docs/SYNTAX.md) | Keywords, operators, literals by layer |
| [docs/VISION.md](docs/VISION.md) | Goals and anti-goals |
| [docs/LANGUAGE.md](docs/LANGUAGE.md) | Syntax and APIs (source of truth) |
| [docs/DESIGN.md](docs/DESIGN.md) | Language design map |
| [docs/SETUP.md](docs/SETUP.md) | Build compiler, SDK, daily workflow |
| [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) | Pipeline and file map |
| [docs/ROADMAP.md](docs/ROADMAP.md) | Prioritized remaining work |
| [docs/INTEROP.md](docs/INTEROP.md) | raylib, bindings, colors |
| [docs/DISTRIBUTION.md](docs/DISTRIBUTION.md) | Shipping releases |
| [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md) | How to add features |

## Current status (June 2026)

| ✅ Works | 🔄 Remaining for v1 ship |
|----------|--------------------------|
| Bundled SDK + release zips (Win/Linux/macOS CI) | **Publish v1.0 GitHub Release** ([PUBLISH.md](docs/PUBLISH.md)) |
| Friendly apps, coin-runner, ECS, `@packed`/`@save` | GitHub Release v1.0.0 |
| Raw raylib + bindgen + `hello_3d` | LSP (hover, go-to-def) |
| `loop` / `defer` / `match`, struct/entity methods | Selective `use raylib { … }` |
| **`Array<T>` / `list T` + `for x in array`** | Codegen error hints (typecheck ✅) |
| Type checker + warnings + friendly error hints | |
| `export fn` / `export let`, frame/level arenas | |
| Windows/Linux/macOS cold-install (CI dist-verify) | |
| `datadream new` + VS Code syntax highlighting | |

## Building from source (contributors)

```bash
go build -o datadream ./cmd/datadream
go test ./internal/...
```

Requires Go 1.22+. Entry point: `cmd/datadream/main.go`. See [docs/SETUP.md](docs/SETUP.md).

Release zip (maintainers):

```powershell
.\scripts\build-dist.ps1
.\scripts\verify-dist.ps1 dist\datadream-windows-amd64.zip
```

## Raylib interop

Two layers — friendly API compiles to raylib; advanced users call C directly:

```dd
use raylib;

fn main() {
    InitWindow(800, 600, "Raw");
    DrawText("Hello", 100, 100, 20, WHITE);
    CloseWindow();
}
```

Plain `use raylib;` puts names in scope (`InitWindow`, not `raylib.InitWindow`).

```bash
datadream bind sdk/raylib/6.0/include/raylib.h --raw --out libs/raylib/raw.dd
```

See `examples/raylib/hello_friendly.dd`, `hello_raw.dd`, `examples/coin-runner/game.dd`.

## Distribution layout

```
datadream/
  bin/datadream[.exe]
  sdk/toolchain/clang/     ← datadream sdk install clang
  sdk/raylib/6.0/          ← headers + lib/<platform>/
  examples/
  libs/raylib/raw.dd
```

## Examples

```bash
datadream check examples/coin-runner/game.dd --codegen
cd examples/coin-runner && ../../datadream build game.dd -o coin-runner
```

| Example | Shows |
|---------|-------|
| `examples/beginner/clicker.dd` | Beginner tutorial — mouse, screen, random |
| `examples/raylib/hello_friendly.dd` | App + friendly draw |
| `examples/raylib/hello_3d.dd` | Raw 3D raylib (v1 target) |
| `examples/raylib/control_flow.dd` | `loop`, `defer`, `match` |
| `examples/raylib/commands.dd` | All friendly namespaces |
| `examples/coin-runner/game.dd` | Sprites, input, collision |
| `examples/raylib/hello_raw.dd` | Raw raylib |
| `examples/raylib/array_demo.dd` | `Array<T>`, push, for-in, struct pointers |
| `examples/raylib/string_for_in.dd` | String byte iteration |
| `examples/colors/css_demo.dd` | CSS color names |
