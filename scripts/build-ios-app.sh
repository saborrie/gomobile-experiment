#!/usr/bin/env bash
# Build the iOS app: bind + xcodegen + xcodebuild for the simulator.
# Assumes macOS + gomobile + Xcode CLI tools + xcodegen.
set -euo pipefail

if [ "$(uname -s)" != "Darwin" ]; then
  echo "ERROR: build-ios-app.sh assumes macOS. From Linux use 'task ios:build' which routes through the buildhost broker." >&2
  exit 1
fi

echo "==> gomobile bind -target=ios"
./scripts/bind-ios.sh

echo "==> xcodegen generate"
(cd ios && xcodegen generate)

echo "==> xcodebuild (iOS Simulator)"
xcodebuild \
  -project ios/GoMobileExperiment.xcodeproj \
  -scheme GoMobileExperiment \
  -destination 'generic/platform=iOS Simulator' \
  -configuration Debug \
  build
