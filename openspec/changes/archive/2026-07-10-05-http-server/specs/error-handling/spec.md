# Error Handling Specification

## Purpose

Standardized JSON error responses and domain-error-to-HTTP-status mapping. Ensures internal errors never leak to clients.

## Requirements

### Requirement: Success Response Envelope

The system SHALL wrap all successful responses in `{ "data": <payload> }`. The `data` field MUST contain the direct payload (slice, object, or scalar).

#### Scenario: Successful rate list

- GIVEN the usecase returns 3 rates
- WHEN the handler responds
- THEN the JSON body is `{ "data": [<rate1>, <rate2>, <rate3>] }`

#### Scenario: Successful empty list

- GIVEN the usecase returns an empty slice
- WHEN the handler responds
- THEN the JSON body is `{ "data": [] }`

### Requirement: Error Response Envelope

The system SHALL wrap all error responses in `{ "error": { "code": "<CODE>", "message": "<message>" } }`. The `code` field MUST be a UPPER_SNAKE_CASE constant.

#### Scenario: Error envelope structure

- GIVEN an error occurs in a handler
- WHEN the error response is serialized
- THEN the JSON contains exactly `error.code` and `error.message` fields

### Requirement: Domain Error to HTTP Status Mapping

The system SHALL map domain errors to HTTP status codes:

| Domain Error | HTTP Status | Code |
|-------------|-------------|------|
| `domain.ErrNotFound` | 404 | NOT_FOUND |
| `domain.ErrInvalidInput` | 400 | BAD_REQUEST |
| `domain.ErrDuplicateRate` | 409 | CONFLICT |
| Any other error | 500 | INTERNAL_ERROR |

#### Scenario: NotFound maps to 404

- GIVEN the usecase returns `domain.ErrNotFound`
- WHEN the error mapper processes it
- THEN HTTP status is 404 and code is "NOT_FOUND"

#### Scenario: InvalidInput maps to 400

- GIVEN the usecase returns `domain.ErrInvalidInput`
- WHEN the error mapper processes it
- THEN HTTP status is 400 and code is "BAD_REQUEST"

#### Scenario: Unknown error maps to 500

- GIVEN the usecase returns a generic `fmt.Errorf("db timeout")`
- WHEN the error mapper processes it
- THEN HTTP status is 500 and code is "INTERNAL_ERROR"

### Requirement: Internal Error Sanitization

The system MUST NOT expose internal error details (SQL queries, stack traces, file paths) in JSON responses. The `message` field for 500 errors MUST be a generic string.

#### Scenario: SQL error is sanitized

- GIVEN the repository returns a SQL syntax error
- WHEN the handler responds
- THEN the response message is "internal server error"
- AND no SQL text appears in the response body

#### Scenario: Panic is sanitized

- GIVEN a handler panics with a detailed error
- WHEN recovery middleware catches it
- THEN the response is HTTP 500 with generic message
- AND no stack trace appears in the response body

## Acceptance Criteria

- [ ] All responses use `{ "data": ... }` or `{ "error": { "code", "message" } }` envelope
- [ ] Domain errors map to correct HTTP status codes
- [ ] 500 responses never contain internal details
- [ ] `go test ./internal/handler/...` passes error scenarios
- [ ] Issue #6 error handling requirements satisfied
