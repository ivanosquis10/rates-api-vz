# Design: 01-project-scaffolding

## Technical Approach

Establish foundational Go packages for domain model, configuration, and SQLite storage. This creates the contract layer that downstream changes (scraper, API, scheduler) will depend on. The design follows clean architecture principles with domain-first design, keeping interfaces in the domain package and implementations in dedicated packages.

## Architecture Decisions

### Decision: Package Structure

**Choice**: `internal/domain/`, `internal/config/`, `internal/store/` with `cmd/api/main.go` entry point  
**Alternatives considered**: Flat package structure, domain package at root  
**Rationale**: Standard Go project layout with `internal/` for private packages. Domain package contains only interfaces and types (no dependencies), config handles env/file loading, store handles database implementation. This separation enables testing each layer independently.

### Decision: SQLite Driver

**Choice**: `modernc.org/sqlite` (pure Go, no CGO)  
**Alternatives considered**: `github.com/mattn/go-sqlite3` (CGO), `crawshaw.io/sqlite`  
**Rationale**: Pure Go eliminates CGO compilation issues and cross-compilation problems. No external C dependencies simplifies deployment and CI. Performance is acceptable for this use case (single-server API).

### Decision: Config Loading Strategy

**Choice**: `joho/godotenv` with env var override  
**Alternatives considered**: Viper, manual os.Getenv, `caarlos0/env`  
**Rationale**: `.env` file loading with env var override is the simplest approach that meets requirements. Viper is overkill for this use case. The pattern: load `.env` first, then env vars override (os.Getenv returns "" if not set, so we need conditional logic).

### Decision: Schema Migration Approach

**Choice**: Idempotent `CREATE TABLE IF NOT EXISTS` with manual SQL  
**Alternatives considered**: Migration libraries (golang-migrate, atlas), goose  
**Rationale**: For initial scaffolding with single table, migration libraries are overkill. Idempotent SQL ensures safe repeated execution. Future changes can adopt a migration library when complexity grows.

### Decision: Repository Interface Location

**Choice**: Domain package (`internal/domain/repository.go`)  
**Alternatives considered**: Store package, separate ports package  
**Rationale**: Interface in domain package keeps the dependency rule clean: domain defines contracts, store implements them. This follows hexagonal architecture principles where the domain is the center with no outward dependencies.

## Data Flow

```
cmd/api/main.go
    │
    ├─→ internal/config.Load()
    │       └─→ .env file + env vars → Config struct
    │
    ├─→ internal/store.New(config.DBPath)
    │       └─→ SQLite connection → schema migration → *sql.DB
    │
    └─→ (future: repository implementation)
            └─→ *sql.DB + domain.Repository interface
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/domain/rate.go` | Create | Rate struct with JSON/SQLite tags, domain error sentinels |
| `internal/domain/errors.go` | Create | Exported error types: ErrNotFound, ErrDuplicateRate, ErrInvalidInput, ErrDatabase |
| `internal/domain/repository.go` | Create | Repository interface with SaveRates, GetLatestRates, GetHistoryRates |
| `internal/config/config.go` | Create | Config struct, Load() function, validation, defaults |
| `internal/store/sqlite.go` | Create | New() function, SQLite connection, schema migration |
| `internal/store/sqlite_test.go` | Create | Unit tests with in-memory SQLite |
| `cmd/api/main.go` | Modify | Wire config loading and DB initialization |
| `go.mod` | Modify | Add dependencies: joho/godotenv, modernc.org/sqlite |

## Interfaces / Contracts

### Domain Errors (internal/domain/errors.go)

```go
package domain

import "errors"

var (
    ErrNotFound      = errors.New("rate not found")
    ErrDuplicateRate = errors.New("duplicate rate")
    ErrInvalidInput  = errors.New("invalid input")
    ErrDatabase      = errors.New("database error")
)
```

### Rate Entity (internal/domain/rate.go)

```go
package domain

import "time"

type Rate struct {
    ID        int64     `json:"id"         sqlite:"id"`
    Currency  string    `json:"currency"   sqlite:"currency"`
    RateType  string    `json:"rate_type"  sqlite:"rate_type"`
    Bank      string    `json:"bank"       sqlite:"bank"`
    Value     float64   `json:"value"      sqlite:"value"`
    ScrapedAt time.Time `json:"scraped_at" sqlite:"scraped_at"`
}
```

### Repository Interface (internal/domain/repository.go)

```go
package domain

import "context"

type Repository interface {
    SaveRates(ctx context.Context, rates []Rate) error
    GetLatestRates(ctx context.Context, currency string) ([]Rate, error)
    GetHistoryRates(ctx context.Context, currency, rateType, from, to string, limit int) ([]Rate, error)
}
```

### Config Struct (internal/config/config.go)

```go
package config

type Config struct {
    APIKey      string // Required
    Port        int    // Default: 8080
    DBPath      string // Default: ./rates.db
    ScrapeHour  int    // Default: 8
    RateLimit   int    // Default: 60
}
```

### SQLite Schema

```sql
CREATE TABLE IF NOT EXISTS rates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    currency TEXT NOT NULL,
    rate_type TEXT NOT NULL,
    bank TEXT,
    value REAL NOT NULL,
    scraped_at DATETIME NOT NULL,
    UNIQUE(currency, rate_type, bank, scraped_at)
);

CREATE INDEX IF NOT EXISTS idx_rates_currency_scraped ON rates(currency, scraped_at DESC);
CREATE INDEX IF NOT EXISTS idx_rates_scraped ON rates(scraped_at DESC);
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Config loading, validation, defaults | Test with explicit env vars (no .env file), verify error on missing API_KEY |
| Unit | Schema migration | Use `:memory:` SQLite, verify table creation, idempotency, unique constraint |
| Unit | Domain errors | Verify errors.Is works with sentinel errors |
| Integration | Repository interface compliance | Mock implementation in store package, verify interface satisfaction |

## Migration / Rollout

No migration required. This is initial scaffolding with no existing data. Schema creation is idempotent and safe for repeated execution.

## Open Questions

- [ ] Should Rate struct have `ID` field (auto-increment) or rely on composite unique constraint?
- [ ] Should `SaveRates` batch insert or individual inserts with conflict handling?
- [ ] Should config validation happen at load time or defer to first use?
