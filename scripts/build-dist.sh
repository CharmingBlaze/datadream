#!/usr/bin/env bash
# Build a DataDream distribution (maintainers only — end users never need Go).
# Output: dist/datadream-<platform>.zip
#
# Usage:
#   ./scripts/build-dist.sh
#   ./scripts/build-dist.sh --skip-clang
#   ./scripts/build-dist.sh --skip-verify

set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

SKIP_CLANG=false
SKIP_VERIFY=false
SKIP_STUDIO=false
for arg in "$@"; do
  case "$arg" in
    --skip-clang) SKIP_CLANG=true ;;
    --skip-verify) SKIP_VERIFY=true ;;
    --skip-studio) SKIP_STUDIO=true ;;
  esac
done

export CGO_ENABLED=0

echo "Building datadream compiler..."
go build -mod=mod -o datadream ./cmd/datadream

echo "Fetching offline Monaco editor..."
chmod +x scripts/fetch-monaco.sh
./scripts/fetch-monaco.sh

echo "Installing raylib SDK..."
./datadream sdk install headers
./datadream sdk install raylib

if [ "$SKIP_CLANG" = false ]; then
  echo "Installing bundled Clang (large download)..."
  ./datadream sdk install clang
else
  echo "Skipping Clang install (--skip-clang)"
fi

echo "SDK doctor..."
./datadream doctor

echo "Smoke build hello_friendly..."
./datadream build examples/raylib/hello_friendly.dd -o hello_dist_smoke

echo "Smoke build hello_raw..."
./datadream build examples/raylib/hello_raw.dd -o hello_raw_dist_smoke

echo "Smoke build coin-runner..."
( cd examples/coin-runner && ../../datadream build game.dd -o coin_runner_smoke )

if [ "$SKIP_STUDIO" = false ]; then
  echo "Building DataDream Studio (Wails)..."
  chmod +x scripts/build-studio.sh
  ./scripts/build-studio.sh
else
  echo "Skipping Studio build (--skip-studio)"
fi

echo "Building packdist tool..."
go build -o packdist ./tools/packdist

PLAT="$(go env GOOS)-$(go env GOARCH)"
OUT="$ROOT/dist/datadream-$PLAT.zip"
mkdir -p "$(dirname "$OUT")"

echo "Packing distribution..."
PACK_ARGS=(--out "$OUT")
if [ "$SKIP_VERIFY" = false ]; then
  PACK_ARGS+=(--verify)
fi
if [ "$SKIP_STUDIO" = true ]; then
  PACK_ARGS+=(--skip-studio)
fi
./packdist "${PACK_ARGS[@]}"

echo "Done: $OUT"
if [ "$SKIP_CLANG" = true ]; then
  echo "Note: zip omits bundled Clang — users must run: datadream sdk install clang"
fi
