# Design: 02 — SQLite Repository Test Coverage

## Technical Approach

Test-only change. Extend `internal/store/sqlite_test.go` with new test functions that close the coverage gap between existing happy-path tests and the 9 spec scenarios in `openspec/changes/02-sqlite-repository/specs/sqlite-store/spec.md`. No production code changes. All tests use in-memory SQLite (`:memory:`).

## Architecture Decisions

### Decision: Table-driven subtests for filter combinations

**Choice**: Use `t.Run` subtests with a shared test-table struct for `GetHistoryRates` filter combos.
**Alternatives considered**: Separate top-level test per scenario (current pattern).
**Rationale**: The spec defines 4 filter scenarios for `GetHistoryRates` (rateType, date range, combined, empty result). Table-driven keeps related scenarios grouped and reduces boilerplate. Existing top-level tests remain unchanged — this adds new structure only where complexity warrants it.

### Decision: Extract `newTestStore` helper

**Choice**: Add a `newTestStore(t *testing.T) *Store` helper that creates in-memory DB + Store.
**Alternatives considered**: Duplicate `New(":memory:")` + `NewStore(db)` in each test.
**Rationale**: Every existing test repeats the same 6-line setup. A helper eliminates duplication and enforces consistent teardown (`t.Cleanup` instead of `defer`). Follows Go testing convention (`t.Helper()`).

### Decision: Test `SaveRates` duplicate via public API

**Choice**: Test duplicate prevention by calling `SaveRates` twice with the same rate, asserting error.
**Alternatives considered**: Skip — already tested via raw `insertRate`.
**Rationale**: The spec explicitly requires "rejects duplicate via public API" as a distinct scenario. The raw `insertRate` path bypasses the transaction logic in `SaveRates`. The public API test verifies the full code path.

## Data Flow

```
Test Helper Setup
     │
     ▼
newTestStore(t) → *Store (in-memory SQLite, schema migrated)
     │
     ├── SaveRates(ctx, rates) ──→ Insert via transaction ──→ Verify row count
     │
     ├── GetLatestRates(ctx, currency) ──→ GROUP BY + HAVING MAX ──→ Assert latest only
     │
     └── GetHistoryRates(ctx, filters) ──→ Dynamic WHERE clause ──→ Assert filtered results
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/store/sqlite_test.go` | Modify | Add `newTestStore` helper + 7 new test functions (256→~450 lines) |

## Test Structure

### Helper: `newTestStore`

```go
func newTestStore(t *testing.T) *Store {
    t.Helper()
    db, err := New(":memory:")
    if err != nil { t.Fatalf(...) }
    t.Cleanup(func() { db.Close() })
    return NewStore(db)
}
```

### New Test Functions

| Function | Spec Scenario | Approach |
|----------|--------------|----------|
| `TestSaveRatesBankSpecific` | Bank-specific rates persisted | Insert mixed reference + bank rates, query raw rows to verify bank values |
| `TestSaveRatesDuplicateViaAPI` | Duplicate via public API returns error | Save same rate twice, assert second call errors |
| `TestGetLatestRatesMultiTimestamp` | Most recent per (currency, rate_type) | Insert 2 timestamps for same type, assert only latest returned |
| `TestGetLatestRatesMultiBank` | Dedup across banks | Insert rates from 2 banks at different times, verify latest-per-type selection |
| `TestGetHistoryRatesTableDriven` | Filter combos (rateType, date range, combined, empty) | Table-driven subtests covering 4 filter scenarios |
| `TestGetHistoryRatesOrdering` | Results ordered by scraped_at DESC | Insert 3 rates at different times, verify reverse chronological order |
| `TestGetHistoryRatesNilSafety` | Empty slice (not nil) for no match | Query nonexistent currency, assert `len(result) == 0` and `result != nil` |

### Table-Driven Test Structure for `GetHistoryRates`

```go
func TestGetHistoryRatesTableDriven(t *testing.T) {
    // Seed: USD reference + parallel on 3 dates, EUR reference on 1 date
    cases := []struct {
        name      string
        currency  string
        rateType  string
        from      string
        to        string
        limit     int
        wantCount int
    }{
        {"rateType filter", "USD", "reference", "", "", 100, 3},
        {"date range filter", "USD", "", "2026-07-03", "2026-07-08", 100, 1},
        {"combined filters", "USD", "reference", "2026-07-01", "2026-07-10", 30, 3},
        {"no match", "XYZ", "", "", "", 100, 0},
    }
    // ...
}
```

## Migration / Rollout

No migration required. Test-only change. No production code modified.

## Open Questions

None — all design decisions are clear from the spec and existing implementation.
