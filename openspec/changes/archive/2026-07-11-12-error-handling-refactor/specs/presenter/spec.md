# Delta Spec: Presenter

## Requirements

### Requirement: Enveloping and Mapping

The presenter SHALL provide helpers `OK`, `Created`, `NoContent`, and `Error` wrapping responses. Success responses MUST include `success`, `data`, and `request_id` fields. Error responses MUST include `success`, `code` (renamed from `error_code`), `error`, and `request_id` fields, and MUST omit the `data` field (using `omitempty`).

(Previously: The presenter SHALL provide helpers `OK`, `Created`, `NoContent`, and `Error` wrapping responses in: `success`, `data`, `error_code`, `error`, `request_id`.)

#### Scenario: Present OK envelope
- GIVEN data payload and request ID "req-2"
- WHEN `presenter.OK(w, r, data)` is executed
- THEN status MUST be 200
- AND response MUST be success envelope containing payload and "req-2"

#### Scenario: Map domain error
- GIVEN a domain not found error
- WHEN `presenter.Error(w, r, err)` is executed
- THEN status MUST be 404
- AND response MUST be error envelope with `success` false, `code` "NOT_FOUND", error message, and a non-empty `request_id`
- AND the response MUST NOT contain a `data` field
