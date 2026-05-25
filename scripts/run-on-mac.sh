#!/usr/bin/env bash
# Run a build script on a Mac — locally if we ARE a Mac, otherwise via the
# buildhost broker. Used by Taskfile to keep individual tasks one-liners.
#
# Usage: ./scripts/run-on-mac.sh <build-script> [extra args...]
set -euo pipefail

SCRIPT="${1:-}"
if [ -z "$SCRIPT" ]; then
  echo "usage: $0 <build-script> [args...]" >&2
  exit 2
fi

if [ "$(uname -s)" = "Darwin" ]; then
  exec "$@"
fi

SERVER="${BUILDHOST:-http://localhost:3003}"
TARBALL="$(mktemp -t source.XXXXXX.tar.gz)"
OUTPUT="$(mktemp -t output.XXXXXX.log)"
trap 'rm -f "$TARBALL" "$OUTPUT"' EXIT

cd "$(dirname "$0")/.."

echo "Tarballing source (mobile + scripts)..."
tar czf "$TARBALL" go.mod go.sum mobile scripts

# Normalize "./scripts/foo.sh" → "scripts/foo.sh" for the broker.
SCRIPT_REL="${SCRIPT#./}"

echo "Submitting to $SERVER (script=$SCRIPT_REL, streaming output)..."
echo "----------------------------------------"
curl -sS -N -X POST --data-binary "@$TARBALL" \
  "$SERVER/build?script=$SCRIPT_REL" | tee "$OUTPUT"
echo "----------------------------------------"

if tail -10 "$OUTPUT" | grep -q "^__BUILD_STATUS:ok__$"; then
  echo "SUCCEEDED"
elif tail -10 "$OUTPUT" | grep -q "^__BUILD_STATUS:fail__"; then
  echo "FAILED"
  exit 1
else
  echo "UNKNOWN status — no marker found (connection dropped?)"
  exit 2
fi
