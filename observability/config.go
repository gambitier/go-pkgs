package commonobservability

import (
	"strings"
)

// Config is the observability settings the caller builds (from any source) and
// passes to Init. This package does not load config files or environment.
type Config struct {
	Enabled      bool              `mapstructure:"enabled"`
	CollectorURL string            `mapstructure:"collector_url"`
	ServiceName  string            `mapstructure:"service_name"`
	Insecure     bool              `mapstructure:"insecure"`
	InsecureMode string            `mapstructure:"insecure_mode"` // legacy string "true"/"false"
	Sampling     SamplingConfig    `mapstructure:"sampling"`
	Resource     ResourceConfig    `mapstructure:"resource"`
	TracesURL    string            `mapstructure:"traces_endpoint"`
	MetricsURL   string            `mapstructure:"metrics_endpoint"`
	LogsURL      string            `mapstructure:"logs_endpoint"`
	Headers      map[string]string `mapstructure:"headers"`
}

// SamplingConfig controls trace head sampling.
type SamplingConfig struct {
	Ratio float64 `mapstructure:"ratio"`
}

// ResourceConfig adds resource attributes beyond service.name.
type ResourceConfig struct {
	DeploymentEnvironment string `mapstructure:"deployment_environment"`
	ServiceVersion        string `mapstructure:"service_version"`
}

// runtimeConfig is the normalized shape used inside Init and exporters.
type runtimeConfig struct {
	ServiceName string
	Enabled     bool

	Endpoint string
	Headers  map[string]string
	Insecure bool

	TracesEndpoint  string
	MetricsEndpoint string
	LogsEndpoint    string

	SamplingRatio float64

	DeploymentEnvironment string
	ServiceVersion        string
}

func normalize(cfg Config) runtimeConfig {
	serviceName := strings.TrimSpace(cfg.ServiceName)

	insecure := cfg.Insecure
	if strings.TrimSpace(cfg.InsecureMode) != "" {
		insecure = parseBoolString(cfg.InsecureMode, insecure)
	}

	ratio := cfg.Sampling.Ratio
	if ratio <= 0 {
		ratio = 1.0
	}
	if ratio > 1 {
		ratio = 1.0
	}

	headers := cfg.Headers
	if headers == nil {
		headers = map[string]string{}
	}

	return runtimeConfig{
		ServiceName:           serviceName,
		Enabled:               cfg.Enabled,
		Endpoint:              strings.TrimSpace(cfg.CollectorURL),
		Headers:               headers,
		Insecure:              insecure,
		TracesEndpoint:        strings.TrimSpace(cfg.TracesURL),
		MetricsEndpoint:       strings.TrimSpace(cfg.MetricsURL),
		LogsEndpoint:          strings.TrimSpace(cfg.LogsURL),
		SamplingRatio:         ratio,
		DeploymentEnvironment: strings.TrimSpace(cfg.Resource.DeploymentEnvironment),
		ServiceVersion:        strings.TrimSpace(cfg.Resource.ServiceVersion),
	}
}

func parseBoolString(raw string, defaultValue bool) bool {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return defaultValue
	}
}
