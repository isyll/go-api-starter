package utils

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsMutationMethod(t *testing.T) {
	tests := []struct {
		method   string
		expected bool
	}{
		{http.MethodPost, true},
		{http.MethodPut, true},
		{http.MethodPatch, true},
		{http.MethodDelete, true},
		{http.MethodGet, false},
		{http.MethodHead, false},
		{http.MethodOptions, false},
	}

	for _, tc := range tests {
		t.Run(tc.method, func(t *testing.T) {
			assert.Equal(t, tc.expected, IsMutationMethod(tc.method))
		})
	}
}

func TestGetAuthMessageKey(t *testing.T) {
	tests := []struct {
		method   string
		expected string
	}{
		{http.MethodPost, "auth.required.write"},
		{http.MethodPut, "auth.required.write"},
		{http.MethodPatch, "auth.required.write"},
		{http.MethodDelete, "auth.required.write"},
		{http.MethodGet, "auth.required.read_only"},
		{http.MethodHead, "auth.required.read_only"},
		{http.MethodOptions, "auth.required.read_only"},
	}

	for _, tc := range tests {
		t.Run(tc.method, func(t *testing.T) {
			assert.Equal(
				t, tc.expected, GetAuthMessageKey(tc.method),
			)
		})
	}
}
