package main

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// This Lua script runs ATOMICALLY on the Redis server.
// It ensures that "Reading", "Refilling", and "Writing" happen
// in a single instant so no two requests can race each other.
// Keys: [1] "rate_limit:<ip>"
// Args: [1] max_tokens, [2] refill_rate, [3] now_timestamp
var requestScript = redis.NewScript(`
    local key = KEYS[1]
    local max_tokens = tonumber(ARGV[1])
    local refill_rate = tonumber(ARGV[2])
    local now = tonumber(ARGV[3])

    -- 1. Get current state from Redis (Hash Map)
    local state = redis.call("HMGET", key, "tokens", "last_refill")
    local tokens = tonumber(state[1])
    local last_refill = tonumber(state[2])

    -- 2. If no state exists, initialize it (Full Bucket)
    if not tokens then
        tokens = max_tokens
        last_refill = now
    end

    -- 3. Refill Logic
    local elapsed = now - last_refill
    local tokens_to_add = elapsed * refill_rate
    tokens = math.min(tokens + tokens_to_add, max_tokens)

    -- 4. Consume Logic
    local allowed = 0
    if tokens >= 1 then
        allowed = 1
        tokens = tokens - 1
    end

    -- 5. Save State (and set expiration to clean up old keys automatically)
    redis.call("HMSET", key, "tokens", tokens, "last_refill", now)
    redis.call("EXPIRE", key, 60) -- Auto-delete key after 60 seconds of inactivity

    -- Return [allowed (1/0), remaining_tokens]
    return { allowed, tokens }
`)

type RateLimiter struct {
	client *redis.Client
}

func NewRateLimiter(client *redis.Client) *RateLimiter {
	return &RateLimiter{client: client}
}

// Allow asks Redis if the user can pass.
// Returns: (allowed bool, remaining_tokens float64)
func (rl *RateLimiter) Allow(ip string, maxTokens float64, refillRate float64) (bool, float64) {
	ctx := context.Background()

	// Convert current time to seconds (Unix Timestamp)
	now := float64(time.Now().UnixNano()) / 1e9

	// Run the Lua script on Redis
	result, err := requestScript.Run(ctx, rl.client, []string{"rate_limit:" + ip}, maxTokens, refillRate, now).Result()
	if err != nil {
		// If Redis is down, we Default to Failing Closed (Block the request)
		// Alternative: You could return "true" to Fail Open (Let traffic through)
		return false, 0
	}

	// Parse the results from Lua
	resSlice := result.([]interface{})
	allowed := resSlice[0].(int64) == 1
	remaining := 0.0

	// Handle Go interface{} type assertion
	// Redis Lua sometimes returns different number types depending on the version
	switch v := resSlice[1].(type) {
	case int64:
		remaining = float64(v)
	case float64:
		remaining = v
	case string:
		// Fallback if needed
	}

	return allowed, remaining
}
