# Apply Progress: 02 — SQLite Repository Test Coverage

## Status: COMPLETE

All 11 tasks implemented and verified. 18/18 tests passing (11 original + 7 new).

## Completed Tasks

### Phase 1: Test Helper Infrastructure
- [x] 1.1 Extract `newTestStore(t *testing.T) *Store` helper — creates in-memory DB + Store, registers `t.Cleanup` for `db.Close()`, uses `t.Helper()`.
- [x] 1.1b Add `createTestRate` helper — returns `domain.Rate` with sensible defaults.

### Phase 2: SaveRates Tests
- [x] 2.1 `TestSaveRatesBankSpecific` — inserts mixed reference (bank="") + bank-specific (bank="Banco de Venezuela") rates via `SaveRates`, queries raw rows to assert both `bank` values persisted correctly.
- [x] 2.2 `TestSaveRatesDuplicateViaAPI` — saves a rate, then calls `SaveRates` again with identical `(currency, rate_type, bank, scraped_at)`. Asserts error returned and original row unchanged.

### Phase 3: GetLatestRates Tests
- [x] 3.1 `TestGetLatestRatesMultiTimestamp` — inserts USD reference at 10:00 and 12:00 same day, calls `GetLatestRates("USD")`, asserts only 12:00 rate returned with correct value (36.50).
- [x] 3.2 `TestGetLatestRatesMultiBank` — inserts rates from "Banco A" at 10:00 and 12:00, and "Banco B" at 11:00 (all USD reference). Asserts GetLatestRates returns latest per (currency, rate_type) — Banco A at 12:00 wins.

### Phase 4: GetHistoryRates Tests
- [x] 4.1 `TestGetHistoryRatesTableDriven` — seeds 3 dates × 2 types × 2 currencies. Table-driven subtests: rateType filter only, date range only, combined filters, no-match. Asserts correct count and filter behavior per case.
- [x] 4.2 `TestGetHistoryRatesOrdering` — inserts 3 USD rates at different timestamps, calls `GetHistoryRates`, verifies results are `scraped_at DESC` with specific value ordering (36.50, 36.25, 36.00).
- [x] 4.3 `TestGetHistoryRatesNilSafety` — queries nonexistent currency, asserts `len(result) == 0` and `result != nil`.

### Phase 5: Verification
- [x] 5.1 `go test ./internal/store/... -v` — all 18 tests pass, no regressions.
- [x] 5.2 `git diff --name-only` shows only `internal/store/sqlite_test.go` — no production code changes.

## TDD Cycle Evidence

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|------|-----------|-------|------------|-----|-------|-------------|----------|
| 1.1 | `sqlite_test.go` | Unit | N/A (new) | ✅ Written | ✅ Passed | ➖ Structural (helper) | ➖ None needed |
| 1.1b | `sqlite_test.go` | Unit | N/A (new) | ✅ Written | ✅ Passed | ➖ Structural (helper) | ➖ None needed |
| 2.1 | `sqlite_test.go` | Integration | ✅ 11/11 | ✅ Written | ✅ Passed | ✅ 2 cases (reference + bank) | ✅ Clean |
| 2.2 | `sqlite_test.go` | Integration | ✅ 11/11 | ✅ Written | ✅ Passed | ✅ 2 cases (error + unchanged) | ✅ Clean |
| 3.1 | `sqlite_test.go` | Integration | ✅ 11/11 | ✅ Written | ✅ Passed | ✅ 2 timestamps | ✅ Clean |
| 3.2 | `sqlite_test.go` | Integration | ✅ 11/11 | ✅ Written | ✅ Passed | ✅ 2 banks | ✅ Clean |
| 4.1 | `sqlite_test.go` | Integration | ✅ 11/11 | ✅ Written | ✅ Passed | ✅ 4 subtests | ✅ Clean |
| 4.2 | `sqlite_test.go` | Integration | ✅ 11/11 | ✅ Written | ✅ Passed | ✅ 3 rates | ✅ Clean |
| 4.3 | `sqlite_test.go` | Integration | ✅ 11/11 | ✅ Written | ✅ Passed | ➖ Single scenario | ✅ Clean |
| 5.1 | — | — | — | — | ✅ All 18 pass | — | — |
| 5.2 | — | — | — | — | ✅ Test-only | — | — |

## Test Summary
- **Total tests written**: 7 new test functions (+ 4 subtests in table-driven)
- **Total tests passing**: 18 (11 original + 7 new)
- **Layers used**: Integration (7 new), Unit (2 helpers)
- **Approval tests** (refactoring): None — no refactoring tasks
- **Pure functions created**: 0 (helpers are not pure but are test-only)

## Discoveries

### SQLite string-based date comparison gotcha
The `GetHistoryRates` date filters use string comparison (`scraped_at >= ?` and `scraped_at <= ?`). Since `time.Time` is stored as `"2026-07-10 00:00:00"` (with space), the comparison `"2026-07-10 00:00:00" <= "2026-07-10"` evaluates to FALSE in SQLite because the stored string is lexicographically greater than the date-only filter string. This means:
- `from="2026-07-01", to="2026-07-10"` returns 2 results (Jul 1 and Jul 5), NOT 3
- `from="2026-07-01", to="2026-07-11"` correctly returns all 3 results

This is a **known limitation** of the current string-based date comparison approach. To properly support inclusive date boundaries, the production code would need to use SQLite date functions or store dates in a format that sorts correctly with string comparison. This is documented but NOT fixed in this test-only change.

## Files Changed

| File | Action | Lines Added | Description |
|------|--------|-------------|-------------|
| `internal/store/sqlite_test.go` | Modified | +200 | Added `newTestStore`, `createTestRate` helpers and 7 new test functions |
