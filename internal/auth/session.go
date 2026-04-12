package auth

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"
	"time"
)

type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]session
	expiry   time.Duration
}

type session struct {
	userID    string
	expiresAt time.Time
}

func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string]session),
		expiry:   30 * time.Minute,
	}
}

func (s *SessionStore) SetExpiry(d time.Duration) {
	s.expiry = d
}

func (s *SessionStore) CreateSession(userID string) string {
	token := generateToken()
	s.mu.Lock()
	s.sessions[token] = session{
		userID:    userID,
		expiresAt: time.Now().Add(s.expiry),
	}
	s.mu.Unlock()
	return token
}

func (s *SessionStore) ValidateSession(token string) (string, bool) {
	s.mu.RLock()
	sess, ok := s.sessions[token]
	s.mu.RUnlock()

	if !ok || time.Now().After(sess.expiresAt) {
		if ok {
			s.mu.Lock()
			delete(s.sessions, token)
			s.mu.Unlock()
		}
		return "", false
	}

	return sess.userID, true
}

func (s *SessionStore) DeleteSession(token string) {
	s.mu.Lock()
	delete(s.sessions, token)
	s.mu.Unlock()
}

func (s *SessionStore) SetCookie(w http.ResponseWriter, r *http.Request, token string, secure bool) {
	cookie := &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(s.expiry),
		Secure:   secure,
	}
	http.SetCookie(w, cookie)
}

func (s *SessionStore) GetSessionToken(r *http.Request) string {
	cookie, err := r.Cookie("session")
	if err != nil {
		return ""
	}
	return cookie.Value
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
