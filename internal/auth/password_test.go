package auth

import (
	"strings"
	"testing"
)

func TestPasswordHashAndVerify(t *testing.T) {
	h := newPasswordHasher(0, 0, 0, 0, 0)

	encoded, err := h.hash("correct horse battery staple")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if !strings.HasPrefix(encoded, "$argon2id$") {
		t.Fatalf("expected argon2id PHC string, got %q", encoded)
	}

	if !verifyPassword(encoded, "correct horse battery staple") {
		t.Fatal("verify should succeed for the correct password")
	}
	if verifyPassword(encoded, "wrong password") {
		t.Fatal("verify should fail for a wrong password")
	}
	if verifyPassword("not-a-hash", "x") {
		t.Fatal("verify should fail for a malformed hash")
	}
}

func TestPasswordHashIsSalted(t *testing.T) {
	h := newPasswordHasher(0, 0, 0, 0, 0)
	a, _ := h.hash("same")
	b, _ := h.hash("same")
	if a == b {
		t.Fatal("two hashes of the same password must differ (random salt)")
	}
}
