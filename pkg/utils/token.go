package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func GenerateSecureToken(length int) (string, error) {
	if length <= 0 {
		length = 48
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf(
			"failed to generate random token: %w",
			err,
		)
	}

	token := base64.URLEncoding.WithPadding(base64.NoPadding).
		EncodeToString(bytes)
	return token, nil
}
