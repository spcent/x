package limiter

import (
	"sync"
	"time"
)

// TokenBucket is a token bucket rate limiter used to limit the frequency of requests.
// Its working principle is: over a period of time, it generates a certain number of tokens.
// For each request, if the number of tokens is ≥1, the request is allowed; otherwise, it is rejected.
// This ensures that the request frequency does not exceed the specified rate over a period of time.
type TokenBucket struct {
	rate     float64    // Number of tokens generated per minute
	capacity float64    // Capacity of the bucket
	tokens   float64    // Current number of tokens
	last     time.Time  // Last time tokens were generated
	lock     sync.Mutex // Mutex to protect concurrent access to tokens
}

func NewTokenBucket(rate float64, capacity int) *TokenBucket {
	return &TokenBucket{
		rate:     rate,
		capacity: float64(capacity),
		tokens:   float64(capacity),
		last:     time.Now(),
	}
}

func (tb *TokenBucket) Allow() bool {
	tb.lock.Lock()
	defer tb.lock.Unlock()

	now := time.Now()
	// Calculate time difference in minutes
	elapsed := now.Sub(tb.last).Minutes()
	tb.last = now

	tb.tokens += elapsed * tb.rate
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}
	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}
	return false
}
