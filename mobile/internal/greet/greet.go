// Package greet holds the greeting logic. Lives in mobile/internal/ so it
// can use full Go (any types, generics, etc.) without worrying about what
// gomobile can cross the FFI boundary. The mobile/core/ package wraps it
// for the apps.
package greet

import "fmt"

// Hello returns a greeting for the given name. Empty name falls back to "world".
func Hello(name string) string {
	if name == "" {
		name = "world"
	}
	return fmt.Sprintf("Hello, %s!", name)
}

// Greet formats a greeting with the given prefix. Free function so the
// FFI-side Greeter wrapper can stay a plain struct without exporting any
// internal types across the boundary.
func Greet(prefix, name string) string {
	return fmt.Sprintf("%s %s", prefix, name)
}
