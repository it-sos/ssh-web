package config

import (
	"encoding/base64"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	key := base64.StdEncoding.EncodeToString([]byte(generateRandomString(32)))
	plaintext := "mysecretpassword"

	encrypted, err := Encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}
	if encrypted == plaintext {
		t.Error("encrypted should differ from plaintext")
	}

	decrypted, err := Decrypt(key, encrypted)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}
	if decrypted != plaintext {
		t.Errorf("expected %q, got %q", plaintext, decrypted)
	}
}

func TestDecrypt_WrongKey(t *testing.T) {
	key1 := base64.StdEncoding.EncodeToString([]byte(generateRandomString(32)))
	key2 := base64.StdEncoding.EncodeToString([]byte(generateRandomString(32)))
	plaintext := "mysecretpassword"

	encrypted, err := Encrypt(key1, plaintext)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Decrypt(key2, encrypted)
	if err == nil {
		t.Error("expected error when decrypting with wrong key")
	}
}
