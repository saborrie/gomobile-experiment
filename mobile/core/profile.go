package core

import "github.com/saborrie/go-mobile-experiment/mobile/internal/profile"

// Profile is the FFI-friendly user profile. Gomobile generates Java/Swift
// classes with getters for Id and Name. Currently identical in shape to
// internal/profile.Profile; the wrapper exists so that shape can diverge
// later (extra fields, computed properties, types unsupported by FFI)
// without forcing the internal model to compromise.
type Profile struct {
	Id   string
	Name string
}

// FetchProfile retrieves a user profile by id. Backend behavior is
// selected at build time by build tags in mobile/internal/profile/.
// On the Java side this method throws an Exception when err is non-nil.
func FetchProfile(id string) (*Profile, error) {
	p, err := profile.Fetch(id)
	if err != nil {
		return nil, err
	}
	return &Profile{Id: p.Id, Name: p.Name}, nil
}
