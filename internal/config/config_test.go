package config

import (
	"encoding/base32"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestLoadConfig_CreatesDefault(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg.Auth.Username != "admin" {
		t.Errorf("expected username 'admin', got %q", cfg.Auth.Username)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Server.Port)
	}
	if cfg.EncryptionKey == "" {
		t.Error("expected encryption_key to be generated")
	}
	if cfg.Auth.TOTPSecret == "" {
		t.Error("expected totp_secret to be generated")
	}
}

func TestLoadConfig_ExistingConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	content := `
server:
  port: 9090
auth:
  username: "testuser"
  password_hash: "$2a$10$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ012"
  totp_secret: "TESTSECRET123"
encryption_key: "dGVzdC1lbmNyeXB0aW9uLWtleS0zMi1ieXRlcy1sb25n"
default_host:
  host: "10.0.0.1"
  port: 2222
  username: "deploy"
  auth_method: "private_key"
  private_key_path: "/home/user/.ssh/id_rsa"
  host_key_check: false
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg.Server.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Server.Port)
	}
	if cfg.Auth.Username != "testuser" {
		t.Errorf("expected username 'testuser', got %q", cfg.Auth.Username)
	}
	if cfg.DefaultHost.Host != "10.0.0.1" {
		t.Errorf("expected host '10.0.0.1', got %q", cfg.DefaultHost.Host)
	}
}

func TestTOTPSecret_IsValidBase32(t *testing.T) {
	dir := t.TempDir()
	cfg, err := LoadConfig(filepath.Join(dir, "config.yaml"))
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	secret := cfg.Auth.TOTPSecret

	// Must match Base32 alphabet: A-Z and 2-7 only
	validBase32 := regexp.MustCompile(`^[A-Z2-7]+$`)
	if !validBase32.MatchString(secret) {
		t.Errorf("TOTPSecret %q contains non-Base32 characters", secret)
	}

	// Must be decodable as Base32
	_, err = base32.StdEncoding.DecodeString(secret)
	if err != nil {
		t.Errorf("TOTPSecret %q is not valid Base32: %v", secret, err)
	}
}
