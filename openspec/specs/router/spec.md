# Router Specification

## Purpose

Define the decoupled routing layer and middleware configuration, securing client IP resolution to prevent spoofing.

## Requirements

### Requirement: Router Decoupling and Dependency Injection

The package `internal/http/router` SHALL expose a constructor `New(deps Deps) http.Handler` that instantiates and configures a Chi router. The `Deps` struct MUST contain `Handler *handler.Handler`, `Config *config.Config`, and `Context context.Context`. The HTTP server in `cmd/api/main.go` MUST instantiate and run this handler.

#### Scenario: Instantiation of router
- GIVEN valid dependencies in `Deps`
- WHEN `router.New` is called
- THEN a non-nil `http.Handler` is returned

### Requirement: Middleware Stack and Secure IP Resolution

The router SHALL configure built-in and custom middleware. It MUST apply go-chi's `middleware.ClientIPFromRemoteAddr` first to resolve the client IP securely, followed by `middleware.RequestID`, then the custom middleware stack in order: panic recovery, request logging, rate limiting, and authentication.

#### Scenario: Secure client IP resolution
- GIVEN a request from a client with potentially spoofed `X-Forwarded-For` headers
- WHEN the request passes through the router
- THEN the router uses `middleware.ClientIPFromRemoteAddr` to derive the real IP from the connection remote address

#### Scenario: Request middleware execution order
- GIVEN the router is serving request GET /rates
- WHEN a request is processed
- THEN recovery middleware is executed first
- AND logging middleware is executed second
- AND rate limiting middleware is executed third
- AND authentication middleware is executed fourth

### Requirement: Route Delegation

The router SHALL register the following routes: `GET /rates`, `GET /rates/history`, and `POST /admin/scrape` (within `/admin` path group). All requests to these endpoints MUST route to their respective handler methods.

#### Scenario: All routes are dispatched
- GIVEN a running server using the decoupled router
- WHEN a client requests GET /rates
- THEN the router dispatches the request to the rates handler

## Acceptance Criteria

- [ ] `internal/http/router` package exposes `New(Deps) http.Handler`
- [ ] `middleware.RealIP` is replaced by `middleware.ClientIPFromRemoteAddr`
- [ ] Middleware execution order is preserved
