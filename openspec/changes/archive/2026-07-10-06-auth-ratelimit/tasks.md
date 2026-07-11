# Tasks: Auth and Rate Limit Middleware

## Review Workload Forecast

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: stacked-to-main
400-line budget risk: Medium

| Field | Value |
|-------|-------|
| Estimated changed lines | ~350 |
| 400-line budget risk | Medium |
| Chained PRs recommended | No |
| Suggested split | single PR |
| Delivery strategy | auto-chain |
| Chain strategy | stacked-to-main |

### Suggested Work Units
| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Middleware implementation & integration | single PR | Low risk |

## Phase 1: Foundation / Infrastructure
- [x] 1.1 Create skeleton files [auth.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/auth.go) and [ratelimit.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/ratelimit.go) under the `middleware` package.

## Phase 2: Core Implementation
- [x] 2.1 Implement timing-attack resistant `Auth` in [auth.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/auth.go) using SHA-256 and `subtle.ConstantTimeCompare`. Return 401 JSON error on failure.
- [x] 2.2 Implement per-IP rate limiting in [ratelimit.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/ratelimit.go).
- [x] 2.3 Validate `limitPerMin > 0` and fallback to 60.
- [x] 2.4 Use `net.SplitHostPort` on `r.RemoteAddr` (fallback to raw remote address if error) to strip port numbers.
- [x] 2.5 Use `limiter.Reserve()` and check delay. Calculate exact `Retry-After` using `delay.Seconds()`. If rejected, call `Cancel()`, set `Retry-After` header, return 429 JSON.
- [x] 2.6 Implement context-aware janitor. Scan inactive clients under `RLock()`, then delete from map under `Lock()` after re-verifying expiration inside write lock. Defer `ticker.Stop()` in the goroutine. Update `lastSeen` with atomic operations.

## Phase 3: Integration / Wiring
- [x] 3.1 Wire middlewares in [main.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/cmd/api/main.go): insert `RateLimit` and `Auth` after recovery/logging. Pass the application context and config values.

## Phase 4: Testing / Verification
- [x] 4.1 Write unit tests in `auth_test.go` to test valid, missing, and invalid keys.
- [x] 4.2 Write unit tests in `ratelimit_test.go` to verify limit capacity, `Retry-After` calculation, map pruning/race conditions, and context-controlled janitor shutdown.
- [x] 4.3 Verify execution order of middlewares using httptest.

## Phase 5: Cleanup
- [x] 5.1 Format Go files (`go fmt`), run the test suite with `-race` flag, and verify no lint issues.
