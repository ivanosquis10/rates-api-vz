# Auth Middleware Specification

## Purpose

Secure API endpoints by validating the `X-API-Key` header against a configured API key.

## Requirements

### Requirement: API Key Validation

The system SHALL intercept incoming HTTP requests and validate the value of the `X-API-Key` header. It MUST perform a constant-time comparison against the `API_KEY` loaded from environment variables using `subtle.ConstantTimeCompare`. If the key is valid, the request MUST proceed to the next handler. If the key is missing or invalid, the request MUST be rejected with HTTP 401 Unauthorized and standard error JSON payload (code `UNAUTHORIZED`).

#### Scenario: Valid API key allows request

- GIVEN a request contains header `X-API-Key` matching `API_KEY`
- WHEN the request passes through the authentication middleware
- THEN the middleware forwards the request to the next handler

#### Scenario: Missing API key is rejected

- GIVEN a request has no `X-API-Key` header
- WHEN the request passes through the authentication middleware
- THEN the middleware returns HTTP status 401
- AND the JSON response body contains error code "UNAUTHORIZED"

#### Scenario: Invalid API key is rejected

- GIVEN a request contains header `X-API-Key` with an incorrect key
- WHEN the request passes through the authentication middleware
- THEN the middleware returns HTTP status 401
- AND the JSON response body contains error code "UNAUTHORIZED"

## Acceptance Criteria

- [ ] Valid API key allows route access
- [ ] Missing/invalid keys return HTTP 401 and JSON envelope with code `UNAUTHORIZED`
- [ ] Safe constant-time string comparison is used
