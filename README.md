# Go Rate Limiter (Distributed) 🛡️

A thread-safe, distributed-ready API Rate Limiter built from scratch in Go.
It uses **Redis + Lua Scripts** to enforce rate limits atomically across multiple server instances.

## 🚀 Features

* **Distributed State:** Uses Redis to share token buckets across multiple server replicas.
* **Atomic Operations:** Custom Lua script ensures `Get-Calculation-Set` operations happen instantly to prevent race conditions.
* **Token Bucket Algorithm:** Allows for traffic bursts while maintaining strict average limits.
* **Fail-Safe:** Defaults to "Fail Closed" if Redis is unreachable (configurable).
* **Observability:** Returns standard `X-RateLimit-*` HTTP headers.

## 🛠️ Tech Stack

* **Language:** Go (Golang)
* **Database:** Redis 7+
* **Libraries:** `go-redis/v9` (Redis Client)

## ⚡ How to Run

1. **Prerequisites**
   You must have Redis installed and running.

   ```bash
   # MacOS
   brew install redis
   brew services start redis
   ```

2. **Clone and Run**

   ```bash
   git clone [https://github.com/nadavramon/rate-limiter.git](https://github.com/nadavramon/rate-limiter.git)
   cd rate-limiter
   go mod tidy  # Download dependencies
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

### Why Redis + Lua?

In a distributed system (e.g., 3 API servers behind a Load Balancer), we cannot store state in local memory because Server A doesn't know about requests sent to Server B. I used Redis as the shared source of truth. To prevent race conditions (where two servers read "5 tokens" simultaneously), I implemented the logic in a Lua script. This guarantees that the read-update-write cycle is atomic—no other request can interleave.

### Why Token Bucket?

I chose the Token Bucket algorithm to allow for short bursts of traffic (e.g., loading a dashboard) while enforcing a long-term rate limit. This provides a better UX than Fixed Window counters.

🔮 Future Improvements
Burst-specific policies: Allow per-endpoint or per-API-key rate settings so critical endpoints get tighter control.

Prometheus metrics: Export limiter stats (token counts, rejections, janitor sweeps) for dashboards and alerting
