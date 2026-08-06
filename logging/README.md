# logging

Package `logging` provides **structured loggers** with independent instances (no global logrus mutation), console/file sinks, correlation IDs, and optional OpenTelemetry trace field attachment.

This package does **not** load YAML/env/viper config. The service builds a `Config` and passes it to `New`.

## Install

```bash
export GOPRIVATE=github.com/gambitier/*
go get github.com/gambitier/go-pkgs/logging@v0.2.0
```

## When to use

- One logger per service/process that does not touch the process-global logrus logger
- JSON or text output to console and/or rotating files
- Propagate **correlation IDs** (`X-Correlation-ID`) and OTel `trace_id` / `span_id` on request-scoped loggers

## Create a logger

```go
import "github.com/gambitier/go-pkgs/logging"

logger, err := logging.New(logging.Config{
    ServiceName: "TradingOps.IdentityService",
    Level:       "info",  // debug, info, warn, error, fatal
    Format:      "json",  // json | text
})
if err != nil {
    return err
}

logger.Info("server started", logging.Fields{"port": 8080})
logger.Error("request failed", err, logging.Fields{"path": "/items"})
```

`NewDefault(serviceName)` builds a console JSON logger at info for early bootstrap failures.

### `Logger` API

```go
type Logger interface {
    WithFields(fields Fields) Logger
    WithCorrelationID(id string) Logger
    WithContext(ctx context.Context) Logger // correlation + TraceFields
    AddHook(hook logrus.Hook)
    Debug(message string, fields Fields)
    Info(message string, fields Fields)
    Warn(message string, fields Fields)
    Error(message string, err error, fields Fields)
    Fatal(message string, fields Fields)
}

type Fields map[string]any
```

`Error` adds a basic `"error"` field from `err.Error()` when missing. Richer domain enrichment (`error_code`, stacks, cause chain) belongs in app glue via `errors.LogFields(err).Map()`.

## Config and sinks

```go
type Config struct {
    ServiceName string
    Level       string
    Format      string // json | text
    Sinks       []SinkConfig
    BaseFields  Fields
    Output      io.Writer // optional override; usually leave nil
    // ErrorFrame* reserved for caller-supplied stack frame filtering in HTTP helpers
}

type SinkConfig struct {
    Type    string // console | file
    Enabled bool
    File    FileSinkConfig
}
```

| Sink | Behavior |
|------|----------|
| `console` | stdout |
| `file` | lumberjack rotation (`Path`, `MaxSizeMB`, `MaxBackups`, `MaxAgeDays`, `Compress`) |
| none | default stdout |

Optional `mapstructure` tags on `Config` exist for service-side unmarshaling only.

## Correlation

Subpackage `github.com/gambitier/go-pkgs/logging/correlation`:

```go
import "github.com/gambitier/go-pkgs/logging/correlation"

ctx, id := correlation.EnsureCorrelationID(r.Context(), r.Header.Get(correlation.HeaderName))
w.Header().Set(correlation.HeaderName, id)
scoped := logger.WithContext(ctx)
```

| API | Behavior |
|-----|----------|
| `HeaderName` | `"X-Correlation-ID"` |
| `GetCorrelationID` / `SetCorrelationID` | Context accessors |
| `EnsureCorrelationID` | Reuse incoming ID or generate a UUID |

`WithContext` also attaches `TraceFields(ctx)` (`trace_id`, `span_id`) when a valid OTel span is present.

## HTTP error logging

```go
payload := logging.BuildHTTPErrorLog(err, logging.HTTPErrorLogInput{
    Method:     r.Method,
    Path:       r.URL.Path,
    StatusCode: status,
    ErrorCode:  "NOT_FOUND",
    ErrorMsg:   "item not found",
})
logging.LogHTTPError(logger, err, payload) // 5xx → Error, 4xx → Warn
```

Domain-code mapping and stack enrichment are caller-supplied; this package stays free of `go-pkgs/errors`.

## Notes

- Does **not** import `errors` or `observability`.
- Tags: `logging/vX.Y.Z`.
- Bridge to OTLP logs in the app by adapting this logger to `observability.Logger` (`AddHook` + `Warn`).
