package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSecureToken(t *testing.T) {
	t.Run("generates non-empty token", func(t *testing.T) {
		token, err := GenerateSecureToken(32)
		require.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("defaults length when zero", func(t *testing.T) {
		token, err := GenerateSecureToken(0)
		require.NoError(t, err)
		assert.NotEmpty(t, token)
		// 48 bytes → base64 = 64 chars
		assert.Len(t, token, 64)
	})

	t.Run("defaults length when negative", func(t *testing.T) {
		token, err := GenerateSecureToken(-1)
		require.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run(
		"different calls produce different tokens",
		func(t *testing.T) {
			t1, _ := GenerateSecureToken(32)
			t2, _ := GenerateSecureToken(32)
			assert.NotEqual(t, t1, t2)
		},
	)

	t.Run("token is url-safe base64", func(t *testing.T) {
		token, err := GenerateSecureToken(48)
		require.NoError(t, err)
		// URL-safe base64 should not contain + or /
		assert.NotContains(t, token, "+")
		assert.NotContains(t, token, "/")
		assert.NotContains(t, token, "=")
	})
}
