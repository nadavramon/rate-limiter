package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/redis/go-redis/v9"
)

// APIResponse is the shape of data we send back to the user
type APIResponse struct {
	Success   bool    `json:"success"`
	Message   string  `json:"message"`
	Remaining float64 `json:"remaining_tokens,omitempty"`
}

// Global Limiter Instance
var globalLimiter *RateLimiter

func rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		// Ask Redis: "Can this IP pass?"
		// Limits: 5 tokens, refill 1 per second
		allowed, remaining := globalLimiter.Allow(ip, 5.0, 1.0)

		if !allowed {
			http.Error(w, "429 Too Many Requests", http.StatusTooManyRequests)
			return
		}

		// Inject Headers
		w.Header().Set("X-RateLimit-Limit", "5")
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%.2f", remaining))

		next(w, r)
	}
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := APIResponse{
		Success: true,
		Message: "Welcome to the Distributed API!",
	}
	json.NewEncoder(w).Encode(resp)
}

func main() {
	// 1. Connect to Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// 2. Initialize our Limiter Wrapper
	globalLimiter = NewRateLimiter(rdb)

	http.HandleFunc("/", rateLimitMiddleware(helloHandler))

	fmt.Println("Server running on :8080 (Redis Backend)")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
