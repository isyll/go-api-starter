package auth

import "golang.org/x/crypto/bcrypt"

const minPasswordLen = 8

func hashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func checkPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

func validatePassword(plain string) error {
	if len(plain) < minPasswordLen {
		return ErrWeakPassword
	}
	return nil
}
