#!/usr/bin/env bash
# Build every .dd example under examples/ (smoke test for releases and CI).
# Usage: ./scripts/build-all-examples.sh

set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if [ ! -x ./datadream ]; then
  echo "Building datadream..."
  go build -o datadream ./cmd/datadream
fi

OUTDIR="$(mktemp -d)"
trap 'rm -rf "$OUTDIR"' EXIT

count=0
while IFS= read -r -d '' f; do
  count=$((count + 1))
  base="$(basename "$f" .dd)"
  dir="$(dirname "$f")"
  echo "build $f"
  if [ "$(basename "$dir")" = "coin-runner" ]; then
    ( cd "$dir" && ../../datadream build "$(basename "$f")" -o "$OUTDIR/$base" )
  else
    ./datadream build "$f" -o "$OUTDIR/$base"
  fi
done < <(find examples -name '*.dd' -print0 | sort -z)

echo "Done: built $count examples"
