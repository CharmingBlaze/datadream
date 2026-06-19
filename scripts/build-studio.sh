#!/usr/bin/env bash
# Build DataDream Studio (Wails) for the current or selected platform.
#
# Usage:
#   ./scripts/build-studio.sh
#   ./scripts/build-studio.sh linux/amd64
#   ./scripts/build-studio.sh darwin/arm64
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

PLATFORM="${1:-$(go env GOOS)/$(go env GOARCH)}"

export GOFLAGS="-mod=mod"
export PATH="$(go env GOPATH)/bin:$PATH"

if ! command -v wails >/dev/null 2>&1; then
  echo "Installing Wails CLI..."
  go install github.com/wailsapp/wails/v2/cmd/wails@latest
fi

STUDIO="$ROOT/cmd/studio"
mkdir -p "$STUDIO/build/windows"

ICON="$STUDIO/build/appicon.png"
if [ ! -f "$ICON" ]; then
  if command -v magick >/dev/null 2>&1; then
    magick -size 256x256 xc:'#1D3A5F' "$ICON"
  elif command -v convert >/dev/null 2>&1; then
    convert -size 256x256 xc:'#1D3A5F' "$ICON"
  else
    echo "error: missing $ICON — install ImageMagick or run build-studio.ps1 once on Windows" >&2
    exit 1
  fi
fi

echo "Building DataDream Studio (Wails) for $PLATFORM..."
cd "$STUDIO"

WAILS_ARGS=(-clean -platform "$PLATFORM")
case "$PLATFORM" in
  windows/*)
    WAILS_ARGS+=(-webview2 embed)
    ;;
  linux/*)
    # Ubuntu 24.04+ and CI use webkit2gtk-4.1
    WAILS_ARGS+=(-tags webkit2_41)
    ;;
esac
wails build "${WAILS_ARGS[@]}"

case "$PLATFORM" in
  darwin/*)
    OUT="$STUDIO/build/bin/datadream-studio.app"
    ;;
  windows/*)
    OUT="$STUDIO/build/bin/datadream-studio.exe"
    ;;
  linux/*)
    chmod +x "$ROOT/scripts/build-studio-appimage.sh"
    "$ROOT/scripts/build-studio-appimage.sh"
    arch="$(uname -m)"
    case "$arch" in
      amd64) arch=x86_64 ;;
    esac
    OUT="$STUDIO/build/bin/datadream-studio-${arch}.AppImage"
    ;;
  *)
    OUT="$STUDIO/build/bin/datadream-studio"
    ;;
esac

if [ ! -e "$OUT" ]; then
  echo "error: build output not found: $OUT" >&2
  exit 1
fi

echo "Built: $OUT"
