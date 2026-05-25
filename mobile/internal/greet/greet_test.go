package greet

import "testing"

func TestHello(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"Steven", "Hello, Steven!"},
		{"", "Hello, world!"},
	}
	for _, c := range cases {
		if got := Hello(c.in); got != c.want {
			t.Errorf("Hello(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestGreet(t *testing.T) {
	if got, want := Greet("Hey", "Steven"), "Hey Steven"; got != want {
		t.Errorf("Greet = %q, want %q", got, want)
	}
}
