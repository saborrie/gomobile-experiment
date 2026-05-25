#!/usr/bin/env bash
# Build build/Core.xcframework. Assumes macOS + gomobile + Xcode CLI tools.
#
# Environment variables (optional):
#   TAGS — comma-separated Go build tags, e.g. "demo"
#   OUT  — output path (default: build/Core.xcframework)
set -euo pipefail

if [ "$(uname -s)" != "Darwin" ]; then
  echo "ERROR: bind-ios.sh assumes macOS. From Linux use 'task ios:build' (or any iOS task) — they route through the buildhost broker to run on a Mac." >&2
  exit 1
fi

TAGS="${TAGS:-}"
OUT="${OUT:-build/Core.xcframework}"

mkdir -p "$(dirname "$OUT")"

EXTRA=()
[ -n "$TAGS" ] && EXTRA+=("-tags=$TAGS")

# ${EXTRA[@]+"${EXTRA[@]}"} expands only if the array is set; needed because
# `set -u` errors on a bare empty-array expansion.
gomobile bind -target=ios ${EXTRA[@]+"${EXTRA[@]}"} -o "$OUT" ./core
echo "Built $OUT"
