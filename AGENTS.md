# Venezuela Rates API — Agent Guide

## Project snapshot

This repository contains a Go 1.25 REST API that scrapes official USD and EUR
reference rates from the Banco Central de Venezuela (BCV), stores them in
SQLite, and exposes authenticated HTTP endpoints.

Main dependencies:

- Chi for HTTP routing.
- modernc.org/sqlite for pure-Go SQLite (CGO is not required).
- GoQuery for HTML scraping.
- robfig/cron for scheduled scraping.
- log/slog for structured JSON logs.

The runtime entrypoint is cmd/api/main.go. It loads configuration, opens the
database, wires dependencies, starts the scheduler, runs an initial scrape, and
starts the HTTP server.

## Source of truth

When documentation and implementation disagree, verify the code in this order:

1. internal/http/router/router.go — mounted routes and middleware order.
2. internal/config/config.go — configuration names and defaults.
3. internal/domain/ — domain entities, errors, and interfaces.
4. internal/store/sqlite.go — database schema and persistence behavior.
5. internal/presenter/presenter.go — HTTP response contract.
6. docs/adr/ and docs/glossary.md — architectural decisions and domain terms.

The README still contains some legacy examples from before the BCV-only API
redesign. Do not copy old /rates, bank-rate, or rate_type examples without
checking the current router and domain model.

## Architecture and boundaries

~~~text
cmd/api
  -> internal/config
  -> internal/store       SQLite adapter and idempotent schema setup
  -> internal/scraper     BCV HTML adapter
  -> internal/usecase     application/business orchestration
  -> internal/handler     HTTP input/output handling
  -> internal/presenter   response mapping and error envelopes
~~~

- internal/domain is the innermost layer and must not depend on adapters or
  HTTP concerns.
- Keep business rules in domain or usecase, not in handlers.
- Define interfaces at the consuming boundary; use small interfaces for
  testability.
- Keep dependency wiring in cmd/api/main.go.
- Keep scraper selectors isolated in internal/scraper/scraper.go; BCV HTML
  changes should normally be localized there and covered by fixture tests.
- The scheduler runs in the America/Caracas timezone, retries failed scrapes,
  and performs an initial background scrape on startup.

## API and runtime conventions

- Current API prefix: /api/v1.
- All routes, including /api/v1/health, require X-API-Key.
- Success responses use { "success": true, "data": ... }.
- Error responses use { "success": false, "code": "...", "error": "..." }.
- Responses expose X-Request-ID; preserve it when changing middleware or
  presenters.
- The router applies CORS, recovery, structured logging, rate limiting, and
  authentication middleware globally. Preserve the order unless there is a
  tested reason to change it.
- Configuration comes from environment variables and optionally .env.
  API_KEY is required. Never commit .env, credentials, or production data.
- SQLite migrations must be idempotent. Use parameterized queries and preserve
  transaction boundaries.

### Current security caveat

The BCV HTTP client currently sets InsecureSkipVerify: true because of the
provider integration. Do not copy this setting to new clients or broaden its
scope; treat removal or isolation of this exception as security-sensitive work.

## Go conventions and good practices

- Run gofmt on changed Go files.
- Prefer clear, small functions and explicit error handling.
- Wrap errors with context using %w; use errors.Is and errors.As for
  classification.
- Propagate context.Context through HTTP, scraping, database, and scheduler
  operations.
- Avoid global mutable state. Use constructors and dependency injection.
- Do not silently change API payloads, routes, environment variables, or the
  SQLite schema. Update tests and documentation when the contract changes.
- Add or update focused tests beside the implementation. Prefer table-driven
  tests for multiple scenarios and deterministic scraper fixtures for HTML.
- Do not add dependencies without checking whether the standard library or an
  existing package is sufficient.

## Local commands

~~~bash
# Install dependencies
go mod download

# Run the API (requires API_KEY)
go run ./cmd/api

# Run tests and race detection
go test ./...
go test -race ./...

# Static checks and build
go vet ./...
go build ./cmd/api

# Format changed Go files
gofmt -w path/to/file.go

# Run with Docker
docker compose up --build
~~~

Before opening a PR, run at least gofmt, go test ./..., go test -race ./...,
go vet ./..., and go build ./cmd/api. CI also runs tests with race detection,
coverage, build, and vet.

## Issue Tracking & GitHub Workflow

- Issue Tracker: GitHub Issues.
- GitHub CLI Integration: when asked to search, review, or view an issue,
  always use the GitHub CLI (gh issue list, gh issue view <id>).
- Conventional Commits: commit messages must follow Conventional Commits,
  such as feat:, fix:, chore:, refactor:, docs:, test:, style:, or perf:.
  Never add AI attribution or Co-Authored-By metadata.
- Git Branching Strategy: always create a dedicated feat/, fix/, or chore/
  branch for new work. Never work or commit directly on main or master.

Useful commands:

~~~bash
gh repo view --json nameWithOwner,url
gh auth status
gh issue list --state open
gh issue view <id>
gh issue create --title "fix: short problem statement" --body-file issue.md
gh pr create --base main --head <branch> --title "feat: short description" --body-file pr.md
gh pr checks --watch
~~~
