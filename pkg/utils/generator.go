package utils

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/google/uuid"
)

func GenerateRequestID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return hex.EncodeToString([]byte(uuid.NewString()))
	}
	return hex.EncodeToString(bytes)
}

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

func NewUUIDNoDash() string {
	return strings.ReplaceAll(uuid.NewString(), "-", "")
}
