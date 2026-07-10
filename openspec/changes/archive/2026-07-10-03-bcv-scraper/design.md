# Design: 03 — BCV Scraper

## Technical Approach

Implement a `BCVScraper` struct in `internal/scraper/` that fetches two BCV pages (homepage for reference rates, statistics page for bank rates), parses HTML with goquery, and returns `[]domain.Rate`. Follows existing project patterns: `fmt.Errorf` with `%w` wrapping, table-driven tests, compile-time interface checks.

## Architecture Decisions

### Decision: Scraper Interface

**Choice**: Define `Scraper` interface with `Scrape(ctx context.Context) ([]domain.Rate, error)`
**Alternatives considered**: Concrete struct only (no interface)
**Rationale**: Enables mocking for future cron integration and allows multiple scraper implementations (BCV, other sources) without changing consumers. Matches `domain.Repository` pattern already in the codebase.

### Decision: Selector Isolation

**Choice**: CSS selectors as package-level `const` values, not inline strings
**Alternatives considered**: Inline selector strings in parse functions
**Rationale**: When BCV changes their HTML (medium risk per proposal), selectors are in one place. Tests validate against fixtures, so selector changes are caught immediately. Adds TODO comment for monitoring.

### Decision: Two-Phase Fetch

**Choice**: Fetch homepage and statistics page sequentially in `Scrape()`, combine results
**Alternatives considered**: Parallel fetches, single-page parse
**Rationale**: BCV reference rates and bank rates live on different pages. Sequential is simpler and sufficient — scraper runs on a schedule (not user-facing latency). Context timeout covers both requests.

### Decision: Error Strategy

**Choice**: Return `fmt.Errorf("parse USD rate: %w", err)` with descriptive context per element
**Alternatives considered**: Sentinel errors only, custom error types
**Rationale**: Matches existing `store` pattern. Each parse step wraps errors with context ("parse USD rate", "fetch reference page") so logs pinpoint failures. No custom types needed for this scope.

## Data Flow

```
Scrape(ctx)
  │
  ├─→ fetchPage(ctx, homepageURL)
  │     └─→ http.Client.Do(req) → *html.Node
  │
  ├─→ parseReferenceRates(doc) → []domain.Rate (USD, EUR)
  │     ├─→ #dolar .strong-tb → USD reference
  │     └─→ #euro .strong-tb  → EUR reference
  │
  ├─→ fetchPage(ctx, statsURL)
  │     └─→ http.Client.Do(req) → *html.Node
  │
  ├─→ parseDate(doc) → time.Time
  │     └─→ .date-display-single[content] → ISO 8601
  │
  ├─→ parseBankRates(doc) → []domain.Rate
  │     └─→ .views-table tbody tr → iterate rows
  │           ├─→ td[0] → bank name
  │           ├─→ td[1] → compra (buy rate)
  │           └─→ td[2] → venta (sell rate)
  │
  └─→ combine rates + set ScrapedAt → []domain.Rate
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `go.mod` | Modify | Add `github.com/PuerkitoBio/goquery` dependency |
| `internal/scraper/scraper.go` | Create | `Scraper` interface, `BCVScraper` struct, `Scrape()`, parse helpers, selector constants |
| `internal/scraper/scraper_test.go` | Create | Table-driven tests with `httptest.NewServer`, fixture HTML |
| `internal/scraper/fixtures/bcv-homepage.html` | Create | Sample BCV homepage HTML for tests |
| `internal/scraper/fixtures/bcv-stats.html` | Create | Sample BCV statistics page HTML for tests |
| `internal/domain/errors.go` | Modify | Add scraper sentinel errors (`ErrScrapeFailed`, `ErrParseFailed`) |

## Interfaces / Contracts

```go
// Scraper defines the contract for exchange rate data sources.
type Scraper interface {
    Scrape(ctx context.Context) ([]domain.Rate, error)
}

// BCVScraper fetches and parses rates from the Banco Central de Venezuela.
type BCVScraper struct {
    client      *http.Client
    homepageURL string  // https://www.bcv.org.ve
    statsURL    string  // https://www.bcv.org.ve/estadisticas/tipo-cambio-de-referencia-smc
}

// Selector constants (isolated for easy update when BCV changes HTML)
const (
    selUSDRate     = "#dolar .strong-tb"
    selEURRate     = "#euro .strong-tb"
    selBankTable   = ".views-table tbody tr"
    selDate        = ".date-display-single"
    selDateAttr    = "content"
)
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | `parseReferenceRates`, `parseBankRates`, `parseDate` | Table-driven subtests with raw HTML strings |
| Integration | Full `Scrape()` flow | `httptest.NewServer` serving fixture HTML from `fixtures/` dir |
| Edge Cases | Missing selectors, non-numeric values, empty table, malformed HTML | Subtests with modified fixtures |

**Test structure**:
- `TestScrapeSuccess` — full scrape returns correct rates
- `TestScrapeMissingUSD` — missing `#dolar` returns error
- `TestScrapeNonNumericValue` — "N/A" in rate returns error
- `TestScrapeEmptyBankTable` — empty table returns only reference rates
- `TestScrapeMalformedHTML` — invalid HTML returns error (not panic)
- `TestScrapeHTTPError` — 500 status returns error
- `TestScrapeContextCancellation` — cancelled context aborts immediately

## Migration / Rollout

No migration required. This change adds a new isolated package with no existing consumers. The scraper will be wired to the cron scheduler in a subsequent change (issue #5).

## Open Questions

- [ ] Should the scraper accept a custom `*http.Client` via constructor (for testing) or always create one internally? **Recommendation**: Accept via `NewBCVScraper(client *http.Client, homepageURL, statsURL string)` — enables `httptest` injection without global state.
- [ ] Should `ErrScrapeFailed` and `ErrParseFailed` live in `domain/errors.go` or `scraper/errors.go`? **Recommendation**: Keep in `domain/errors.go` since other packages (cron, API) may need to check them.
