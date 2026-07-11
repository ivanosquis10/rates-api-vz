# Design: Error Handling Refactor

## Technical Approach
Standardize API error handling and routing boundaries by:
- Updating `ResponseEnvelope` in `internal/presenter/presenter.go` to rename `error_code` to `code` and use `omitempty` on optional response fields (`data`, `code`, `error`).
- Updating presenter functions (`OK`, `Created`, `Error`) to align with the new envelope layout.
- Wiring Chi's custom `r.NotFound` handler in `internal/http/router/router.go` to intercept unregistered routes.
- Adapting the test suites to assert against the updated JSON format and verify that error payloads omit the `data` field.

## Architecture Decisions

### Decision: Omitting the `data` field in error payloads
**Choice**: Apply the Go `json:"data,omitempty"` tag on the `Data any` field.
**Alternatives considered**: Introducing separate success and error envelope structs.
**Rationale**: In Go's `encoding/json` package, an `any` (interface) field is omitted when its value is `nil`. When the presenter returns an error, `Data` is omitted because it is `nil`. When the presenter returns success, `Data` is populated with a non-nil slice/object (even empty slices have type metadata), ensuring that `data` is included in the JSON payload (e.g., `"data": []`). This maintains a single struct definition while satisfying the specification.

### Decision: Struct field renaming
**Choice**: Rename the struct field `ErrorCode` to `Code` with JSON tag `json:"code,omitempty"`.
**Alternatives considered**: Keeping the struct field name `ErrorCode` and changing only the JSON tag to `code`.
**Rationale**: Renaming the Go struct field to match the JSON key improves code readability and eliminates potential developer confusion.

### Decision: Router NotFound interceptor
**Choice**: Wire a custom route interceptor using Chi's `r.NotFound` handler.
**Alternatives considered**: Custom middleware to inspect route registration.
**Rationale**: Utilizing Chi's native `r.NotFound` Hook ensures the route lookup is performed natively by Chi's trie router, running only when no endpoint matches.

## Data Flow
```
Client Request (e.g., GET /non-existent)
     │
     ▼
[Router Middleware Chain]
 ├── Recovery
 ├── Logging
 ├── RateLimiter
 └── Auth (Validates X-API-Key header)
     │
     ▼ (Passes Auth)
[Chi Route Matcher] ──(No Match)──► [Custom NotFound Handler]
                                            │
                                            ▼
                                    [presenter.Error()]
                                            │
                                            ▼
                                    [Write JSON 404 Response]
```

## File Changes
| File | Action | Description |
|------|--------|-------------|
| [presenter.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/presenter/presenter.go) | Modify | Rename `ErrorCode` to `Code` and update JSON tags to `data,omitempty`, `code,omitempty`, and `error,omitempty`. Update `OK`, `Created`, and `Error` to match. |
| [router.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/http/router/router.go) | Modify | Import `apierrors` and `presenter`, and configure `r.NotFound(...)`. |
| [presenter_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/presenter/presenter_test.go) | Modify | Update assertions to check `Code` / `code` instead of `ErrorCode` / `error_code`. |
| [rate_handlers_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/handler/rate_handlers_test.go) | Modify | Update `verificationEnvelope` layout, verify `code` field, and assert absence of `"data"` key on errors. |
| [auth_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/auth_test.go) | Modify | Update package-level `responseEnvelope` definition and assertions to test for `code`. |
| [integration_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/integration_test.go) | Modify | Update response checks to use `Code` instead of `ErrorCode`. |
| [ratelimit_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/ratelimit_test.go) | Modify | Update response checks to use `Code` instead of `ErrorCode`. |
| [router_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/http/router/router_test.go) | Modify | Add `TestRouter_NotFound` to request an invalid route (with a valid API key) and assert HTTP 404 and structured JSON error response. |

## Interfaces / Contracts
`ResponseEnvelope` JSON contract:
```go
type ResponseEnvelope struct {
	Success   bool             `json:"success"`
	Data      any              `json:"data,omitempty"`
	Code      *apierrors.Code  `json:"code,omitempty"`
	Error     *string          `json:"error,omitempty"`
	RequestID string           `json:"request_id"`
}
```

## Testing Strategy
| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit (Presenter) | Envelope mapping | Use `httptest.NewRecorder` in `presenter_test.go` to verify helpers write correct status codes, `code` tag is mapped correctly, and `data` is omitted on error. |
| Integration (Router) | Route Not Found fallback | In `router_test.go`, make a request to a non-existent route (`GET /unknown`) using `httptest.NewRecorder`. Pass a valid `X-API-Key` header (required since Auth middleware is global and runs first) and assert status is 404 with a JSON body mapping to `NOT_FOUND` code and omitting `data`. |
| Integration (Middleware & Handlers) | JSON response consistency | Update `rate_handlers_test.go`, `auth_test.go`, `integration_test.go`, and `ratelimit_test.go` to assert `"code"` instead of `"error_code"`. Assert the absence of `"data"` on errors by checking raw `map[string]any` keys. |

## Migration / Rollout
No database migration is required. This is a breaking API change for API clients expecting `error_code` or a `data` field on error responses. Rollout requires updating API client libraries, documentation, and coordinating with consumers.

## Open Questions
- [ ] None
