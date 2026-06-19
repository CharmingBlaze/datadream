#!/usr/bin/env bash
# Download Monaco editor into internal/ide/web/vendor for offline IDE use.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DEST="$ROOT/internal/ide/web/vendor/monaco/min"
MARKER="$DEST/vs/loader.js"

if [ -f "$MARKER" ]; then
  echo "Monaco already present: $MARKER"
  exit 0
fi

echo "Fetching monaco-editor@0.45.0..."
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT
cd "$TMP"
npm pack monaco-editor@0.45.0 >/dev/null
tar -xf monaco-editor-*.tgz
mkdir -p "$(dirname "$DEST")"
rm -rf "$DEST"
cp -R package/min "$(dirname "$DEST")/min"
if [ ! -f "$MARKER" ]; then
  echo "error: Monaco install incomplete" >&2
  exit 1
fi
echo "Installed Monaco to $DEST"
