# Tasks: 01-project-scaffolding

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 350–420 |
| 400-line budget risk | Medium |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 (domain + config) → PR 2 (store + main wiring) |
| Delivery strategy | ask-on-risk |
| Chain strategy | pending |

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: pending
400-line budget risk: Medium

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Domain package + config package | PR 1 | Independent foundation — no external deps beyond godotenv |
| 2 | Store package + main.go wiring + tests | PR 2 | Depends on PR 1; includes SQLite and integration |

## Phase 1: Domain Package

- [x] 1.1 Create `internal/domain/rate.go` — Rate struct with JSON tags, SQLite struct tags, `ID/Currency/RateType/Bank/Value/ScrapedAt` fields
- [x] 1.2 Create `internal/domain/errors.go` — Sentinel errors: `ErrNotFound`, `ErrDuplicateRate`, `ErrInvalidInput`, `ErrDatabase`
- [x] 1.3 Create `internal/domain/repository.go` — Repository interface: `SaveRates`, `GetLatestRates`, `GetHistoryRates`

## Phase 2: Config Package

- [x] 2.1 Create `internal/config/config.go` — Config struct (`APIKey`, `Port`, `DBPath`, `ScrapeHour`, `RateLimit`) with defaults
- [x] 2.2 Implement `Load()` function — godotenv .env loading, env var override, validation (fail fast on missing API_KEY)

## Phase 3: SQLite Store

- [x] 3.1 Create `internal/store/sqlite.go` — `New(dbPath)` opens SQLite via modernc.org/sqlite, runs idempotent migration
- [x] 3.2 Schema migration — `CREATE TABLE IF NOT EXISTS rates`, UNIQUE constraint, indexes on `(currency, scraped_at DESC)` and `(scraped_at DESC)`

## Phase 4: Tests

- [x] 4.1 Create `internal/store/sqlite_test.go` — Test in-memory DB creation, migration idempotency, unique constraint enforcement

## Phase 5: Wiring & Dependencies

- [x] 5.1 Modify `cmd/api/main.go` — Load config, initialize DB via `store.New()`, wire dependencies
- [x] 5.2 Update `go.mod` — Add `joho/godotenv`, `modernc.org/sqlite`, run `go mod tidy`

## Implementation Order

Domain first (no dependencies), then config (depends on godotenv only), then store (depends on domain + sqlite driver), then tests (depends on store), then main.go wiring (depends on everything). Phase 1 is fully independent and can be PR'd alone.

## Dependencies Between Tasks

| Task | Depends On |
|------|-----------|
| 1.1–1.3 | None |
| 2.1–2.2 | None (independent of domain) |
| 3.1–3.2 | 1.1 (Rate struct for migration target) |
| 4.1 | 3.1–3.2 |
| 5.1 | 1.1–1.3, 2.1–2.2, 3.1–3.2 |
| 5.2 | All previous tasks |
