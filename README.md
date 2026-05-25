# go-mobile-experiment

A gomobile-based experiment: one shared Go core powers both Android and
iOS apps, talking to a gRPC backend.

## Structure

```
.
├── mobile/                  # Everything that ships in the mobile apps
│   ├── apps/
│   │   ├── android/         # Kotlin + Compose
│   │   └── ios/             # SwiftUI
│   ├── core/                # FFI wrapper layer — gomobile binds THIS.
│   │                        # Thin: each exported function/type is a
│   │                        # gomobile-friendly shape delegating to
│   │                        # mobile/internal/* for actual logic.
│   └── internal/            # Mobile-only Go. Full Go (maps, slices of
│       ├── greet/           # structs, interfaces — all fine here).
│       ├── profile/         # Go's `internal/` rule means nothing outside
│       └── rpcclient/       # mobile/ can import these.
│
├── server/                  # gRPC backend
│   ├── main.go              # entry — binds 0.0.0.0:7777 by default
│   └── svc/                 # service implementations (importable —
│                            # integration tests reach in here)
│
├── api/                     # gRPC service definitions
│   ├── *.proto              # source of truth
│   └── gen/go/              # generated Go stubs (imported by both
│                            # mobile/internal/rpcclient + server/svc)
│
├── integration/             # End-to-end tests: spin up real server,
│                            # call through mobile/core, assert.
│
├── tools/                   # Dev tooling (not shipped, not the product)
│   └── buildhost/           # Build broker — lets Linux devs run macOS-only
│                            # builds remotely on a Mac in runner mode.
│
├── scripts/                 # Mobile build + dispatch scripts
├── Taskfile.yml
├── go.mod
└── go.sum
```

## How the boundaries work

Go's `internal/foo` is importable only by code in the directory containing
`internal/`. So `mobile/internal/greet` is reachable only from `mobile/*` —
the server physically cannot import it.

`server/svc/` is intentionally *not* under `internal/`: the handlers are the
server's public mounting points, and the top-level integration test needs to
mount them on a test gRPC server. Anything truly server-private (db pool,
auth middleware, config) will live in a future `server/internal/`.

The only place mobile and server both see is `api/gen/go/` — the gRPC
contract. If something genuinely needs to be shared (rare), it lives in
`api/` or a top-level non-`internal/` package.

## The wrapper + internal pattern

`mobile/core/` is constrained by what gomobile can cross the FFI boundary:
no maps, no `[]Struct` (only `[]byte`), no rich interfaces, no channels,
no generics. If business logic lived in `mobile/core/`, those constraints
would contaminate it.

So `mobile/core/` is a **thin translation layer**. Each exported function
and type is a gomobile-shaped surface that delegates into `mobile/internal/*`,
where Go is unconstrained.

```go
// mobile/core/profile.go — gomobile-friendly shape
package core

import "github.com/saborrie/go-mobile-experiment/mobile/internal/profile"

type Profile struct {
    Id   string  // exported fields with simple types
    Name string
}

func FetchProfile(id string) (*Profile, error) {
    p, err := profile.Fetch(id)
    if err != nil { return nil, err }
    return &Profile{Id: p.Id, Name: p.Name}, nil
}
```

```go
// mobile/internal/profile/profile.go — rich Go can live here
package profile

type Profile struct { ... }
func Fetch(id string) (*Profile, error) { ... }
```

For a `Profile` with two string fields, the wrapper looks redundant. The
moment it grows a `[]Friend` field or a `map[string]string` metadata bag,
the wrapper translates them into FFI-friendly shapes (joined string,
flattened keys/values arrays, etc.) without forcing those compromises
into the real model.

## Rationale per directory

| Dir                | Why                                                                                                                             |
| ------------------ | ------------------------------------------------------------------------------------------------------------------------------- |
| `mobile/`          | Everything shipping in the mobile apps. A parallel `web/` could appear later at the same level.                                 |
| `mobile/apps/`     | Platform-specific apps (Android, iOS). Distinguishes them from `mobile/core/` (the shared library they consume).                |
| `mobile/core/`     | FFI surface gomobile binds. Kept thin; just shape-translation.                                                                  |
| `mobile/internal/` | Mobile-only Go code. Sub-packages by domain. Scoped via Go's `internal/` rule.                                                  |
| `tools/`           | Dev tooling — used during development but not shipped or run as the product.                                                    |
| `scripts/`         | Build + dispatch scripts. Currently all mobile-specific. Split into `scripts/{mobile,backend}/` if/when backend scripts appear. |

## Bind / build / run

Daily loops:

```
task test               # Go unit tests, prod + server (~1 ms)
task test:demo          # same with -tags=demo, exercises scenarios
task test:integration   # spin up real server, call via mobile/core, assert
task server             # run the gRPC backend on 0.0.0.0:7777
task android:run        # build + install + launch on Android phone
task ios:run            # build + install + launch on iOS simulator (Mac satellite)
```

Less frequent:

```
task buildhost      # start the iOS build broker (Linux dev side)
task bind:android   # produce build/core.aar only
task bind:ios       # produce build/Core.xcframework only
task android:test   # instrumented tests against the demo flavor
task proto:gen      # regenerate api/gen/go/* from api/*.proto
```

## Pointing the apps at a dev backend

Both apps read the gRPC endpoint at build time. Set `BACKEND_ENDPOINT` in
your shell before running the app task:

```bash
export BACKEND_ENDPOINT=192.168.1.111:7777   # your LAN IP + server port
task server                                  # in one terminal
task android:run                             # in another
task ios:run                                 # or this
```

- **Android**: Gradle reads `BACKEND_ENDPOINT` and exposes it via
  `BuildConfig.BACKEND_ENDPOINT`. `MainActivity.kt` calls
  `Core.setEndpoint(BuildConfig.BACKEND_ENDPOINT)` at startup. Default
  (when unset): `10.0.2.2:7777` — the emulator's host-loopback alias.
- **iOS**: the dispatcher (`scripts/run-on-mac.sh`) generates a one-off
  wrapper script that exports `BACKEND_ENDPOINT` in the Mac runner's env
  before invoking `scripts/run-ios-app.sh`. xcodegen substitutes the env
  var into `Info.plist` (`INFOPLIST_KEY_BackendEndpoint`).
  `GoMobileExperimentApp.swift` reads it via `Bundle.main` and calls
  `CoreSetEndpoint`. Default: `localhost:7777`.

The wrapper trick keeps the broker protocol unchanged — no env-var
passthrough plumbing in `tools/buildhost/`; the dispatcher bakes config
in on the Linux side and ships it inside the tarball.

## Where the gRPC step landed

- Generated code lives in `api/gen/go/` (committed); regenerate with
  `task proto:gen` after editing `api/*.proto`.
- Mobile apps never see gRPC types. `mobile/internal/rpcclient/` calls
  gRPC; `mobile/core/` exposes only FFI-friendly `Profile{Id,Name}` etc.
  to the apps.
- Single `go.mod` at root so `mobile/internal/rpcclient` and `server/svc`
  share generated proto code.
- Endpoint configurable at runtime via `Core.SetEndpoint(addr)` (the apps
  call this at startup with their dev/prod backend URL).

## Single-binary-per-domain assumption

`server/main.go` and `tools/buildhost/main.go` skip the standard Go
`cmd/<binary>/main.go` nesting because each domain currently has exactly
one binary. If a second server-side binary appears (admin CLI, migration
tool), we'd restructure to `server/cmd/server/main.go` +
`server/cmd/<other>/main.go`. Easy switch later; verbose now.
