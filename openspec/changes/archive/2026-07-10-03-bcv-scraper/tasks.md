# Tasks: 03 — BCV Scraper

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~340 (all new files) |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | ask-on-risk |
| Chain strategy | pending |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: pending
400-line budget risk: Low

## Phase 1: Foundation (~25 lines)

- [x] 1.1 Add `github.com/PuerkitoBio/goquery` dependency to `go.mod` via `go get`
- [x] 1.2 Add sentinel errors `ErrScrapeFailed` and `ErrParseFailed` to `internal/domain/errors.go`
- [x] 1.3 Update `internal/domain/errors_test.go` with tests for new sentinel errors

## Phase 2: Scraper Implementation (~180 lines)

- [x] 2.1 Create `internal/scraper/scraper.go` — `Scraper` interface with `Scrape(ctx) ([]domain.Rate, error)`
- [x] 2.2 Implement `BCVScraper` struct with `NewBCVScraper(client, homepageURL, statsURL)` constructor
- [x] 2.3 Implement `fetchPage(ctx, url)` helper — HTTP GET, status check, goquery parse
- [x] 2.4 Implement `parseReferenceRates(doc)` — extract USD/EUR from `#dolar .strong-tb`, `#euro .strong-tb`
- [x] 2.5 Implement `parseBankRates(doc)` — iterate `.views-table tbody tr`, extract bank/compra/venta
- [x] 2.6 Implement `parseDate(doc)` — extract ISO 8601 from `.date-display-single[content]`
- [x] 2.7 Implement `Scrape(ctx)` orchestration — fetch homepage, parse refs, fetch stats, parse banks+date, combine

## Phase 3: Test Fixtures (~60 lines)

- [x] 3.1 Create `internal/scraper/fixtures/bcv-homepage.html` — sample with USD/EUR reference rates
- [x] 3.2 Create `internal/scraper/fixtures/bcv-stats.html` — sample with bank rates table and date element

## Phase 4: Tests (~80 lines)

- [x] 4.1 Create `internal/scraper/scraper_test.go` — `TestScrapeSuccess` with httptest serving both fixtures
- [x] 4.2 Add `TestScrapeMissingUSD` — missing `#dolar` returns error
- [x] 4.3 Add `TestScrapeNonNumericValue` — "N/A" in rate returns error
- [x] 4.4 Add `TestScrapeEmptyBankTable` — empty tbody returns only reference rates
- [x] 4.5 Add `TestScrapeMalformedHTML` — invalid HTML returns error (not panic)
- [x] 4.6 Add `TestScrapeHTTPError` — 500 status returns error
- [x] 4.7 Add `TestScrapeContextCancellation` — cancelled context aborts immediately
