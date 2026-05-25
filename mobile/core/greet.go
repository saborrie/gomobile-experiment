// Package core is the shared library exposed to Android/iOS via gomobile bind.
//
// gomobile only exports a restricted subset of Go: signed ints/floats, bool,
// string, []byte, structs, and interfaces with simple method sets. No maps,
// no channels, no slices of non-byte types across the boundary.
package core

import "fmt"

// Hello returns a greeting. Exported free functions become static methods
// on the generated Java/Swift class.
func Hello(name string) string {
	if name == "" {
		name = "world"
	}
	return fmt.Sprintf("Hello, %s!", name)
}

// Greeter demonstrates a struct crossing the FFI boundary. The generated
// binding exposes it as a class with a constructor and methods.
type Greeter struct {
	Prefix string
}

func NewGreeter(prefix string) *Greeter {
	return &Greeter{Prefix: prefix}
}

func (g *Greeter) Greet(name string) string {
	return fmt.Sprintf("%s %s", g.Prefix, name)
}
