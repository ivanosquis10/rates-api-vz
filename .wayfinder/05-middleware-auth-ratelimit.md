# Ticket 05: Auth Middleware & Rate Limiter

**Type**: Grilling (HITL)
**Blocked by**: None (frontier)
**Status**: RESOLVED

## Question

How should the API Key authentication and rate limiter be implemented?

## Resolution

### Auth Middleware

```go
// middleware/auth.go
package middleware

func Auth(apiKey string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            key := r.Header.Get("X-API-Key")
            
            // Constant-time comparison to prevent timing attacks
            if subtle.ConstantTimeCompare([]byte(key), []byte(apiKey)) != 1 {
                respondWithError(w, http.StatusUnauthorized, "Invalid or missing API key")
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

**Key decisions:**
- API Key stored in env var `API_KEY`
- Constant-time comparison (`subtle.ConstantTimeCompare`) to prevent timing attacks
- 401 response: `{ "error": { "code": "UNAUTHORIZED", "message": "Invalid or missing API key" } }`

### Rate Limiter Middleware

```go
// middleware/ratelimit.go
package middleware

import "golang.org/x/time/rate"

func RateLimit(rps int) func(http.Handler) http.Handler {
    // Per-IP map of limiters
    var (
        mu       sync.Mutex
        limiters = make(map[string]*rate.Limiter)
    )
    
    getLimiter := func(ip string) *rate.Limiter {
        mu.Lock()
        defer mu.Unlock()
        
        l, exists := limiters[ip]
        if !exists {
            l = rate.NewLimiter(rate.Limit(rps), rps*2) // burst = 2x rps
            limiters[ip] = l
        }
        return l
    }
    
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ip := extractIP(r)
            limiter := getLimiter(ip)
            
            if !limiter.Allow() {
                w.Header().Set("Retry-After", "60")
                respondWithError(w, http.StatusTooManyRequests, "Rate limit exceeded")
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

**Key decisions:**
- **Library**: `golang.org/x/time/rate` (stdlib, no external deps)
- **Algorithm**: Token bucket per IP
- **Scope**: Per-IP (each IP gets its own limiter)
- **Configurable**: `RATE_LIMIT` env var (default: 60 requests/min)
- **Burst**: 2x the rate limit (allows short spikes)
- **Response**: 429 Too Many Requests + `Retry-After` header
- **Cleanup**: Periodic goroutine to remove inactive IPs (prevent memory leak)

### Middleware chain order

```go
// handler/routes.go
func RegisterRoutes(r *chi.Mux, uc RateUseCase, apiKey string, rateLimit int) {
    r.Use(middleware.RateLimit(rateLimit))  // 1st: rate limit
    r.Use(middleware.Auth(apiKey))          // 2nd: auth
    
    r.Get("/rates", handler.GetCurrentRates(uc))
    r.Get("/rates/history", handler.GetHistory(uc))
    r.Post("/admin/scrape", handler.TriggerScrape(uc))
}
```

Order matters: **rate limit → auth → handler**. This prevents unauthenticated requests from consuming rate limit tokens.

### Summary

| Component | Implementation | Config |
|-----------|---------------|--------|
| Auth | `X-API-Key` header, constant-time compare | `API_KEY` env var |
| Rate Limiter | Token bucket per IP | `RATE_LIMIT` env var (default 60/min) |
| 401 Response | `{ "error": { "code": "UNAUTHORIZED", ... } }` | — |
| 429 Response | `{ "error": { "code": "RATE_LIMITED", ... } }` + `Retry-After` | — |
