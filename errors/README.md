# errors

Package `errors` defines **domain/application errors**: stable codes, opaque client-safe messages, structured fields, and helpers for logging wrapped causes without leaking internals over HTTP.

HTTP Problem Details and status mapping live in [`apiresponse`](../apiresponse) ([RFC 9457](https://www.rfc-editor.org/rfc/rfc9457.html)), not here.

## Install

```bash
export GOPRIVATE=github.com/gambitier/*
go get github.com/gambitier/go-pkgs/errors@v0.4.0
```

Import the module root (there is no `domainerr` subpackage). Alias if you also import the standard library:

```go
import pkgerrors "github.com/gambitier/go-pkgs/errors"
```

## When to use

- Return stable **codes** (`NOT_FOUND`, `INVALID_ARGUMENT`, …) for APIs and metrics
- Keep **messages** constant and client-safe; put dynamic values in **Fields**
- Log infrastructure causes (S3, Mongo, timeouts) without changing what `err.Error()` returns to clients

## Core type

```go
type Error struct {
    Code    Code           // stable machine-readable code
    Message string         // opaque client-safe text; never concatenate causes
    Err     error          // wrapped cause (stack preserved via cockroachdb/errors)
    Fields  map[string]any // dynamic attributes for APIs and logs
}
```

| Method / rule | Behavior |
|---------------|----------|
| `Error()` | Returns `Message` only — never the wrapped cause |
| `Unwrap()` | Returns `Err` |
| `Is(*Error)` | Matches by `Code` |

Prefer constant messages. Put IDs, field names, and other dynamic data in `Fields`.

## Creating errors

```go
err := pkgerrors.NotFound("item not found", nil, map[string]any{"id": id})
err = pkgerrors.InvalidArgumentWithFields("name is required", map[string]any{"field": "name"})
err = pkgerrors.Internal("failed to list items", cause, nil)
```

Common constructors: `InvalidArgument`, `NotFound`, `Unauthorized`, `Forbidden`, `Conflict`, `RateLimited`, `Internal`, plus HTTP-shaped helpers (`MethodNotAllowed`, `UnprocessableEntity`, …). Use `NewFromCode` for less common codes.

Fluent builders are available when you prefer chaining:

```go
err := pkgerrors.NotFoundError().
    Message("item not found").
    Fields(map[string]any{"id": id})
```

## Inspecting and wrapping

```go
if pkgerrors.Is(err, pkgerrors.CodeNotFound) { /* ... */ }
code := pkgerrors.CodeOf(err)
de, ok := pkgerrors.As(err)

// Preserve domain code while adding a safe outer message
wrapped := pkgerrors.Wrap(err, "failed to load item")
```

| Helper | Purpose |
|--------|---------|
| `CauseChain(err)` | Human-readable wrap chain for logs |
| `RootCause(err)` | Deepest leaf message |
| `ErrorContext(err)` | Extra context string when present |
| `CollectFields(err)` | Merge `Fields` along the unwrap chain (outer wins) |

## Logging: `LogFields`

`LogFields(err)` returns a typed `LogAttrs` view for structured logging. This package does **not** import `logging`.

```go
type LogAttrs struct {
    Error   string         // preferred detail (often CauseChain)
    Cause   string         // RootCause when it differs from Error()
    Context string
    Code    string         // domain code when present
    Message string         // opaque domain message
    Source  string         // one-line stack source when available
    Fields  map[string]any // CollectFields
    Stack   []string       // AppStackTraceLines for INTERNAL only
}

attrs := pkgerrors.LogFields(err)
logger.Error(attrs.Message, logging.Fields(attrs.Map()))
```

`Map()` produces keys: `error`, `error_cause`, `error_context`, `error_code`, `error_msg`, `error_source`, `stack_trace`, plus field entries. Returns `nil` when empty.

## Stack helpers

For INTERNAL diagnostics: `OneLineSource`, `StackFrames`, `AppStackTraceLines`, `AppStackTraceString`, plus `StackFrameOptions` (include/strip prefixes, max depth). Defaults filter to the main module path when possible.

## Notes

- Independent of `logging`, `observability`, and `apiresponse` (`apiresponse` depends on this module).
- Tags: `errors/vX.Y.Z`.
- Map to HTTP with `go-pkgs/apiresponse.BuildProblem`.
