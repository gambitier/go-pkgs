# errors

Structured domain/application errors (stable codes, opaque client messages, fields).

## What / why

Use this package when you want:

- Stable error **codes** for APIs and observability grouping
- Constant user-facing **messages** with dynamic data in **fields**
- Stack-aware wrap/cause helpers for logging without leaking internals to clients

HTTP Problem Details and status mapping live in [`apiresponse`](../apiresponse) ([RFC 9457](https://www.rfc-editor.org/rfc/rfc9457.html)), not here.

## Install

```bash
export GOPRIVATE=github.com/gambitier/*
go get github.com/gambitier/go-pkgs/errors@v0.2.0
```

Import path is the module root (no `domainerr` subpackage):

```go
import pkgerrors "github.com/gambitier/go-pkgs/errors"

err := pkgerrors.NotFound("item not found", nil, map[string]any{"id": id})
```

Alias the import if you also use the standard library `errors` package.

## Usage

```go
import pkgerrors "github.com/gambitier/go-pkgs/errors"

err := pkgerrors.InvalidArgumentWithFields("name is required", map[string]any{"field": "name"})
if pkgerrors.Is(err, pkgerrors.CodeInvalidArgument) {
  // ...
}
```

## Config

None.

## Important notes

- Independent of `logging`, `observability`, and `apiresponse` (apiresponse depends on this module).
- Version with tags `errors/vX.Y.Z`.
- Prefer constant messages; put dynamic values in `Fields`.

## Composition

Map to HTTP with `go-pkgs/apiresponse`. Enrich logs in your app (`internal/platform`), not inside this module.
