# Tasks: Rate Limiter Lifecycle Refactor

## Review Workload Forecast

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: stacked-to-main
400-line budget risk: Low

| Field | Value |
|-------|-------|
| Estimated changed lines | ~60 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | single PR |
| Delivery strategy | auto-chain |
| Chain strategy | stacked-to-main |

### Suggested Work Units
| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Refactor Rate Limiter and Router deps | PR 1 | Promotes rateLimiter, updates deps, wires tests |

## Phase 1: Foundation / Infrastructure
- [x] 1.1 In [ratelimit.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/ratelimit.go), promote unexported `rateLimiter` to public `RateLimiter` and add constructor `NewRateLimiter(ctx context.Context, limitPerMin int) *RateLimiter`.

## Phase 2: Core Implementation
- [x] 2.1 In [router.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/http/router/router.go), modify `Deps` struct to accept `RateLimiter *middleware.RateLimiter` instead of `Context context.Context`.
- [x] 2.2 In [router.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/http/router/router.go), replace call to `middleware.RateLimit(deps.Context, deps.Config.RateLimit)` with `deps.RateLimiter.Handler`.
- [x] 2.3 In [main.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/cmd/api/main.go), instantiate `RateLimiter` using `middleware.NewRateLimiter(ctx, cfg.RateLimit)` and pass it into `router.Deps`.

## Phase 3: Integration / Wiring
- [x] 3.1 In [router_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/http/router/router_test.go), import `github.com/ivanosquis10/api-rates-venezuela/internal/middleware` and instantiate `RateLimiter` in dependencies configuration for `TestRouter_New` and `TestRouter_Middleware_Auth`.

## Phase 4: Testing / Verification
- [x] 4.1 In [ratelimit_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/ratelimit_test.go), refactor to use `NewRateLimiter` and verify that `defer cancel()` is called immediately after `context.WithCancel(...)` or other cancellable contexts.
- [x] 4.2 In [integration_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/integration_test.go), update instantiation of rate limiter using `NewRateLimiter` and verify immediate `defer cancel()` usage.
- [x] 4.3 Run `go test ./...` in the terminal to ensure all tests compile and pass.

## Phase 5: Cleanup
- [x] 5.1 Remove any unused imports and double check for context leaks or goroutine leaks across all modified code.
