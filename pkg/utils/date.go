package utils

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	// TimezoneUTC is the canonical name for Coordinated Universal Time.
	TimezoneUTC = "UTC"
	// TimezoneLocal represents the Go runtime's local timezone.
	TimezoneLocal = "Local"
)

// NormalizeTimezone canonicalizes a timezone string to a valid IANA location
// name, falling back to "UTC" for empty, invalid, or unresolvable inputs.
func NormalizeTimezone(tz string) string {
	tz = strings.TrimSpace(tz)

	if tz == "" || strings.EqualFold(tz, TimezoneUTC) {
		return TimezoneUTC
	}

	if strings.EqualFold(tz, TimezoneLocal) {
		loc := time.Now().Location()
		if loc != nil && loc.String() != "" &&
			loc.String() != TimezoneLocal {
			return loc.String()
		}
		return TimezoneUTC
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		parts := strings.Split(tz, "/")
		if len(parts) == 2 {
			normalized := fmt.Sprintf(
				"%s/%s",
				cases.Title(language.English).String(parts[0]),
				cases.Title(language.English).String(parts[1]),
			)

			normalizedLoc, normalizedErr := time.LoadLocation(
				normalized,
			)
			if normalizedErr == nil {
				return normalizedLoc.String()
			}
		}

		return TimezoneUTC
	}

	return loc.String()
}
