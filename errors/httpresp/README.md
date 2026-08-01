# `httpresp` package

Transport-agnostic helpers for converting application errors into a stable HTTP response envelope.

This package is intentionally small and focused:
- map `domainerr` codes to HTTP status codes
- build success/error envelopes
- enforce safe client-facing error messages

## Response shape

Error responses use:

```json
{
  "error": {
    "errorCode": "NOT_FOUND",
    "message": "customer not found",
    "fields": {
      "resource": "customer"
    },
    "requestId": "req-123"
  },
  "requestId": "req-123"
}
```

Success responses use:

```json
{
  "data": { "...": "..." },
  "requestId": "req-123"
}
```

## Public API

- `BuildError(err, requestID)` -> `ErrorBuildResult`
- `BuildErrorEnvelope(err, requestID)` -> `(status, envelope, code, msg, fields)`
- `BuildOKEnvelope(data, requestID)`
- `BuildCreatedEnvelope(data, requestID)`
- `StatusFromCode(code)`

## Mapping rules

### Domain error (`domainerr.As(err)` succeeds)

- status is derived from `domainerr.Code`
- `errorCode` is the domain code
- `fields` are taken from domain error fields
- message behavior:
  - for non-internal codes: use domain error message
  - for `INTERNAL`: always return `"internal server error"` (sanitized)

### Non-domain error

- fallback to:
  - status: `500`
  - code: `INTERNAL`
  - message: `"internal server error"`
  - fields: `nil`

## Wrapping semantics validated by tests

The tests in `httpresp_test.go` define expected behavior for common chains:

- domain error maps directly (`INVALID_ARGUMENT` -> 400 with message/fields)
- unknown raw error falls back to generic internal 500
- `domainerr.Internal(...)` always returns generic internal client message
- `domainerr.Wrap(...)` preserves original domain code/message for response
- `fmt.Errorf("...: %w", domainErr)` also preserves domain mapping
- multi-level domain error chains resolve to the top-level domain message/code
- mixed chains (internal at lower layer, not found at higher layer) resolve to higher-layer semantic code/message

## Recommended usage pattern

1. Convert infra errors to `domainerr` in repository/application layers.
2. Add context in higher layers (`domainerr.Wrap` or explicit higher-level domain error when semantics change).
3. At HTTP boundary, call service-local wrapper that delegates to this package.
4. Do not return raw infra/internal messages to clients.

## Notes

- This package is transport-agnostic (no Fiber/HTTP framework dependency).
- Framework-specific writing/logging belongs in each service wrapper (e.g. `internal/presentation/http/response`).
