# DataDream Studio (Desktop)

Turnkey IDE included in every release zip — **extract and double-click**, no installs.

## End users

| Platform | What to open |
|----------|----------------|
| Windows | `Start DataDream Studio.bat` or `DataDream Studio.exe` |
| Linux | `./start-studio.sh` |
| macOS | `DataDream Studio.app` |

See **GETTING_STARTED.txt** in the zip.

The IDE automatically finds the bundled SDK (`sdk/`), loads the editor **offline** (Monaco is embedded), and opens a sample project. Press **Ctrl+Enter** to run.

## What's bundled

- DataDream Studio (Wails desktop app)
- `datadream` CLI + Clang + raylib in `sdk/`
- Examples under `examples/`
- On Windows: WebView2 runtime embedded in the IDE build (`-webview2 embed`)

## Maintainers

Release builds run:

```powershell
.\scripts\build-dist.ps1      # Windows
./scripts/build-dist.sh       # Linux / macOS
```

This fetches Monaco (`scripts/fetch-monaco`), builds Studio with Wails, bundles Clang, and packs launchers via `tools/packdist`.

Skip IDE from a zip (smaller): `-SkipStudio` / `--skip-studio`
