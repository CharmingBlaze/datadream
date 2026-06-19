#!/usr/bin/env bash
# Verify a release zip: unzip to temp dir, run packdist --verify-only.
# Usage: ./scripts/verify-dist.sh dist/datadream-linux-amd64.zip

set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

ZIP="${1:?usage: $0 dist/datadream-<platform>.zip}"
if [ ! -f "$ZIP" ]; then
  echo "error: zip not found: $ZIP" >&2
  exit 1
fi

if [ ! -x ./packdist ] && [ ! -x ./packdist.exe ]; then
  echo "Building packdist..."
  go build -o packdist ./tools/packdist
fi
PACKDIST=./packdist
if [ -x ./packdist.exe ]; then
  PACKDIST=./packdist.exe
fi

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

echo "Extracting $ZIP..."
unzip -q "$ZIP" -d "$TMP"

echo "Verifying unpacked tree..."
"$PACKDIST" --verify-only "$TMP"

echo "Done: $ZIP"
