package utils

import (
	"regexp"
	"strings"
)

var (
	multiSpace  = regexp.MustCompile(`\s+`)
	nonAlphaNum = regexp.MustCompile(`[^a-z0-9\s]`)
)

// SanitizeString lowercases, trims, strips non-alphanumerics, and
// collapses whitespace.
// e.g. "  Hello, World!  " -> "hello world".
func SanitizeString(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)

	// remove special characters
	s = nonAlphaNum.ReplaceAllString(s, "")

	// collapse multiple spaces
	s = multiSpace.ReplaceAllString(s, " ")

	return s
}
