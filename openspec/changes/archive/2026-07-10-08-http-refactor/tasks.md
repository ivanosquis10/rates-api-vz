# Tasks: 08-http-refactor

## Review Workload Forecast

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: stacked-to-main
400-line budget risk: Medium

| Field | Value |
|-------|-------|
| Estimated changed lines | ~400 |
| 400-line budget risk | Medium |
| Chained PRs recommended | No |
| Suggested split | single PR |
| Delivery strategy | auto-chain |
| Chain strategy | stacked-to-main |

### Suggested Work Units
| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Foundation/Infrastructure Packages | PR 1 | Base package setup |
| 2 | Middleware Refactoring & Testing | PR 1 | Apply new structures and logger request ID |
| 3 | Handlers & Clean-up | PR 1 | Switch to presenter, remove old responses |

## Phase 1: Foundation / Infrastructure
- [x] 1.1 Create [apierrors.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/apierrors/apierrors.go) with API-specific errors.
- [x] 1.2 Create [httpx.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/httpx/httpx.go) providing `WriteJSON` and `GetRequestID`.
- [x] 1.3 Create [presenter.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/presenter/presenter.go) with `ResponseEnvelope` using local pointer helpers to satisfy pointer constraints safely.

## Phase 2: Core Implementation
- [x] 2.1 Refactor [auth.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/auth.go) using `presenter.Error`.
- [x] 2.2 Refactor [ratelimit.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/ratelimit.go) using `presenter.Error` and headers.
- [x] 2.3 Refactor recovery handler in [recovery.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/recovery.go) to return standard envelope error.
- [x] 2.4 Refactor logging in [logging.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/logging.go) to fetch `request_id` via `httpx.GetRequestID` and output it in structured attributes.

## Phase 3: Integration / Wiring
- [x] 3.1 Refactor [rate_handlers.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/handler/rate_handlers.go) to use presenter helpers.
- [x] 3.2 Update handler initialization in [handler.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/handler/handler.go) removing responses references.

## Phase 4: Testing / Verification
- [x] 4.1 Update [auth_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/auth_test.go) asserting response envelope structure.
- [x] 4.2 Update [ratelimit_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/ratelimit_test.go) asserting response envelope structure.
- [x] 4.3 Update [rate_handlers_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/handler/rate_handlers_test.go) asserting response envelope structure.
- [x] 4.4 Run validation test suite via `go test ./...`.

## Phase 5: Cleanup
- [x] 5.1 Delete [responses.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/handler/responses.go) and [responses_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/handler/responses_test.go).
