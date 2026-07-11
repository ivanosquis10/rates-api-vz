# API Endpoints Specification

## Purpose

RESTful HTTP handlers that expose RateUsecase operations. Each endpoint maps to a usecase method, parses query parameters, and returns standardized JSON responses.

## Requirements

### Requirement: GET /rates
The system SHALL serve `GET /rates` enveloped using presenter helpers.

#### Scenario: Get all rates
- GIVEN rates exist
- WHEN GET /rates is requested
- THEN HTTP 200 is returned
- AND response is success envelope with rates list
  (Previously: - AND response body is `{ "data": [<all rates>] }`)

#### Scenario: Filter by currency
- GIVEN rates exist
- WHEN GET /rates?currency=USD
- THEN HTTP 200 is returned
- AND response is success envelope with USD rates
  (Previously: - AND response contains only USD rates)

#### Scenario: Filter by currency and type
- GIVEN rates exist
- WHEN GET /rates?currency=USD&type=reference
- THEN HTTP 200 is returned
- AND response is success envelope with USD reference rates
  (Previously: - AND response contains only USD reference rates)

### Requirement: GET /rates/history
The system SHALL serve `GET /rates/history` enveloped using presenter helpers.

#### Scenario: History with all filters
- GIVEN historical rates
- WHEN GET /rates/history?currency=USD&type=buy&from=2026-01-01&to=2026-07-01&limit=50
- THEN HTTP 200 is returned
- AND response is success envelope with up to 50 rates
  (Previously: - AND response contains up to 50 matching rates)

#### Scenario: History with no data
- GIVEN no rates match
- WHEN GET /rates/history is requested
- THEN HTTP 200 is returned
- AND response is success envelope with empty data list
  (Previously: - AND response body is `{ "data": [] }`)

### Requirement: POST /admin/scrape
The system SHALL serve `POST /admin/scrape` enveloped using presenter helpers.
(Previously: Scrape failures only returned HTTP 500 with "INTERNAL_ERROR")

#### Scenario: Trigger scrape successfully
- GIVEN repository is operational
- WHEN POST /admin/scrape
- THEN HTTP 202 is returned
- AND response is success envelope with scrape message

#### Scenario: Scrape fails with provider timeout
- GIVEN scraper returns a timeout error
- WHEN POST /admin/scrape
- THEN HTTP 504 is returned
- AND response is error envelope with code "PROVIDER_ERROR"
  (Previously: - THEN HTTP 500 is returned
  - AND response is error envelope with code "INTERNAL_ERROR")

#### Scenario: Scrape fails with provider network error
- GIVEN scraper returns a non-timeout network error
- WHEN POST /admin/scrape
- THEN HTTP 502 is returned
- AND response is error envelope with code "PROVIDER_ERROR"
  (Previously: - THEN HTTP 500 is returned
  - AND response is error envelope with code "INTERNAL_ERROR")

#### Scenario: Scrape fails with internal database error
- GIVEN database is offline when saving rates
- WHEN POST /admin/scrape
- THEN HTTP 500 is returned
- AND response is error envelope with code "INTERNAL_ERROR"

### Requirement: Request Validation
The system MUST reject invalid parameters returning HTTP 400 with error envelope.

#### Scenario: Bad limit parameter
- GIVEN limit=abc
- WHEN GET /rates/history?limit=abc
- THEN HTTP 400 is returned
- AND response is error envelope with code "BAD_REQUEST"
  (Previously: - AND response body contains `{ "error": { "code": "BAD_REQUEST", "message": "..." } }`)

## Acceptance Criteria

- [ ] GET /rates returns filtered rates in `{ "data": [...] }` envelope
- [ ] GET /rates/history respects all query params
- [ ] POST /admin/scrape returns 202 with confirmation
- [ ] Invalid params return 400 with error envelope
- [ ] All 7 handler test scenarios pass
- [ ] Issue #6 endpoint requirements satisfied
