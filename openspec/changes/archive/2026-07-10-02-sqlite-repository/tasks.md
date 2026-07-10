# Tasks: 02 — SQLite Repository Test Coverage

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | +180–220 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | ask-on-risk |
| Chain strategy | size-exception |

Decision needed before apply: Yes
Chained PRs recommended: No
Chain strategy: size-exception
400-line budget risk: Low

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | All new tests in single PR | PR 1 | ~200 lines added to `sqlite_test.go`; test-only, no prod changes |

## Phase 1: Test Helper Infrastructure

- [x] 1.1 Extract `newTestStore(t *testing.T) *Store` helper at top of `internal/store/sqlite_test.go` — creates in-memory DB + Store, registers `t.Cleanup` for `db.Close()`. Use `t.Helper()`. (~12 lines)
- [x] 1.1b Add `createTestRate` helper — returns `domain.Rate` with sensible defaults.

## Phase 2: SaveRates Tests

- [x] 2.1 Add `TestSaveRatesBankSpecific` — insert mixed reference (bank="") + bank-specific (bank="Banco de Venezuela") rates via `SaveRates`, query raw rows to assert both `bank` values persisted correctly. (~30 lines)
- [x] 2.2 Add `TestSaveRatesDuplicateViaAPI` — save a rate, then call `SaveRates` again with identical `(currency, rate_type, bank, scraped_at)`. Assert error returned and original row unchanged. (~25 lines)

## Phase 3: GetLatestRates Tests

- [x] 3.1 Add `TestGetLatestRatesMultiTimestamp` — insert USD reference at 10:00 and 12:00 same day, call `GetLatestRates("USD")`, assert only 12:00 rate returned. (~25 lines)
- [x] 3.2 Add `TestGetLatestRatesMultiBank` — insert rates from "Banco A" at 10:00 and 12:00, and "Banco B" at 11:00 (all USD reference). Assert GetLatestRates returns latest per bank. (~35 lines)

## Phase 4: GetHistoryRates Tests

- [x] 4.1 Add `TestGetHistoryRatesTableDriven` — seed 3 dates × 2 types × 2 currencies. Table-driven subtests: rateType filter only, date range only, combined filters, no-match. Assert correct count and filter behavior per case. (~60 lines)
- [x] 4.2 Add `TestGetHistoryRatesOrdering` — insert 3 USD rates at different timestamps, call `GetHistoryRates`, verify results are `scraped_at DESC`. (~20 lines)
- [x] 4.3 Add `TestGetHistoryRatesNilSafety` — query nonexistent currency, assert `len(result) == 0` and `result != nil`. (~12 lines)

## Phase 5: Verification

- [x] 5.1 Run `go test ./internal/store/... -v` — all tests pass, no regressions.
- [x] 5.2 Verify no production code changes (`git diff --name-only` shows only `sqlite_test.go`).
