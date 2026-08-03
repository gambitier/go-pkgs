# go-pkgs

Shared Go modules for services built from the hexagonal boilerplate.

Each package is an **independent Go module** (own `go.mod` / `go.sum`). There is no root module. Use `go.work` only for local multi-module development.

## Packages

| Package | Module path | Docs |
|---------|-------------|------|
| [errors](./errors) | `github.com/gambitier/go-pkgs/errors` | [README](./errors/README.md) |
| [apiresponse](./apiresponse) | `github.com/gambitier/go-pkgs/apiresponse` | [README](./apiresponse/README.md) |
| [logging](./logging) | `github.com/gambitier/go-pkgs/logging` | [README](./logging/README.md) |
| [observability](./observability) | `github.com/gambitier/go-pkgs/observability` | [README](./observability/README.md) |
| [lifecycle](./lifecycle) | `github.com/gambitier/go-pkgs/lifecycle` | [README](./lifecycle/README.md) |
| [mongodb](./mongodb) | `github.com/gambitier/go-pkgs/mongodb` | [README](./mongodb/README.md) |

`apiresponse` depends on `errors`. `mongodb` implements `lifecycle.Component` by structural typing (no module dependency). `logging` and `observability` stay independent of the others. Wire packages together in your service (for example `internal/platform`).

## Tagging

Tags are per package:

```
errors/v0.2.0
apiresponse/v0.1.0
logging/v0.1.0
observability/v0.1.0
lifecycle/v0.1.0
mongodb/v0.1.0
```

Cut and bump tags independently.

```bash
export GOPRIVATE=github.com/gambitier/*
go get github.com/gambitier/go-pkgs/errors@v0.2.0
go get github.com/gambitier/go-pkgs/apiresponse@v0.1.0
go get github.com/gambitier/go-pkgs/lifecycle@v0.1.0
go get github.com/gambitier/go-pkgs/mongodb@v0.1.0
```

## Make

```bash
make test   # go test in each module
make tidy   # go mod tidy in each module
make vet    # go vet in each module
make fmt    # gofmt every module
make hooks  # enable .githooks (runs make fmt on every commit)
make check  # vet + test
```

After clone, run `make hooks` once so commits auto-format via the pre-commit hook.
