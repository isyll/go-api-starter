package utils

import (
	"regexp"
	"strings"
)

var (
	multiSpace  = regexp.MustCompile(`\s+`)
	nonAlphaNum = regexp.MustCompile(`[^a-z0-9\s]`)
)

func SanitizeString(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)

	s = nonAlphaNum.ReplaceAllString(s, "")

	s = multiSpace.ReplaceAllString(s, " ")

	return s
}
