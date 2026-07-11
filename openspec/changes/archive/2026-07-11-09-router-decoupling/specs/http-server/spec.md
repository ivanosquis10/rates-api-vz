# Delta for http-server

## MODIFIED Requirements

### Requirement: Chi Router Setup

The system SHALL use `github.com/go-chi/chi/v5` as the HTTP router. The router MUST be configured in the dedicated `internal/http/router` package and instantiated inside `cmd/api/main.go` via dependency injection. It MUST serve on the configured port.

(Previously: The system SHALL use `github.com/go-chi/chi/v5` as the HTTP router. The router MUST be configured in `cmd/api/main.go` and serve on the configured port.)

#### Scenario: Router compiles and starts

- GIVEN the application binary is built with `go build ./cmd/api`
- WHEN the binary starts
- THEN the Chi router begins listening on the configured port
- AND the process exits cleanly on interrupt
