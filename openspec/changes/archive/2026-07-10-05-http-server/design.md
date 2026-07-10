# Design: 05 — HTTP Server & API Endpoints

## Technical Approach

Replace the placeholder `http.NewServeMux` in `cmd/api/main.go` with Chi router, add handler package that delegates to existing `RateUsecase`, implement standardized JSON response envelope and error mapping, and add handler tests with mocked usecase. The design follows existing project conventions: internal packages, constructor injection, and table‑driven tests.

## Architecture Decisions

### Decision: Use Chi router over standard library mux

| Option | Tradeoff | Decision |
|--------|----------|----------|
| `net/http.ServeMux` (stdlib) | No external dependency; limited route grouping, no built‑in middleware chaining | Rejected |
| `github.com/go-chi/chi/v5` | Adds dependency; expressive route groups, middleware chaining, lightweight | **Selected** |

**Rationale**: Chi provides route grouping (`/admin`), middleware chaining, and clean handler signatures while staying idiomatic Go. The dependency is well‑maintained and aligns with the proposal.

### Decision: Handler depends on `*usecase.RateUsecase` directly

| Option | Tradeoff | Decision |
|--------|----------|----------|
| Define a `RateUsecase` interface | Decouples handler from concrete type; extra interface to maintain | Rejected for now |
| Depend on concrete `*usecase.RateUsecase` | Simpler; test mock replaces the whole struct via `httptest` with a custom handler function | **Selected** |

**Rationale**: The usecase layer is stable and has a single implementation. For testing we can create a mock handler function that returns pre‑canned data; we don’t need a full interface mock. This keeps the codebase minimal.

### Decision: Centralized error mapper in `responses.go`

| Option | Tradeoff | Decision |
|--------|----------|----------|
| Map errors inline in each handler | Duplicated switch statements; easy to miss a case | Rejected |
| `mapError(err) (int, string)` helper | Single source of truth; all handlers use same mapping | **Selected** |

**Rationale**: The error‑to‑HTTP‑status mapping is defined in the spec and should not be scattered. A helper ensures consistency and simplifies future changes.

## Data Flow

```
HTTP Request → Chi Router → Middleware (recovery → logging)
                ↓
         Handler method (e.g., GetRates)
                ↓
         Parse query params
                ↓
         Call RateUsecase method
                ↓
         Receive []domain.Rate or error
                ↓
         mapError / respondJSON
                ↓
         JSON response { "data": ... } or { "error": { "code", "message" } }
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/handler/handler.go` | Create | `Handler` struct holding `*usecase.RateUsecase`; `NewHandler` constructor |
| `internal/handler/rate_handlers.go` | Create | `GetRates`, `GetHistory`, `TriggerScrape` HTTP handler methods |
| `internal/handler/responses.go` | Create | `respondJSON`, `respondError`, `mapError` helpers |
| `internal/middleware/logging.go` | Create | `Logging` middleware – logs method, path, status via `slog` |
| `internal/middleware/recovery.go` | Create | `Recovery` middleware – catches panics, returns 500 |
| `cmd/api/main.go` | Modify | Replace mux with Chi, wire dependencies, register routes |
| `go.mod` | Modify | Add `github.com/go-chi/chi/v5` dependency |
| `internal/handler/handler_test.go` | Create | Table‑driven tests using `httptest.NewRecorder` |

## Interfaces / Contracts

### Handler struct

```go
// Handler holds the usecase dependency for all HTTP handlers.
type Handler struct {
    uc *usecase.RateUsecase
}

// NewHandler creates a Handler with the given usecase.
func NewHandler(uc *usecase.RateUsecase) *Handler {
    return &Handler{uc: uc}
}
```

### Response helpers

```go
// respondJSON writes a JSON response with the given status code.
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(payload)
}

// respondError writes a standardized error envelope.
func respondError(w http.ResponseWriter, status int, code, message string) {
    respondJSON(w, status, map[string]interface{}{
        "error": map[string]string{"code": code, "message": message},
    })
}

// mapError converts a domain error to HTTP status and error code.
func mapError(err error) (int, string) {
    switch {
    case errors.Is(err, domain.ErrNotFound):
        return http.StatusNotFound, "NOT_FOUND"
    case errors.Is(err, domain.ErrInvalidInput):
        return http.StatusBadRequest, "BAD_REQUEST"
    case errors.Is(err, domain.ErrDuplicateRate):
        return http.StatusConflict, "CONFLICT"
    default:
        return http.StatusInternalServerError, "INTERNAL_ERROR"
    }
}
```

### Route registration (in `cmd/api/main.go`)

```go
r := chi.NewRouter()
r.Use(middleware.Recovery)
r.Use(middleware.Logging)

h := handler.NewHandler(uc)
r.Get("/rates", h.GetRates)
r.Get("/rates/history", h.GetHistory)
r.Route("/admin", func(r chi.Router) {
    r.Post("/scrape", h.TriggerScrape)
})
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Handler logic (param parsing, usecase delegation, response formatting) | `httptest.NewRecorder`, mock usecase via function fields or struct replacement |
| Integration | Chi routing, middleware chaining | Not needed for this change; covered by unit tests |
| E2E | Full HTTP stack | Out of scope (deferred to future change) |

**Test scenarios** (per spec):

1. `GET /rates` – no filter → 200 with `{ "data": [...] }`
2. `GET /rates?currency=USD` – filtered → 200 with USD only
3. `GET /rates?currency=USD&type=reference` – double filter → 200
4. `GET /rates/history` – all filters → 200
5. `GET /rates/history` – empty result → 200 with `{ "data": [] }`
6. `POST /admin/scrape` – success → 202 with confirmation
7. `GET /rates/history?limit=abc` – invalid param → 400

## Migration / Rollout

No data migration required. The change is additive: new packages, new routes, modified `main.go`. The existing usecase layer remains untouched. Rollback: delete `internal/handler/` and `internal/middleware/`, revert `cmd/api/main.go`, remove Chi from `go.mod`.

## Open Questions

- Should we define a `RateUsecase` interface for better testability, or keep concrete dependency for now? (Current decision: concrete; revisit if multiple usecase implementations appear)
- Should the recovery middleware log the full panic stack trace or just the error message? (Spec says log via `slog`; stack trace omitted for security)
- Should we add a `/health` endpoint for load balancers? (Not in scope for this change)
