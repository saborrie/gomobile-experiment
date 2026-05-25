package core

// Profile is a user profile fetched from the backend. Exported so that
// gomobile generates a Java/Swift class with getters for Id and Name.
type Profile struct {
	Id   string
	Name string
}

// FetchProfile retrieves a user profile by id. The actual backend behavior
// is selected at build time by build tags: the production build talks to
// the real backend (see backend.go); the -tags=demo build returns canned
// scenarios (see backend_demo.go).
//
// On the Java side this method throws an Exception when err is non-nil.
func FetchProfile(id string) (*Profile, error) {
	return fetchProfile(id)
}
