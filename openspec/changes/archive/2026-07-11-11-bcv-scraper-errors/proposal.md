<Proposal: 11-bcv-scraper-errors>
## Intent
Resolve live Banco Central de Venezuela (BCV) scraping failures caused by connection/TLS handshake timeouts and stats page 404s, while ensuring diagnostic errors are descriptively propagated to API clients.

## Scope
### In Scope
- Add [PROVIDER_ERROR](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/apierrors/apierrors.go#L11) error code and wrapper [NewProviderError](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/apierrors/apierrors.go) in [apierrors.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/apierrors/apierrors.go).
- Map provider errors in [presenter.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/presenter/presenter.go) to `504 Gateway Timeout` (for TLS/connection timeouts) or `502 Bad Gateway` (for other failures) without masking the raw error message.
- Modify [scraper.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/scraper/scraper.go) to parse USD/EUR reference rates, date, and bank rates entirely from the homepage in a single HTTP request. Remove the `statsURL` from [NewBCVScraper](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/scraper/scraper.go#L37).
- Inject browser-like headers in [fetchPage](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/scraper/scraper.go#L91) to prevent firewall throttling.
- Instantiate a custom `*http.Client` with timeouts (15s total, 10s TLS handshake) in [main.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/cmd/api/main.go).
- Update unit tests in [scraper_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/scraper/scraper_test.go).

### Out of Scope
- Automatic scraper retries.
- Support for alternate providers.

## Capabilities
### New Capabilities
- None.
### Modified Capabilities
- `exchange-rate-scraping`: Scrapes references, dates, and bank rates in a single request to the homepage, avoiding 404s on the stats page, with custom HTTP timeouts and headers.
- `error-reporting`: Returns detailed raw provider error messages instead of generic internal errors, returning 502/504 status codes.

## Approach
1. Add `PROVIDER_ERROR` to [apierrors.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/apierrors/apierrors.go).
2. Intercept `PROVIDER_ERROR` in [presenter.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/presenter/presenter.go). If the error message indicates a timeout, respond with `504 Gateway Timeout`, else `502 Bad Gateway`. Write the raw error message.
3. Consolidate selector parsing inside [scraper.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/scraper/scraper.go) to use only the homepage. Add browser headers in [fetchPage](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/scraper/scraper.go#L91).
4. Update client construction in [main.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/cmd/api/main.go).
5. Modify [scraper_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/scraper/scraper_test.go) to mock only the homepage and assert values.

## Affected Areas
| Area | Impact | Description |
|------|--------|-------------|
| [apierrors.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/apierrors/apierrors.go) | Modified | Define error code and provider error wrapper. |
| [presenter.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/presenter/presenter.go) | Modified | Support 502/504 mapping with raw messages. |
| [scraper.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/scraper/scraper.go) | Modified | Consolidate parsing on homepage request, add custom headers. |
| [main.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/cmd/api/main.go) | Modified | Inject custom timeouts and update scraper initialization. |
| [scraper_test.go](file:///C:/Users/Windows%2011/OneDrive/Documentos/projects/backend-projects/go-projects/venezuela-rates-api/internal/scraper/scraper_test.go) | Modified | Update tests to mock single homepage request. |

## Risks
| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Homepage HTML structure changes break selectors | Medium | Selectors are isolated; unit tests cover extraction logic. |

## Rollback Plan
Revert changes using git: `git checkout main -- cmd/api/main.go internal/`

## Dependencies
- Public availability of Banco Central de Venezuela homepage.

## Success Criteria
- [ ] Scraper fetches all data successfully from the homepage.
- [ ] TLS timeouts and HTTP 404s are returned to clients as 504 and 502 responses containing the raw error message.
- [ ] All unit tests pass.
