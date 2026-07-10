# Proposal: 02 — SQLite Repository Implementation

## Intent

Issue #3 requires a complete, tested SQLite repository that implements `domain.Repository`. The scaffolding (issue #1/9) delivered the schema, migration, and a working `Store` with all three methods implemented. However, the existing test suite only covers basic happy paths — it lacks coverage for bank-specific rates, date-range filtering, rateType filtering, and the duplicate-prevention behavior through the public `SaveRates` API. This change hardens the repository with the full acceptance criteria from the PRD.

## Scope

### In Scope

- Validate existing `SaveRates` handles bank-specific rates (bank ≠ "") and reference rates (bank = NULL/empty)
- Add test: `SaveRates` with mixed reference + bank rates, verify all rows persisted
- Add test: duplicate insert via `SaveRates` returns error (unique constraint through public API, not just raw insert)
- Add test: `GetLatestRates` returns only the most recent per `(currency, rate_type)` when multiple timestamps exist
- Add test: `GetHistoryRates` with `rateType` filter
- Add test: `GetHistoryRates` with `from`/`to` date range filters
- Add test: `GetHistoryRates` returns empty slice for no matches (nil safety)
- Verify `GetHistoryRates` ordering is `scraped_at DESC`
- Verify interface compile-time satisfaction (already exists, keep)

### Out of Scope

- Schema changes (UNIQUE constraint already in place)
- New repository methods beyond the three in `domain.Repository`
- Performance/index testing (deferred)
- Integration tests with real database file (in-memory only per criteria)

## Capabilities

### New Capabilities

None — repository methods are already implemented in `internal/store/sqlite.go`.

### Modified Capabilities

- `sqlite-store`: Expanding test coverage for the existing Repository Implementation requirement (all four scenarios in the spec are already defined; this change fills the gaps in test assertions)

## Approach

1. **Gap analysis** — compare existing tests against spec scenarios in `openspec/specs/sqlite-store/spec.md`
2. **Add missing tests** — table-driven subtests in `sqlite_test.go`:
   - `TestSaveRates_BankAndReference` — mixed bank/reference rates
   - `TestSaveRates_DuplicatePrevention` — unique violation through public API
   - `TestGetLatestRates_MultipleTimestamps` — correct dedup behavior
   - `TestGetHistoryRates_RateTypeFilter` — filtered by rateType
   - `TestGetHistoryRates_DateRangeFilter` — from/to boundaries
   - `TestGetHistoryRates_EmptyResult` — nil-safety check
3. **Verify all pass** — `go test ./internal/store/... -v`
4. **No production code changes** — implementation is complete; this is a test-only deliverable

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/store/sqlite_test.go` | Modified | Add 6 new test functions covering acceptance criteria gaps |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| `GetLatestRates` GROUP BY + HAVING MAX behavior is SQLite-specific | Low | Pure-Go driver (`modernc.org/sqlite`), no CGO; test covers behavior directly |
| Date filter uses string comparison on DATETIME | Low | ISO 8601 strings sort lexicographically; tests verify with explicit timestamps |

## Rollback Plan

Revert `internal/store/sqlite_test.go` to pre-change state. No production code is modified, so rollback is zero-risk.

## Dependencies

- Issue #9 (completed): domain types, config, store scaffolding
- `modernc.org/sqlite` pure-Go driver (already in go.mod)

## Success Criteria

- [ ] `go test ./internal/store/... -v` passes (all existing + new tests)
- [ ] All 7 acceptance criteria from issue #3 are covered by at least one test
- [ ] No production code changes required (implementation already complete)
- [ ] Interface compile-time check remains passing
