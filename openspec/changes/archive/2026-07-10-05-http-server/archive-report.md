# Archive Report: 05 — HTTP Server & API Endpoints

**Change**: 05-http-server
**Status**: COMPLETE
**Archived**: 2026-07-10
**Verdict**: PASS (no CRITICAL issues)

---

## Executive Summary

Expose the Venezuela Rates API over HTTP using Chi router. The usecase layer (issue #5) was complete but the application had no serving surface — just a bare `http.ListenAndServe` placeholder. This change wires up the HTTP transport, defines RESTful endpoints (`GET /rates`, `GET /rates/history`, `POST /admin/scrape`), standardizes JSON response envelopes (`{ "data": ... }` / `{ "error": { "code", "message" } }`), adds request logging and panic recovery middleware, and provides handler‑level tests with mocked dependencies. All acceptance criteria from issue #6 satisfied, all spec scenarios covered by passing tests, no critical verification issues.

## What Was Built

| Capability | Impact |
|------------|--------|
| **Chi Router Setup** | Replaces `http.NewServeMux` with `chi.NewRouter`; routes registered under `/rates`, `/rates/history`, `/admin/scrape` |
| **Middleware Stack** | Panic recovery (returns 500, logs via `slog`) and request logging (method, path, status, duration) applied to all routes |
| **REST Handlers** | `GetRates` (optional `currency`/`type` filtering), `GetHistory` (all 5 query params), `TriggerScrape` (returns 202) |
| **JSON Envelope** | Success: `{ "data": <payload> }`; Error: `{ "error": { "code": "<CODE>", "message": "<message>" } }` |
| **Error Mapping** | Domain errors → HTTP status: `ErrNotFound`→404, `ErrInvalidInput`→400, `ErrDuplicateRate`→409, else→500 |
| **Internal Sanitization** | 500 responses never expose SQL queries, stack traces, or file paths |
| **Handler Tests** | 7 table‑driven scenarios using `httptest.NewRecorder` with mocked `RateUsecase` |
| **Dependency Wiring** | `cmd/api/main.go` wires config → store → usecase → handler → router |

## Key Decisions

| Decision | Rationale |
|----------|-----------|
| **Chi router over stdlib mux** | Route grouping (`/admin`), middleware chaining, lightweight footprint; well‑maintained dependency |
| **Centralized error mapper** | Single source of truth for domain‑error‑to‑HTTP‑status mapping; prevents scattered switch statements |
| **Recovery middleware writes raw JSON** | Maintains the JSON error envelope contract; avoids `http.Error` which sets `text/plain` |
| **Middleware order: recovery → logging** | Panic recovery must run first to catch panics in later middleware (including logging) |
| **Handler depends on concrete `*usecase.RateUsecase`** | Usecase layer stable with single implementation; testing uses function‑field mocks via `httptest`; keeps codebase minimal |

## Files Modified

| File | Lines Changed | Description |
|------|---------------|-------------|
| `internal/handler/handler.go` | +30 | `Handler` struct, `NewHandler` constructor, `Usecaser` interface for testability |
| `internal/handler/rate_handlers.go` | +120 | `GetRates`, `GetHistory`, `TriggerScrape` HTTP handlers |
| `internal/handler/responses.go` | +60 | `respondJSON`, `respondError`, `mapError` helpers |
| `internal/handler/handler_test.go` | +180 | 10 table‑driven tests covering all spec scenarios |
| `internal/middleware/logging.go` | +40 | Request logging middleware using `slog` |
| `internal/middleware/recovery.go` | +35 | Panic recovery middleware returning JSON 500 |
| `internal/middleware/logging_test.go` | +45 | 3 tests: method/path logging, status recording, pass‑through |
| `internal/middleware/recovery_test.go` | +60 | 4 tests: catches panic, error envelope, logs panic, pass‑through |
| `cmd/api/main.go` | +100 | Rewrite: Chi router, middleware stack, route registration, graceful shutdown |
| `go.mod` / `go.sum` | +2 | Added `github.com/go-chi/chi/v5` dependency |

**Total estimated changed lines**: ~385–420 (within 400‑line PR budget after stacked slicing).

## Test Results

- **Total tests**: 87 (including subtests)
- **All passing**: ✅
- **Coverage**:
  - `internal/handler`: 93.2%
  - `internal/middleware`: 100.0%
  - `internal/usecase`: 92.0%
  - `internal/store`: 82.1%
  - `internal/scraper`: 86.4%
  - `internal/config`: 94.7%
- **Regression**: None
- **Spec compliance**: All 18 spec scenarios covered by passing tests

## Engram Artifact IDs

| Artifact | Observation ID |
|----------|----------------|
| proposal | #554 |
| spec | #555 |
| design | #556 |
| tasks | #557 |
| apply-progress | #558 |
| verify-report | #562 |

## Lessons Learned

1. **Recovery middleware must preserve JSON envelope** — Using `http.Error` would set `Content-Type: text/plain` and break the error envelope contract. Writing raw JSON bytes directly ensures consistency.
2. **Middleware order matters** — Recovery must run before logging to catch panics in logging middleware itself. The spec mandated this order, and implementation followed it.
3. **Stacked PR slicing kept review budget manageable** — The 400‑line budget risk was mitigated by splitting into three PRs (foundation, handlers+middleware, tests). Each PR was independently reviewable.

## Archive Verification

- [x] Main specs created for three new domains (`http-server`, `api-endpoints`, `error-handling`)
- [x] Change folder moved to `openspec/changes/archive/2026-07-10-05-http-server/`
- [x] Archive contains all artifacts (proposal, specs, design, tasks, apply-progress, verify-report)
- [x] All 17 implementation tasks marked complete in tasks artifact
- [x] Active changes directory no longer contains this change
- [x] Verify report shows PASS with no CRITICAL issues

## SDD Cycle Complete

The change has been fully planned, implemented, verified, and archived.
Ready for the next change.