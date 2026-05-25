# go-mobile-experiment

A gomobile-based experiment: one shared Go core powers both Android and
iOS apps, with a gRPC backend planned as the next step.

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
│       └── profile/         # Go's `internal/` rule means nothing outside
│                            # mobile/ can import these.
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

Planned, not yet present:

```
├── server/                  # gRPC backend
│   ├── main.go
│   └── internal/{svc,db}/
└── api/                     # gRPC service definitions
    ├── *.proto
    └── gen/go/              # generated Go stubs
```

## Two `internal/` directories, no shared one

Go's `internal/foo` is only importable by code in the directory containing
`internal/`. So `mobile/internal/greet` is reachable only from `mobile/*` —
the server (once it exists) physically cannot import it. Same in reverse
once `server/internal/` lands.

The only thing both sides see is `api/gen/go/` — the gRPC contract. If
something genuinely needs to be shared (rare), it lives in `api/` or a
new top-level non-`internal/` package.

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

| Dir | Why |
|---|---|
| `mobile/` | Everything shipping in the mobile apps. A parallel `web/` could appear later at the same level. |
| `mobile/apps/` | Platform-specific apps (Android, iOS). Distinguishes them from `mobile/core/` (the shared library they consume). |
| `mobile/core/` | FFI surface gomobile binds. Kept thin; just shape-translation. |
| `mobile/internal/` | Mobile-only Go code. Sub-packages by domain. Scoped via Go's `internal/` rule. |
| `tools/` | Dev tooling — used during development but not shipped or run as the product. |
| `scripts/` | Build + dispatch scripts. Currently all mobile-specific. Split into `scripts/{mobile,backend}/` if/when backend scripts appear. |

## Bind / build / run

Daily loops:

```
task test           # Go unit tests (~1 ms)
task test:demo      # same with -tags=demo, exercises scenarios
task android:run    # build + install + launch on Android phone
task ios:run        # build + install + launch on iOS simulator (Mac satellite)
```

Less frequent:

```
task buildhost      # start the iOS build broker (Linux dev side)
task bind:android   # produce build/core.aar only
task bind:ios       # produce build/Core.xcframework only
task android:test   # instrumented tests against the demo flavor
```

## Open questions for the gRPC step

1. **Where does generated code live?** `api/gen/go/` (proposed) keeps
   source and output together. Alternative: regenerate-on-the-fly via
   `go generate` with no committed artefacts.
2. **Will mobile core's gRPC client be visible to the apps?** Proposal:
   no — `mobile/internal/rpcclient/` calls gRPC, `mobile/core/` exposes a
   simpler Go API to gomobile. gomobile's FFI rules would make raw gRPC
   types painful to surface anyway.
3. **One module or multiple?** Single `go.mod` at root is simplest while
   `mobile/core` and `server/` share generated proto code.

## Single-binary-per-domain assumption

`server/main.go` and `tools/buildhost/main.go` skip the standard Go
`cmd/<binary>/main.go` nesting because each domain currently has exactly
one binary. If a second server-side binary appears (admin CLI, migration
tool), we'd restructure to `server/cmd/server/main.go` +
`server/cmd/<other>/main.go`. Easy switch later; verbose now.
