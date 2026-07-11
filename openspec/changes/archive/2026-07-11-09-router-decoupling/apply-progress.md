# Apply Progress: Router Decoupling

All phases of the Router Decoupling task have been successfully implemented, integrated, and verified.

## Progress Details

- **Phase 1: Foundation / Infrastructure**
  - Updated `internal/middleware/ratelimit.go` to import `chimw "github.com/go-chi/chi/v5/middleware"`.
  - Refactored IP resolution in the rate limiting middleware to use `chimw.GetClientIP(r.Context())` first, with fallback to parsing `r.RemoteAddr` via `net.SplitHostPort`.

- **Phase 2: Core Implementation**
  - Recreated `internal/http/router/router.go`.
  - Declared `Deps` struct containing `Handler`, `Config`, and `Context`.
  - Implemented `New(Deps) http.Handler` constructor to instantiate the Chi router.
  - Placed built-in and custom middlewares in the correct order: `ClientIPFromRemoteAddr`, `RequestID`, `Recovery`, `Logging`, `RateLimit`, and `Auth`.
  - Registered all rates endpoints (`/rates`, `/rates/history`, and `/admin/scrape`).

- **Phase 3: Integration / Wiring**
  - Refactored `cmd/api/main.go` to import and use `internal/http/router`.
  - Instantiated the router with `router.New` and injected dependency parameters cleanly.
  - Passed the instantiated handler to the HTTP server.
  - Cleaned up direct Chi middleware references and inline route registration.

- **Phase 4: Testing / Verification**
  - Created `internal/http/router/router_test.go`.
  - Verified non-nil router initialization with a mock `handler.Usecaser` implementation.
  - Added unit/integration tests using `httptest.NewRecorder` for all routes and verified unauthorized requests return `401 Unauthorized`.

- **Phase 5: Cleanup**
  - Executed `go test ./...` and confirmed all tests passed.
  - Verified no unused imports exist.
