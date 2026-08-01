package commonobservability

import (
	"testing"

	"github.com/spf13/viper"
)

func resetViper(t *testing.T) {
	t.Helper()
	viper.Reset()
}

func TestInitConfig_missingSectionDisabled(t *testing.T) {
	resetViper(t)

	cfg, err := InitConfig("go-pkgs-test")
	if err != nil {
		t.Fatalf("InitConfig: %v", err)
	}
	if cfg.Enabled {
		t.Fatal("expected enabled=false when opentel section missing")
	}
	if cfg.ServiceName != "go-pkgs-test" {
		t.Fatalf("service name = %q, want go-pkgs-test", cfg.ServiceName)
	}
}

func TestInitConfig_yamlOnlyNoEnvOverride(t *testing.T) {
	resetViper(t)
	t.Setenv("OTEL_ENABLED", "true")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "env-should-not-win:4317")

	viper.Set("opentel.enabled", false)
	viper.Set("opentel.collector_url", "yaml-collector:4317")

	cfg, err := InitConfig("go-pkgs-test")
	if err != nil {
		t.Fatalf("InitConfig: %v", err)
	}
	if cfg.Enabled {
		t.Fatal("expected YAML enabled=false, not env OTEL_ENABLED")
	}
	if cfg.Endpoint != "yaml-collector:4317" {
		t.Fatalf("endpoint = %q, want yaml-collector:4317", cfg.Endpoint)
	}
}

func TestInitConfig_samplingAndLegacyInsecureMode(t *testing.T) {
	resetViper(t)
	viper.Set("opentel.enabled", true)
	viper.Set("opentel.collector_url", "collector:4317")
	viper.Set("opentel.insecure_mode", "true")
	viper.Set("opentel.sampling.ratio", 0.2)
	viper.Set("opentel.resource.deployment_environment", "test")
	viper.Set("opentel.resource.service_version", "9.9.9")

	cfg, err := InitConfig("fallback-name")
	if err != nil {
		t.Fatalf("InitConfig: %v", err)
	}
	if !cfg.Enabled {
		t.Fatal("expected enabled=true")
	}
	if !cfg.Insecure {
		t.Fatal("expected insecure=true from insecure_mode")
	}
	if cfg.SamplingRatio != 0.2 {
		t.Fatalf("sampling ratio = %v, want 0.2", cfg.SamplingRatio)
	}
	if cfg.DeploymentEnvironment != "test" {
		t.Fatalf("deployment env = %q", cfg.DeploymentEnvironment)
	}
	if cfg.ServiceVersion != "9.9.9" {
		t.Fatalf("service version = %q", cfg.ServiceVersion)
	}
}

func TestNormalizeConfig_samplingClamped(t *testing.T) {
	cfg := normalizeConfig("svc", YAMLConfig{
		Enabled: true,
		Sampling: SamplingConfig{
			Ratio: 2.5,
		},
	})
	if cfg.SamplingRatio != 1.0 {
		t.Fatalf("ratio = %v, want clamped to 1.0", cfg.SamplingRatio)
	}
}
