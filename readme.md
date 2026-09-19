<div align="center">

# Venezuela Rates API

**Official USD and EUR exchange rates published by the Banco Central de Venezuela.**

[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![CI](https://github.com/ivanosquis10/rates-api-vz/actions/workflows/ci.yml/badge.svg)](https://github.com/ivanosquis10/rates-api-vz/actions/workflows/ci.yml)
[![Issues](https://img.shields.io/github/issues/ivanosquis10/rates-api-vz?style=flat-square)](https://github.com/ivanosquis10/rates-api-vz/issues)

</div>

---

## Overview

Venezuela Rates API is a Go service that:

1. Fetches official USD and EUR reference rates from the BCV website.
2. Stores the rates and their publication timestamps in SQLite.
3. Exposes authenticated HTTP endpoints for current and historical data.
4. Refreshes rates automatically using scheduled jobs in the Caracas timezone.

The service is designed to be small, self-contained, and easy to run with either Go or Docker.

## Features

- Official USD and EUR reference rates from the BCV.
- Automatic startup scrape plus scheduled refreshes.
- Historical queries with date ranges and limits.
- SQLite persistence with idempotent schema initialization.
- API key authentication through the X-API-Key header.
- Per-IP rate limiting.
- CORS, panic recovery, request logging, and X-Request-ID responses.
- JSON structured logs through log/slog.
- Pure-Go SQLite driver; CGO is not required.

## Tech stack

| Area | Technology |
| --- | --- |
| Language | Go 1.25 |
| HTTP router | Chi |
| Database | SQLite with modernc.org/sqlite |
| HTML scraping | GoQuery |
| Scheduler | robfig/cron |
| Rate limiting | golang.org/x/time/rate |
| Logging | log/slog |

## Architecture

The project follows a lightweight Clean Architecture structure:

~~~text
cmd/api/main.go
  Application entrypoint and dependency wiring

internal/domain
  Domain entities, errors, and repository interfaces

internal/usecase
  Application workflows and business rules

internal/store
  SQLite repository and schema initialization

internal/scraper
  BCV HTTP client and HTML parsing

internal/scheduler
  Scheduled scraping, retries, and startup execution

internal/handler
  HTTP handlers and request parsing

internal/presenter
  HTTP response mapping and error envelopes

internal/middleware
  Authentication, CORS, logging, recovery, and rate limiting
~~~

The dependency direction is inward: domain code must not depend on HTTP,
SQLite, scraping, or other infrastructure details. Dependency wiring belongs
in cmd/api/main.go.

## Getting started

### Requirements

- Go 1.25 or newer.
- Docker and Docker Compose are optional.
- A local API key for authenticated requests.

### Install

~~~bash
git clone https://github.com/ivanosquis10/rates-api-vz.git
cd rates-api-vz
go mod download
~~~

### Configure

Copy the example environment file:

~~~bash
cp .env.example .env
~~~

Set API_KEY before starting the service. The .env file is ignored by Git and
must never contain values that are committed to the repository.

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| PORT | No | 8080 | HTTP server port |
| DB_PATH | No | ./rates.db | SQLite database path |
| API_KEY | Yes | — | Key expected in X-API-Key |
| SCRAPE_CRON_MAINTENANCE | No | 0 8 * * * | Daily maintenance scrape |
| SCRAPE_CRON_WINDOW | No | */5 8-18 * * 1-5 | Weekday refresh window |
| RATE_LIMIT | No | 60 | Requests per minute per IP |

Cron schedules use the America/Caracas timezone.

### Run locally

~~~bash
go run ./cmd/api
~~~

For live reload during development, install Air and run:

~~~bash
go install github.com/air-verse/air@latest
air
~~~

### Run with Docker

~~~bash
docker compose up --build
~~~

The Compose configuration persists SQLite data in the rates_data volume and
uses the health endpoint for container checks.

## API

All endpoints require:

~~~http
X-API-Key: your-api-key
~~~

The base path is /api/v1.

### Endpoints

| Method | Path | Description |
| --- | --- | --- |
| GET | /api/v1/health | Service health and API version |
| GET | /api/v1/dollars | Latest USD rate as an array |
| GET | /api/v1/dollars/official | Latest USD rate as an object |
| GET | /api/v1/euros | Latest EUR rate as an array |
| GET | /api/v1/euros/official | Latest EUR rate as an object |
| GET | /api/v1/history/dollars | Historical USD rates |
| GET | /api/v1/history/euros | Historical EUR rates |
| POST | /api/v1/admin/scrape | Trigger an on-demand scrape |

History endpoints accept from, to, and limit query parameters:

~~~http
GET /api/v1/history/dollars?from=2026-07-01&to=2026-07-10&limit=30
~~~

The admin scrape endpoint executes the scrape and returns the number of rates
saved. Scheduled and startup scrapes use the scheduler retry policy.

### Success response

~~~json
{
  "success": true,
  "data": {
    "id": 1,
    "currency": "USD",
    "average": 123.45,
    "updated_at": "2026-07-10T08:00:00Z"
  }
}
~~~

The list endpoints return an array in data. Every response includes an
X-Request-ID header for tracing.

### Error response

~~~json
{
  "success": false,
  "code": "UNAUTHORIZED",
  "error": "invalid or missing API key"
}
~~~

Common error codes include UNAUTHORIZED, BAD_REQUEST, RATE_LIMITED,
NOT_FOUND, PROVIDER_ERROR, and INTERNAL_ERROR.

## Testing and quality

Run the same checks used by CI before opening a pull request:

~~~bash
go test ./...
go test -race ./...
go vet ./...
go build ./cmd/api
~~~

When changing behavior:

- Add or update focused tests next to the implementation.
- Use table-driven tests for multiple scenarios.
- Use deterministic HTML fixtures for scraper behavior.
- Test both successful and failing paths.
- Run gofmt on changed Go files.

## Contributing

Contributions should start from a GitHub Issue describing the problem, scope,
acceptance criteria, and expected verification.

~~~bash
gh issue list --state open
gh issue view <id>
gh issue create --title "feat: short description" --body-file issue.md
~~~

Always create a dedicated branch for new work:

~~~text
feat/123-short-description
fix/123-short-description
chore/123-short-description
~~~

Use Conventional Commits, for example:

~~~text
feat: add EUR history endpoint
fix: handle duplicate scraped rates
test: cover scheduler retry cancellation
docs: improve API documentation
~~~

Never work or commit directly on main or master. Pull requests should link
the issue, explain the change, document verification commands, and mention
any API, configuration, schema, or operational impact.

## Development notes

- Keep BCV selectors isolated in internal/scraper/scraper.go.
- Preserve the response envelope and X-Request-ID contract.
- Keep SQLite migrations idempotent and queries parameterized.
- Do not commit .env, credentials, rates.db, or other runtime data.
- Review docs/adr/ before changing the domain model or API shape.

## License

This project is distributed under the MIT License. See [LICENSE.md](LICENSE.md)
for the full text.
