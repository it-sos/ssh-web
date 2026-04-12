package ssh

import (
	"testing"
)

func TestNewClientConfig_Password(t *testing.T) {
	cfg, err := NewClientConfig(Config{
		Host:         "127.0.0.1",
		Port:         22,
		Username:     "test",
		AuthMethod:   "password",
		Password:     "testpass",
		HostKeyCheck: false,
	})
	if err != nil {
		t.Fatalf("NewClientConfig() error = %v", err)
	}
	if cfg.User != "test" {
		t.Errorf("expected username 'test', got %q", cfg.User)
	}
}
