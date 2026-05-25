#!/usr/bin/env bash
# Build the iOS app, install it on a booted iPhone simulator, and launch it.
# Assumes macOS + gomobile + Xcode CLI tools + xcodegen.
set -euo pipefail

if [ "$(uname -s)" != "Darwin" ]; then
  echo "ERROR: run-ios-app.sh assumes macOS. From Linux use 'task ios:run'." >&2
  exit 1
fi

# xcodegen substitutes ${BACKEND_ENDPOINT} into the generated Info.plist;
# ensure it's set so generation doesn't produce an empty value.
export BACKEND_ENDPOINT="${BACKEND_ENDPOINT:-localhost:7777}"

echo "==> bind + xcodegen"
./scripts/bind-ios.sh
(cd mobile/apps/ios && xcodegen generate)

echo "==> xcodebuild (iOS Simulator)"
DERIVED="$(pwd)/mobile/apps/ios/build/derived"
xcodebuild \
  -project mobile/apps/ios/GoMobileExperiment.xcodeproj \
  -scheme GoMobileExperiment \
  -destination 'generic/platform=iOS Simulator' \
  -configuration Debug \
  -derivedDataPath "$DERIVED" \
  build

APP=$(find "$DERIVED" -name 'GoMobileExperiment.app' -type d | head -1)
[ -n "$APP" ] || { echo "couldn't find .app bundle under $DERIVED"; exit 1; }
echo "==> built: $APP"

BOOTED=$(xcrun simctl list devices booted 2>/dev/null \
  | grep -oE '[A-F0-9-]{36}' | head -1 || true)
if [ -z "$BOOTED" ]; then
  BOOTED=$(xcrun simctl list devices available \
    | grep -E '^\s+iPhone' | grep -oE '\(([A-F0-9-]+)\)' \
    | head -1 | tr -d '()')
  [ -n "$BOOTED" ] || { echo "no iPhone simulator available"; exit 1; }
  echo "==> booting $BOOTED"
  xcrun simctl boot "$BOOTED"
fi

open -a Simulator
xcrun simctl install "$BOOTED" "$APP"
xcrun simctl launch "$BOOTED" com.example.GoMobileExperiment
echo "==> launched on $BOOTED"
