// Package codes is the registry of machine-readable API error codes.
package codes

import "fmt"

// Code is a machine-readable error code; the unexported field blocks outside construction.
type Code struct {
	value string
}

// String returns the UPPER_SNAKE_CASE wire format.
func (c Code) String() string { return c.value }

// IsZero reports whether the code is the zero Code.
func (c Code) IsZero() bool { return c.value == "" }

var allCodes = map[string]struct{}{}

// register enforces wire-format and uniqueness invariants; violations panic at init.
func register(value string) Code {
	if value == "" {
		panic("codes: empty code value")
	}
	if len(value) > 64 {
		panic(
			fmt.Sprintf("codes: code %q exceeds 64-char limit", value),
		)
	}
	for i := range len(value) {
		ch := value[i]
		switch {
		case ch >= 'A' && ch <= 'Z':
		case ch >= '0' && ch <= '9':
		case ch == '_':
		default:
			panic(
				fmt.Sprintf(
					"codes: code %q must be UPPER_SNAKE_CASE",
					value,
				),
			)
		}
	}
	if _, dup := allCodes[value]; dup {
		panic(fmt.Sprintf("codes: duplicate registration of %q", value))
	}
	allCodes[value] = struct{}{}
	return Code{value: value}
}

// All returns a snapshot of every registered code value.
func All() []string {
	out := make([]string, 0, len(allCodes))
	for k := range allCodes {
		out = append(out, k)
	}
	return out
}
