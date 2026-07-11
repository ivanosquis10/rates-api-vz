# Proposal: Rate Limiter Lifecycle Refactor

## Intent
Decouple rate limiter initialization and its background cleanup janitor from the router's context propagation. This simplifies router dependencies and ensures the router focuses solely on HTTP request routing, rather than managing active background goroutines.

## Scope
### In Scope
- Refactor [ratelimit.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/ratelimit.go) to expose `RateLimiter` and `NewRateLimiter(ctx context.Context, limitPerMin int) *RateLimiter`.
- Expose the public method `Handler(next http.Handler) http.Handler` on `RateLimiter`.
- Update [router.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/http/router/router.go) to remove `Context` and inject `RateLimiter *middleware.RateLimiter` in `Deps`.
- Update [main.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/cmd/api/main.go) to construct `RateLimiter` with application `ctx` and pass it to the router.
- Update [router_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/http/router/router_test.go) to pass a `RateLimiter` instance in `router.Deps`.

### Out of Scope
- Changing the rate limiting algorithm (Token Bucket) or standard configuration variables.
- Modifying authentication, routing endpoints, or logging middleware behaviors.

## Capabilities
### New Capabilities
- None.

### Modified Capabilities
- `RateLimiter Initialization`: The rate limiter lifecycle is managed at the application root rather than within the router package.

## Approach
1. Modify [ratelimit.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/ratelimit.go):
   - Rename `rateLimiter` to `RateLimiter`.
   - Implement `NewRateLimiter(ctx, limitPerMin)` that starts the janitor goroutine.
   - Expose `Handler(next http.Handler) http.Handler`.
2. Modify [router.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/http/router/router.go):
   - Remove `Context` from `Deps`.
   - Add `RateLimiter *middleware.RateLimiter` to `Deps`.
   - Replace `r.Use(middleware.RateLimit(...))` with `r.Use(deps.RateLimiter.Handler)`.
3. Modify [main.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/cmd/api/main.go) to instantiate `RateLimiter` using `middleware.NewRateLimiter(ctx, cfg.RateLimit)` and pass it in `router.Deps`.
4. Modify [router_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/http/router/router_test.go) to instantiate `RateLimiter` inside tests.

## Affected Areas
| Area | Impact | Description |
|------|--------|-------------|
| [ratelimit.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/ratelimit.go) | Modified | Exports `RateLimiter` and its constructor; exposes `Handler` |
| [router.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/http/router/router.go) | Modified | Removes `Context` dependency, accepts `RateLimiter` |
| [router_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/http/router/router_test.go) | Modified | Passes `RateLimiter` in `router.Deps` during test setup |
| [main.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/cmd/api/main.go) | Modified | Instantiates `RateLimiter` and wires it to `router.New` |

## Risks
| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Memory leaks if the passed context is not cancelled | Low | Ensure context cancellation is deferred in `main.go` |
| Test suite compilation failures due to mismatched dependency types | Low | Update all test setups with valid `RateLimiter` instances |

## Rollback Plan
Revert changes using git:
```bash
git checkout main -- internal/middleware/ratelimit.go internal/http/router/router.go internal/http/router/router_test.go cmd/api/main.go
```

## Dependencies
- None.

## Success Criteria
- [ ] Refactored code compiles and all existing tests pass (`go test ./...`).
- [ ] Rate limiter background goroutine gracefully terminates when application context is cancelled.
- [ ] API successfully rate limits incoming requests and cleans up inactive clients after 5 minutes of inactivity.
