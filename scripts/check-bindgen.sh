#!/usr/bin/env bash
# Verify libs/raylib/raw.dd matches bindgen output from raylib.h.
# Usage: ./scripts/check-bindgen.sh

set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if [ ! -x ./datadream ]; then
  echo "Building datadream..."
  go build -o datadream ./cmd/datadream
fi

./datadream sdk install headers

GEN="$(mktemp)"
trap 'rm -f "$GEN"' EXIT

echo "Regenerating raw bindings..."
./datadream bind sdk/raylib/6.0/include/raylib.h --raw --out "$GEN"

if ! diff -q "$GEN" libs/raylib/raw.dd >/dev/null; then
  echo "error: libs/raylib/raw.dd is out of date. Regenerate with:" >&2
  echo "  datadream bind sdk/raylib/6.0/include/raylib.h --raw --out libs/raylib/raw.dd" >&2
  diff -u libs/raylib/raw.dd "$GEN" | head -80
  exit 1
fi

./datadream check --codegen libs/raylib/raw.dd

echo "Verifying infer return-type map..."
INF="$(mktemp)"
go run ./tools/infergen/main.go -raw libs/raylib/raw.dd -out "$INF"
if ! diff -q "$INF" internal/infer/raylib_returns_gen.go >/dev/null; then
  echo "error: internal/infer/raylib_returns_gen.go is out of date. Regenerate with:" >&2
  echo "  cd internal/infer && go generate ." >&2
  diff -u internal/infer/raylib_returns_gen.go "$INF" | head -40
  exit 1
fi

echo "Done: raw.dd matches bindgen output"
