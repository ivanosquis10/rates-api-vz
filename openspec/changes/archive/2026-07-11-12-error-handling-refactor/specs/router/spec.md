# Delta Spec: Router

## Requirements

### Requirement: Route Delegation

The router SHALL register the following routes: `GET /rates`, `GET /rates/history`, and `POST /admin/scrape` (within `/admin` path group). All requests to these endpoints MUST route to their respective handler methods. If a client requests a non-existent route, the router MUST return a structured JSON response matching the new error envelope with HTTP status 404 and code `NOT_FOUND`.

(Previously: The router SHALL register the following routes: `GET /rates`, `GET /rates/history`, and `POST /admin/scrape` (within `/admin` path group). All requests to these endpoints MUST route to their respective handler methods.)

#### Scenario: All routes are dispatched
- GIVEN a running server using the decoupled router
- WHEN a client requests GET /rates
- THEN the router dispatches the request to the rates handler

#### Scenario: Route non-existent endpoint
- GIVEN a running server using the decoupled router
- WHEN a client requests a non-existent route
- THEN the router MUST return HTTP status 404
- AND the response MUST be a JSON error envelope with `success` false, `code` "NOT_FOUND", error message, and a non-empty `request_id`
- AND the response MUST NOT contain a `data` field
