# Design: Router Decoupling

## Technical Approach
Decouple the HTTP routing and middleware chain assembly from `cmd/api/main.go` into a dedicated package `internal/http/router`. The router package will expose a `Deps` struct and a constructor `New(Deps) http.Handler` that sets up the go-chi router. This decouples the server's HTTP routing from the server initialization, making it easier to test routing and middleware in isolation.

To enhance security against IP spoofing, we will replace the deprecated `middleware.RealIP` with `middleware.ClientIPFromRemoteAddr` in the router stack.

## Architecture Decisions
### Decision: Decouple router into a dedicated package
**Choice**: Create `internal/http/router` package with structure `Deps` and function `New(Deps) http.Handler`.
**Alternatives considered**: Keep routing in `cmd/api/main.go` or split routes into separate handler files.
**Rationale**: Keeping routing in `main.go` bloats it and prevents testing the routing logic in isolation. A separate router package clearly separates routing concerns and allows manual dependency injection of configuration and handlers.

### Decision: Swapping RealIP with ClientIPFromRemoteAddr
**Choice**: Use `github.com/go-chi/chi/v5/middleware.ClientIPFromRemoteAddr` first in the middleware chain.
**Alternatives considered**: Keep `middleware.RealIP`, or use a custom IP resolution header parser.
**Rationale**: `middleware.RealIP` is deprecated due to security vulnerabilities (specifically IP spoofing where clients can supply arbitrary headers like `X-Forwarded-For` to bypass rate limits or controls). `ClientIPFromRemoteAddr` retrieves the IP address directly from the TCP/IP connection's RemoteAddr, which is secure and cannot be spoofed by HTTP headers in standard deployments not behind reverse proxies.

## Data Flow
```
Client Request 
     │
     ▼ (TCP connection)
[http.Server]
     │
     ▼
[router.New()] 
     │
     ├─► [chimw.ClientIPFromRemoteAddr] (Resolves client IP securely from RemoteAddr)
     ├─► [chimw.RequestID] (Generates/propagates unique correlation ID)
     ├─► [middleware.Recovery] (Catches panics, logs stack trace, returns 500)
     ├─► [middleware.Logging] (Logs HTTP request details & timing)
     ├─► [middleware.RateLimit] (Checks rate limits against client IP)
     ├─► [middleware.Auth] (Validates X-API-Key header using ConstantTimeCompare)
     │
     ▼
[Handler] ──► [Usecase] ──► [Store/Scraper]
```

## File Changes
| File | Action | Description |
|------|--------|-------------|
| `internal/http/router/router.go` | Modify/Recreate | Implement `Deps` struct and `New(deps Deps) http.Handler` constructor with Chi router, secure IP middleware, and custom middleware stack. |
| `internal/http/router/router_test.go` | Create | Unit test for the router initialization to verify that valid dependencies return a non-nil handler and dispatch routes. |
| `cmd/api/main.go` | Modify | Remove direct chi dependencies and route definition. Instantiate router via `router.New` with dependencies and pass to `http.Server`. |

## Interfaces / Contracts
```go
package router

import (
	"context"
	"net/http"

	"github.com/ivanosquis10/api-rates-venezuela/internal/config"
	"github.com/ivanosquis10/api-rates-venezuela/internal/handler"
)

// Deps holds router dependencies
type Deps struct {
	Handler *handler.Handler
	Config  *config.Config
	Context context.Context
}

// New constructs the decoupled chi router
func New(deps Deps) http.Handler
```

## Testing Strategy
| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit (Router) | Router Instantiation | Verify `router.New` returns a non-nil `http.Handler` when given valid `Deps`. |
| Integration | Routing & Middleware Order | Use `httptest.NewRecorder` to verify that requests to endpoints (`/rates`, `/rates/history`, `/admin/scrape`) are successfully dispatched or intercepted by middleware (e.g. rate limit, auth) based on request attributes. |

## Migration / Rollout
No migration required. This is a refactoring change that preserves existing endpoint contracts.

## Open Questions
- None
