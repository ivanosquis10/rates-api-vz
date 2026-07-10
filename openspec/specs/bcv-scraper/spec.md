# BCV Scraper Specification

## Purpose

Implement a web scraper that fetches exchange rates from the Banco Central de Venezuela (BCV) website, parses the HTML response, and returns domain `Rate` structs. This is the first data ingestion pipeline for the Venezuela Rates API.

## Requirements

### Requirement: Fetch BCV Pages

The scraper SHALL use `net/http` with context-aware requests to fetch the BCV homepage (`https://www.bcv.org.ve`) and the reference rate page (`/estadisticas/tipo-cambio-de-referencia-smc`). The HTTP client MUST support configurable timeouts.

#### Scenario: Successful fetch of both pages

- GIVEN a valid BCV website responding with HTML
- WHEN `Scrape(ctx)` is called
- THEN both pages are fetched without error
- AND the response bodies are passed to the HTML parser

#### Scenario: Network timeout

- GIVEN a BCV website that does not respond within the configured timeout
- WHEN `Scrape(ctx)` is called
- THEN an error is returned indicating a timeout
- AND no partial results are returned

#### Scenario: Context cancellation

- GIVEN a cancelled context
- WHEN `Scrape(ctx)` is called
- THEN the request is aborted immediately
- AND an error is returned

### Requirement: Parse USD Reference Rate

The scraper SHALL extract the USD reference rate from the BCV homepage using the CSS selector `#dolar .strong-tb`. The extracted value MUST be a valid numeric string parseable to float64.

#### Scenario: Parse valid USD rate

- GIVEN a BCV HTML page with `<div id="dolar"><span class="strong-tb">36.50</span></div>`
- WHEN the scraper parses the USD reference rate
- THEN a Rate with Currency="USD", RateType="reference", Bank="", Value=36.50 is produced

#### Scenario: Missing USD selector

- GIVEN a BCV HTML page without the `#dolar .strong-tb` element
- WHEN the scraper parses the USD reference rate
- THEN an error is returned indicating the USD rate could not be found

#### Scenario: Non-numeric USD value

- GIVEN a BCV HTML page with `<div id="dolar"><span class="strong-tb">N/A</span></div>`
- WHEN the scraper parses the USD reference rate
- THEN an error is returned indicating the value is not a valid number

### Requirement: Parse EUR Reference Rate

The scraper SHALL extract the EUR reference rate from the BCV homepage using the CSS selector `#euro .strong-tb`. The extracted value MUST be a valid numeric string parseable to float64.

#### Scenario: Parse valid EUR rate

- GIVEN a BCV HTML page with `<div id="euro"><span class="strong-tb">38.20</span></div>`
- WHEN the scraper parses the EUR reference rate
- THEN a Rate with Currency="EUR", RateType="reference", Bank="", Value=38.20 is produced

#### Scenario: Missing EUR selector

- GIVEN a BCV HTML page without the `#euro .strong-tb` element
- WHEN the scraper parses the EUR reference rate
- THEN an error is returned indicating the EUR rate could not be found

### Requirement: Parse Bank Rates

The scraper SHALL extract buy and sell rates from the `.views-table tbody tr` table. Each row produces two Rate structs: one for `RateType="buy"` and one for `RateType="sell"`. The bank name MUST be extracted from the first cell of each row.

#### Scenario: Parse valid bank rates

- GIVEN a BCV HTML table with a row: `<tr><td>Banesco</td><td>35.80</td><td>36.50</td></tr>`
- WHEN the scraper parses the bank rates table
- THEN two Rates are produced: (Currency="USD", RateType="buy", Bank="Banesco", Value=35.80) and (Currency="USD", RateType="sell", Bank="Banesco", Value=36.50)

#### Scenario: Empty bank rates table

- GIVEN a BCV HTML page with an empty `.views-table tbody`
- WHEN the scraper parses the bank rates table
- THEN zero bank rates are returned (not an error)
- AND the reference rates are still returned

#### Scenario: Missing bank name cell

- GIVEN a BCV HTML table row without the first cell (bank name)
- WHEN the scraper parses the bank rates table
- THEN that row is skipped with no error
- AND valid rows are still parsed

### Requirement: Parse Scrape Date

The scraper SHALL extract the publication date from the BCV page using the `.date-display-single` element's `content` attribute (ISO 8601 format). This date MUST be assigned as `ScrapedAt` on all returned Rate structs.

#### Scenario: Parse valid ISO 8601 date

- GIVEN a BCV HTML page with `<span class="date-display-single" content="2026-07-10T00:00:00-04:00">10 julio 2026</span>`
- WHEN the scraper parses the date
- THEN all returned rates have ScrapedAt set to 2026-07-10 00:00:00 -04:00

#### Scenario: Missing date element

- GIVEN a BCV HTML page without the `.date-display-single` element
- WHEN the scraper parses the date
- THEN an error is returned indicating the date could not be found

### Requirement: Return Domain Rate Structs

The scraper SHALL return `[]domain.Rate` with all required fields populated: `Currency`, `RateType`, `Bank`, `Value`, `ScrapedAt`. The returned slice MUST contain at least the USD and EUR reference rates when parsing succeeds.

#### Scenario: Full successful scrape

- GIVEN a well-formed BCV HTML page with USD reference, EUR reference, and bank rates
- WHEN `Scrape(ctx)` is called
- THEN the returned slice contains at least 2 reference rates (USD, EUR)
- AND any bank rates found are included
- AND all rates have ScrapedAt set to the parsed date

#### Scenario: Partial scrape (bank table missing)

- GIVEN a BCV HTML page with USD/EUR reference rates but no bank table
- WHEN `Scrape(ctx)` is called
- THEN the returned slice contains only the 2 reference rates
- AND no error is returned for the missing bank table

### Requirement: Graceful Error Handling

The scraper MUST return descriptive errors (never panic) for: malformed HTML, missing required selectors, network failures, and invalid numeric values. Errors MUST include context about which element or page failed.

#### Scenario: Malformed HTML

- GIVEN a BCV page returning invalid HTML
- WHEN `Scrape(ctx)` is called
- THEN an error is returned (not a panic)
- AND the error message indicates HTML parsing failure

#### Scenario: HTTP 500 response

- GIVEN a BCV page returning HTTP 500
- WHEN `Scrape(ctx)` is called
- THEN an error is returned indicating a non-200 status code
- AND no partial results are returned

## Acceptance Criteria

- [ ] `go test ./internal/scraper/... -v` passes with all test cases
- [ ] Scraper returns correct `[]domain.Rate` for well-formed BCV HTML
- [ ] Scraper returns descriptive error (not panic) for malformed/missing HTML
- [ ] USD and EUR reference rates parsed correctly
- [ ] Bank rates parsed with correct currency, bank name, buy/sell values
- [ ] Date parsed from ISO 8601 attribute
- [ ] `go build ./...` succeeds
