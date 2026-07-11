# Delta for presenter

## MODIFIED Requirements

### Requirement: Enveloping and Mapping

The presenter SHALL provide helpers `OK`, `Created`, `NoContent`, and `Error` wrapping responses. Success responses MUST include `success` and `data` fields. Error responses MUST include `success`, `code`, and `error` fields, and MUST omit the `data` field (using `omitempty`). The JSON response envelope MUST NOT contain the `request_id` field. Instead, all successful and failed HTTP responses MUST carry the `X-Request-ID` HTTP header containing the request's ID from the context.

(Previously: The presenter SHALL provide helpers `OK`, `Created`, `NoContent`, and `Error` wrapping responses. Success responses MUST include `success`, `data`, and `request_id` fields. Error responses MUST include `success`, `code` (renamed from `error_code`), `error`, and `request_id` fields, and MUST omit the `data` field (using `omitempty`).)

#### Scenario: Present OK envelope
- GIVEN a request context containing request ID "req-2" and a data payload
- WHEN `presenter.OK(w, r, data)` is executed
- THEN the HTTP response status code MUST be 200
- AND the response body MUST be a JSON success envelope containing `success` true and `data` matching the payload
- AND the response body MUST NOT contain the `request_id` field
- AND the HTTP response MUST carry the `X-Request-ID` header set to "req-2"

#### Scenario: Map domain error
- GIVEN a request context containing request ID "req-err-1" and a domain not found error
- WHEN `presenter.Error(w, r, err)` is executed
- THEN the HTTP response status code MUST be 404
- AND the response body MUST be a JSON error envelope containing `success` false, `code` "NOT_FOUND", and the error message
- AND the response body MUST NOT contain the `request_id` field
- AND the response body MUST NOT contain a `data` field
- AND the HTTP response MUST carry the `X-Request-ID` header set to "req-err-1"
