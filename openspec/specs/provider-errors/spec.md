# Provider Errors Specification

## Purpose
Define requirements for wrapping, mapping, and returning errors from external rates providers (e.g., BCV) to prevent masking diagnostics.

## Requirements

### Requirement: Mapping Provider Errors
The system MUST wrap scraper or external provider failures in a custom provider error with the code `PROVIDER_ERROR`.
The system MUST map the custom provider error to appropriate HTTP status codes based on the inner error details:
- If the error message indicates a timeout (e.g., context deadline exceeded, connection timeout, or TLS handshake timeout), the system MUST return `504 Gateway Timeout`.
- For other provider failures (e.g., HTTP 404, invalid response status, or network connection refused), the system MUST return `502 Bad Gateway`.
The system MUST return the raw, unmodified error message from the provider in the error response payload.

#### Scenario: Provider timeout mapping
- GIVEN a scraping operation fails with a connection or TLS handshake timeout
- WHEN the presenter processes the provider error
- THEN HTTP status 504 is returned
- AND the JSON response payload contains error code "PROVIDER_ERROR" and the raw timeout message

#### Scenario: Provider non-timeout mapping
- GIVEN a scraping operation fails with a 404 Not Found error
- WHEN the presenter processes the provider error
- THEN HTTP status 502 is returned
- AND the JSON response payload contains error code "PROVIDER_ERROR" and the raw error message

## Acceptance Criteria
- [ ] Provider error wraps raw scraping errors under "PROVIDER_ERROR".
- [ ] Timeout errors map to HTTP 504 containing the raw timeout message.
- [ ] Non-timeout provider errors map to HTTP 502 containing the raw message.
