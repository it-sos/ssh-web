package auth

import (
	"github.com/pquerna/otp/totp"
)

func VerifyTOTP(secret, code string) bool {
	if secret == "" || len(code) != 6 {
		return false
	}
	return totp.Validate(code, secret)
}
