package auth

import "golang.org/x/crypto/bcrypt"

// minPasswordLen is the minimum accepted password length.
const minPasswordLen = 8

// hashPassword returns the bcrypt hash of a plaintext password.
func hashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// checkPassword reports whether plain matches the stored bcrypt hash.
func checkPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

// validatePassword enforces the minimum strength policy.
func validatePassword(plain string) error {
	if len(plain) < minPasswordLen {
		return ErrWeakPassword
	}
	return nil
}
