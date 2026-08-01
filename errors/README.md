# errors

Structured domain/application errors and HTTP response mapping.

## What / why

Use this package when you want:

- Stable error **codes** for APIs and observability grouping
- Constant user-facing **messages** with dynamic data in **fields**
- Consistent HTTP envelopes (`success`, `message`, `code`, `fields`, `requestId`)

Advantages vs ad-hoc `fmt.Errorf` / status switches in every handler: one mapping table, stack-aware helpers, and a shared client contract.

## Install

```bash
export GOPRIVATE=github.com/gambitier/*
go get github.com/gambitier/go-pkgs/errors@v0.1.0
```

## Usage

```go
import (
  "github.com/gambitier/go-pkgs/errors/domainerr"
  "github.com/gambitier/go-pkgs/errors/httpresp"
)

err := domainerr.NotFound("item not found", map[string]any{"id": id})
_ = httpresp.Write(c, err, requestID) // Fiber/HTTP adapter in your presentation layer
```

Subpackages:

- `domainerr` — create/wrap typed errors
- `httpstatus` — map codes to HTTP status
- `httpresp` — build OK / error envelopes

## Config

None.

## Important notes

- Independent of `logging` and `observability`.
- Version with tags `errors/vX.Y.Z`.
- Prefer constant messages; put dynamic values in `Fields`.

## Composition

Enrich logs with domainerr fields in your app (`internal/platform`), not inside this module.
