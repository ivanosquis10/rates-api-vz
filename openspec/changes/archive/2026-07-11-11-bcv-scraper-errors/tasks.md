# Tasks: BCV Scraper Errors and Single Request Layout

## Review Workload Forecast

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: stacked-to-main
400-line budget risk: Low

| Field | Value |
|-------|-------|
| Estimated changed lines | 200 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | single PR |
| Delivery strategy | auto-chain |
| Chain strategy | stacked-to-main |

### Suggested Work Units
| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Setup custom HTTP client & error types | PR 1 | Define provider error and custom client |
| 2 | Consolidate scraping on homepage | PR 1 | Modify scraper structure and parsing logic |
| 3 | Wiring and tests update | PR 1 | Adapt presenter, main, and tests |

## Phase 1: Foundation / Infrastructure
- [x] 1.1 Add `PROVIDER_ERROR` Code and `NewProviderError` in [internal/apierrors/apierrors.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/apierrors/apierrors.go)
- [x] 1.2 Setup custom client transport with `TLSHandshakeTimeout = 10 * time.Second` and `Timeout = 15 * time.Second` in [cmd/api/main.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/cmd/api/main.go)

## Phase 2: Core Implementation
- [x] 2.1 Update `NewBCVScraper` signature and remove `statsURL` from `BCVScraper` in [internal/scraper/scraper.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/scraper/scraper.go)
- [x] 2.2 Consolidate scraping on BCV homepage (fetching homepage only, parsing references, date, and bank rates) in [internal/scraper/scraper.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/scraper/scraper.go)
- [x] 2.3 Inject browser headers and wrap errors in `NewProviderError` in [internal/scraper/scraper.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/scraper/scraper.go)
- [x] 2.4 Consolidate HTML fixtures in [internal/scraper/fixtures/bcv-homepage.html](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/scraper/fixtures/bcv-homepage.html) and remove [internal/scraper/fixtures/bcv-stats.html](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/scraper/fixtures/bcv-stats.html)

## Phase 3: Integration / Wiring
- [x] 3.1 Extract `apierrors.Error` using `errors.As(err, &apiErr)` in `presenter.Error` in [internal/presenter/presenter.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/presenter/presenter.go)
- [x] 3.2 Add robust timeout checks (checking `net.Error`, context deadline, and "timeout"/"deadline"/"handshake" substrings) to return 504 (or 502) without masking in [internal/presenter/presenter.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/presenter/presenter.go)
- [x] 3.3 Update `NewBCVScraper` instantiation in [cmd/api/main.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/cmd/api/main.go)

## Phase 4: Testing / Verification
- [x] 4.1 Update all 7 invocations of `NewBCVScraper` in [internal/scraper/scraper_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/scraper/scraper_test.go) to pass 2 arguments
- [x] 4.2 Run `go test ./...` to verify compilation and unit tests

## Phase 5: Cleanup
- [x] 5.1 Remove unused code and clean up scrape endpoints logic
