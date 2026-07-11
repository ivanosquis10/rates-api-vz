# API Errors Spec

## Purpose
Define API error catalog and error representation.

## Requirements

### Requirement: Error Code representation
The system SHALL define `apierrors.Error` and standard codes: `UNAUTHORIZED`, `RATE_LIMITED`, `NOT_FOUND`, `BAD_REQUEST`, `INTERNAL_ERROR`.

#### Scenario: Build error
- GIVEN code NOT_FOUND
- WHEN `apierrors.New(NOT_FOUND, "msg")` is called
- THEN error code MUST be "NOT_FOUND"
- AND message MUST be "msg"
