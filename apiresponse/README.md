# apiresponse

Package `apiresponse` builds [RFC 9457](https://www.rfc-editor.org/rfc/rfc9457.html) **Problem Details** from `go-pkgs/errors` and maps domain codes ↔ HTTP status codes.

Success responses are **not** wrapped: return the resource JSON with `Content-Type: application/json`. Correlation belongs on the **`X-Correlation-ID`** header (see `go-pkgs/logging/correlation`), not in the JSON body.

## Install

```bash
export GOPRIVATE=github.com/gambitier/*
go get github.com/gambitier/go-pkgs/apiresponse@v0.1.0
```

Requires `github.com/gambitier/go-pkgs/errors`.

## When to use

- Emit `application/problem+json` instead of proprietary error envelopes
- Keep a single mapping between domain `Code` and HTTP status
- Ensure INTERNAL errors never leak implementation messages to clients

## Build a problem

```go
import (
    "net/http"

    "github.com/gambitier/go-pkgs/apiresponse"
    pkgerrors "github.com/gambitier/go-pkgs/errors"
)

err := pkgerrors.NotFound("item not found", nil, map[string]any{"id": id})
built := apiresponse.BuildProblem(err, apiresponse.BuildOptions{
    Instance: c.Path(), // request path or URI reference
})

c.Set("Content-Type", apiresponse.ContentTypeProblemJSON)
_ = c.Status(built.Status).JSON(built.Problem)

// Success: bare resource
_ = c.Status(http.StatusOK).JSON(itemDTO)
```

### `Problem` shape

```go
type Problem struct {
    Type     string         `json:"type"`
    Title    string         `json:"title"`
    Status   int            `json:"status"`
    Detail   string         `json:"detail,omitempty"`
    Instance string         `json:"instance,omitempty"`
    Code     string         `json:"code,omitempty"`   // extension: domain code
    Fields   map[string]any `json:"fields,omitempty"` // extension: domain fields
}
```

Example JSON:

```json
{
  "type": "about:blank",
  "title": "Not Found",
  "status": 404,
  "detail": "item not found",
  "instance": "/items/abc",
  "code": "NOT_FOUND",
  "fields": { "id": "abc" }
}
```

| Rule | Behavior |
|------|----------|
| Default `type` | `about:blank` (override via `BuildOptions.Type`) |
| `CodeInternal` | `detail` is always `"internal server error"` — never the domain `Message` or cause |
| Framework glue | Fiber/chi adapters live in the app (`internal/.../response`), not here |

Constants: `ContentTypeProblemJSON`, `DefaultProblemType`.

## Status ↔ code

```go
status := apiresponse.StatusFromCode(pkgerrors.CodeNotFound)           // 404
code := apiresponse.CodeFromHTTPStatus(http.StatusTooManyRequests)     // RATE_LIMITED
de := apiresponse.ToDomainError(http.StatusUnsupportedMediaType, "", cause)
```

| Function | Behavior |
|----------|----------|
| `StatusFromCode` | Unknown code → `500` |
| `CodeFromHTTPStatus` | Unknown 4xx → `INVALID_ARGUMENT`; else `INTERNAL` |
| `ToDomainError` | Build `*errors.Error` from status + optional message/cause |

## Notes

- Depends only on `go-pkgs/errors` (not `logging` / `observability`).
- Tags: `apiresponse/vX.Y.Z`.
- Register concrete problem type URIs later via `BuildOptions.Type` if your API needs them.
