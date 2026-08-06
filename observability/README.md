# observability

Package `observability` bootstraps **OpenTelemetry** (traces, metrics, logs over OTLP/gRPC) and exposes small stdlib HTTP / gRPC instrumentation helpers.

This package does **not** load YAML/env/viper config. The service builds a `Config` and passes it to `Init`. Framework middleware (Fiber, etc.) stays in the consuming app.

## Install

```bash
export GOPRIVATE=github.com/gambitier/*
go get github.com/gambitier/go-pkgs/observability@v0.3.0
```

## When to use

- Shared OTLP SDK setup instead of duplicating providers in every service
- Instrument `net/http` and `google.golang.org/grpc` with health/swagger path filters
- Optionally bridge logrus → OTLP logs via a tiny local `Logger` interface

## Init

```go
import "github.com/gambitier/go-pkgs/observability"

shutdown, err := observability.Init(ctx, observability.Config{
    Enabled:      true,
    ServiceName:  "TradingOps.IdentityService",
    CollectorURL: "localhost:4317",
    Insecure:     true,
    Sampling:     observability.SamplingConfig{Ratio: 1.0},
}, otelLogger) // otelLogger may be nil
if err != nil {
    return err
}
defer shutdown(context.Background())
```

When `Enabled` is `false`, `Init` is a no-op and returns a nil shutdown function.

### Config

| Field | Purpose |
|-------|---------|
| `Enabled` | Master switch |
| `CollectorURL` | Default OTLP/gRPC endpoint |
| `TracesURL` / `MetricsURL` / `LogsURL` | Optional per-signal overrides |
| `ServiceName` | Resource `service.name` |
| `Insecure` | Disable TLS on exporters |
| `Sampling.Ratio` | Trace head sampling; clamped to `(0,1]`; `≤0` → `1.0` |
| `Resource.DeploymentEnvironment` / `ServiceVersion` | Extra resource attributes |
| `Headers` | OTLP headers |

### Optional log bridge

```go
type Logger interface {
    AddHook(hook logrus.Hook)
    Warn(message string, fields map[string]any)
}
```

This is **not** `go-pkgs/logging`. Adapt your app logger in `internal/shared/platform` (e.g. wrap `logging.Logger` so OTel can attach a logrus hook). Pass `nil` to skip the bridge.

## Instrumentation helpers

No-ops when observability is not enabled (`!IsEnabled()`).

```go
handler = observability.WrapHTTPHandler("identity-http", handler)
client := observability.NewHTTPClient(http.DefaultClient)

srv := grpc.NewServer(observability.GRPCServerOptions()...)
conn, err := grpc.NewClient(target, observability.GRPCDialOptions()...)
```

| Helper | Role |
|--------|------|
| `WrapHTTPHandler` | otelhttp server middleware |
| `NewHTTPClient` | otelhttp round-tripper on a base client |
| `GRPCServerOptions` / `GRPCDialOptions` | otelgrpc interceptors |
| `IsEnabled` | Whether `Init` installed providers |

HTTP spans skip: `/health`, `/livez`, `/healthz`, `/swagger`, `/swagger/*`.

## Notes

- Does **not** import `logging` or `errors`.
- Tags: `observability/vX.Y.Z`.
- Typical process wiring: `lifecycle` component that calls `Init` on Start and `shutdown` on Stop (see service `OTelComponent`).
