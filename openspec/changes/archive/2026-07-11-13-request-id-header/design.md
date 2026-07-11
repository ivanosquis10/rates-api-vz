# Design: 13-request-id-header

## Technical Approach
Refactor request ID propagation by migrating the request ID from the JSON response body envelope to a standardized `X-Request-ID` HTTP header for all HTTP API responses.

## Architecture Decisions
### Decision: Inject X-Request-ID Header in Presenter Helpers
**Choice**: Modify `presenter.OK`, `presenter.Created`, and `presenter.Error` to call `w.Header().Set("X-Request-ID", reqID)` and remove the `RequestID` field from the JSON `ResponseEnvelope` struct.
**Alternatives considered**: Injecting the header via a separate middleware.
**Rationale**: Injecting in a middleware is standard, but the presenter already maps errors to HTTP response codes and serializes JSON. Since writing JSON body commits headers (locks headers once `w.WriteHeader` or `w.Write` is called), placing `w.Header().Set("X-Request-ID", reqID)` inside presenter helpers immediately prior to JSON serialization guarantees that all handlers consistently emit the `X-Request-ID` header.

## Data Flow
```
Client ──► [HTTP Request] ──► [Chi router (chimw.RequestID)] ──► Handler ─┐
                                                                           │
                                                                           ▼
Client ◄── [w.Header().Set("X-Request-ID", reqID)] ◄── Presenter ◄── Usecase/Err
```

## File Changes
| File | Action | Description |
|------|--------|-------------|
| [internal/presenter/presenter.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/presenter/presenter.go) | Modified | Remove `RequestID` struct field from `ResponseEnvelope`. Call `w.Header().Set("X-Request-ID", reqID)` inside `OK`, `Created`, and `Error` functions before calling `httpx.WriteJSON`. |
| [internal/presenter/presenter_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/presenter/presenter_test.go) | Modified | Inject dummy request ID into request context using `context.WithValue` and assert `w.Header().Get("X-Request-ID")` matches. Confirm `request_id` is absent in decoded JSON body. |
| [internal/handler/rate_handlers_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/handler/rate_handlers_test.go) | Modified | Remove `RequestID` from local `verificationEnvelope` struct. Inject `chimw.RequestID` middleware in `newTestRouter`. Check for `X-Request-ID` header presence in test assertions. |
| [internal/middleware/auth_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/auth_test.go) | Modified | Remove `RequestID` field from package-level `responseEnvelope` struct. Inject test request ID and assert `X-Request-ID` is in header. |
| [internal/middleware/integration_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/integration_test.go) | Modified | Remove `RequestID` field from local `responseEnvelope` struct. Add assertion for header validation. |
| [internal/middleware/ratelimit_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/ratelimit_test.go) | Modified | Assert `X-Request-ID` response header. `responseEnvelope` updated via package-level definition. |
| [internal/http/router/router_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/http/router/router_test.go) | Modified | Replace assertion checking for `"request_id"` in JSON map with check ensuring `"request_id"` is not present in body, and that the `"X-Request-ID"` header exists. |

## Interfaces / Contracts
### HTTP API Response Envelope Contract
All JSON responses will conform to `ResponseEnvelope` without `request_id`:
```go
type ResponseEnvelope struct {
	Success bool            `json:"success"`
	Data    any             `json:"data,omitempty"`
	Code    *apierrors.Code `json:"code,omitempty"`
	Error   *string         `json:"error,omitempty"`
}
```
All successful and failure responses will carry the HTTP header:
`X-Request-ID: <string>`

## Testing Strategy
| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit / presenter | Header injection and payload cleanliness | Assert presence of `X-Request-ID` in headers, and absence of `request_id` in response body. |
| Unit / middleware | Auth & Rate Limiting propagation | Check that headers contain request ID during error scenarios (e.g. 401, 429). |
| Integration / router | End-to-end Router execution | Verify header is propagated successfully on default/not-found endpoints. |

## Migration / Rollout
No migration required. Client applications must be notified of the deprecation of `request_id` in the JSON response envelope.

## Open Questions
- None
