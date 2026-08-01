package commonobservability

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/bridges/otellogrus"
	runtimeinstrumentation "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
)

const initTimeout = 30 * time.Second

var telemetryEnabled atomic.Bool

// Logger is an optional sink for OTLP log bridging and init warnings.
// Implementations need not come from any sibling go-pkgs module; the app
// typically adapts its logging package in platform glue.
type Logger interface {
	AddHook(hook logrus.Hook)
	Warn(message string, fields map[string]any)
}

// Init configures global OpenTelemetry providers. Pass nil logger to skip the
// logrus OTLP bridge and runtime-metric warnings.
func Init(ctx context.Context, cfg Config, logger Logger) (func(context.Context) error, error) {
	if !cfg.Enabled {
		telemetryEnabled.Store(false)
		return func(context.Context) error { return nil }, nil
	}

	initCtx, cancel := context.WithTimeout(ctx, initTimeout)
	defer cancel()

	res, err := buildResource(initCtx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create otel resource: %w", err)
	}

	traceExporter, err := otlptracegrpc.New(initCtx, traceClientOptions(cfg)...)
	if err != nil {
		return nil, fmt.Errorf("create trace exporter: %w", err)
	}
	meterExporter, err := otlpmetricgrpc.New(initCtx, metricClientOptions(cfg)...)
	if err != nil {
		return nil, fmt.Errorf("create metric exporter: %w", err)
	}
	logExporter, err := otlploggrpc.New(initCtx, logClientOptions(cfg)...)
	if err != nil {
		return nil, fmt.Errorf("create log exporter: %w", err)
	}

	sampler := sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.SamplingRatio))

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sampler),
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(meterExporter)),
		metric.WithResource(res),
	)
	loggerProvider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
		sdklog.WithResource(res),
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetMeterProvider(meterProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	global.SetLoggerProvider(loggerProvider)
	telemetryEnabled.Store(true)

	if logger != nil {
		logger.AddHook(otellogrus.NewHook(
			cfg.ServiceName,
			otellogrus.WithLoggerProvider(loggerProvider),
		))
	}

	if err := runtimeinstrumentation.Start(runtimeinstrumentation.WithMeterProvider(meterProvider)); err != nil && logger != nil {
		logger.Warn("failed to start runtime metrics", map[string]any{"error": err.Error()})
	}

	return func(shutdownCtx context.Context) error {
		var joined error
		if err := loggerProvider.Shutdown(shutdownCtx); err != nil {
			joined = errors.Join(joined, err)
		}
		if err := meterProvider.Shutdown(shutdownCtx); err != nil {
			joined = errors.Join(joined, err)
		}
		if err := tracerProvider.Shutdown(shutdownCtx); err != nil {
			joined = errors.Join(joined, err)
		}
		telemetryEnabled.Store(false)
		return joined
	}, nil
}

func IsEnabled() bool {
	return telemetryEnabled.Load()
}

func buildResource(ctx context.Context, cfg Config) (*resource.Resource, error) {
	attrs := []resource.Option{
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithHost(),
		resource.WithContainer(),
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithAttributes(semconv.ServiceName(cfg.ServiceName)),
	}
	if cfg.DeploymentEnvironment != "" {
		attrs = append(attrs, resource.WithAttributes(semconv.DeploymentEnvironmentName(cfg.DeploymentEnvironment)))
	}
	if cfg.ServiceVersion != "" {
		attrs = append(attrs, resource.WithAttributes(semconv.ServiceVersion(cfg.ServiceVersion)))
	}
	return resource.New(ctx, attrs...)
}

func traceClientOptions(cfg Config) []otlptracegrpc.Option {
	return signalClientOptions(cfg, cfg.TracesEndpoint)
}

func metricClientOptions(cfg Config) []otlpmetricgrpc.Option {
	return signalMetricClientOptions(cfg, cfg.MetricsEndpoint)
}

func logClientOptions(cfg Config) []otlploggrpc.Option {
	return signalLogClientOptions(cfg, cfg.LogsEndpoint)
}
