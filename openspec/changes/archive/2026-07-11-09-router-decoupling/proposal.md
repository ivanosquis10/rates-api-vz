# Proposal: Router Decoupling

## Intent
Decouple HTTP routing and middleware setup from `cmd/api/main.go` into a dedicated package `internal/http/router`. This improves testability, maintainability, and clean separation of concerns. Additionally, it addresses security by swapping deprecated `middleware.RealIP` with `middleware.ClientIPFromRemoteAddr` to prevent IP spoofing.

## Scope
### In Scope
- Create `internal/http/router/router.go` with structures `Config`, `Deps` (Handler, Config, Context), and constructor `New(deps Deps) http.Handler`.
- Replace deprecated `middleware.RealIP` with `middleware.ClientIPFromRemoteAddr` in Chi configuration.
- Refactor `cmd/api/main.go` to inject dependencies and delegate routing to `internal/http/router`.
- Adapt and fix unit and integration tests to ensure compilation and successful execution.

### Out of Scope
- Modifying application core logic (use cases, repositories, scrapers).
- Adding new endpoints or changing existing API routes.

## Capabilities
### New Capabilities
- None (refactoring change).
### Modified Capabilities
- `http-server-initialization`: Initial routing definition and middleware chain assembly are delegated from main to `internal/http/router`.
- `secure-ip-resolution`: Middleware uses `middleware.ClientIPFromRemoteAddr` to mitigate client IP spoofing risks.

## Approach
1. Define package `router` in `internal/http/router/router.go`.
2. Define `Config` containing config options relevant to routing (e.g. rate limit, API key).
3. Define `Deps` containing `Handler *handler.Handler`, `Config *config.Config`, and `Context context.Context`.
4. Implement `New(deps Deps) http.Handler` to instantiate chi router, configure middlewares (recovery, logging, rate limiting, auth) in correct order, set up routes, and return the handler.
5. In `cmd/api/main.go`, construct dependencies, call `router.New(router.Deps{...})`, and pass the returned handler to `http.Server`.
6. Fix any compiling issues and verify all tests run correctly.

## Affected Areas
| Area | Impact | Description |
|------|--------|-------------|
| `cmd/api/main.go` | Modified | Delegate router configuration, inject dependencies. |
| `internal/http/router/router.go` | New/Modified | Package implementation containing `Config`, `Deps`, and `New()`. |
| `internal/middleware/integration_test.go` | Modified | Update tests to reflect the routing changes. |
| `cmd/api/main_test.go` | Modified | Update tests if affected by main signature changes. |

## Risks
| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Client IP spoofing if RemoteAddr configuration is incorrect | Low | Use standard `middleware.ClientIPFromRemoteAddr` and test local/integration. |
| Breaking existing integration tests due to middleware adjustments | Medium | Verify integration tests carefully and adjust mocks/dependencies. |

## Rollback Plan
Revert to the latest commit on `feat/08-http-refactor` or restore `cmd/api/main.go` from git backup.

## Dependencies
- None.

## Success Criteria
- [ ] `go test ./...` passes.
- [ ] Application compiles successfully with `go build ./cmd/api`.
- [ ] `middleware.RealIP` is removed and replaced by `middleware.ClientIPFromRemoteAddr`.
