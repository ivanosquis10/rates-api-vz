# Domain Model Specification

## Purpose

Define the core domain types for the Venezuela Rates API: the Rate entity, domain error sentinels, and the repository interface contract that storage and scraper implementations must satisfy.

## Requirements

### Requirement: Rate Entity

The system SHALL define a `Rate` struct in `internal/domain/` with fields: `Currency` (string), `RateType` (string), `Bank` (string, nullable), `Value` (float64), `ScrapedAt` (time.Time). The struct MUST include JSON tags for API serialization and struct tags for SQLite column mapping.

#### Scenario: Create a reference rate

- GIVEN a Rate with Currency="USD", RateType="reference", Bank="", Value=36.5, ScrapedAt=<timestamp>
- WHEN the Rate is serialized to JSON
- THEN the output contains all five fields with correct types
- AND Bank is omitted or null when empty

#### Scenario: Create a bank-specific rate

- GIVEN a Rate with Currency="EUR", RateType="buy", Bank="Banesco", Value=38.2, ScrapedAt=<timestamp>
- WHEN the Rate is serialized to JSON
- THEN Bank is "Banesco" in the output

#### Scenario: Reject invalid currency

- GIVEN a Rate with Currency="CNY" (not USD or EUR)
- WHEN the Rate is validated
- THEN an ErrInvalidInput error is returned

### Requirement: Domain Error Types

The system SHALL define exported sentinel errors in `internal/domain/errors.go`: `ErrNotFound`, `ErrDuplicateRate`, `ErrInvalidInput`, `ErrDatabase`. These MUST be comparable with `errors.Is`.

#### Scenario: Check ErrNotFound

- GIVEN a returned error
- WHEN `errors.Is(err, domain.ErrNotFound)` is called
- THEN the result is true if and only if the error is ErrNotFound

#### Scenario: Check ErrDuplicateRate

- GIVEN a returned error from a duplicate insert
- WHEN `errors.Is(err, domain.ErrDuplicateRate)` is called
- THEN the result is true

### Requirement: Repository Interface

The system SHALL define a `Repository` interface in `internal/domain/repository.go` with methods: `SaveRates(ctx, []Rate) error`, `GetLatestRates(ctx, currency string) ([]Rate, error)`, `GetHistoryRates(ctx, currency, rateType, from, to string, limit int) ([]Rate, error)`.

#### Scenario: SaveRates accepts multiple rates

- GIVEN a slice of 3 Rate values
- WHEN `SaveRates(ctx, rates)` is called
- THEN all 3 rates are persisted without error

#### Scenario: GetLatestRates filters by currency

- GIVEN rates exist for USD and EUR
- WHEN `GetLatestRates(ctx, "USD")` is called
- THEN only USD rates are returned

#### Scenario: GetHistoryRates returns empty slice for no data

- GIVEN no rates exist for the given filters
- WHEN `GetHistoryRates(ctx, ...)` is called
- THEN an empty slice (not nil) is returned with no error

## Acceptance Criteria

- [ ] `Rate` struct compiles with correct JSON/SQL tags (maps to issue #9 — rate entity)
- [ ] Domain errors are exported and comparable with `errors.Is`
- [ ] Repository interface compiles and all three methods have correct signatures
- [ ] `go build ./internal/domain/...` succeeds
