# Tasks: Error Handling Refactor

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
| 1 | Foundation & Core: Update ResponseEnvelope & route fallback | PR 1 | Base changes in presenter and router. |
| 2 | Testing: Update and add test assertions across files | PR 1 | Verifies updated JSON error structure. |

## Phase 1: Foundation / Infrastructure
- [x] 1.1 Update `ResponseEnvelope` in [presenter.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/presenter/presenter.go): rename `ErrorCode` to `Code` and adjust JSON tags to `data,omitempty`, `code,omitempty`, and `error,omitempty`.
- [x] 1.2 Update presentation helper functions `OK`, `Created`, and `Error` in [presenter.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/presenter/presenter.go) to match the new struct and omit optional fields.

## Phase 2: Core Implementation
- [x] 2.1 Import `apierrors` and `presenter` in [router.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/http/router/router.go).
- [x] 2.2 Wire Chi's custom `r.NotFound` handler in [router.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/http/router/router.go) to return a structured JSON response using `presenter.Error()`.

## Phase 3: Integration / Wiring
- [x] 3.1 Verify routing integration by ensuring global middlewares pass request flow correctly down to the NotFound handler.

## Phase 4: Testing / Verification
- [x] 4.1 Update assertions in [presenter_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/presenter/presenter_test.go) to check `Code` / `code` mapping and verify `data` is omitted.
- [x] 4.2 Update verification structs and assertions in [rate_handlers_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/handler/rate_handlers_test.go) to test for `code` and the absence of `data` on errors.
- [x] 4.3 Update package-level envelope and response checks in [auth_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/auth_test.go) to use `code` / `Code`.
- [x] 4.4 Update integration checks in [integration_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/integration_test.go) and [ratelimit_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/ratelimit_test.go).
- [x] 4.5 Add `TestRouter_NotFound` to [router_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/http/router/router_test.go) to request an unregistered route and verify HTTP 404 and structured JSON error response.

## Phase 5: Cleanup
- [x] 5.1 Run all tests locally to verify clean compilation and zero regression.
