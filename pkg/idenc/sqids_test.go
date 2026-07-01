package idenc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAlphabet = "abcdefghijklmnopqrstuvwxyz1234567890"

func TestNewSqidsEncoder(t *testing.T) {
	encoder := NewSqidsEncoder(testAlphabet, 9)
	require.NotNil(t, encoder)
}

func TestSqidsEncoder_EncodeAndDecode(t *testing.T) {
	encoder := NewSqidsEncoder(testAlphabet, 9)

	tests := []struct {
		name string
		id   int64
	}{
		{"encode id 1", 1},
		{"encode id 42", 42},
		{"encode id 1000", 1000},
		{"encode id 999999", 999999},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			encoded := encoder.Encode(tc.id)
			assert.NotEmpty(t, encoded)
			assert.GreaterOrEqual(t, len(encoded), 9)

			decoded, err := encoder.Decode(encoded)
			require.NoError(t, err)
			assert.Equal(t, tc.id, decoded)
		})
	}
}

func TestSqidsEncoder_DifferentIdsProduceDifferentHashes(
	t *testing.T,
) {
	encoder := NewSqidsEncoder(testAlphabet, 9)

	h1 := encoder.Encode(1)
	h2 := encoder.Encode(2)
	assert.NotEqual(t, h1, h2)
}

func TestSqidsEncoder_DifferentAlphabetsProduceDifferentIds(
	t *testing.T,
) {
	enc1 := NewSqidsEncoder(
		"abcdefghijklmnopqrstuvwxyz1234567890", 9,
	)
	enc2 := NewSqidsEncoder(
		"0987654321zyxwvutsrqponmlkjihgfedcba", 9,
	)

	h1 := enc1.Encode(42)
	h2 := enc2.Encode(42)
	assert.NotEqual(t, h1, h2)
}

func TestSqidsEncoder_DecodeInvalid(t *testing.T) {
	encoder := NewSqidsEncoder(testAlphabet, 9)

	t.Run("invalid string returns empty", func(t *testing.T) {
		_, err := encoder.Decode("!!invalid!!")
		assert.Error(t, err)
	})
}

func TestSqidsEncoder_MinLength(t *testing.T) {
	encoder := NewSqidsEncoder(testAlphabet, 12)

	encoded := encoder.Encode(1)
	assert.GreaterOrEqual(t, len(encoded), 12)
}
