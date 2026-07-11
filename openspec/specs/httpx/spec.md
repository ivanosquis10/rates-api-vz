# HTTP Helpers Spec

## Purpose
Provide helper utilities for JSON writing and request context.

## Requirements

### Requirement: Write JSON response
The system SHALL define `httpx.WriteJSON` to render JSON.

#### Scenario: Write JSON
- GIVEN payload
- WHEN `httpx.WriteJSON(w, 200, payload)` is executed
- THEN status MUST be 200
- AND Content-Type header MUST be "application/json"

### Requirement: Request ID context
The system MUST extract Chi request ID from context.

#### Scenario: Get request ID
- GIVEN context with request ID "req-1"
- WHEN `httpx.GetRequestID(ctx)` is called
- THEN "req-1" MUST be returned
