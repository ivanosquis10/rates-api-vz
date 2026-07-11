# Proposal: 13-request-id-header

## Intent
Refactor request ID handling by removing the `request_id` field from the JSON response envelope and instead injecting it into the HTTP response headers as `X-Request-ID`. This reduces response payload size and aligns with REST API best practices.

## Scope
### In Scope
- Remove the `RequestID` field from the `ResponseEnvelope` struct in [presenter.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/presenter/presenter.go).
- Modify `OK`, `Created`, and `Error` functions in [presenter.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/presenter/presenter.go) to retrieve the request ID using `httpx.GetRequestID` and write it to the `X-Request-ID` header before response delivery.
- Update assertions in unit and integration tests: [presenter_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/presenter/presenter_test.go), [rate_handlers_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/handler/rate_handlers_test.go), [auth_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/auth_test.go), [integration_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/integration_test.go), [ratelimit_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/ratelimit_test.go), and [router_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/http/router/router_test.go) to remove JSON `request_id` checks and verify the `X-Request-ID` header.

### Out of Scope
- Modifying request ID logging formats or middleware logic.
- Adding or changing other headers or envelope fields.

## Capabilities
### New Capabilities
- None

### Modified Capabilities
- `presenter`: Response formatting logic updated to output the request ID as an HTTP header (`X-Request-ID`) instead of a JSON property.

## Approach
1. Edit [presenter.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/presenter/presenter.go) to remove the `RequestID` struct field.
2. In `OK`, `Created`, and `Error`, inject the retrieved request ID into the header: `w.Header().Set("X-Request-ID", reqID)`.
3. Update each target test file to replace assertions for `request_id` in the JSON body with assertions checking for a non-empty `X-Request-ID` response header.

## Affected Areas
| Area | Impact | Description |
|------|--------|-------------|
| [presenter.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/presenter/presenter.go) | Modified | Removed struct field, header injection in helper methods. |
| [presenter_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/presenter/presenter_test.go) | Modified | Assert header output on error responses. |
| [rate_handlers_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/handler/rate_handlers_test.go) | Modified | Remove `RequestID` field from validation struct, assert header. |
| [auth_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/auth_test.go) | Modified | Remove `RequestID` field from test envelope, assert header. |
| [integration_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/integration_test.go) | Modified | Remove `RequestID` field from test envelope, assert header. |
| [ratelimit_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/ratelimit_test.go) | Modified | Assert header presence on rate-limit responses. |
| [router_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/http/router/router_test.go) | Modified | Assert header presence on standard fallback/not-found. |

## Risks
| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Client integration issues | Low | Clients should parse HTTP response headers for tracking request IDs. |

## Rollback Plan
Run `git checkout -- .` to discard local changes and restore the previous codebase state.

## Dependencies
- None

## Success Criteria
- [ ] `X-Request-ID` HTTP header is present and populated on all standard, created, and error API responses.
- [ ] `request_id` is absent from all JSON response bodies.
- [ ] All updated tests pass successfully.
