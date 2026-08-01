# observability

OpenTelemetry setup for traces, metrics, and logs (OTLP/gRPC), plus Fiber and gorilla/mux helpers.

## What / why

Use this package when you want:

- YAML-driven OTLP exporters without scattering SDK setup
- HTTP/gRPC instrumentation with health/swagger path filters
- Optional logrus → OTLP log bridge via a small `Logger` interface

Advantages vs copy-pasting otel SDK init: one config shape, shared middleware, and a disable-by-default path when `opentel` is absent.

## Install

```bash
export GOPRIVATE=github.com/gambitier/*
go get github.com/gambitier/go-pkgs/observability@v0.1.0
```

## Usage

```go
import commonobservability "github.com/gambitier/go-pkgs/observability"

cfg, err := commonobservability.InitConfig("golang-service-template")
shutdown, err := commonobservability.Init(ctx, cfg, otelLogger) // otelLogger may be nil
defer shutdown(context.Background())

app.Use(commonobservability.FiberMiddleware("golang-service-template-http"))
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

```yaml
opentel:
  enabled: false
  collector_url: "localhost:4317"
  service_name: "golang-service-template"
  insecure: true
  sampling:
    ratio: 1.0
  resource:
    deployment_environment: "development"
    service_version: "1.0.0"
```

Exporter endpoints come from YAML only (`collector_url`, optional per-signal endpoints). `OTEL_*` env vars do not override exporters. Missing `opentel` section ⇒ telemetry disabled.

## Important notes

- Does **not** import `logging` or `errors`.
- Health/swagger paths skipped for HTTP spans: `/health`, `/livez`, `/healthz`, `/swagger`.
- Version with tags `observability/vX.Y.Z`.

## Composition

Build logger + call `Init` + register Fiber middleware from the consuming service’s `internal/platform` / `main`.
