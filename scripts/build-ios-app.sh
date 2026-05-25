#!/usr/bin/env bash
# Build the iOS app: bind + xcodegen + xcodebuild for the simulator.
# Assumes macOS + gomobile + Xcode CLI tools + xcodegen.
set -euo pipefail

if [ "$(uname -s)" != "Darwin" ]; then
  echo "ERROR: build-ios-app.sh assumes macOS. From Linux use 'task ios:build' which routes through the buildhost broker." >&2
  exit 1
fi

# xcodegen substitutes ${BACKEND_ENDPOINT} into the generated Info.plist;
# ensure it's set so generation doesn't produce an empty value.
export BACKEND_ENDPOINT="${BACKEND_ENDPOINT:-localhost:7777}"

echo "==> gomobile bind -target=ios"
./scripts/bind-ios.sh

echo "==> xcodegen generate"
(cd mobile/apps/ios && xcodegen generate)

echo "==> xcodebuild (iOS Simulator)"
xcodebuild \
  -project mobile/apps/ios/GoMobileExperiment.xcodeproj \
  -scheme GoMobileExperiment \
  -destination 'generic/platform=iOS Simulator' \
  -configuration Debug \
  build
