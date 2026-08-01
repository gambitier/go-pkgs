# go-pkgs

Shared Go modules for services built from the hexagonal boilerplate.

Each package is an **independent Go module** (own `go.mod` / `go.sum`). There is no root module. Use `go.work` only for local multi-module development.

## Packages

| Package | Module path | Docs |
|---------|-------------|------|
| [errors](./errors) | `github.com/gambitier/go-pkgs/errors` | [README](./errors/README.md) |
| [logging](./logging) | `github.com/gambitier/go-pkgs/logging` | [README](./logging/README.md) |
| [observability](./observability) | `github.com/gambitier/go-pkgs/observability` | [README](./observability/README.md) |

These packages do **not** depend on each other. Wire them together in your service (for example `internal/platform`).

## Tagging

Tags are per package:

```
errors/v0.1.0
logging/v0.1.0
observability/v0.1.0
```

Cut and bump tags independently.

```bash
export GOPRIVATE=github.com/gambitier/*
go get github.com/gambitier/go-pkgs/errors@v0.1.0
go get github.com/gambitier/go-pkgs/logging@v0.1.0
go get github.com/gambitier/go-pkgs/observability@v0.1.0
```
