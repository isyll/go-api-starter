package utils

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/google/uuid"
)

// GenerateRequestID returns a 32-character hex-encoded random
// identifier suitable for request correlation. Pulls 16 bytes from
// crypto/rand; on the unlikely RNG failure it falls back to a hex-
// encoded UUID v4 so callers never see an empty string.
func GenerateRequestID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return hex.EncodeToString([]byte(uuid.NewString()))
	}
	return hex.EncodeToString(bytes)
}

// GenerateNumericCode returns a numeric string of the requested
// length using crypto/rand for each digit. Used for OTP codes —
// math/rand is unsuitable because the codes are security-sensitive.
func GenerateNumericCode(length int) (string, error) {
	var sb strings.Builder
	sb.Grow(length)

	maxVal := big.NewInt(10)
	for range length {
		digit, err := rand.Int(rand.Reader, maxVal)
		if err != nil {
			return "", err
		}

		_, _ = sb.WriteString(digit.String())
	}
	return sb.String(), nil
}

// NewUUIDNoDash returns a UUID v4 with the dashes stripped, yielding
// a compact 32-character identifier. Used for Asynq task IDs and
// idempotency keys where shorter is preferable.
func NewUUIDNoDash() string {
	return strings.ReplaceAll(uuid.NewString(), "-", "")
}
