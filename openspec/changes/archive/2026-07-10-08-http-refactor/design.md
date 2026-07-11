# Design: 08-http-refactor

## Technical Approach
We will standardize HTTP presentation, request ID extraction, error representation, and routing response structures using three new packages (`internal/apierrors`, `internal/httpx`, and `internal/presenter`). This ensures all API endpoints and custom middleware return a standard response envelope, preventing internal implementation details (e.g. SQLite error details) from leaking to client-facing APIs.

```
[cmd/api/main.go]
      │
      ├──► [internal/middleware] ──► [internal/presenter] ──► [internal/httpx]
      │          │                                                │
      │          ▼                                                │
      └───────────────────────────────────────────────────────────┴► [internal/apierrors]
```

## Architecture Decisions

### Decision: Decouple Middleware and Presenter Errors via apierrors package
**Choice**: Define middleware sentinel errors in `internal/apierrors` so that both `internal/middleware` and `internal/presenter` can refer to them.
**Alternatives considered**:
1. *Define errors inside `internal/middleware`*: Creates a circular import since `internal/middleware` needs to call `presenter.Error()` to render error responses, and `internal/presenter` would need to import `internal/middleware` to check errors with `errors.Is`.
2. *Interface-based reflection / duck-typing*: Makes error mapping less explicit and harder to track.
**Rationale**: By placing shared HTTP-level middleware errors in `internal/apierrors`, both `middleware` and `presenter` can access them without circular references.

### Decision: Pointers for Optional JSON Fields in Envelope Struct
**Choice**: Use pointers for `error_code` and `error` in the `ResponseEnvelope` struct.
**Alternatives considered**: Using plain string types.
**Rationale**: Pointers serialize to `null` in JSON when unset (i.e. on successful responses), which explicitly fulfills client requirements for empty error values rather than sending empty strings (`""`).

## Data Flow
1. Incoming HTTP Request -> Chi router extracts/injects `RequestID` via `chimw.RequestID`.
2. Request passes through `middleware.RateLimit` and `middleware.Auth`.
3. If middleware rejects request:
   - Call `presenter.Error(w, r, apierrors.ErrRateLimitExceeded)` or `presenter.Error(w, r, apierrors.ErrUnauthorized)`.
   - `presenter.Error` maps the error to status code/API error code and writes the standard envelope using `httpx.WriteJSON`.
4. Router routes to `handler.GetRates`, `handler.GetHistory`, or `handler.TriggerScrape`.
5. Handler calls `RateUsecase`.
   - On success: Handler calls `presenter.OK` or `presenter.Created` which packages data into the standard success envelope and renders via `httpx.WriteJSON`.
   - On error: Handler calls `presenter.Error` which converts domain error to safe client API errors and writes the standard error envelope.

## File Changes
| File | Action | Description |
|------|--------|-------------|
| [internal/apierrors/apierrors.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/apierrors/apierrors.go) | Create | Custom `apierrors.Error` struct, `apierrors.Code` type, standard error code constants, and middleware errors `ErrUnauthorized`/`ErrRateLimitExceeded`. |
| [internal/httpx/httpx.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/httpx/httpx.go) | Create | Helper function `WriteJSON` and `GetRequestID` context wrapper around `chi/middleware.GetReqID`. |
| [internal/presenter/presenter.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/presenter/presenter.go) | Create | Response envelope struct and helpers `OK`, `Created`, `NoContent`, and `Error` implementing centralized error mapping. |
| [internal/middleware/auth.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/auth.go) | Modify | Refactor to use `presenter.Error(w, r, apierrors.ErrUnauthorized)` instead of writing inline JSON. |
| [internal/middleware/ratelimit.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/ratelimit.go) | Modify | Refactor to use `presenter.Error(w, r, apierrors.ErrRateLimitExceeded)` and set retry headers properly. |
| [internal/handler/rate_handlers.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/handler/rate_handlers.go) | Modify | Refactor handlers (`GetRates`, `GetHistory`, `TriggerScrape`) to return enveloped responses using presenter helpers. |
| [internal/handler/handler.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/handler/handler.go) | Modify | Remove references/imports to old responses helpers. |
| [internal/handler/responses.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/handler/responses.go) | Delete | Remove deprecated responses file. |
| [internal/handler/responses_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/handler/responses_test.go) | Delete | Remove deprecated responses test file. |
| [internal/handler/rate_handlers_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/handler/rate_handlers_test.go) | Modify | Update and rewrite tests to assert against the new JSON envelope structure (success/data/error/error_code/request_id). |

## Interfaces / Contracts

### apierrors Contract
```go
package apierrors

type Code string

const (
	UNAUTHORIZED   Code = "UNAUTHORIZED"
	RATE_LIMITED   Code = "RATE_LIMITED"
	NOT_FOUND      Code = "NOT_FOUND"
	BAD_REQUEST    Code = "BAD_REQUEST"
	INTERNAL_ERROR Code = "INTERNAL_ERROR"
)

type Error struct {
	Code    Code
	Message string
}

func (e *Error) Error() string { return e.Message }

func New(code Code, message string) *Error {
	return &Error{Code: code, Message: message}
}

var (
	ErrUnauthorized      = New(UNAUTHORIZED, "unauthorized")
	ErrRateLimitExceeded = New(RATE_LIMITED, "too many requests")
)
```

### Response Envelope Contract
```go
package presenter

type ResponseEnvelope struct {
	Success   bool             `json:"success"`
	Data      interface{}      `json:"data"`
	ErrorCode *apierrors.Code  `json:"error_code"`
	Error     *string          `json:"error"`
	RequestID string           `json:"request_id"`
}
```

## Testing Strategy
| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit / Handler | Envelope parsing, status codes, request ID extraction | Rewrite `rate_handlers_test.go` and use `httptest.ResponseRecorder` to assert envelope structure and fields. |
| Middleware | Auth and Rate limit output envelopes | Assert that HTTP 401 and 429 return correct `success: false` envelopes with corresponding error codes. |

## Migration / Rollout
No database migration required. Simple package drop-in and routing update.

## Open Questions
- None.
