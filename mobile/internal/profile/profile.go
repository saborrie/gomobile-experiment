// Package profile owns the user-profile data model and the fetch operation.
// The actual fetching backend is selected at build time by build tags:
// backend.go for production, backend_demo.go for -tags=demo (canned scenarios).
package profile

// Profile is the user-profile shape used inside Go. mobile/core/ wraps it
// for the FFI boundary; backend implementations and tests use it directly.
type Profile struct {
	Id   string
	Name string
}

// Fetch retrieves a profile by id. Delegates to fetch() which is provided
// by the build-tagged backend.go / backend_demo.go file.
func Fetch(id string) (*Profile, error) {
	return fetch(id)
}
