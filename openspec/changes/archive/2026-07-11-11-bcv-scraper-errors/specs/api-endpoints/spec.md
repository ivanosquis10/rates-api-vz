# Delta for api-endpoints

## MODIFIED Requirements

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
