package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/isyll/go-grpc-starter/pkg/utils"
)

func TestGenerateRequestID(t *testing.T) {
	t.Run("generates non-empty string", func(t *testing.T) {
		id := utils.GenerateRequestID()
		assert.NotEmpty(t, id)
	})

	t.Run(
		"generates hex string of expected length",
		func(t *testing.T) {
			id := utils.GenerateRequestID()
			assert.Len(t, id, 32)
		},
	)

	t.Run("generates unique IDs", func(t *testing.T) {
		ids := make(map[string]bool)
		for range 100 {
			id := utils.GenerateRequestID()
			assert.False(t, ids[id], "duplicate ID generated")
			ids[id] = true
		}
	})
}

func TestGenerateNumericCode(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{name: "4 digits", length: 4},
		{name: "6 digits", length: 6},
		{name: "8 digits", length: 8},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			code, err := utils.GenerateNumericCode(tc.length)

			require.NoError(t, err)
			assert.Len(t, code, tc.length)

			for _, c := range code {
				assert.True(t, c >= '0' && c <= '9',
					"expected digit, got %c", c)
			}
		})
	}

	t.Run("generates unique codes", func(t *testing.T) {
		codes := make(map[string]bool)
		for range 100 {
			code, err := utils.GenerateNumericCode(6)
			require.NoError(t, err)
			codes[code] = true
		}
		assert.Greater(t, len(codes), 90)
	})
}

func TestNewUUIDNoDash(t *testing.T) {
	t.Run("generates 32 character string", func(t *testing.T) {
		uuid := utils.NewUUIDNoDash()
		assert.Len(t, uuid, 32)
	})

	t.Run("contains no dashes", func(t *testing.T) {
		uuid := utils.NewUUIDNoDash()
		assert.NotContains(t, uuid, "-")
	})

	t.Run("generates unique UUIDs", func(t *testing.T) {
		uuids := make(map[string]bool)
		for range 100 {
			uuid := utils.NewUUIDNoDash()
			assert.False(t, uuids[uuid], "duplicate UUID generated")
			uuids[uuid] = true
		}
	})
}
