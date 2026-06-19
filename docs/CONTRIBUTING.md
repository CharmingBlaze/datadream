# Contributing to DataDream

How to add features without derailing the project. **Read [VISION.md](VISION.md) first.**

---

## Golden rules

1. **Example first** — every language feature needs a `.dd` example that `datadream check` passes.
2. **Build second** — if it should run, `datadream build` must succeed (or document why not yet).
3. **Go only** — no Python in tooling. Embed tables in Go (see `internal/colors/namespace.go`).
4. **Small diffs** — one feature per change. Match existing style in the file you edit.
5. **No Go for users** — distribution must not require end users to install Go.
6. **Raylib 6.0** — pin version in `sdk/manifest.json`, `internal/sdk/version.go`, `libs/raylib/package.json`.
7. **Update docs** — [LANGUAGE.md](LANGUAGE.md), [SYNTAX.md](SYNTAX.md) (keywords/operators), [ROADMAP.md](ROADMAP.md), and [HANDOFF.md](HANDOFF.md) when status changes.

---

## Adding a new language feature

### Step 1 — Design check

- Friendly API, raw interop, or compiler core?
- Violates [VISION.md](VISION.md) anti-goals?
- Can existing C API exposure suffice?

### Step 2 — Implementation order

```
lexer → parser → ast (if new node) → codegen → example → test → docs
```

| Step | File(s) |
|------|---------|
| New keyword | `internal/lexer/lexer.go` |
| New syntax | `internal/parser/*.go` |
| New AST node | `internal/ast/*.go` |
| C output | `internal/codegen/*.go` |
| C runtime helper | `game_runtime.go`, `runtime.go`, `color_runtime.go` |
| Example | `examples/...` |
| Test | `internal/codegen/*_test.go` |
| Docs | `docs/LANGUAGE.md`, `docs/SYNTAX.md`, `docs/ROADMAP.md`, `docs/HANDOFF.md` |

### Step 3 — Friendly builtin pattern

For `foo.bar(...)` namespace APIs:

1. Parser: `expr.go` — namespace roots (`draw`, `input`, …)
2. Codegen: `expr.go` → `genNamespaceCall` or `draw.go` → `genDrawCall`
3. C helper if needed: `game_runtime.go` or `runtime.go`

### Step 4 — Verify

```bash
go build -o datadream ./cmd/datadream
datadream check path/to/example.dd
datadream check path/to/example.dd --codegen
datadream build path/to/example.dd -o /tmp/test
go test ./internal/...
```

---

## Adding a CLI command

1. Create `internal/cli/mycommand.go` with `func cmdMyCommand(args []string) int`
2. Register in `internal/cli/cli.go`
3. Document in `internal/cli/help.go`, `docs/LANGUAGE.md`, `docs/HANDOFF.md`

---

## Regenerating raylib bindings

```bash
datadream bind sdk/raylib/6.0/include/raylib.h \
  --raw --out libs/raylib/raw.dd

datadream check --codegen libs/raylib/raw.dd   # should finish quickly (<1s)

./scripts/check-bindgen.sh   # CI: verify raw.dd matches bindgen output
```

**Note:** Generated bindings may use C field names that match DataDream keywords (`shader`, `data`). The parser accepts these via `IsBindingName`. Regenerating bindings should not reintroduce infinite-parse bugs — run `go test ./internal/parser/... -run TestParseRaylibRawBindings`.

---

## PR / commit checklist

- [ ] `datadream check` on touched examples
- [ ] `datadream check --codegen` on touched examples (if codegen changed)
- [ ] `datadream build` on at least one affected example (or `./scripts/build-all-examples.sh` before release)
- [ ] `go test ./internal/...` passes
- [ ] [LANGUAGE.md](LANGUAGE.md) updated if user-visible
- [ ] [ROADMAP.md](ROADMAP.md) item marked done
- [ ] No Python files added
- [ ] No secrets committed
- [ ] Do not commit `sdk/toolchain/clang/` or large raylib libs

---

## Common mistakes

| Mistake | Fix |
|---------|-----|
| `internal/pkg` imports `internal/sdk` | Breaks build — keep pkg free of sdk |
| `check` only in CI | Also run `build-all-examples.sh` on examples |
| MSVC Clang on Windows without MinGW target | `datadream sdk install clang` |
| English-sentence draw syntax | `draw.text(..., { ... })` |
| Global `let x = sprite(...)` as C initializer | Use deferred `datadream_init_globals()` pattern |
| Distributing `go.mod` in release zip | `packdist` ships bin + sdk only |

---

## Who to ask (conceptually)

Finish [ROADMAP.md](ROADMAP.md) P0 (multi-platform builds + release) and P1 before editor, WASM, or web targets.

The language must **build and run games on all platforms** before new surfaces.
