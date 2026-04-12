package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSession_CreateAndValidate(t *testing.T) {
	store := NewSessionStore()

	token := store.CreateSession("admin")
	if token == "" {
		t.Fatal("expected session token")
	}

	userID, ok := store.ValidateSession(token)
	if !ok || userID != "admin" {
		t.Errorf("expected user 'admin', got %q, ok=%v", userID, ok)
	}
}

func TestSession_Expired(t *testing.T) {
	store := NewSessionStore()
	store.SetExpiry(1 * time.Millisecond)

	token := store.CreateSession("admin")
	time.Sleep(10 * time.Millisecond)

	_, ok := store.ValidateSession(token)
	if ok {
		t.Error("expected session to be expired")
	}
}

func TestSession_SetCookie(t *testing.T) {
	store := NewSessionStore()
	token := store.CreateSession("admin")

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	store.SetCookie(w, req, token, false)

	resp := w.Result()
	cookies := resp.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	if !cookies[0].HttpOnly {
		t.Error("expected HttpOnly cookie")
	}
	if cookies[0].SameSite != http.SameSiteLaxMode {
		t.Error("expected SameSite=Lax")
	}
}
