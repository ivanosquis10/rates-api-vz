# Proposal: 05 — HTTP Server & API Endpoints

## Intent

Expose the Venezuela Rates API over HTTP using Chi router. The usecase layer (issue #5) is complete but the application has no serving surface — just a bare `http.ListenAndServe` placeholder. This change wires up the HTTP transport, defines RESTful endpoints, standardizes JSON responses, and adds handler-level tests with mocked dependencies.

## Scope

### In Scope

- Chi router replacing `http.NewServeMux` in `cmd/api/main.go`
- `GET /rates` — latest rates, optional `currency` and `type` query params
- `GET /rates/history` — historical rates, query params: `currency`, `type`, `from`, `to`, `limit`
- `POST /admin/scrape` — trigger scrape, return immediate `202 Accepted`
- Standard JSON envelope: `{ "data": ... }` for success, `{ "error": { "code": "...", "message": "..." } }` for errors
- HTTP status mapping: 200, 400, 404, 500
- Internal errors (SQL, stack traces) never leak in responses
- Handler tests using `httptest.NewRecorder` with mocked `RateUsecase`
- Test coverage: current rates, filtered current, history, filtered history, trigger scrape, bad request, not found

### Out of Scope

- Authentication/authorization on `/admin/scrape` (deferred to future change)
- Rate limiting or CORS configuration
- OpenAPI/Swagger documentation
- Graceful shutdown (current `log.Fatal` pattern stays)
- Structured logging middleware (slog already used in usecase layer)

## Capabilities

### New Capabilities

- `http-server`: Chi router setup, middleware stack (logging, recovery), route registration
- `api-endpoints`: GET /rates, GET /rates/history, POST /admin/scrape handlers
- `error-handling`: JSON error envelope, domain-error-to-HTTP-status mapping, internal error sanitization

### Modified Capabilities

- None — usecase layer interface is unchanged

## Approach

1. Add `github.com/go-chi/chi/v5` dependency
2. Create `internal/handler/` package with:
   - `handler.go` — `Handler` struct holding `*usecase.RateUsecase`, `NewHandler` constructor
   - `rate_handlers.go` — `GetRates`, `GetHistory`, `TriggerScrape` methods
   - `responses.go` — `respondJSON`, `respondError`, domain-error mapper
3. Create `internal/middleware/` package:
   - `logging.go` — request logging middleware
   - `recovery.go` — panic recovery middleware
4. Rewrite `cmd/api/main.go`:
   - Wire `config` → `store` → `usecase` → `handler` → `chi.Router`
   - Register routes: `r.Get("/rates", ...)`, `r.Get("/rates/history", ...)`, `r.Route("/admin", admin.Group(...))`
5. Write handler tests in `internal/handler/` using `httptest` + mocked usecase interface

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `cmd/api/main.go` | Modified | Replace mux with Chi router, wire all dependencies |
| `internal/handler/` | New | Handler package with endpoint logic, response helpers |
| `internal/middleware/` | New | Logging and recovery middleware |
| `go.mod` / `go.sum` | Modified | Add `github.com/go-chi/chi/v5` dependency |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Chi adds external dependency | Low | Well-established Go HTTP router, minimal footprint |
| Mock complexity for handler tests | Low | Define a thin usecase interface; mock in test files only |
| Admin endpoint exposed without auth | Medium | Explicit out-of-scope; document in README that auth is deferred |

## Rollback Plan

Remove `internal/handler/`, `internal/middleware/`, revert `cmd/api/main.go` to placeholder mux, remove Chi from `go.mod`. The usecase and store layers remain untouched.

## Dependencies

- Issue #5 (Rate Usecase) — completed
- `github.com/go-chi/chi/v5` — new dependency to add

## Success Criteria

- [ ] `go build ./cmd/api` compiles without errors
- [ ] `go test ./internal/handler/...` passes all tests
- [ ] All 7 handler test scenarios pass
- [ ] Internal errors never appear in JSON responses
- [ ] GET /rates returns `{ "data": [...] }` with correct filtering
- [ ] GET /rates/history respects all query params
- [ ] POST /admin/scrape returns 202 with confirmation JSON
