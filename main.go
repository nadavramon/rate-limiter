package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
)

// APIResponse is the shape of data we send back to the user
type APIResponse struct {
	Success   bool    `json:"success"`
	Message   string  `json:"message"`
	Remaining float64 `json:"remaining_tokens,omitempty"`
}

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
		// 1. Extract the IP, stripping the port
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			// If there is no port (rare), use the whole string
			ip = r.RemoteAddr
		}

		// Print it to prove it works (check your server terminal!)
		fmt.Println("User IP:", ip)

		limiter := getLimiter(ip)

		// 1. Check if allowed
		if !limiter.Allow() {
			http.Error(w, "429 Too Many Requests", http.StatusTooManyRequests)
			return
		}

		// 2. INJECT HEADERS (The new part)
		// We need to peek inside the bucket (Thread-safe read)
		limiter.mu.Lock()
		remaining := limiter.tokens
		limiter.mu.Unlock()

		// Set the headers
		w.Header().Set("X-RateLimit-Limit", "5")
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%.2f", remaining))

		// 3. Run the next handler
		next(w, r)
	}
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Tell the browser/client: "I am sending you JSON, not text"
	w.Header().Set("Content-Type", "application/json")

	// 2. Create the data object
	resp := APIResponse{
		Success: true,
		Message: "Welcome to the secret API citadel!",
		// (Optional: In a real app, you'd fetch the actual remaining tokens from the limiter)
	}

	// 3. Encode it into JSON and send it
	json.NewEncoder(w).Encode(resp)
}

func main() {
	// Wrap the helloHandler with our Rate Limiter Middleware
	http.HandleFunc("/", rateLimitMiddleware(helloHandler))

	fmt.Println("Server running on :8080")
	http.ListenAndServe(":8080", nil)
}
