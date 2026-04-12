package auth

import (
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
)

func TestVerifyTOTP(t *testing.T) {
	secret := "JBSWY3DPEHPK3PXP" // Valid base32 secret

	code, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	if !VerifyTOTP(secret, code) {
		t.Error("expected TOTP verification to succeed")
	}

	if VerifyTOTP(secret, "000000") {
		t.Error("expected TOTP verification to fail with wrong code")
	}
}

func TestVerifyTOTP_InvalidSecret(t *testing.T) {
	if VerifyTOTP("", "123456") {
		t.Error("expected TOTP verification to fail with empty secret")
	}
}
