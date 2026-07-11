# Tasks: Request ID Header Migration

## Review Workload Forecast

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: stacked-to-main
400-line budget risk: Low

| Field | Value |
|-------|-------|
| Estimated changed lines | ~120 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | single PR |
| Delivery strategy | auto-chain |
| Chain strategy | stacked-to-main |

### Suggested Work Units
| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Relocate request ID in presenter and wire test router | PR 1 | presenter.go core changes and rate_handlers_test.go router setup |
| 2 | Assert header propagation and envelope cleanliness in tests | PR 1 | Updates to presenter, auth, integration, ratelimit, and router tests |

## Phase 1: Foundation / Infrastructure
- [x] 1.1 Verify `internal/http/httpx/httpx.go` exports `GetRequestID` using `go-chi` middleware helpers.

## Phase 2: Core Implementation
- [x] 2.1 Edit [presenter.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/presenter/presenter.go):
  - Remove `RequestID` string field from `ResponseEnvelope` struct.
  - In `OK` and `Created` helper functions, call `w.Header().Set("X-Request-ID", reqID)` immediately prior to writing JSON.
  - In `Error` helper function, call `w.Header().Set("X-Request-ID", reqID)` immediately prior to writing JSON.

## Phase 3: Integration / Wiring
- [x] 3.1 Verify that `internal/http/router/router.go` mounts the Chi request ID middleware (`chimw.RequestID`) before other middlewares.
- [x] 3.2 Update `newTestRouter` in [rate_handlers_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/handler/rate_handlers_test.go) to import `chimw "github.com/go-chi/chi/v5/middleware"` and mount `r.Use(chimw.RequestID)` so that request headers are generated in handler tests.

## Phase 4: Testing / Verification
- [x] 4.1 Update [presenter_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/presenter/presenter_test.go):
  - Inject a dummy request ID into the request context using `context.WithValue(r.Context(), chimw.RequestIDKey, "test-req-id")` or similar.
  - Assert response headers contain `X-Request-ID: test-req-id`.
  - Assert that `"request_id"` is not present in the unmarshaled JSON body.
- [x] 4.2 Update [rate_handlers_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/handler/rate_handlers_test.go):
  - Remove `RequestID` field from local `verificationEnvelope` struct.
  - Assert presence of `X-Request-ID` header in test responses.
- [x] 4.3 Update [auth_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/auth_test.go):
  - Remove `RequestID` field from package-level `responseEnvelope` struct.
  - Assert presence of `X-Request-ID` header in error responses.
- [x] 4.4 Update [integration_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/integration_test.go):
  - Remove `RequestID` field from local `responseEnvelope` struct.
  - Add assertion validating `X-Request-ID` header is propagated.
- [x] 4.5 Update [router_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/http/router/router_test.go):
  - Replace assertion checking for `"request_id"` in JSON map with checks ensuring `"request_id"` is absent from JSON body and that `"X-Request-ID"` exists in headers.

## Phase 5: Cleanup
- [x] 5.1 Run all tests locally (`go test ./...`) to verify compilation and that all tests pass.
- [x] 5.2 Format modified files (`go fmt ./...`).
