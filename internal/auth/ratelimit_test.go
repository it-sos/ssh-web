package auth

import (
	"testing"
	"time"
)

func TestRateLimiter_AllowsBeforeLimit(t *testing.T) {
	rl := NewRateLimiter(5, 15*time.Minute)

	for i := 0; i < 5; i++ {
		if !rl.Allow("user1") {
			t.Errorf("request %d should be allowed", i+1)
		}
	}
}

func TestRateLimiter_BlocksAfterLimit(t *testing.T) {
	rl := NewRateLimiter(3, 15*time.Minute)

	for i := 0; i < 3; i++ {
		rl.Allow("user1")
	}

	if rl.Allow("user1") {
		t.Error("expected request to be blocked after limit")
	}
}

func TestRateLimiter_ResetsAfterTimeout(t *testing.T) {
	rl := NewRateLimiter(2, 10*time.Millisecond)

	rl.Allow("user1")
	rl.Allow("user1")
	// 3rd call should be blocked
	if rl.Allow("user1") {
		t.Error("expected 3rd request to be blocked")
	}
	time.Sleep(20 * time.Millisecond)

	if !rl.Allow("user1") {
		t.Error("expected request to be allowed after timeout")
	}
}
