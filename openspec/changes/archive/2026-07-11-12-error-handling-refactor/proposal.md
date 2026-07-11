# Proposal: Error Handling Refactor

## Intent
Resolve Issues #13 and #14 by standardizing the error response structure, returning structured JSON 404 responses for non-existent endpoints, and optimizing the error response envelope.

## Scope
### In Scope
- Rename `error_code` to `code` and omit the empty `data` field in error responses.
- Configure a custom Not Found handler in `internal/http/router/router.go`.
- Refactor all presenter helpers (`OK`, `Created`, `Error`) to align with the new envelope struct fields.
- Refactor unit/integration tests to assert the new envelope fields and verify structured 404 behavior.

### Out of Scope
- Adding any new routes or domain logic.
- Changing success response structure except ensuring `data` is omitted only when empty if applicable.

## Capabilities
### New Capabilities
- `<json-404-fallback>`: API returns a structured JSON error response for invalid endpoints with status 404 and code `NOT_FOUND`.
### Modified Capabilities
- `<structured-error-responses>`: Error responses serialize using `code` instead of `error_code`, and omit the `data` field when empty/erroring.

## Approach
1. Refactor `ResponseEnvelope` struct in `internal/presenter/presenter.go` to update JSON tags.
2. Register `r.NotFound(func(w http.ResponseWriter, r *http.Request) { presenter.Error(w, r, ...) })` in `internal/http/router/router.go`.
3. Update all Go test files to check `"code"` instead of `"error_code"` and assert that `"data"` is omitted from error responses.

## Affected Areas
| Area | Impact | Description |
|------|--------|-------------|
| `internal/presenter/presenter.go` | Modified | Refactor `ResponseEnvelope` fields and JSON tags. |
| `internal/http/router/router.go` | Modified | Add custom `r.NotFound` handler. |
| `internal/presenter/presenter_test.go` | Modified | Align assertions with the new error envelope structure. |
| `internal/handler/rate_handlers_test.go` | Modified | Update test response envelope types and assertions. |
| `internal/middleware/auth_test.go` | Modified | Update test response envelope types and assertions. |
| `internal/middleware/integration_test.go` | Modified | Update test response envelope types and assertions. |
| `internal/middleware/ratelimit_test.go` | Modified | Update test response envelope types and assertions. |
| `internal/http/router/router_test.go` | Modified | Add test case for non-existent endpoint. |

## Risks
| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Breaking clients expecting `error_code` | High | Document API change. |
| Regression in existing error formatting | Low | Run all unit/integration tests locally to verify compatibility. |

## Rollback Plan
Perform a git revert of the refactoring commit: `git revert HEAD`.

## Dependencies
- None.

## Success Criteria
- [ ] Non-existent endpoints return 404 JSON with code `NOT_FOUND`.
- [ ] Error responses omit `data` and use `code` instead of `error_code`.
- [ ] `go test ./...` runs successfully.
