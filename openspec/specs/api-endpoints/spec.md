# API Endpoints Specification

## Purpose

RESTful HTTP handlers that expose RateUsecase operations. Each endpoint maps to a usecase method, parses query parameters, and returns standardized JSON responses.

## Requirements

### Requirement: GET /rates

The system SHALL serve `GET /rates` with optional query parameters `currency` (string) and `type` (string). The handler MUST call `RateUsecase.GetCurrentRates(ctx, currency, rateType)` and return the result.

#### Scenario: Get all rates

- GIVEN rates exist for USD and EUR
- WHEN GET /rates is requested with no query params
- THEN HTTP 200 is returned
- AND response body is `{ "data": [<all rates>] }`

#### Scenario: Filter by currency

- GIVEN rates exist for USD and EUR
- WHEN GET /rates?currency=USD is requested
- THEN HTTP 200 is returned
- AND response contains only USD rates

#### Scenario: Filter by currency and type

- GIVEN USD reference, buy, and sell rates exist
- WHEN GET /rates?currency=USD&type=reference is requested
- THEN HTTP 200 is returned
- AND response contains only USD reference rates

### Requirement: GET /rates/history

The system SHALL serve `GET /rates/history` with query parameters: `currency`, `type`, `from`, `to`, `limit`. The handler MUST call `RateUsecase.GetHistoryRates(ctx, ...)` with all parameters.

#### Scenario: History with all filters

- GIVEN historical USD buy rates exist
- WHEN GET /rates/history?currency=USD&type=buy&from=2026-01-01&to=2026-07-01&limit=50 is requested
- THEN HTTP 200 is returned
- AND response contains up to 50 matching rates

#### Scenario: History with no data

- GIVEN no rates match the filters
- WHEN GET /rates/history is requested
- THEN HTTP 200 is returned
- AND response body is `{ "data": [] }`

### Requirement: POST /admin/scrape

The system SHALL serve `POST /admin/scrape`. The handler MUST call `RateUsecase.ScrapeRates(ctx)` and return HTTP 202 Accepted with a confirmation message.

#### Scenario: Trigger scrape successfully

- GIVEN the scraper and repository are operational
- WHEN POST /admin/scrape is requested
- THEN HTTP 202 is returned
- AND response body is `{ "data": { "message": "scrape triggered", "rates_scraped": <count> } }`

#### Scenario: Scrape fails

- GIVEN the scraper returns an error
- WHEN POST /admin/scrape is requested
- THEN HTTP 500 is returned
- AND response body is `{ "error": { "code": "INTERNAL_ERROR", "message": "..." } }`

### Requirement: Request Validation

The system MUST reject requests with invalid query parameters by returning HTTP 400. Invalid parameters include non-numeric `limit` values and malformed date strings.

#### Scenario: Bad limit parameter

- GIVEN a request with limit=abc
- WHEN GET /rates/history?limit=abc is processed
- THEN HTTP 400 is returned
- AND response body contains `{ "error": { "code": "BAD_REQUEST", "message": "..." } }`

## Acceptance Criteria

- [ ] GET /rates returns filtered rates in `{ "data": [...] }` envelope
- [ ] GET /rates/history respects all query params
- [ ] POST /admin/scrape returns 202 with confirmation
- [ ] Invalid params return 400 with error envelope
- [ ] All 7 handler test scenarios pass
- [ ] Issue #6 endpoint requirements satisfied
