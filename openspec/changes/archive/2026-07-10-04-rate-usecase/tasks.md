# Tasks: Rate Usecase (Business Logic)

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 200–250 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | ask-on-risk |
| Chain strategy | pending |

Decision needed before apply: Yes
Chained PRs recommended: No
Chain strategy: pending
400-line budget risk: Low

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Complete usecase layer (struct + methods + tests) | PR 1 | All three files, fully self-contained |

## Phase 1: Package Foundation

- [ ] 1.1 Create `internal/usecase/doc.go` with package documentation (~15 lines)
- [ ] 1.2 Create `internal/usecase/rate_usecase.go`: define `RateUsecase` struct with `domain.Repository` and `scraper.Scraper` fields, add `NewRateUsecase` constructor (~20 lines)

## Phase 2: Core Implementation

- [ ] 2.1 Implement `ScrapeRates(ctx)` — call `scraper.Scrape`, persist via `repo.SaveRates`, log errors with `slog`, wrap with `fmt.Errorf` (~25 lines)
- [ ] 2.2 Implement `GetCurrentRates(ctx, currency, rateType)` — call `repo.GetLatestRates`, filter by `rateType` using `strings.EqualFold` when non-empty (~25 lines)
- [ ] 2.3 Implement `GetHistoryRates(ctx, currency, rateType, from, to, limit)` — delegate to `repo.GetHistoryRates` with all params, log errors (~15 lines)

## Phase 3: Testing

- [ ] 3.1 Create `internal/usecase/rate_usecase_test.go` with `mockScraper` and `mockRepository` structs implementing both interfaces (~60 lines)
- [ ] 3.2 Test ScrapeRates: success (returns count), scraper error (no SaveRates call), repo error after scrape (~30 lines)
- [ ] 3.3 Test GetCurrentRates: no filter (all returned), with filter (subset returned), empty result (~25 lines)
- [ ] 3.4 Test GetHistoryRates: delegation success, repo error (~15 lines)

## Phase 4: Verification

- [ ] 4.1 Run `go test ./internal/usecase/...` — all tests pass
- [ ] 4.2 Run `go vet ./internal/usecase/...` — no issues
- [ ] 4.3 Verify `go build ./...` — no compilation errors across project
