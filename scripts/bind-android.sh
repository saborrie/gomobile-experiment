#!/usr/bin/env bash
# Build an Android .aar from the core Go package.
#
# Environment variables (optional):
#   TAGS  — comma-separated build tags (e.g. "demo")
#   OUT   — output path (defaults to build/core.aar)
#
# Examples:
#   ./scripts/bind-android.sh                                # production
#   TAGS=demo OUT=build/core-demo.aar ./scripts/bind-android.sh
set -euo pipefail

cd "$(dirname "$0")/.."
mkdir -p build

TAGS="${TAGS:-}"
OUT="${OUT:-build/core.aar}"

EXTRA_ARGS=()
if [ -n "$TAGS" ]; then
  EXTRA_ARGS+=("-tags=$TAGS")
fi

gomobile bind \
  -target=android \
  -androidapi=24 \
  ${EXTRA_ARGS[@]+"${EXTRA_ARGS[@]}"} \
  -o "$OUT" \
  ./mobile/core

echo "Built $OUT"
