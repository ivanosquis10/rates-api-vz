# Design: Rate Limiter Lifecycle Refactor

## Technical Approach
Refactor the rate limiting mechanism to decouple it from the router package's context and configuration. We will promote the unexported `rateLimiter` struct in [ratelimit.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/ratelimit.go) to a public `RateLimiter` struct and expose a constructor `NewRateLimiter(ctx context.Context, limitPerMin int) *RateLimiter`. The constructor will spin up the background janitor goroutine, bound to the provided application context.
The router's dependency structure in [router.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/http/router/router.go) will accept `RateLimiter *middleware.RateLimiter` instead of `Context context.Context`. This keeps the router thin, free of background goroutine orchestration, and easily testable.

## Architecture Decisions
### Decision: Inject RateLimiter Dependency directly into Router
**Choice**: Inject a pre-initialized `*middleware.RateLimiter` instance into `router.Deps`.
**Alternatives considered**: Keep passing `context.Context` and `Config` to the router to build the rate limiter inside `router.New()`.
**Rationale**: The router should not be responsible for managing background task lifecycles (like the rate limiter's janitor goroutine). Moving construction to `main.go` ensures proper dependency inversion and cleaner separation of concerns.

### Decision: Clean Background Janitor Cancellation
**Choice**: Ticker with select block on `ctx.Done()`.
**Alternatives considered**: Passing a channel or running an un-cancellable goroutine.
**Rationale**: By binding the janitor to the application context passed to `NewRateLimiter`, calling `cancel()` in `main.go` propagates down to `<-ctx.Done()`, which breaks the loop and stops the ticker, preventing goroutine leaks.

## Data Flow
1. An incoming HTTP request hits the Chi Router.
2. The router passes the request to the `RateLimiter.Handler` middleware.
3. The middleware resolves the client's IP and fetches/creates their rate limiter.
4. If rate-limited, the middleware returns `429 Too Many Requests`. Otherwise, it propagates the request to subsequent handlers.
5. Concurrently, a background janitor tick periodically (every 10 seconds) purges clients inactive for >= 5 minutes.

## File Changes
| File | Action | Description |
|------|--------|-------------|
| [ratelimit.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/ratelimit.go) | Modify | Rename `rateLimiter` to `RateLimiter`, add constructor `NewRateLimiter`, expose `Handler` method. |
| [router.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/http/router/router.go) | Modify | Update `router.Deps` to accept `RateLimiter *middleware.RateLimiter` instead of `Context`, update middleware initialization. |
| [main.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/cmd/api/main.go) | Modify | Instantiate `RateLimiter` using `middleware.NewRateLimiter(ctx, cfg.RateLimit)` and pass to `router.Deps`. |
| [router_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/http/router/router_test.go) | Modify | Update test setups to instantiate and pass `RateLimiter` in `Deps`. |
| [ratelimit_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/ratelimit_test.go) | Modify | Update tests to use `NewRateLimiter` and `RateLimiter.Handler`. |
| [integration_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/integration_test.go) | Modify | Update tests to use `NewRateLimiter` and `RateLimiter.Handler`. |

## Interfaces / Contracts
```go
package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	mu          sync.RWMutex
	clients     map[string]*client
	limitPerMin int
	pruneAge    time.Duration
}

func NewRateLimiter(ctx context.Context, limitPerMin int) *RateLimiter
func (rl *RateLimiter) Handler(next http.Handler) http.Handler
```

## Testing Strategy
| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit (Middleware) | Janitor Lifecycle | Assert janitor stops running after context cancellation. |
| Unit (Middleware) | Client Pruning | Assert inactive client keys are pruned after `pruneAge` duration. |
| Unit (Middleware) | Rate Limiting | Verify correct HTTP status codes (200, 429) and headers (`Retry-After`). |
| Unit (Router) | Dependency Injection | Confirm `router.New` works correctly with valid `Deps`. |
| Integration | Middlewares Order | Assert rate limiter runs before auth check. |

## Migration / Rollout
No database migration is required. This is a pure codebase refactoring. The change will deploy seamlessly as part of standard application lifecycle.

## Open Questions
- [x] None.
