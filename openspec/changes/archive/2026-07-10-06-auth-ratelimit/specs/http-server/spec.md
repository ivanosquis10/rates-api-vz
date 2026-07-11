# Delta for http-server

## MODIFIED Requirements

### Requirement: Middleware Stack

The system SHALL apply four middleware functions to all routes: panic recovery, request logging, rate limiting, and authentication. Middleware MUST be applied in order: recovery first, then logging, then rate limiting, then authentication.

(Previously: The system SHALL apply two middleware functions to all routes: request logging and panic recovery. Middleware MUST be applied in order: recovery first, then logging.)

#### Scenario: Request is logged

- GIVEN a GET /rates request is received
- WHEN the request passes through middleware
- THEN the logging middleware records method, path, and status code via `slog`

#### Scenario: Panic does not crash the server

- GIVEN a handler panics during request processing
- WHEN the recovery middleware catches the panic
- THEN the server returns HTTP 500 and continues serving
- AND the panic is logged via `slog`

#### Scenario: Middleware execution order

- GIVEN a request to GET /rates
- WHEN the request is received by the server
- THEN recovery middleware is executed first
- AND logging middleware is executed second
- AND rate limiting middleware is executed third
- AND authentication middleware is executed fourth
