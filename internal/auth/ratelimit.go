package auth

import (
	"sync"
	"time"
)

type RateLimiter struct {
	mu           sync.Mutex
	attempts     map[string]*attemptInfo
	maxAttempts  int
	lockDuration time.Duration
}

type attemptInfo struct {
	count     int
	firstFail time.Time
	locked    bool
	lockUntil time.Time
}

func NewRateLimiter(maxAttempts int, lockDuration time.Duration) *RateLimiter {
	return &RateLimiter{
		attempts:     make(map[string]*attemptInfo),
		maxAttempts:  maxAttempts,
		lockDuration: lockDuration,
	}
}

func (rl *RateLimiter) Allow(identifier string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	info, exists := rl.attempts[identifier]
	if !exists {
		rl.attempts[identifier] = &attemptInfo{
			count:     1,
			firstFail: time.Now(),
		}
		return true
	}

	if info.locked {
		if time.Now().After(info.lockUntil) {
			info.count = 1
			info.locked = false
			info.firstFail = time.Now()
			return true
		}
		return false
	}

	if info.count >= rl.maxAttempts {
		info.locked = true
		info.lockUntil = time.Now().Add(rl.lockDuration)
		return false
	}

	info.count++
	return true
}
