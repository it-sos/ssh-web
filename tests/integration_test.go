package tests

import (
	"testing"
	"time"

	"github.com/ssh-web/internal/auth"
)

func TestSessionExpiry(t *testing.T) {
	store := auth.NewSessionStore()
	store.SetExpiry(10 * time.Millisecond)

	token := store.CreateSession("admin")
	time.Sleep(20 * time.Millisecond)

	_, ok := store.ValidateSession(token)
	if ok {
		t.Error("Session should be expired")
	}
}
