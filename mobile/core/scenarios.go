//go:build demo

package core

import "github.com/saborrie/go-mobile-experiment/mobile/internal/profile"

// SetScenario controls which canned response FetchProfile returns next.
// Only present in the demo build (-tags=demo); calling this from a
// production-flavor build fails at compile time, by design.
func SetScenario(name string) {
	profile.SetScenario(name)
}
