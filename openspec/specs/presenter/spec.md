# Presenter Spec

## Purpose
Enveloping responses and mapping domain errors.

## Requirements

### Requirement: Enveloping and Mapping
The presenter SHALL provide helpers `OK`, `Created`, `NoContent`, and `Error` wrapping responses in: `success`, `data`, `error_code`, `error`, `request_id`.

#### Scenario: Present OK envelope
- GIVEN data payload and request ID "req-2"
- WHEN `presenter.OK(w, r, data)` is executed
- THEN status MUST be 200
- AND response MUST be success envelope containing payload and "req-2"

#### Scenario: Map domain error
- GIVEN a domain not found error
- WHEN `presenter.Error(w, r, err)` is executed
- THEN status MUST be 404
- AND response MUST be error envelope with code "NOT_FOUND"
