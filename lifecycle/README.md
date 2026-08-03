# lifecycle

Minimal ordered Start / reverse Stop runner for long-lived process components.

## What / why

Use this when services share the same bootstrap pattern (Mongo → OTel → HTTP) and you want a tiny stdlib runner instead of a DI framework (uber/fx, etc.).

## Install

```bash
export GOPRIVATE=github.com/gambitier/*
go get github.com/gambitier/go-pkgs/lifecycle@v0.1.0
```

## Usage

```go
app := lifecycle.New(10 * time.Second)
app.Add(mongoComp, otelComp, httpComp)
return app.Run(ctx) // blocks until ctx cancelled, then Stop reverse-order
```

Implement `lifecycle.Component` (`Name` / `Start` / `Stop`) in adapters. Signal handling stays in `cmd` (`signal.NotifyContext`).

## Important notes

- Stdlib only — no logging or framework deps.
- On Start failure, already-started components are stopped in reverse order.
- This is process glue, not a DI container.
