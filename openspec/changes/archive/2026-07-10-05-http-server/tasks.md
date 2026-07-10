# Tasks: HTTP Server & API Endpoints (05-http-server)

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 385–420 |
| 400-line budget risk | Medium |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 → PR 2 → PR 3 |
| Delivery strategy | ask-on-risk |
| Chain strategy | stacked-to-main |

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Handler package foundation + response helpers | PR 1 | base: main; chi dep + handler.go + responses.go + mapError tests |
| 2 | Middleware + HTTP handlers + main.go wiring | PR 2 | base: PR 1 branch; logging, recovery, rate_handlers, main.go rewrite |
| 3 | Handler integration tests (7 scenarios) | PR 3 | base: PR 2 branch; httptest + mocked usecase |

## Phase 1: Foundation

- [x] 1.1 Add `github.com/go-chi/chi/v5` dependency via `go get`
- [x] 1.2 Create `internal/handler/handler.go` — `Handler` struct holding `*usecase.RateUsecase`, `NewHandler` constructor
- [x] 1.3 Create `internal/handler/responses.go` — `respondJSON`, `respondError`, `mapError` helpers (domain error → HTTP status mapping per spec table)

## Phase 2: Core Implementation

- [x] 2.1 Create `internal/handler/rate_handlers.go` — `GetRates` (parse `currency`/`type` query params, call `GetCurrentRates`), `GetHistory` (parse all 5 params, call `GetHistoryRates`), `TriggerScrape` (call `ScrapeRates`, return 202)
- [x] 2.2 Create `internal/middleware/logging.go` — request logging middleware using `slog` (method, path, status code)
- [x] 2.3 Create `internal/middleware/recovery.go` — panic recovery middleware, returns HTTP 500, logs via `slog`
- [x] 2.4 Rewrite `cmd/api/main.go` — replace `http.NewServeMux` with `chi.NewRouter`, wire config→store→usecase→handler→router, register routes (`r.Get("/rates", ...)`, `r.Get("/rates/history", ...)`, `r.Route("/admin", ...)`)

## Phase 3: Testing

- [x] 3.1 Create `internal/handler/handler_test.go` — table-driven tests with `httptest.NewRecorder`, mock `RateUsecase` responses
- [x] 3.2 Test: `GET /rates` no filter → 200 with `{ "data": [...] }`
- [x] 3.3 Test: `GET /rates?currency=USD` → 200 with USD only
- [x] 3.4 Test: `GET /rates?currency=USD&type=reference` → 200 double-filtered
- [x] 3.5 Test: `GET /rates/history` with all filters → 200
- [x] 3.6 Test: `GET /rates/history` empty result → 200 with `{ "data": [] }`
- [x] 3.7 Test: `POST /admin/scrape` success → 202 with `{ "data": { "message": "scrape triggered", "rates_scraped": N } }`
- [x] 3.8 Test: `GET /rates/history?limit=abc` → 400 with error envelope
- [x] 3.9 Run `go test ./internal/handler/...` — all 7+ scenarios pass

## Phase 4: Verification

- [x] 4.1 Run `go build ./cmd/api` — compiles without errors
- [x] 4.2 Verify: panic recovery returns 500 without crashing server
- [x] 4.3 Verify: 500 responses never contain internal error details (SQL, stack traces)
- [x] 4.4 Verify: all responses use `{ "data": ... }` or `{ "error": { "code", "message" } }` envelope
