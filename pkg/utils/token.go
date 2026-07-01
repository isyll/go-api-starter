package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// GenerateSecureToken returns a cryptographically secure random
// token encoded in URL-safe base64 with no padding.
func GenerateSecureToken(length int) (string, error) {
	if length <= 0 {
		length = 48 // default to 48 bytes (64 chars when base64 encoded)
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf(
			"failed to generate random token: %w",
			err,
		)
	}

	// Use URL-safe base64 encoding without padding
	token := base64.URLEncoding.WithPadding(base64.NoPadding).
		EncodeToString(bytes)
	return token, nil
}
