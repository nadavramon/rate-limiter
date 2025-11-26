package main

import (
	"fmt"
	"net/http"
	"sync"
)

// 1. Create a global map to store rate limiters for each IP
// We need a Mutex for the map itself, because multiple requests
// might try to create a NEW bucket at the same time.
var (
	clients = make(map[string]*RateLimiter)
	mu      sync.Mutex
)

// 2. This helper function gets the limiter for a specific IP
// If the user is new, it creates a bucket for them.
func getLimiter(ip string) *RateLimiter {
	mu.Lock()
	defer mu.Unlock()

	limiter, exists := clients[ip]
	if !exists {
		// Create a new bucket:
		// Capacity = 5 tokens
		// Refill Rate = 1 token per second
		limiter = NewRateLimiter(5, 1)
		clients[ip] = limiter
	}

	return limiter
}

// 3. Middleware: This function intercepts the request BEFORE the helloHandler
func rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the user's IP address
		// (In a real app, you'd parse X-Forwarded-For headers)
		ip := r.RemoteAddr

		// Get their bucket
		limiter := getLimiter(ip)

		// Check if allowed
		if !limiter.Allow() {
			http.Error(w, "429 Too Many Requests", http.StatusTooManyRequests)
			return // Stop here! Don't run helloHandler
		}

		// If allowed, run the next handler (helloHandler)
		next(w, r)
	}
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Request Allowed! Welcome to the API.")
}

func main() {
	// Wrap the helloHandler with our Rate Limiter Middleware
	http.HandleFunc("/", rateLimitMiddleware(helloHandler))

	fmt.Println("Server running on :8080")
	http.ListenAndServe(":8080", nil)
}
