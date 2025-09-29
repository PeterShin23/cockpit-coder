package relay

import (
	"sync"
	"time"
)

type RateLimiter struct {
	mu           sync.Mutex
	tokens       float64
	maxTokens    float64
	refillRate   float64 // tokens per second
	lastRefill   time.Time
}

func NewRateLimiter(bytesPerSec int) *RateLimiter {
	return &RateLimiter{
		tokens:     float64(bytesPerSec),
		maxTokens:  float64(bytesPerSec),
		refillRate: float64(bytesPerSec),
		lastRefill: time.Now(),
	}
}

func (rl *RateLimiter) Allow(n int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()
	added := elapsed * rl.refillRate
	rl.tokens = rl.tokens + added
	if rl.tokens > rl.maxTokens {
		rl.tokens = rl.maxTokens
	}
	rl.lastRefill = now

	if rl.tokens >= float64(n) {
		rl.tokens -= float64(n)
		return true
	}
	return false
}
