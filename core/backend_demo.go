//go:build demo

package core

import "errors"

var currentScenario = "happy"

// SetScenario controls which canned response FetchProfile returns next.
// Only present in the demo build (-tags=demo). Production code that
// tries to call this fails to compile, by design.
func SetScenario(name string) {
	currentScenario = name
}

func fetchProfile(id string) (*Profile, error) {
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
