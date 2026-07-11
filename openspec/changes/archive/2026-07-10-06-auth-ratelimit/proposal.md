# Proposal: Auth & Rate Limiter Middleware

## Intent
Implement authentication and rate limiting middleware to secure and protect the API from abuse.

## Scope
### In Scope
- Auth middleware validating `X-API-Key` against the `API_KEY` environment variable.
- Token bucket rate limiting middleware per client IP using `golang.org/x/time/rate`.
- Configurable rate via `RATE_LIMIT` (default 60 req/min) and burst (2x).
- Background cleanup worker for inactive IP limiters.
- Pipeline order: Rate limit -> Auth -> Handler.
- Unit and integration tests for both middlewares.

### Out of Scope
- Distributed rate limiting (e.g., Redis).
- User/API key specific rate limit tiers.
- Other authentication methods (JWT, OAuth).

## Capabilities
### New Capabilities
- `auth-middleware`: Secures routes using the `X-API-Key` header with constant-time comparison.
- `rate-limit-middleware`: Rate limits API consumers per client IP.
### Modified Capabilities
- `http-server`: The middleware chain integrates rate limiting and auth before endpoints are executed.

## Approach
1. Implement auth middleware in [auth.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/auth.go) using `subtle.ConstantTimeCompare`. Return `401 Unauthorized` with JSON error payload.
2. Implement rate limiting in [ratelimit.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/ratelimit.go) using a `sync.Map` of client IP to `*rate.Limiter`. Start a background janitor goroutine that periodically cleans up limiters older than 5 minutes. Return `429 Too Many Requests` with a `Retry-After` header and JSON error payload.
3. Wire both middlewares in the Chi router inside [main.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/cmd/api/main.go).

## Affected Areas
| Area | Impact | Description |
|------|--------|-------------|
| [auth.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/auth.go) | New | Authentication middleware checking `X-API-Key` |
| [ratelimit.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/ratelimit.go) | New | IP-based rate limiting middleware |
| [main.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/cmd/api/main.go) | Modified | Apply middleware to router stack |

## Risks
| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Memory leaks from tracking client IPs | Medium | Implement background cleaner that deletes inactive IPs (no requests for 5+ min) |
| Incorrect client IP detection | Low | Utilize `chimw.RealIP` to resolve proxies |
| Performance overhead from mutex locks | Low | Protect map with `sync.RWMutex` or use `sync.Map` |

## Rollback Plan
Revert changes in [main.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/cmd/api/main.go) to remove the middlewares, and delete [auth.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/auth.go) and [ratelimit.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/middleware/ratelimit.go).

## Dependencies
- `golang.org/x/time/rate`

## Success Criteria
- [ ] Valid requests pass to handlers.
- [ ] Requests without valid key fail with 401 and JSON error body.
- [ ] Requests exceeding rate limit fail with 429, standard JSON error, and `Retry-After` header.
- [ ] Inactive IPs are purged by the janitor goroutine.
