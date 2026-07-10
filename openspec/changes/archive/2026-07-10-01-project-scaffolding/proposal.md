# Proposal: 01 — Project Scaffolding: Domain, Config & SQLite Setup

## Intent

Establish the foundational layer of the Venezuela Rates API. Currently the project has only a bare `main.go` with no packages, no dependencies, and no database. This change creates the domain model, config loading, SQLite connection with schema, and repository interface — everything downstream (scraper, API, scheduler) depends on these pieces existing first.

## Scope

### In Scope
- **Domain package** (`internal/domain/`): `Rate` entity (currency, rate_type, bank, value, scraped_at) and domain error types (`ErrNotFound`, `ErrDuplicateRate`, `ErrInvalidInput`, `ErrDatabase`)
- **Config package** (`internal/config/`): Load from `.env` via `joho/godotenv`, env var override, validate required vars (`API_KEY`), fail fast on startup, sensible defaults for optional vars
- **SQLite setup** (`internal/store/`): `modernc.org/sqlite` (pure Go, no CGO), schema migration creating `rates` table with unique constraint `(currency, rate_type, bank, scraped_at)`, indexes on `(currency, scraped_at DESC)` and `(scraped_at DESC)`
- **Repository interface** in domain package: `SaveRates`, `GetLatestRates`, `GetHistoryRates`
- **Unit tests**: config validation tests, schema creation tests
- **go.mod updates**: add all required dependencies, `go mod tidy`

### Out of Scope
- Repository implementation (separate change)
- Scraper, API handlers, scheduler (downstream changes)
- Rate limiting, auth middleware, cron scheduling

## Capabilities

### New Capabilities
- `domain-model`: Rate entity, domain errors, repository interface
- `config-loading`: .env-based config with validation and defaults
- `sqlite-store`: Database connection, schema migration, repository interface contract

### Modified Capabilities
None — this is the initial scaffolding.

## Approach

1. Create `internal/domain/rate.go` — `Rate` struct with JSON/SDB tags, domain error types
2. Create `internal/domain/errors.go` — sentinel errors for the domain
3. Create `internal/domain/repository.go` — `Repository` interface with three methods
4. Create `internal/config/config.go` — `Load()` function using godotenv, validation, `Config` struct with defaults
5. Create `internal/store/sqlite.go` — `New()` opens DB, runs `CREATE TABLE IF NOT EXISTS` migration, returns `*sql.DB`
6. Create `internal/store/sqlite_test.go` — test schema creation and config validation
7. Update `go.mod` with dependencies: `joho/godotenv`, `modernc.org/sqlite`
8. Update `cmd/api/main.go` to load config and initialize DB on startup
9. Run `go mod tidy` to resolve all dependencies

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/domain/` | New | Rate entity, errors, repository interface |
| `internal/config/` | New | Config loading and validation |
| `internal/store/` | New | SQLite connection and schema migration |
| `cmd/api/main.go` | Modified | Wire up config + DB init |
| `go.mod` | Modified | Add dependencies |
| `.env.example` | No change | Already documents all required vars |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| `modernc.org/sqlite` schema differences from CGO sqlite | Low | Use standard SQL only; test with in-memory DB |
| `godotenv` not loading `.env` in test environments | Low | Config tests use explicit env var overrides, not .env files |
| Schema migration fails on existing DB | Low | `IF NOT EXISTS` is idempotent; no destructive migrations |

## Rollback Plan

Delete the new packages (`internal/domain/`, `internal/config/`, `internal/store/`), revert `cmd/api/main.go` to the bare skeleton, run `go mod tidy`. No data to lose since no production DB exists yet.

## Dependencies

- `joho/godotenv` — .env file loading
- `modernc.org/sqlite` — pure Go SQLite driver
- `github.com/mattn/go-sqlite3` — NOT used (CGO dependency avoided)

## Success Criteria

- [ ] `go build ./...` succeeds with zero errors
- [ ] `go test ./internal/config/...` passes — config loads from env, validates API_KEY, fails on missing required var
- [ ] `go test ./internal/store/...` passes — schema creates rates table, unique constraint enforced
- [ ] `go mod tidy` exits cleanly with all dependencies resolved
- [ ] `Rate` struct has fields: currency, rate_type, bank (nullable), value, scraped_at
- [ ] Domain errors are exported sentinel values
- [ ] Repository interface defines SaveRates, GetLatestRates, GetHistoryRates
