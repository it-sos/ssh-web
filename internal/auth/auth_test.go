package auth

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestVerifyPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)

	if !VerifyPassword("correct", string(hash)) {
		t.Error("expected password verification to succeed")
	}

	if VerifyPassword("wrong", string(hash)) {
		t.Error("expected password verification to fail")
	}
}

func TestVerifyPassword_InvalidHash(t *testing.T) {
	if VerifyPassword("test", "invalid_hash") {
		t.Error("expected verification to fail with invalid hash")
	}
}
