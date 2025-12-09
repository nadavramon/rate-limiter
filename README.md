# Go Rate Limiter 🛡️

A thread-safe, distributed-ready API Rate Limiter built from scratch in Go.
It implements the **Token Bucket Algorithm** to handle traffic bursts while maintaining strict average rate limits.

## 🚀 Features

* **Token Bucket Algorithm:** Implements O(1) lazy-refill logic to manage request tokens efficiently.
* **Thread Safety:** Uses `sync.Mutex` to prevent race conditions during concurrent request processing.
* **Smart IP Handling:** Correctly identifies users by IP address (automatically stripping dynamic ports).
* **Observability:** Returns standard `X-RateLimit-*` HTTP headers for client transparency.
* **Memory Management:** Includes a background "Janitor" Goroutine that automatically cleans up stale clients to prevent memory leaks.

## 🛠️ Tech Stack

* **Language:** Go (Golang) 1.25
* **Standard Lib:** `net/http`, `sync`, `time`, `net` (Zero external dependencies).

## ⚡ How to Run

1. **Clone the repository**

   ```bash
   git clone [https://github.com/nadavramon/rate-limiter.git](https://github.com/nadavramon/rate-limiter.git)
   cd rate-limiter
   ```

2. **Start the server**

   ```bash
   go run .
   ```

3. **Test with curl**

   To see the headers and the rate limiting in action, run this command in a separate terminal:

   ```bash
   curl -v http://localhost:8080
   ```

To simulate a burst of traffic (and trigger a 429 error), chain multiple requests:

```bash
curl -v http://localhost:8080 && curl -v http://localhost:8080 && curl -v http://localhost:8080
```

## 🧠 Design Decisions

### Why Token Bucket?

I chose the Token Bucket algorithm over a standard Fixed Window counter because it allows for short bursts of traffic (e.g., a user loading a page with 5 assets simultaneously) while still enforcing a strict long-term average rate. This prevents the "use-it-or-lose-it" rigidity of Fixed Window algorithms and provides a smoother experience for real users who rarely send requests at a perfectly constant pace.

### Concurrency Strategy

Since Go's http handlers run in parallel Goroutines, accessing a shared map of clients is not thread-safe by default. I implemented a sync.Mutex to lock the client map during read/write operations. This ensures that two requests arriving from the same IP at the exact same microsecond do not cause race conditions or incorrect token calculations.

### Resource Management

To prevent Unbounded Memory Growth (a potential DDoS vector), I implemented a "Janitor" process. It runs on a separate Goroutine and periodically scans the client map to delete records of users who haven't made a request in over 3 minutes.

## 🔮 Future Improvements

* **Persistent storage backend:** Swap the in-memory map with Redis or another distributed cache to share limits across multiple instances.
* **Burst-specific policies:** Allow per-endpoint or per-API-key rate settings so critical endpoints get tighter control.
* **Prometheus metrics:** Export limiter stats (token counts, rejections, janitor sweeps) for dashboards and alerting.
