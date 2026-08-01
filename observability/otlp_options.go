package commonobservability

import (
	"strings"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
)

func signalClientOptions(cfg runtimeConfig, signalEndpoint string) []otlptracegrpc.Option {
	endpoint := endpointOrDefault(signalEndpoint, cfg.Endpoint)
	opts := []otlptracegrpc.Option{}
	if endpoint != "" {
		opts = append(opts, otlptracegrpc.WithEndpoint(endpoint))
	}
	if cfg.Insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}
	if len(cfg.Headers) > 0 {
		opts = append(opts, otlptracegrpc.WithHeaders(cfg.Headers))
	}
	return opts
}

func signalMetricClientOptions(cfg runtimeConfig, signalEndpoint string) []otlpmetricgrpc.Option {
	endpoint := endpointOrDefault(signalEndpoint, cfg.Endpoint)
	opts := []otlpmetricgrpc.Option{}
	if endpoint != "" {
		opts = append(opts, otlpmetricgrpc.WithEndpoint(endpoint))
	}
	if cfg.Insecure {
		opts = append(opts, otlpmetricgrpc.WithInsecure())
	}
	if len(cfg.Headers) > 0 {
		opts = append(opts, otlpmetricgrpc.WithHeaders(cfg.Headers))
	}
	return opts
}

func signalLogClientOptions(cfg runtimeConfig, signalEndpoint string) []otlploggrpc.Option {
	endpoint := endpointOrDefault(signalEndpoint, cfg.Endpoint)
	opts := []otlploggrpc.Option{}
	if endpoint != "" {
		opts = append(opts, otlploggrpc.WithEndpoint(endpoint))
	}
	if cfg.Insecure {
		opts = append(opts, otlploggrpc.WithInsecure())
	}
	if len(cfg.Headers) > 0 {
		opts = append(opts, otlploggrpc.WithHeaders(cfg.Headers))
	}
	return opts
}

func endpointOrDefault(specific, fallback string) string {
	specific = strings.TrimSpace(specific)
	if specific != "" {
		return specific
	}
	return strings.TrimSpace(fallback)
}
