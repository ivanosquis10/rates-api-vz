# Apply Progress: 05-http-server — All Slices (PR 1+2+3)

**Date**: 2026-07-10
**Mode**: Strict TDD
**Delivery**: stacked-to-main

## Applied Tasks

### PR 1 — Foundation (already done)
- [x] 1.1 Add `github.com/go-chi/chi/v5` dependency via `go get`
- [x] 1.2 Create `internal/handler/handler.go` — `Handler` struct holding `*usecase.RateUsecase`, `NewHandler` constructor
- [x] 1.3 Create `internal/handler/responses.go` — `respondJSON`, `respondError`, `mapError` helpers

### PR 2 — Handlers + Middleware + Wiring (already done)
- [x] 2.1 Create `internal/handler/rate_handlers.go` — GetRates, GetHistory, TriggerScrape
- [x] 2.2 Create `internal/middleware/logging.go` — slog-based request logging
- [x] 2.3 Create `internal/middleware/recovery.go` — panic recovery with JSON 500
- [x] 2.4 Rewrite `cmd/api/main.go` — Chi router, dependency wiring, graceful shutdown

### PR 3 — Integration Verification Tests (this batch)
- [x] 3.1 handler_test.go with httptest.NewRecorder + mock (created in PR 1)
- [x] 3.2 Test GET /rates no filter → 200 (TestGetRates_NoFilter)
- [x] 3.3 Test GET /rates?currency=USD → 200 (TestGetRates_FilterByCurrency)
- [x] 3.4 Test GET /rates?currency=USD&type=reference → 200 (TestGetRates_FilterByCurrencyAndType)
- [x] 3.5 Test GET /rates/history with all filters → 200 (TestGetHistory_WithAllFilters)
- [x] 3.6 Test GET /rates/history empty result → 200 (TestGetHistory_EmptyResult)
- [x] 3.7 Test POST /admin/scrape success → 202 (TestTriggerScrape_Success)
- [x] 3.8 Test GET /rates/history?limit=abc → 400 (TestGetHistory_InvalidLimit)
- [x] 3.9 go test ./internal/handler/... — all pass (17 tests)
- [x] 4.1 go build ./cmd/api — compiles without errors
- [x] 4.2 Panic recovery returns 500 without crashing (TestVerification_PanicRecoveryReturns500)
- [x] 4.3 500 responses never contain internal details (TestVerification_500ResponsesSanitized)
- [x] 4.4 All responses use correct envelope (TestVerification_ResponseEnvelopeConsistency)

## TDD Cycle Evidence

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|------|-----------|-------|------------|-----|-------|-------------|----------|
| 1.1 | N/A | N/A | N/A (new dep) | ➖ Structural | ✅ `go get` succeeded | ➖ Single | ➖ None needed |
| 1.2 | handler_test.go | Unit | N/A (new) | ✅ Written | ✅ 2/2 passed | ➖ Single scenario | ➖ None needed |
| 1.3 | responses_test.go | Unit | N/A (new) | ✅ Written | ✅ 7/7 passed | ✅ 4 error mappings + 2 respondJSON + 1 respondError | ➖ None needed |
| 2.1 | rate_handlers_test.go | Unit | ✅ 9/9 | ✅ Written | ✅ 8/8 passed | ✅ 8 cases (3+3+2) | ✅ Clean |
| 2.2 | logging_test.go | Unit | N/A (new) | ✅ Written | ✅ 3/3 passed | ✅ 3 cases | ✅ Clean |
| 2.3 | recovery_test.go | Unit | N/A (new) | ✅ Written | ✅ 4/4 passed | ✅ 4 cases | ✅ Clean |
| 2.4 | N/A (wiring) | N/A | N/A | N/A | ✅ Compiles | N/A | N/A |
| 3.1-3.8 | rate_handlers_test.go | Unit | ✅ 8/8 | ✅ Written (PR 1) | ✅ Passed | ✅ 8 cases | ✅ Clean |
| 3.9 | — | Unit | ✅ 17/17 | N/A | ✅ Passed | N/A | N/A |
| 4.1 | — | Build | N/A | N/A | ✅ Compiles | N/A | N/A |
| 4.2 | rate_handlers_test.go | Integration | ✅ 1/1 | ✅ Written | ✅ Passed | ✅ 2 requests | ✅ Clean |
| 4.3 | rate_handlers_test.go | Integration | ✅ 1/1 | ✅ Written | ✅ Passed | ✅ SQL+detail checks | ✅ Clean |
| 4.4 | rate_handlers_test.go | Integration | ✅ 4/4 | ✅ Written | ✅ Passed | ✅ 4 scenarios | ✅ Clean |

## Test Summary

- Total tests written: 27 (9 PR1 + 15 PR2 + 3 verification functions with 11 sub-tests)
- Total tests passing: 60 (across all packages)
- Layers used: Unit (24), Integration (3 functions, 11 sub-tests)

## Files Changed

| File | Action | What Was Done |
|------|--------|---------------|
| `go.mod` | Modified | Added `github.com/go-chi/chi/v5` dependency |
| `internal/handler/handler.go` | Created+Modified | Handler struct, NewHandler, Usecaser interface, NewHandlerFromUsecaser |
| `internal/handler/responses.go` | Created | `respondJSON`, `respondError`, `mapError` helpers |
| `internal/handler/rate_handlers.go` | Created | GetRates, GetHistory, TriggerScrape handlers |
| `internal/handler/handler_test.go` | Created | Tests for NewHandler (nil + with usecase) |
| `internal/handler/responses_test.go` | Created | Tests for respondJSON, respondError, mapError |
| `internal/handler/rate_handlers_test.go` | Created+Modified | 8 unit tests + 3 integration verification tests |
| `internal/middleware/logging.go` | Created | Logging middleware with slog |
| `internal/middleware/logging_test.go` | Created | 3 tests for logging middleware |
| `internal/middleware/recovery.go` | Created | Recovery middleware with JSON 500 |
| `internal/middleware/recovery_test.go` | Created | 4 tests for recovery middleware |
| `cmd/api/main.go` | Rewritten | Chi router, middleware stack, route registration, graceful shutdown |

## Deviations from Design

- Added `Usecaser` interface to handler.go for test mockability (design said concrete dependency, but test mock requires interface). Kept `NewHandler(*usecase.RateUsecase)` for backward compatibility.
- Recovery middleware writes raw JSON bytes instead of using `http.Error` to maintain the JSON error envelope contract.

## Status

17/17 tasks complete. All phases done. Ready for sdd-verify.
