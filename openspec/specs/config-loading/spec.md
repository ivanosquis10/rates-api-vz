# Config Loading Specification

## Purpose

Load application configuration from environment variables with `.env` file support, validate required variables, apply sensible defaults, and fail fast on startup if required config is missing.

## Requirements

### Requirement: Load Configuration from Environment

The system SHALL load configuration via `joho/godotenv` from a `.env` file (if present), then allow environment variable overrides. The `Config` struct MUST hold: `APIKey` (string, required), `Port` (int, default 8080), `DBPath` (string, default `./rates.db`), `ScrapeHour` (int, default 8), `RateLimit` (int, default 60).

#### Scenario: Load from .env file

- GIVEN a `.env` file with `API_KEY=test-key` and `PORT=9090`
- WHEN `config.Load()` is called
- THEN `Config.APIKey` is "test-key" and `Config.Port` is 9090

#### Scenario: Environment variable overrides .env

- GIVEN a `.env` file with `PORT=9090` and env var `PORT=3000`
- WHEN `config.Load()` is called
- THEN `Config.Port` is 3000 (env wins)

#### Scenario: Missing .env file uses defaults

- GIVEN no `.env` file exists and `API_KEY` env var is set
- WHEN `config.Load()` is called
- THEN optional fields use defaults (Port=8080, DBPath=./rates.db, ScrapeHour=8, RateLimit=60)

### Requirement: Validate Required Variables

The system SHALL return an error if `API_KEY` is empty or missing after loading. Validation MUST happen during `Load()`.

#### Scenario: Missing API_KEY fails fast

- GIVEN no `API_KEY` env var and no `.env` file
- WHEN `config.Load()` is called
- THEN an error is returned with a message indicating API_KEY is required

#### Scenario: Empty API_KEY fails fast

- GIVEN `API_KEY=` (empty string)
- WHEN `config.Load()` is called
- THEN an error is returned

### Requirement: Defaults for Optional Variables

The system SHALL apply defaults for optional variables: `PORT` → 8080, `DB_PATH` → `./rates.db`, `SCRAPE_HOUR` → 8, `RATE_LIMIT` → 60. Defaults MUST be applied only when the variable is not set (not when set to empty string).

#### Scenario: All defaults applied

- GIVEN only `API_KEY=valid-key` is set
- WHEN `config.Load()` is called
- THEN Port=8080, DBPath="./rates.db", ScrapeHour=8, RateLimit=60

#### Scenario: Partial override

- GIVEN `API_KEY=valid-key` and `SCRAPE_HOUR=12`
- WHEN `config.Load()` is called
- THEN ScrapeHour=12 and all other optionals use defaults

## Acceptance Criteria

- [ ] `config.Load()` returns populated Config on valid input
- [ ] Missing API_KEY causes immediate error return
- [ ] Defaults match `.env.example` documentation
- [ ] `go test ./internal/config/...` passes with table-driven tests
- [ ] .env file loading works; env var override works; missing .env is non-fatal
