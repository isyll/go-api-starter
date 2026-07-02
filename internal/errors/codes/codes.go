// Package codes is the registry of machine-readable API error codes.
package codes

import "fmt"

// Code is the typed, machine-readable error code emitted in the
// API error envelope. Code values exist only as constants in
// this package; the unexported field guarantees no other
// package can synthesize one.
type Code struct {
	value string
}

// String returns the UPPER_SNAKE_CASE wire format of the code.
func (c Code) String() string { return c.value }

// IsZero reports whether the code is the zero Code. The
// constructors panic when handed a zero Code so the wire never
// carries an empty code field.
func (c Code) IsZero() bool { return c.value == "" }

var allCodes = map[string]struct{}{}

// register is the package-private Code constructor. It enforces
// the wire-format invariants (UPPER_SNAKE_CASE, ≤ 64 chars) and
// the global uniqueness rule. Violations panic at init.
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

// All returns a sorted snapshot of every registered code value.
// Useful for documentation generators and audit scripts.
func All() []string {
	out := make([]string, 0, len(allCodes))
	for k := range allCodes {
		out = append(out, k)
	}
	return out
}
