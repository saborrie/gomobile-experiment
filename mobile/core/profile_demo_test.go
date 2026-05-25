//go:build demo

package core

import "testing"

func TestFetchProfile_Scenarios(t *testing.T) {
	SetScenario("happy")
	p, err := FetchProfile("user-1")
	if err != nil {
		t.Fatalf("happy: unexpected error: %v", err)
	}
	if p.Name != "Demo User" || p.Id != "user-1" {
		t.Errorf("happy: got %+v", p)
	}

	SetScenario("not-found")
	if _, err := FetchProfile("x"); err == nil {
		t.Error("not-found: expected error, got nil")
	}

	SetScenario("network-error")
	if _, err := FetchProfile("x"); err == nil {
		t.Error("network-error: expected error, got nil")
	}
}
