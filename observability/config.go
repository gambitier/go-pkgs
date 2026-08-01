package commonobservability

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// YAMLConfig is the opentel block in service config files (YAML-only; no OTEL_* env overrides).
type YAMLConfig struct {
	Enabled       bool            `mapstructure:"enabled"`
	CollectorURL  string          `mapstructure:"collector_url"`
	ServiceName   string          `mapstructure:"service_name"`
	Insecure      bool            `mapstructure:"insecure"`
	InsecureMode  string          `mapstructure:"insecure_mode"` // legacy string "true"/"false"
	Sampling      SamplingConfig  `mapstructure:"sampling"`
	Resource      ResourceConfig  `mapstructure:"resource"`
	TracesURL     string          `mapstructure:"traces_endpoint"`
	MetricsURL    string          `mapstructure:"metrics_endpoint"`
	LogsURL       string          `mapstructure:"logs_endpoint"`
	Headers       map[string]string `mapstructure:"headers"`
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

// Config is the normalized runtime observability configuration.
type Config struct {
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

// InitConfig reads the opentel section from viper. When the section is missing, telemetry is disabled.
func InitConfig(serviceName string) (Config, error) {
	cfg := Config{
		ServiceName:   strings.TrimSpace(serviceName),
		Enabled:       false,
		SamplingRatio: 1.0,
	}

	subv := viper.Sub("opentel")
	if subv == nil {
		return cfg, nil
	}

	var yamlCfg YAMLConfig
	if err := subv.Unmarshal(&yamlCfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal opentel config: %w", err)
	}

	return normalizeConfig(cfg.ServiceName, yamlCfg), nil
}

func normalizeConfig(fallbackServiceName string, yamlCfg YAMLConfig) Config {
	serviceName := strings.TrimSpace(yamlCfg.ServiceName)
	if serviceName == "" {
		serviceName = strings.TrimSpace(fallbackServiceName)
	}

	insecure := yamlCfg.Insecure
	if strings.TrimSpace(yamlCfg.InsecureMode) != "" {
		insecure = parseBoolString(yamlCfg.InsecureMode, insecure)
	}

	ratio := yamlCfg.Sampling.Ratio
	if ratio <= 0 {
		ratio = 1.0
	}
	if ratio > 1 {
		ratio = 1.0
	}

	headers := yamlCfg.Headers
	if headers == nil {
		headers = map[string]string{}
	}

	return Config{
		ServiceName:           serviceName,
		Enabled:               yamlCfg.Enabled,
		Endpoint:              strings.TrimSpace(yamlCfg.CollectorURL),
		Headers:               headers,
		Insecure:              insecure,
		TracesEndpoint:        strings.TrimSpace(yamlCfg.TracesURL),
		MetricsEndpoint:       strings.TrimSpace(yamlCfg.MetricsURL),
		LogsEndpoint:          strings.TrimSpace(yamlCfg.LogsURL),
		SamplingRatio:         ratio,
		DeploymentEnvironment: strings.TrimSpace(yamlCfg.Resource.DeploymentEnvironment),
		ServiceVersion:        strings.TrimSpace(yamlCfg.Resource.ServiceVersion),
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
