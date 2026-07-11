# Tasks: Router Decoupling

## Review Workload Forecast

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: stacked-to-main
400-line budget risk: Low

| Field | Value |
|-------|-------|
| Estimated changed lines | 160-180 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | auto-chain |
| Chain strategy | stacked-to-main |

### Suggested Work Units
| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Refactor Rate Limiter & Setup Router | PR 1 | Implement IP resolution fallback and correct router.go |
| 2 | Refactor Main & Implement Router Tests | PR 1 | Decouple cmd/api/main.go and write router_test.go |

## Phase 1: Foundation / Infrastructure
- [x] 1.1 Update `internal/middleware/ratelimit.go` imports to alias Chi middleware as `chimw`.
- [x] 1.2 Refactor IP resolution in `internal/middleware/ratelimit.go` to use `chimw.GetClientIP(r.Context())` first, falling back to `net.SplitHostPort(r.RemoteAddr)`.

## Phase 2: Core Implementation
- [x] 2.1 Update `internal/http/router/router.go` to resolve namespace conflict: import Chi middleware as `chimw` and local as `middleware`.
- [x] 2.2 Define `Deps` struct in `internal/http/router/router.go` containing `Handler`, `Config`, and `Context`.
- [x] 2.3 Refactor `New(Deps) http.Handler` in `internal/http/router/router.go` to construct Chi router using secure IP resolver, logging, custom recovery, rate limiting, authentication, and correct endpoints.
- [x] 2.4 Remove duplicate RequestID, Logger, and Recoverer registrations in `internal/http/router/router.go`.

## Phase 3: Integration / Wiring
- [x] 3.1 Modify `cmd/api/main.go` to import the new `internal/http/router` package.
- [x] 3.2 Update `cmd/api/main.go` to instantiate the router via `router.New` and pass `router.Deps`.
- [x] 3.3 Remove direct Chi dependencies and route setups in `cmd/api/main.go`.

## Phase 4: Testing / Verification
- [x] 4.1 Create `internal/http/router/router_test.go` to test successful router initialization with mock dependencies.
- [x] 4.2 Add integration tests in `internal/http/router/router_test.go` using `httptest.NewRecorder` to verify routing to rates, history, and admin scrape.
- [x] 4.3 Add test cases verifying authentication middleware blocks requests with invalid API keys with 401.

## Phase 5: Cleanup
- [x] 5.1 Run `go test ./...` to verify all tests pass.
- [x] 5.2 Remove any remaining unused imports and format the codebase.
