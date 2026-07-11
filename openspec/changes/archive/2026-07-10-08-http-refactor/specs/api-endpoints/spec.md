# Delta for api-endpoints

## MODIFIED Requirements

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

#### Scenario: Trigger scrape successfully
- GIVEN repository is operational
- WHEN POST /admin/scrape
- THEN HTTP 202 is returned
- AND response is success envelope with scrape message
  (Previously: - AND response body is `{ "data": { "message": "scrape triggered", "rates_scraped": <count> } }`)

#### Scenario: Scrape fails
- GIVEN scraper returns error
- WHEN POST /admin/scrape
- THEN HTTP 500 is returned
- AND response is error envelope with code "INTERNAL_ERROR"
  (Previously: - AND response body is `{ "error": { "code": "INTERNAL_ERROR", "message": "..." } }`)

### Requirement: Request Validation
The system MUST reject invalid parameters returning HTTP 400 with error envelope.

#### Scenario: Bad limit parameter
- GIVEN limit=abc
- WHEN GET /rates/history?limit=abc
- THEN HTTP 400 is returned
- AND response is error envelope with code "BAD_REQUEST"
  (Previously: - AND response body contains `{ "error": { "code": "BAD_REQUEST", "message": "..." } }`)
