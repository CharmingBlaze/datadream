# DataDream Studio (Desktop)

Turnkey IDE included in every release zip — **extract and double-click**, no installs.

## End users

| Platform | What to open |
|----------|----------------|
| Windows | `Start DataDream Studio.bat` or `DataDream Studio.exe` |
| Linux | `./start-studio.sh` or `datadream-studio-x86_64.AppImage` |
| macOS | `DataDream Studio.app` |

See **GETTING_STARTED.txt** in the zip.

The IDE automatically finds the bundled SDK (`sdk/`), loads the editor **offline** (Monaco is embedded), and opens a sample project. Press **Ctrl+Enter** to run.

## What's bundled

- DataDream Studio (Wails desktop app)
- `datadream` CLI + Clang + raylib in `sdk/`
- Examples under `examples/`
- On **Windows**: WebView2 runtime embedded in the IDE build (`-webview2 embed`)
- On **Linux**: self-contained **AppImage** with GTK + WebKit (no system packages required)

## Maintainers

Release builds run:

```powershell
.\scripts\build-dist.ps1      # Windows
./scripts/build-dist.sh       # Linux / macOS
```

On Linux this:

1. Fetches Monaco (`scripts/fetch-monaco`)
2. Builds Studio with Wails (`scripts/build-studio.sh`)
3. Bundles GTK/WebKit into an AppImage (`scripts/build-studio-appimage.sh`)
4. Packs launchers via `tools/packdist`

Studio only (dev):

```bash
./scripts/build-studio.sh
./scripts/build-studio-appimage.sh   # Linux only — after wails build
```

Skip IDE from a zip (smaller): `-SkipStudio` / `--skip-studio`
