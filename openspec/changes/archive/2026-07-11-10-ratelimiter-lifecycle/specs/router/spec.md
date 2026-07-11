# Delta Specification: Router

## Modified Requirements

### Requirement: Router Decoupling and Dependency Injection

The package `internal/http/router` SHALL expose a constructor `New(deps Deps) http.Handler` that instantiates and configures a Chi router. The `Deps` struct MUST contain `Handler *handler.Handler`, `Config *config.Config`, and `RateLimiter *middleware.RateLimiter`. The HTTP server in `cmd/api/main.go` MUST instantiate and run this handler.

(Previously: The package `internal/http/router` SHALL expose a constructor `New(deps Deps) http.Handler` that instantiates and configures a Chi router. The `Deps` struct MUST contain `Handler *handler.Handler`, `Config *config.Config`, and `Context context.Context`. The HTTP server in `cmd/api/main.go` MUST instantiate and run this handler.)

#### Scenario: Instantiation of router
- GIVEN valid dependencies in `Deps`
- WHEN `router.New` is called
- THEN a non-nil `http.Handler` is returned
