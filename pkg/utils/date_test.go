package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/isyll/go-api-starter/pkg/utils"
)

func TestNormalizeTimezone(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string returns UTC",
			input:    "",
			expected: "UTC",
		},
		{
			name:     "UTC is preserved",
			input:    "UTC",
			expected: "UTC",
		},
		{
			name:     "utc lowercase returns UTC",
			input:    "utc",
			expected: "UTC",
		},
		{
			name:     "valid timezone preserved",
			input:    "America/New_York",
			expected: "America/New_York",
		},
		{
			name:     "Africa/Dakar preserved",
			input:    "Africa/Dakar",
			expected: "Africa/Dakar",
		},
		{
			name:     "Europe/Paris preserved",
			input:    "Europe/Paris",
			expected: "Europe/Paris",
		},
		{
			name:     "invalid timezone returns UTC",
			input:    "Invalid/Timezone",
			expected: "UTC",
		},
		{
			name:     "whitespace trimmed",
			input:    "  America/New_York  ",
			expected: "America/New_York",
		},
		{
			name:     "only whitespace returns UTC",
			input:    "   ",
			expected: "UTC",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := utils.NormalizeTimezone(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
