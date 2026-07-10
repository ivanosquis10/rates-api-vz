# HTTP Server Specification

## Purpose

Chi-based HTTP router with middleware stack that serves the Venezuela Rates API. Replaces the placeholder `http.NewServeMux` with a production-grade routing layer.

## Requirements

### Requirement: Chi Router Setup

The system SHALL use `github.com/go-chi/chi/v5` as the HTTP router. The router MUST be configured in `cmd/api/main.go` and serve on the configured port.

#### Scenario: Router compiles and starts

- GIVEN the application binary is built with `go build ./cmd/api`
- WHEN the binary starts
- THEN the Chi router begins listening on the configured port
- AND the process exits cleanly on interrupt

### Requirement: Middleware Stack

The system SHALL apply two middleware functions to all routes: request logging and panic recovery. Middleware MUST be applied in order: recovery first, then logging.

#### Scenario: Request is logged

- GIVEN a GET /rates request is received
- WHEN the request passes through middleware
- THEN the logging middleware records method, path, and status code via `slog`

#### Scenario: Panic does not crash the server

- GIVEN a handler panics during request processing
- WHEN the recovery middleware catches the panic
- THEN the server returns HTTP 500 and continues serving
- AND the panic is logged via `slog`

### Requirement: Route Registration

The system SHALL register three route groups: `GET /rates`, `GET /rates/history`, and `POST /admin/scrape`. The `/admin` group MUST be isolated via `chi.Route("/admin", ...)`.

#### Scenario: All routes are reachable

- GIVEN the server is running
- WHEN GET /rates is requested
- THEN the router dispatches to the GetRates handler

- WHEN GET /rates/history is requested
- THEN the router dispatches to the GetHistory handler

- WHEN POST /admin/scrape is requested
- THEN the router dispatches to the TriggerScrape handler

### Requirement: Dependency Wiring

The system SHALL inject `*usecase.RateUsecase` into the handler via constructor. The `cmd/api/main.go` MUST wire: config → store → usecase → handler → router.

#### Scenario: Handler receives usecase

- GIVEN a `RateUsecase` with valid repo and scraper
- WHEN `NewHandler(uc)` is called
- THEN a non-nil Handler is returned
- AND subsequent handler calls use the injected usecase

## Acceptance Criteria

- [ ] `go build ./cmd/api` succeeds
- [ ] Chi router replaces `http.NewServeMux`
- [ ] Middleware applies to all routes
- [ ] Panic recovery returns 500, does not crash
- [ ] Issue #6 routing requirements satisfied
