# logging

Structured logging with independent logger instances, sinks, and correlation IDs.

## What / why

Use this package when you want:

- Per-service loggers that do not mutate the global logrus logger
- JSON/text format, level config, console/file sinks
- Correlation ID and optional trace fields on context-scoped loggers

Advantages vs raw logrus in every service: consistent config loading, sink wiring, and field conventions.

## Install

```bash
export GOPRIVATE=github.com/gambitier/*
go get github.com/gambitier/go-pkgs/logging@v0.1.0
```

## Usage

```go
import "github.com/gambitier/go-pkgs/logging"

logger, err := logging.New(logging.Config{
  ServiceName: "golang-service-template",
  Level:       "info",
  Format:      "json",
})
logger.Info("server started", logging.Fields{"port": 8080})
logger.Error("request failed", err, logging.Fields{"path": "/items"})
```

`BuildHTTPErrorLog` / `LogHTTPError` accept caller-supplied status/code/message. Domain-error mapping belongs in app glue.

## Config

Viper-friendly via `logging.InitConfig` / `logging.Config` (level, format, service_name, sinks). Typical YAML:

```yaml
logging:
  level: info
  format: json
  service_name: golang-service-template
  sinks:
    - type: console
      enabled: true
```

## Important notes

- Does **not** import `errors` or `observability`.
- `Logger.Error` records a basic `error` string; richer domain enrichment is app-side.
- Version with tags `logging/vX.Y.Z`.

## Composition

Wire OTel log export in `internal/platform` by adapting this logger to `observability.Logger` (AddHook + Warn).
