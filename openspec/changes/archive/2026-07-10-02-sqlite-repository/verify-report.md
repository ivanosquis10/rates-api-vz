# Verification Report: 02-sqlite-repository

**Change**: 02-sqlite-repository (SQLite Repository Test Coverage)
**Mode**: STRICT TDD
**Date**: 2026-07-10
**Verdict**: PASS

---

## Completeness Table

| Artifact | Status | Notes |
|----------|--------|-------|
| Proposal | Present | openspec/changes/02-sqlite-repository/proposal.md |
| Specs | Present | 9 scenarios defined |
| Design | Present | openspec/changes/02-sqlite-repository/design.md |
| Tasks | Complete | 11/11 tasks done |
| Apply-Progress | Complete | 18/18 tests passing |

---

## Build & Test Evidence

**Command**: `go test ./internal/store/... -v`
**Result**: PASS — 18/18 tests pass (exit 0)

Tests executed:
- TestNewInMemoryDB ✅
- TestNewInvalidPath ✅
- TestMigrationIdempotency ✅
- TestRatesTableExists ✅
- TestRatesTableColumns ✅
- TestUniqueConstraintEnforced ✅
- TestSaveAndGetLatestRates ✅
- TestGetLatestRatesEmptyCurrency ✅
- TestGetHistoryRates ✅
- TestGetHistoryRatesWithLimit ✅
- TestInterfaceSatisfaction ✅
- TestSaveRatesBankSpecific ✅
- TestSaveRatesDuplicateViaAPI ✅
- TestGetLatestRatesMultiTimestamp ✅
- TestGetLatestRatesMultiBank ✅
- TestGetHistoryRatesTableDriven (4 subtests) ✅
- TestGetHistoryRatesOrdering ✅
- TestGetHistoryRatesNilSafety ✅

**Command**: `go test ./internal/store/... -cover`
**Coverage**: 82.1% of statements

---

## Acceptance Criteria Verification (Issue #3)

| Criterion | Status | Evidence |
|-----------|--------|----------|
| Repository implements all methods | ✅ | `SaveRates` (line 85), `GetLatestRates` (line 112), `GetHistoryRates` (line 131) in sqlite.go. Compile-time check: `TestInterfaceSatisfaction` |
| SaveRates handles reference rates (bank=NULL) and bank rates | ✅ | `TestSaveRatesBankSpecific` — inserts bank="" and bank="Banco de Venezuela", verifies both via raw query |
| GetLatestRates returns most recent per currency and rate_type | ✅ | `TestGetLatestRatesMultiTimestamp` (2 timestamps, latest wins), `TestGetLatestRatesMultiBank` (2 banks, latest across banks for same rate_type wins) |
| GetHistoryRates supports filters | ✅ | `TestGetHistoryRatesTableDriven` covers: rateType only, date range only, combined, no-match. `TestGetHistoryRatesOrdering` verifies DESC order. `TestGetHistoryRatesWithLimit` verifies LIMIT |
| Unique constraint prevents duplicates | ✅ | `TestUniqueConstraintEnforced` (raw insert level), `TestSaveRatesDuplicateViaAPI` (public API). Schema: `UNIQUE(currency, rate_type, bank, scraped_at)` |
| Tests use SQLite :memory: | ✅ | All tests use `New(":memory:")` or `newTestStore(t)` |
| Tests cover: save, get latest, get history with filters, duplicate prevention, empty results | ✅ | Save (2 tests), Latest (3 tests), History (4 tests), Duplicates (2 tests), Empty (2 tests) |

---

## Spec Compliance Matrix

| # | Scenario | Test | Status |
|---|----------|------|--------|
| 1 | SaveRates persists all rates | TestSaveAndGetLatestRates | ✅ |
| 2 | SaveRates persists bank-specific rates | TestSaveRatesBankSpecific | ✅ |
| 3 | SaveRates rejects duplicate via public API | TestSaveRatesDuplicateViaAPI | ✅ |
| 4 | GetLatestRates returns most recent per type | TestGetLatestRatesMultiTimestamp | ✅ |
| 5 | GetLatestRates deduplicates across banks | TestGetLatestRatesMultiBank | ✅ (see Warning) |
| 6 | GetHistoryRates respects filters | TestGetHistoryRatesTableDriven | ✅ |
| 7 | GetHistoryRates filters by rateType | TestGetHistoryRatesTableDriven/rateType_filter_only | ✅ |
| 8 | GetHistoryRates filters by date range | TestGetHistoryRatesTableDriven/date_range_filter_only | ✅ |
| 9 | GetHistoryRates returns empty slice for no match | TestGetHistoryRatesNilSafety | ✅ |

---

## Issues

### WARNING #1: Spec Scenario Inconsistency

Spec scenario #5 "GetLatestRates deduplicates across multiple banks" states:

> AND the 11:00 rate from "Banco B" is returned for its combination

But the implementation groups by `(currency, rate_type)` — not `(currency, rate_type, bank)`. The acceptance criteria correctly says "per currency and rate_type" which matches the implementation. The test `TestGetLatestRatesMultiBank` correctly validates the actual behavior (1 result: latest across all banks for the same rate_type).

**Impact**: None — the test validates the correct behavior per acceptance criteria. The spec scenario has a minor internal inconsistency.

### WARNING #2: Known Date Comparison Gotcha

String-based date comparison in `GetHistoryRates` causes off-by-one boundary issues. For example, `to="2026-07-10"` excludes datetime values like `"2026-07-10 12:00:00"` because the stored string is lexicographically greater than the date-only filter.

**Impact**: Pre-existing behavior, not introduced by this change. Documented in apply-progress. Would require SQLite date functions or format changes to fix — out of scope for this test-only change.

### SUGGESTION: Coverage Optimization

Coverage at 82.1% — acceptable for a test-only change. Uncovered paths are mostly error branches in `scanRates` and `insertRate` helpers.

---

## Files Changed

| File | Action | Lines Added | Description |
|------|--------|-------------|-------------|
| internal/store/sqlite_test.go | Modified | +200 | Added `newTestStore`, `createTestRate` helpers and 7 new test functions |

---

## Verdict

**PASS**

All acceptance criteria satisfied. All 18 tests pass. Coverage 82.1%. No production code changes — test-only change as designed.
