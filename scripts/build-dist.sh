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
for arg in "$@"; do
  case "$arg" in
    --skip-clang) SKIP_CLANG=true ;;
    --skip-verify) SKIP_VERIFY=true ;;
  esac
done

export CGO_ENABLED=0

echo "Building datadream compiler..."
go build -o datadream ./cmd/datadream

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
./packdist "${PACK_ARGS[@]}"

echo "Done: $OUT"
if [ "$SKIP_CLANG" = true ]; then
  echo "Note: zip omits bundled Clang — users must run: datadream sdk install clang"
fi
