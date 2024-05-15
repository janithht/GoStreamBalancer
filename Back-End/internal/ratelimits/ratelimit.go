package ratelimits

import (
	"sync"
	"time"
)

type RateLimiter struct {
	tokens    int
	capacity  int
	mu        sync.Mutex
	rate      time.Duration
	lastCheck time.Time
}

func NewRateLimiter(limit int, interval time.Duration) *RateLimiter {
	return &RateLimiter{
		capacity:  limit,
		tokens:    limit,
		rate:      interval,
		lastCheck: time.Now(),
	}
}

func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastCheck)

	regenTokens := int(elapsed / rl.rate)
	if regenTokens > 0 {
		rl.tokens = min(rl.capacity, rl.tokens+regenTokens)
		rl.lastCheck = now
	}

	if rl.tokens > 0 {
		rl.tokens--
		return true
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
