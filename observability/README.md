# observability

OpenTelemetry setup for traces, metrics, and logs (OTLP/gRPC), plus stdlib HTTP and gRPC helpers.

## What / why

Use this package when you want:

- Shared OTLP SDK bootstrap without scattering provider setup
- HTTP/gRPC instrumentation with health/swagger path filters
- Optional logrus → OTLP log bridge via a small `Logger` interface

This package does **not** load config (YAML/JSON/env/viper). The service builds a `Config` and passes it to `Init`.

## Install

```bash
export GOPRIVATE=github.com/gambitier/*
go get github.com/gambitier/go-pkgs/observability@v0.3.0
```

## Usage

```go
import "github.com/gambitier/go-pkgs/observability"

cfg := observability.Config{
  Enabled:      true,
  ServiceName:  "golang-service-template",
  CollectorURL: "localhost:4317",
  Insecure:     true,
  Sampling:     observability.SamplingConfig{Ratio: 1.0},
}
shutdown, err := observability.Init(ctx, cfg, otelLogger) // otelLogger may be nil
defer shutdown(context.Background())

// Framework middleware (e.g. Fiber) belongs in the consuming service.
handler = observability.WrapHTTPHandler("my-service-http", handler)
```

`Logger` interface (package-local — not the logging module):

```go
type Logger interface {
  AddHook(hook logrus.Hook)
  Warn(message string, fields map[string]any)
}
```

Adapt your app logger in `internal/platform`.

## Config

`Config` fields the caller fills (names are hints for mapstructure-style unmarshaling in the service, not a loader):

| Field | Purpose |
|-------|---------|
| `Enabled` | When false, `Init` is a no-op |
| `CollectorURL` | Default OTLP/gRPC endpoint |
| `TracesURL` / `MetricsURL` / `LogsURL` | Optional per-signal overrides |
| `ServiceName` | Resource `service.name` |
| `Insecure` | bool; TLS off for exporters when true |
| `Sampling.Ratio` | Trace head sampling (clamped to `(0,1]`; default 1.0) |
| `Resource.*` | Extra resource attributes |
| `Headers` | OTLP headers |

## Important notes

- Does **not** import `logging` or `errors`.
- Health/swagger paths skipped for HTTP spans: `/health`, `/livez`, `/healthz`, `/swagger`.
- Version with tags `observability/vX.Y.Z`.

## Composition

Service loads its own config → builds `observability.Config` → `Init`. Register HTTP framework middleware (Fiber, etc.) in the service.
