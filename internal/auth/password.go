package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const minPasswordLen = 8

// Verify reads cost params from the hash, so old hashes stay valid.
type passwordHasher struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

func newPasswordHasher(memory, iterations uint32, parallelism uint8, saltLength, keyLength uint32) passwordHasher {
	h := passwordHasher{memory, iterations, parallelism, saltLength, keyLength}
	if h.memory == 0 {
		h.memory = 64 * 1024
	}
	if h.iterations == 0 {
		h.iterations = 3
	}
	if h.parallelism == 0 {
		h.parallelism = 2
	}
	if h.saltLength == 0 {
		h.saltLength = 16
	}
	if h.keyLength == 0 {
		h.keyLength = 32
	}
	return h
}

func (h passwordHasher) hash(plain string) (string, error) {
	salt := make([]byte, h.saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("argon2: read salt: %w", err)
	}
	key := argon2.IDKey([]byte(plain), salt, h.iterations, h.memory, h.parallelism, h.keyLength)
	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, h.memory, h.iterations, h.parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

func verifyPassword(encoded, plain string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false
	}
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return false
	}
	var memory, iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}
	got := argon2.IDKey([]byte(plain), salt, iterations, memory, parallelism, uint32(len(want))) //nolint:gosec // decoded hash length, never near uint32 range
	return subtle.ConstantTimeCompare(got, want) == 1
}

func validatePassword(plain string) error {
	if len(plain) < minPasswordLen {
		return ErrWeakPassword
	}
	return nil
}
