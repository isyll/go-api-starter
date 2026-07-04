// Package id generates opaque identifiers.
package id

import (
	"strings"

	"github.com/google/uuid"
)

// NewUUIDNoDash returns a random UUID v4 with the dashes stripped.
func NewUUIDNoDash() string {
	return strings.ReplaceAll(uuid.NewString(), "-", "")
}
