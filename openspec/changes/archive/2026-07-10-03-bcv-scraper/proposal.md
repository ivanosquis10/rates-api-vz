# Proposal: 03 — BCV Scraper

## Intent

The API has domain types and storage but no data source. This change implements the BCV (Banco Central de Venezuela) web scraper that fetches exchange rates from the official BCV website, parses the HTML, and returns domain `Rate` structs. This is the first data ingestion pipeline — without it, the API serves nothing.

## Scope

### In Scope

- Add `goquery` dependency (HTML parsing library)
- Create `internal/scraper/` package with `BCVScraper` type
- Implement HTTP fetch of BCV homepage (`https://www.bcv.org.ve`) and reference page (`/estadisticas/tipo-cambio-de-referencia-smc`)
- Parse USD reference rate from `#dolar .strong-tb` selector
- Parse EUR reference rate from `#euro .strong-tb` selector
- Parse bank rates from `.views-table tbody tr` (columns: Banco, Compra, Venta)
- Parse date from `.date-display-single[content]` (ISO 8601)
- Return `[]domain.Rate` with correct `Currency`, `RateType`, `Bank`, `Value`, `ScrapedAt`
- Graceful error handling: return error on malformed HTML, missing selectors, network failure (no panic)
- Tests with `httptest.NewServer` serving sample BCV HTML
- Test coverage: successful parse, missing element, malformed HTML, empty table

### Out of Scope

- Scheduling/cron integration (issue #5 territory)
- Rate storage/persistence after scraping
- Rate normalization or business logic
- Multiple data source scrapers (only BCV)
- Caching or rate-limiting HTTP calls

## Capabilities

### New Capabilities

- `bcv-scraper`: BCV web scraper — fetches and parses exchange rates from the official BCV website, returning domain Rate structs

### Modified Capabilities

None — no existing spec-level behavior changes.

## Approach

1. **Dependency**: `go get github.com/PuerkitoBio/goquery` (pure-Go HTML parser, already identified in project planning)
2. **Package**: `internal/scraper/scraper.go` with `BCVScraper` struct holding an `*http.Client`
3. **Method**: `Scrape(ctx context.Context) ([]domain.Rate, error)` — single entry point
4. **Fetch**: Use `net/http` with context-aware request to both BCV pages
5. **Parse**: goquery selectors for each data point:
   - Reference rates: `#dolar .strong-tb`, `#euro .strong-tb`
   - Bank rates table: `.views-table tbody tr` → iterate cells for bank name, compra, venta
   - Date: `.date-display-single` attribute `content`
6. **Validation**: Check each extracted value before building Rate structs; return `fmt.Errorf` on missing required data
7. **Tests**: `internal/scraper/scraper_test.go` with `httptest.NewServer` serving fixture HTML, table-driven subtests

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `go.mod` | Modified | Add `goquery` + transitive deps |
| `internal/scraper/scraper.go` | New | BCV scraper implementation |
| `internal/scraper/scraper_test.go` | New | Tests with httptest fixtures |
| `internal/scraper/fixtures/` | New | Sample BCV HTML for tests |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| BCV website HTML structure changes (selectors break) | Medium | Isolate selectors as constants; tests validate against known fixtures; add TODO for monitoring |
| goquery selector mismatch (real site differs from expectations) | Medium | Use multiple selector strategies as fallback; document expected HTML structure |
| Network/timeout issues in production | Low | Context-based timeout on HTTP client; scraper is called by scheduler, not user-facing endpoint |

## Rollback Plan

Delete `internal/scraper/` package and remove `goquery` from `go.mod`. No other packages depend on the scraper yet (storage and API are independent), so rollback is clean.

## Dependencies

- Issue #9 (completed): domain types (`Rate`, `Repository` interface)
- `goquery` library (new dependency, `github.com/PuerkitoBio/goquery`)

## Success Criteria

- [ ] `go test ./internal/scraper/... -v` passes with all test cases
- [ ] Scraper returns correct `[]domain.Rate` for well-formed BCV HTML
- [ ] Scraper returns descriptive error (not panic) for malformed/missing HTML
- [ ] USD and EUR reference rates parsed correctly
- [ ] Bank rates parsed with correct currency, bank name, buy/sell values
- [ ] Date parsed from ISO 8601 attribute
- [ ] `go build ./...` succeeds
