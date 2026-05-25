// Package core is the FFI surface gomobile binds for the mobile apps.
//
// Every exported function and type here is shaped to fit gomobile's
// supported subset (no maps, no [] of non-byte types, no generics, etc.).
// Real logic lives in mobile/internal/* — this package is just a thin
// translation layer between the FFI-friendly shapes and the rich Go API.
package core

import "github.com/saborrie/go-mobile-experiment/mobile/internal/greet"

// Hello returns a greeting. Exposed as a static method on the generated
// Core class in Kotlin (Core.hello) / Swift (CoreHello).
func Hello(name string) string {
	return greet.Hello(name)
}

// Greeter demonstrates a struct crossing the FFI boundary. The generated
// binding exposes it as a class with a constructor, a prefix property,
// and a greet(name) method.
type Greeter struct {
	Prefix string
}

func NewGreeter(prefix string) *Greeter {
	return &Greeter{Prefix: prefix}
}

func (g *Greeter) Greet(name string) string {
	return greet.Greet(g.Prefix, name)
}
