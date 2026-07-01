package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/isyll/go-api-starter/pkg/utils"
)

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "trims and lowercases input",
			input:    "  Hello World  ",
			expected: "hello world",
		},
		{
			name:     "removes special characters",
			input:    "Hello, World! #2026",
			expected: "hello world 2026",
		},
		{
			name:     "collapses mixed whitespace",
			input:    "hello\t\n  world",
			expected: "hello world",
		},
		{
			name:     "returns empty string for punctuation only",
			input:    " !@#$%^&*() ",
			expected: "",
		},
		{
			name:     "preserves alphanumeric words",
			input:    "App Ride 123",
			expected: "app_owner ride 123",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := utils.SanitizeString(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
