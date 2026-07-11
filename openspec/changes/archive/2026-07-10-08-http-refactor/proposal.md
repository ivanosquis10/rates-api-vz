# Proposal: 08-http-refactor

## Intent
Standardize HTTP response wrapping and centralize error mapping using patterns from `ivanopay`. This improves API consistency, enforces response contract validation, and ensures internal error details do not leak to client-facing responses.

## Scope
### In Scope
- Create `internal/apierrors` package containing standard API error codes (`UNAUTHORIZED`, `RATE_LIMITED`, `NOT_FOUND`, `BAD_REQUEST`, `INTERNAL_ERROR`) and a custom `apierrors.Error` type.
- Create `internal/httpx` package with `WriteJSON` helper and context-aware request ID extractor wrapping Chi's middleware.
- Create `internal/presenter` package with standard JSON envelope (fields: `success`, `data`, `error_code`, `error`, `request_id`) and wrappers `OK`, `Created`, `NoContent`, and `Error`.
- Map domain errors (rate limit, missing keys, not found) to HTTP statuses and log unexpected Internal Server Errors using `slog`.
- Refactor rate handlers in `rate_handlers.go` to use presenter functions.
- Update tests in `rate_handlers_test.go` and remove deprecated `responses.go` and `responses_test.go`.

### Out of Scope
- DB schema modifications or migration changes.
- Scraper schedule logic or adding new API endpoints.

## Capabilities
### New Capabilities
- `apierrors`: Custom error type and API error code catalog for client-safe responses.
- `httpx`: Utility helpers for standard JSON rendering and context request ID retrieval.
- `presenter`: Unified HTTP JSON presentation layer wrapping responses and centralizing domain error mapping.

### Modified Capabilities
- `http-server-handlers`: Standardized envelope and error codes for all rates API endpoints.

## Approach
1. Implement `internal/apierrors` with codes and error struct.
2. Implement `internal/httpx` with `WriteJSON` and `RequestID` extractor.
3. Implement `internal/presenter` mapping domain errors (e.g., `domain.ErrNotFound` -> `NOT_FOUND`) and wrapping responses.
4. Refactor `rate_handlers.go` endpoint functions to use the new presenter methods.
5. Clean up old code by deleting `responses.go` and `responses_test.go`.
6. Refactor integration and handler tests in `rate_handlers_test.go` to match the new envelope payload structures.

## Affected Areas
| Area | Impact | Description |
|------|--------|-------------|
| `internal/apierrors` | New | API error codes and custom error type |
| `internal/httpx` | New | HTTP JSON rendering and request ID extraction |
| `internal/presenter` | New | Response wrapper envelope and error mapper |
| `internal/handler/rate_handlers.go` | Modified | Use presenter helpers |
| `internal/handler/handler.go` | Modified | Remove deprecated responder function dependencies |
| `internal/handler/responses.go` | Removed | Delete obsolete helper file |
| `internal/handler/rate_handlers_test.go` | Modified | Update tests to assert new envelope structure |
| `internal/handler/responses_test.go` | Removed | Delete obsolete helper tests |

## Risks
| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Breaking client expectations | Low | API endpoints are updated under a standard layout; verify with tests. |

## Rollback Plan
Discard branch changes using `git checkout -- .` and check out the `main` branch.

## Dependencies
- Go 1.25.0
- `github.com/go-chi/chi/v5`

## Success Criteria
- [ ] New packages `apierrors`, `httpx`, and `presenter` build clean.
- [ ] Endpoint handlers in `rate_handlers.go` return the new standard envelope.
- [ ] Obsolete files `responses.go` and `responses_test.go` are removed.
- [ ] All handler and integration tests pass successfully using updated assertions.
