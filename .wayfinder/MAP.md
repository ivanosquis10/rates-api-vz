# Wayfinder Map — Venezuela Rates API

## Destination

A production-ready Go API that scrapes daily exchange rates (USD/EUR) from BCV (Banco Central de Venezuela), stores them in SQLite, and exposes them via authenticated Chi endpoints — following Clean Architecture, with unit tests and rate limiting.

## Notes

- **Stack**: Go 1.25, Chi router, SQLite, Clean Architecture
- **Scraping target**: BCV (bcv.org.ve) — USD and EUR reference rates + bank rates
- **Auth**: API Key via `X-API-Key` header (internal use)
- **CORS**: Disabled (backend-to-backend only)
- **Scheduler**: Daily scraping at configurable time
- **Tests**: Unit tests required
- **Currencies**: USD and EUR only
- **Rates to expose**: Both BCV reference rates AND bank buy/sell rates
- **DI**: Manual injection in main.go
- **Config**: Env vars only (os.Getenv + validation)
- **Logging**: slog (JSON structured, stdlib)
- **Rate limiter**: Per-IP, configurable via env var

## Decisions so far

- [Scraping library research](01-scraping-library-research.md) — `net/http` + `goquery`, Drupal 7 static HTML, selectors mapped
- [SQLite schema design](02-sqlite-schema-design.md) — Single `rates` table, nullable bank, multiple entries/day, unique constraint
- [Clean Architecture structure](03-clean-architecture-structure.md) — domain/usecase/repository/handler layers, manual DI, env vars
- [Scheduler design](04-scheduler-design.md) — robfig/cron, 3 retries + exponential backoff, manual trigger endpoint, America/Caracas timezone
- [Auth + Rate Limiter](05-middleware-auth-ratelimit.md) — X-API-Key header, constant-time compare, token bucket per IP, configurable via RATE_LIMIT env var
- [Config + Error Handling](06-config-errorhandling.md) — Env vars only, standard error envelope, slog JSON logging, domain errors
- [Testing Strategy](07-testing-strategy.md) — Manual mocks, table-driven tests, SQLite in-memory for repo tests, stdlib testing

## Not yet specified

None — all key decisions resolved. Ready for implementation.

## Out of scope

- Multi-bank support (future effort)
- Frontend/dashboard
- Authentication via JWT or OAuth
- CORS for browser access
- CNY, TRY, RUB currencies
