# Venezuela Rates API — Spec (PRD)

## Problem Statement

Venezuela's official exchange rates are published daily by the Banco Central de Venezuela (BCV) on their website bcv.org.ve. Currently, there is no programmatic way to access these rates — they must be manually checked on the website. This creates friction for internal systems that need up-to-date exchange rate data for USD and EUR, and there is no historical record of rate changes over time.

## Solution

Build a Go API that automatically scrapes the BCV website daily to collect exchange rates (USD and EUR), stores them in a SQLite database, and exposes them through authenticated HTTP endpoints. The API will provide both current rates and historical data, following Clean Architecture principles with unit test coverage.

## User Stories

### Core Scraping

1. As a system operator, I want the API to automatically scrape BCV exchange rates daily at a configurable hour, so that I always have up-to-date rate data without manual intervention.
2. As a system operator, I want the scraper to collect the BCV reference rate (weighted average) for USD and EUR, so that I have the official exchange rate.
3. As a system operator, I want the scraper to collect buy and sell rates from participating banks (Banesco, Mercantil, Exterior, R4), so that I have a complete picture of the market.
4. As a system operator, I want the scraper to handle BCV website downtime gracefully with 3 retry attempts and exponential backoff, so that transient failures don't cause data loss.
5. As a system operator, I want to manually trigger a scraping run via an admin endpoint, so that I can test the scraper or recover from a missed scheduled run.
6. As a system operator, I want the scraping schedule to respect Caracas timezone (UTC-4), so that rates are collected at the correct local time.

### Data Storage

7. As a system, I want to store all scraped rates in a single SQLite table, so that data access is simple and performant.
8. As a system, I want to allow multiple scraping entries per day, so that retry attempts and manual triggers are captured in the audit trail.
9. As a system, I want a unique constraint on (currency, rate_type, bank, scraped_at), so that duplicate entries are prevented.
10. As a system, I want to store both reference rates (no bank) and bank-specific rates (buy/sell per bank) in the same table, so that the schema is simple and queries are straightforward.

### API — Current Rates

11. As an internal service consumer, I want to call `GET /rates` and receive the most recent USD and EUR rates (reference + bank buy/sell), so that I can use current exchange data in my application.
12. As an internal service consumer, I want the response to include the rate type (reference/buy/sell), currency, value, and timestamp, so that I know what each rate represents and when it was collected.
13. As an internal service consumer, I want to filter rates by currency (`?currency=USD`), so that I can get only the rates I need.
14. As an internal service consumer, I want to filter rates by type (`?type=reference`), so that I can get only reference rates or only bank rates.

### API — Historical Rates

15. As an internal service consumer, I want to call `GET /rates/history` and receive historical rate data, so that I can analyze rate trends over time.
16. As an internal service consumer, I want to filter historical rates by currency (`?currency=USD`), so that I can get history for a specific currency.
17. As an internal service consumer, I want to filter historical rates by date range (`?from=2026-07-01&to=2026-07-10`), so that I can get rates for a specific period.
18. As an internal service consumer, I want to filter historical rates by type (`?type=reference`), so that I can see only reference rate history or only bank rate history.
19. As an internal service consumer, I want to limit the number of results (`?limit=30`), so that I can control response size.
20. As an internal service consumer, I want historical rates ordered by date descending (newest first), so that the most recent data appears first.

### API — Admin

21. As a system operator, I want to call `POST /admin/scrape` with a valid API key, so that I can trigger an immediate scraping run.
22. As a system operator, I want the admin endpoint to return immediately with a confirmation, while scraping runs in the background, so that I'm not blocked waiting for the scrape to complete.

### Authentication & Security

23. As a system operator, I want all API endpoints to require an API key via the `X-API-Key` header, so that unauthorized access is prevented.
24. As a system operator, I want API key validation to use constant-time comparison, so that timing attacks are prevented.
25. As a system operator, I want unauthorized requests to receive a 401 response with a clear error message, so that consumers know what went wrong.

### Rate Limiting

26. As a system operator, I want rate limiting per IP address, so that abuse or infinite loops from a single source are prevented.
27. As a system operator, I want the rate limit to be configurable via environment variable, so that I can adjust it without code changes.
28. As a system operator, I want rate-limited requests to receive a 429 response with a `Retry-After` header, so that consumers know when to retry.

### Error Handling

29. As an internal service consumer, I want all error responses to follow a consistent envelope format (`{ "error": { "code": "...", "message": "..." } }`), so that error handling in my client is predictable.
30. As an internal service consumer, I want HTTP status codes to accurately reflect the error type (400, 401, 404, 429, 500), so that I can handle errors programmatically.
31. As a system operator, I want internal errors (stack traces, SQL errors) to never be exposed in API responses, so that no sensitive information leaks.

### Configuration

32. As a system operator, I want configuration to be loaded from a `.env` file in development and environment variables in production, so that local development is easy and production follows 12-factor principles.
33. As a system operator, I want the API to fail fast on startup if required environment variables (like API_KEY) are missing, so that misconfigurations are caught immediately.
34. As a system operator, I want sensible defaults for optional configuration (port, DB path, scrape hour, rate limit), so that the API works out of the box.

### Logging

35. As a system operator, I want structured JSON logs to stdout, so that logs are parseable by container runtimes and log aggregators.
36. As a system operator, I want each scraping run to be logged with currency, rate value, and duration, so that I can monitor scraping performance.
37. As a system operator, I want scraping failures to be logged with error details and attempt number, so that I can diagnose issues.

### Testing

38. As a developer, I want unit tests for business logic (use cases), so that I can verify correctness without external dependencies.
39. As a developer, I want repository tests using SQLite in-memory, so that I can verify SQL queries without touching a real database.
40. As a developer, I want manual mocks (no external mocking libraries), so that the test suite has minimal dependencies.
41. As a developer, I want table-driven tests following Go conventions, so that test scenarios are clear and maintainable.

## Implementation Decisions

### Architecture

- **Clean Architecture** with 4 layers: Domain, Use Cases, Repository (adapter), Handler (HTTP)
- **Manual dependency injection** in `cmd/api/main.go` — no DI frameworks (Wire, Fx)
- **Interfaces in `domain` package**, implementations in sub-packages
- Dependency rule: domain depends on nothing; everything depends on domain

### Scraping

- **Library**: `net/http` + `goquery` — BCV site is Drupal 7 with static server-rendered HTML, no JavaScript required
- **Target URLs**: `https://www.bcv.org.ve` (homepage) and `https://www.bcv.org.ve/estadisticas/tipo-cambio-de-referencia-smc`
- **CSS selectors**: `#dolar .strong-tb` (USD reference), `#euro .strong-tb` (EUR reference), `.views-table tbody tr` (bank rates table), `.date-display-single[content]` (ISO date)
- **Currencies**: USD and EUR only (not CNY, TRY, RUB)
- **Rate types**: BCV reference (weighted average) AND bank buy/sell rates

### Database

- **Engine**: SQLite with `modernc.org/sqlite` (pure Go, no CGO)
- **Schema**: Single `rates` table with nullable `bank` column
- **Unique constraint**: `(currency, rate_type, bank, scraped_at)` prevents duplicates
- **Multiple entries per day**: Supports retry attempts and manual triggers
- **Indexes**: On `(currency, scraped_at DESC)` and `(scraped_at DESC)` for query performance

### API Endpoints

- `GET /rates` — current rates (latest per currency/type)
- `GET /rates/history` — historical rates with filters
- `POST /admin/scrape` — manual trigger (requires auth)

### Authentication

- API Key via `X-API-Key` header
- Constant-time comparison using `subtle.ConstantTimeCompare`
- Key stored in `API_KEY` environment variable

### Rate Limiting

- Token bucket algorithm per IP using `golang.org/x/time/rate`
- Configurable via `RATE_LIMIT` env var (default: 60 req/min)
- 429 response with `Retry-After` header

### Scheduler

- `robfig/cron` library with `America/Caracas` timezone
- Configurable hour via `SCRAPE_HOUR` env var (default: 8)
- 3 retry attempts with exponential backoff (1min, 2min, 4min)
- Single goroutine execution to prevent overlapping scrapes

### Configuration

- `.env` file for local development, environment variables for production
- `.env.example` documents all required variables with descriptions and defaults
- `.env` is gitignored to prevent leaking secrets
- Required: `API_KEY`
- Optional with defaults: `PORT` (8080), `DB_PATH` (./rates.db), `SCRAPE_HOUR` (8), `RATE_LIMIT` (60)
- Fail-fast validation on startup

### Error Handling

- Standard envelope: `{ "error": { "code": "ERROR_CODE", "message": "Human-readable message" } }`
- Domain errors defined in `domain/errors.go`
- Handler maps domain errors to HTTP status codes
- Internal errors logged but never exposed in responses

### Logging

- `log/slog` (Go stdlib, JSON handler)
- Structured fields: currency, rate, duration_ms, attempt, error
- Output to stdout (container-friendly)

## Testing Decisions

### Testing seams (verified with user)

1. **HTTP API endpoints** — test request/response contracts, auth, rate limiting
2. **Scraper** — test HTML parsing with mocked HTTP responses
3. **Repository** — test SQL queries with SQLite in-memory
4. **Scheduler** — test retry logic and scheduling

### Testing approach

- **Manual mocks** — no external mocking libraries (mockery, gomock, testify)
- **Table-driven tests** — standard Go pattern for all test functions
- **SQLite in-memory** — `:memory:` database for repository tests
- **`httptest`** — for handler and scraper tests with mocked HTTP
- **`_test.go` alongside source** — not separate test directory
- **No testify** — stdlib `testing` + `reflect.DeepEqual` + simple assertion helpers

### Coverage targets

- Use cases: 90%+ (core business logic)
- Repository: 80%+ (SQL correctness)
- Handler: 80%+ (HTTP contracts)
- Scraper: 70%+ (HTML parsing)

## Out of Scope

- Multi-bank support (only BCV, not other Venezuelan banks)
- Frontend or dashboard
- JWT or OAuth authentication (API Key only)
- CORS configuration (backend-to-backend only)
- CNY, TRY, RUB currencies (USD and EUR only)
- Database migrations tooling (manual schema creation)
- Docker/containerization (can be added later)
- CI/CD pipeline (can be added later)
- Performance benchmarking or load testing

## Further Notes

- The BCV website is a Drupal 7 site that may change its HTML structure. The scraper should be designed with isolated CSS selectors for easy maintenance.
- The API is for internal use only — no public-facing consumers.
- All timestamps should be in Caracas timezone (UTC-4) for consistency with BCV publications.
- The project module path is `github.com/ivanosquis10/api-rates-venezuela`.
