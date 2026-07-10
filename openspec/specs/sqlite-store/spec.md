# SQLite Store Specification

## Purpose

Provide a SQLite database connection using the pure-Go driver, run idempotent schema migration on startup, and expose a `Repository` implementation that satisfies the domain interface contract.

## Requirements

### Requirement: Database Connection

The system SHALL open a SQLite database using `modernc.org/sqlite` (pure Go, no CGO). The `New(dbPath string) (*sql.DB, error)` function MUST return a usable `*sql.DB` on success.

#### Scenario: Open new database file

- GIVEN a valid file path `./test-rates.db`
- WHEN `store.New("./test-rates.db")` is called
- THEN a non-nil `*sql.DB` is returned with no error
- AND the database file exists on disk

#### Scenario: Open in-memory database

- GIVEN the path `:memory:`
- WHEN `store.New(":memory:")` is called
- THEN a non-nil `*sql.DB` is returned with no error

#### Scenario: Invalid path returns error

- GIVEN a path under a non-existent directory `/no/such/dir/test.db`
- WHEN `store.New(...)` is called
- THEN an error is returned

### Requirement: Schema Migration

The system SHALL create the `rates` table on connection if it does not exist. The schema MUST include columns: `id` (INTEGER PRIMARY KEY AUTOINCREMENT), `currency` (TEXT NOT NULL), `rate_type` (TEXT NOT NULL), `bank` (TEXT, nullable), `value` (REAL NOT NULL), `scraped_at` (DATETIME NOT NULL). A UNIQUE constraint MUST exist on `(currency, rate_type, bank, scraped_at)`.

#### Scenario: First run creates table

- GIVEN no `rates` table exists
- WHEN `store.New(":memory:")` is called
- THEN the `rates` table exists with all required columns

#### Scenario: Idempotent migration

- GIVEN a database with the `rates` table already created
- WHEN `store.New(...)` is called again on the same DB
- THEN no error is returned and the table is unchanged

#### Scenario: Unique constraint prevents duplicates

- GIVEN the `rates` table with a unique constraint on `(currency, rate_type, bank, scraped_at)`
- WHEN inserting two rows with identical values for all four constrained columns
- THEN the second insert fails with a unique constraint violation

### Requirement: Query Indexes

The system SHALL create indexes on `(currency, scraped_at DESC)` and `(scraped_at DESC)` to support efficient current-rate and history queries.

#### Scenario: Index exists for current rates query

- GIVEN a database after migration
- WHEN querying with `WHERE currency = ? ORDER BY scraped_at DESC LIMIT 1`
- THEN the query uses the index (verified by EXPLAIN QUERY PLAN in tests)

### Requirement: Repository Implementation

The system SHALL implement the `domain.Repository` interface. `SaveRates` MUST insert all rates in a transaction. `GetLatestRates` MUST return the most recent rate per `(currency, rate_type)` combination. `GetHistoryRates` MUST support filtering by currency, rate_type, and date range with configurable limit.

#### Scenario: SaveRates persists all rates

- GIVEN 3 Rate values in a slice
- WHEN `repo.SaveRates(ctx, rates)` is called
- THEN 3 rows exist in the `rates` table

#### Scenario: SaveRates persists bank-specific rates

- GIVEN rates with `Bank` set to "Banco de Venezuela" and `Bank` empty (reference)
- WHEN `repo.SaveRates(ctx, rates)` is called
- THEN both rows are persisted with correct `bank` values
- AND reference rates have NULL/empty `bank` in the database

#### Scenario: SaveRates rejects duplicate via public API

- GIVEN a rate with `(currency=USD, rate_type=reference, bank="", scraped_at=T1)` already saved
- WHEN `repo.SaveRates(ctx, [same rate])` is called
- THEN an error is returned
- AND the original row is unchanged

#### Scenario: GetLatestRates returns most recent per type

- GIVEN rates for USD reference at 10:00 and 12:00 on the same day
- WHEN `repo.GetLatestRates(ctx, "USD")` is called
- THEN only the 12:00 rate is returned

#### Scenario: GetLatestRates deduplicates across multiple banks

- GIVEN rates for USD reference from "Banco A" at 10:00 and 12:00, and from "Banco B" at 11:00
- WHEN `repo.GetLatestRates(ctx, "USD")` is called
- THEN only the 12:00 rate from "Banco A" is returned for rate_type=reference
- AND the 11:00 rate from "Banco B" is returned for its combination

#### Scenario: GetHistoryRates respects filters

- GIVEN rates for USD and EUR across multiple days
- WHEN `repo.GetHistoryRates(ctx, "USD", "reference", "2026-07-01", "2026-07-10", 30)` is called
- THEN only USD reference rates within the date range are returned
- AND results are ordered by scraped_at descending
- AND at most 30 results are returned

#### Scenario: GetHistoryRates filters by rateType

- GIVEN USD rates with both "reference" and "parallel" types
- WHEN `repo.GetHistoryRates(ctx, "USD", "reference", "", "", 100)` is called
- THEN only "reference" type rates are returned

#### Scenario: GetHistoryRates filters by date range

- GIVEN USD rates on 2026-07-01, 2026-07-05, and 2026-07-10
- WHEN `repo.GetHistoryRates(ctx, "USD", "", "2026-07-03", "2026-07-08", 100)` is called
- THEN only the 2026-07-05 rate is returned

#### Scenario: GetHistoryRates returns empty slice for no match

- GIVEN no rates match the filters
- WHEN `repo.GetHistoryRates(ctx, ...)` is called
- THEN an empty slice (not nil) is returned with no error

## Acceptance Criteria

- [ ] `go test ./internal/store/...` passes with SQLite in-memory tests
- [ ] Schema migration is idempotent (run twice without error)
- [ ] Unique constraint enforced on `(currency, rate_type, bank, scraped_at)`
- [ ] Repository methods satisfy the domain.Repository interface (compile-time check)
- [ ] Table-driven tests for all repository methods
