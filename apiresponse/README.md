# apiresponse

HTTP API response helpers: [RFC 9457](https://www.rfc-editor.org/rfc/rfc9457.html) Problem Details for errors, plus domain-code ↔ HTTP status mapping.

## What / why

Use this package when you want:

- Standards-based error bodies (`application/problem+json`) instead of proprietary envelopes
- Stable mapping between `github.com/gambitier/go-pkgs/errors` codes and HTTP status codes
- Safe client-facing details (internal errors never leak implementation messages)

Success responses are **not** wrapped: return the resource JSON directly with `Content-Type: application/json`. Correlation belongs on the **`X-Correlation-ID`** header (see `go-pkgs/logging/correlation`), not in the JSON body.

## Install

```bash
export GOPRIVATE=github.com/gambitier/*
go get github.com/gambitier/go-pkgs/apiresponse@v0.1.0
```

Requires `github.com/gambitier/go-pkgs/errors`.

## Usage

```go
import (
  "net/http"

  "github.com/gambitier/go-pkgs/apiresponse"
  pkgerrors "github.com/gambitier/go-pkgs/errors"
)

err := pkgerrors.NotFound("item not found", nil, map[string]any{"id": id})
built := apiresponse.BuildProblem(err, apiresponse.BuildOptions{Instance: c.Path()})

c.Set("Content-Type", apiresponse.ContentTypeProblemJSON)
_ = c.Status(built.Status).JSON(built.Problem)

// Success: bare resource body
_ = c.Status(http.StatusOK).JSON(itemDTO)
```

Example problem JSON:

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

`code` and `fields` are RFC 9457 extension members for domain contracts.

## Config

None.

## Important notes

- Independent module; depends only on `go-pkgs/errors` (not logging/observability).
- Version with tags `apiresponse/vMAJOR.MINOR.PATCH`.
- Default `type` is `about:blank`; register concrete problem type URIs later if needed.

## Composition

Fiber (or other) adapters that set status, content type, and JSON live in the consuming app (e.g. `internal/presentation/http/response`), not in this module.
