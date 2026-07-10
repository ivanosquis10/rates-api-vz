# Archive Report: 02 — SQLite Repository Test Coverage

**Change**: 02-sqlite-repository
**Status**: COMPLETE
**Archived**: 2026-07-10
**Verdict**: PASS (no CRITICAL issues)

---

## Executive Summary

Expanded test coverage for the existing SQLite repository implementation from 11 basic happy-path tests to 18 comprehensive tests covering all 9 spec scenarios. Test-only change — no production code modified. All acceptance criteria from issue #3 satisfied.

## What Was Built

| Capability | Impact |
|------------|--------|
| `TestSaveRatesBankSpecific` | Verifies mixed reference + bank-specific rates persist correctly |
| `TestSaveRatesDuplicateViaAPI` | Confirms duplicate prevention through public API (not just raw insert) |
| `TestGetLatestRatesMultiTimestamp` | Ensures most-recent-per-type selection works across timestamps |
| `TestGetLatestRatesMultiBank` | Validates latest-per-type deduplication across multiple banks |
| `TestGetHistoryRatesTableDriven` | 4 subtests covering rateType, date range, combined, and no-match filters |
| `TestGetHistoryRatesOrdering` | Confirms `scraped_at DESC` ordering |
| `TestGetHistoryRatesNilSafety` | Ensures empty results return non-nil slice |
| `newTestStore` helper | Reduces boilerplate across all tests |
| `createTestRate` helper | Standardizes rate fixture creation |

## Key Decisions

| Decision | Rationale |
|----------|-----------|
| Test-only scope | Production code already complete; only test coverage gaps existed |
| Table-driven subtests for GetHistoryRates | 4 filter scenarios grouped naturally; reduces boilerplate |
| Extract `newTestStore` helper | Every test repeated 6-line setup; follows Go testing convention |
| Test duplicate via public API | Spec explicitly requires "rejects duplicate via public API" as distinct from raw insert |

## Files Modified

| File | Lines Changed | Description |
|------|---------------|-------------|
| `internal/store/sqlite_test.go` | +200 | 7 new test functions + 2 helpers |

**No production code was modified.**

## Test Results

- **Total tests**: 18 (11 original + 7 new)
- **All passing**: ✅
- **Coverage**: 82.1% of statements
- **Regression**: None

## Engram Artifact IDs

| Artifact | Observation ID |
|----------|----------------|
| proposal | #524 |
| spec | #525 |
| design | #526 |
| tasks | #527 |
| apply-progress | #529 |
| verify-report | #531 |

## Lessons Learned

1. **SQLite string-based date comparison is tricky** — `GetHistoryRates` uses `scraped_at >= ?` with string comparison. The stored format `"2026-07-10 00:00:00"` (with space) is lexicographically greater than `"2026-07-10"`, causing off-by-one boundary issues. This is a known limitation of the current implementation, not a regression from this change.

2. **Spec scenario #5 had a minor internal inconsistency** — It stated "the 11:00 rate from 'Banco B' is returned for its combination" but implementation groups by `(currency, rate_type)`, not `(currency, rate_type, bank)`. The acceptance criteria and tests correctly validate the actual behavior.

3. **Test helpers pay for themselves quickly** — `newTestStore` and `createTestRate` reduced boilerplate by ~12 lines per test and made test intent clearer.

## Archive Verification

- [x] Main spec updated (5 new scenarios added to "Repository Implementation")
- [x] Change folder moved to `openspec/changes/archive/2026-07-10-02-sqlite-repository/`
- [x] Archive contains all artifacts (proposal, specs, design, tasks, apply-progress, verify-report)
- [x] All 11 implementation tasks marked complete in tasks artifact
- [x] Active changes directory no longer contains this change

## SDD Cycle Complete

The change has been fully planned, implemented, verified, and archived.
Ready for the next change.
