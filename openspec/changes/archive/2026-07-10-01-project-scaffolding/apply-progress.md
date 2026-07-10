# Apply Progress — 01-project-scaffolding (PR 1 + PR 2)

## Scope

Full project scaffolding: domain, config, SQLite store, tests, wiring.

## Tasks Completed

### PR 1 — Domain + Config
- [x] 1.1 Create `internal/domain/rate.go` — Rate struct with JSON/SQLite tags
- [x] 1.2 Create `internal/domain/errors.go` — Sentinel errors
- [x] 1.3 Create `internal/domain/repository.go` — Repository interface
- [x] 2.1 Create `internal/config/config.go` — Config struct with defaults
- [x] 2.2 Implement `Load()` function — godotenv .env loading, env var override, validation

### PR 2 — Store + Tests + Wiring
- [x] 3.1 Create `internal/store/sqlite.go` — `New(dbPath)` opens SQLite via modernc.org/sqlite, runs idempotent migration
- [x] 3.2 Schema migration — `CREATE TABLE IF NOT EXISTS rates`, UNIQUE constraint, indexes
- [x] 4.1 Create `internal/store/sqlite_test.go` — 11 tests with in-memory SQLite
- [x] 5.1 Modify `cmd/api/main.go` — Wire config + store into HTTP server
- [x] 5.2 Update `go.mod` — Added `joho/godotenv v1.5.1` + `modernc.org/sqlite` (pure Go, no CGO)

## Files Created/Modified

| File | Action | Description |
|------|--------|-------------|
| `internal/domain/rate.go` | Created | Rate struct: ID, Currency, RateType, Bank, Value, ScrapedAt |
| `internal/domain/errors.go` | Created | Sentinel errors: ErrNotFound, ErrDuplicateRate, ErrInvalidInput, ErrDatabase |
| `internal/domain/repository.go` | Created | Repository interface: SaveRates, GetLatestRates, GetHistoryRates |
| `internal/domain/rate_test.go` | Created | 2 tests: JSON serialization, struct fields |
| `internal/domain/errors_test.go` | Created | 4 tests: comparability, distinctness, wrapping |
| `internal/domain/repository_test.go` | Created | 2 tests: interface satisfaction, method signatures |
| `internal/config/config.go` | Created | Config struct, Load() with godotenv, getEnvString, getEnvInt helpers |
| `internal/config/config_test.go` | Created | 8 tests: defaults, required key, override, empty key, partial, .env loading, env override, missing .env |
| `internal/store/sqlite.go` | Created | New(), migrate(), Store, SaveRates, GetLatestRates, GetHistoryRates |
| `internal/store/sqlite_test.go` | Created | 11 tests: in-memory DB, migration, constraints, queries |
| `cmd/api/main.go` | Modified | Config + store wiring, HTTP server |
| `cmd/api/main_test.go` | Created | 1 test: component wiring validation |
| `go.mod` | Modified | Added `joho/godotenv v1.5.1` + `modernc.org/sqlite v1.53.0` + indirect deps |
| `go.sum` | Modified | Updated checksums |

## Test Results

```
ok  github.com/ivanosquis10/api-rates-venezuela/cmd/api         0.213s
ok  github.com/ivanosquis10/api-rates-venezuela/internal/config  0.068s
ok  github.com/ivanosquis10/api-rates-venezuela/internal/domain  (cached)
ok  github.com/ivanosquis10/api-rates-venezuela/internal/store   (cached)
```

- **Total tests**: 29 (1 wiring + 8 config + 8 domain + 11 store + 1 combined)
- **All passing**: ✅
- **go vet clean**: ✅

## TDD Cycle Evidence — godotenv Fix

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|------|-----------|-------|------------|-----|-------|-------------|----------|
| 2.2 (fix) | `config/config_test.go` | Unit | ✅ 5/5 | ✅ Written | ✅ Passed | ✅ 3 cases | ✅ Clean |

### Test Summary
- **Total tests written**: 3 (new for godotenv fix)
- **Total tests passing**: 29
- **Layers used**: Unit (29)
- **Approval tests** (refactoring): None — no refactoring tasks
- **Pure functions created**: 0 (godotenv integration is side-effect based)

## Lines Changed Estimate

- Domain package: ~264 lines (6 files, including tests)
- Config package: ~240 lines (2 files, including tests — 3 new tests added)
- Store package: ~449 lines (2 files, including tests)
- Wiring + deps: ~127 lines (4 files)
- **Total PR 2**: ~611 lines
- **Cumulative total**: ~1,080 lines

## PR 2 Status

- **Ready for commit**: ✅
- **Tests passing**: ✅ `go test ./...` all green
- **Build clean**: ✅ `go build ./...` succeeds
- **go vet clean**: ✅ `go vet ./...` no issues
- **Ready for verify**: ✅

## Deviations from Design

None — implementation matches design exactly.

## Issues Found

- `TestNewInvalidPath` initially failed because SQLite accepts empty string. Fixed by adding explicit empty-path validation in `New()`.
- **godotenv gap**: Config package was missing `.env` file loading. Fixed by adding `godotenv.Load()` call and 3 new tests.

## Remaining Tasks

None. All 10/10 tasks complete. Ready for verify → archive.
