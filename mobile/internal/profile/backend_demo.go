//go:build demo

package profile

import "errors"

var currentScenario = "happy"

// SetScenario picks which canned response the next Fetch will return.
// Only present in the demo build. Production code that tries to import or
// call this fails at compile time, by design.
func SetScenario(name string) {
	currentScenario = name
}

func fetch(id string) (*Profile, error) {
	switch currentScenario {
	case "happy":
		return &Profile{Id: id, Name: "Demo User"}, nil
	case "not-found":
		return nil, errors.New("profile not found")
	case "network-error":
		return nil, errors.New("network error: connection refused")
	default:
		return nil, errors.New("unknown scenario: " + currentScenario)
	}
}
