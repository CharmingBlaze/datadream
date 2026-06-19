# Publishing DataDream v1.0

End users download prebuilt zips from GitHub Releases — no Go required.

## Prerequisites

- Push all changes to `main`
- GitHub CLI authenticated: `gh auth login`
- Maintainer machine with Go 1.22+ (only for triggering the workflow; artifacts are built on GitHub runners)

## Option A — GitHub Actions (recommended)

1. Open **Actions → Release → Run workflow**
2. Set **tag** to `v1.0.0`
3. Leave **skip_clang** and **skip_verify** as `false`
4. Set **draft** to `false` when ready to ship publicly
5. Wait for Windows, Linux, and macOS jobs + **Publish GitHub Release**

The `publish` job uploads:

- `datadream-windows-amd64.zip`
- `datadream-linux-amd64.zip`
- `datadream-darwin-arm64.zip`

Each zip includes **DataDream Studio** (desktop IDE), root launchers, and `GETTING_STARTED.txt`. End users can double-click the Studio launcher without adding anything to PATH.

Use **skip_studio: true** only for smaller test artifacts.

## Option B — Tag push

```bash
git tag v1.0.0
git push origin v1.0.0
```

Pushing a `v*` tag triggers the same release workflow (non-draft).

## Option C — Local smoke test before shipping

```powershell
# Windows
.\scripts\build-dist.ps1
.\scripts\verify-dist.ps1 dist\datadream-windows-amd64.zip
```

```bash
# Linux/macOS
./scripts/build-dist.sh
./scripts/verify-dist.sh dist/datadream-linux-amd64.zip
```

Then upload zips manually in GitHub **Releases → Draft a new release**.

## After publish

Update the root README link to point at the new release URL if needed:

https://github.com/CharmingBlaze/datadream/releases/latest

Verify cold install (CLI):

```bash
datadream doctor
datadream new my-game
cd my-game && datadream run game.dd
```

Verify Studio (recommended on Windows):

1. Unzip the release zip.
2. Double-click **`Start DataDream Studio.bat`** or **`DataDream Studio.exe`**.
3. Confirm `examples/beginner/clicker.dd` opens and **Ctrl+Enter** runs it.
