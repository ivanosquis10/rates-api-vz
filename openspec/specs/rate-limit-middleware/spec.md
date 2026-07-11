# Rate Limit Middleware Specification

## Purpose

Protect the API from abuse by limiting the rate of requests per client IP address.

## Requirements

### Requirement: Token Bucket Rate Limiting

The system SHALL rate limit incoming requests per client IP. The limit rate (requests per minute) MUST be configurable via `RATE_LIMIT` (defaulting to 60). The bucket burst capacity MUST be twice the rate limit. When a client IP exceeds the limit, the system MUST return HTTP 429 Too Many Requests, set the `Retry-After` header (in seconds), and return the standard JSON error payload (code `TOO_MANY_REQUESTS`).

#### Scenario: Request within rate limit succeeds

- GIVEN a client IP that has not exceeded its rate limit
- WHEN the client sends a request
- THEN the server returns HTTP 200 (or other successful status)

#### Scenario: Request exceeding rate limit fails

- GIVEN a client IP has consumed its rate limit and burst capacity
- WHEN the client sends another request
- THEN the server returns HTTP status 429
- AND the response contains a non-empty `Retry-After` header
- AND the JSON response body contains error code "TOO_MANY_REQUESTS"

### Requirement: Inactive Limiter Cleanup

The system SHALL run a background cleanup goroutine that periodically removes rate limiters for client IPs that have been inactive (no requests received) for 5 minutes or more.

#### Scenario: Inactive limiter is cleaned up

- GIVEN a rate limiter for an IP has had no activity for over 5 minutes
- WHEN the background cleanup janitor runs
- THEN the rate limiter for that IP is deleted from memory

## Acceptance Criteria

- [ ] Rate limiting applied per client IP
- [ ] Configurable limit via `RATE_LIMIT` with default of 60 req/min and 2x burst
- [ ] Rejections return HTTP 429, `Retry-After` header, and JSON code `TOO_MANY_REQUESTS`
- [ ] Inactive limiters are cleaned up after 5 minutes of silence
