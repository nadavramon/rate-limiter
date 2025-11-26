package main

import (
	"sync"
	"time"
)

// RateLimiter holds the state for a single user
type RateLimiter struct {
	tokens     float64    // Current number of tokens in the bucket
	maxTokens  float64    // Maximum bucket size (burst capacity)
	refillRate float64    // Tokens added per second
	lastRefill time.Time  // The last time we calculated the tokens
	mu         sync.Mutex // The lock to prevent race conditions
}

// NewRateLimiter creates a new bucket for a user
func NewRateLimiter(maxTokens, refillRate float64) *RateLimiter {
	return &RateLimiter{
		tokens:     maxTokens, // Start full
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request can pass.
// Returns true if allowed, false if rejected.
func (rl *RateLimiter) Allow() bool {
	// 1. LOCK the bucket so no other request can touch it
	rl.mu.Lock()
	defer rl.mu.Unlock() // Ensure we unlock even if the function crashes

	// 2. REFILL Logic (Lazy)
	now := time.Now()
	// Calculate how many seconds passed since last refill
	elapsed := now.Sub(rl.lastRefill).Seconds()

	// Calculate tokens to add: (seconds * rate)
	tokensToAdd := elapsed * rl.refillRate

	// Update tokens (but don't exceed maxTokens)
	if tokensToAdd > 0 {
		rl.tokens = rl.tokens + tokensToAdd
		if rl.tokens > rl.maxTokens {
			rl.tokens = rl.maxTokens
		}
		// Update the timestamp
		rl.lastRefill = now
	}

	// 3. CONSUME Logic
	if rl.tokens >= 1.0 {
		rl.tokens--
		return true // Allowed
	}

	return false // Rejected
}
