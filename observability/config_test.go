package commonobservability

import "testing"

func TestNormalize_defaultsWhenDisabled(t *testing.T) {
	rc := normalize(Config{ServiceName: "go-pkgs-test"})
	if rc.Enabled {
		t.Fatal("expected enabled=false by default")
	}
	if rc.ServiceName != "go-pkgs-test" {
		t.Fatalf("service name = %q, want go-pkgs-test", rc.ServiceName)
	}
	if rc.SamplingRatio != 1.0 {
		t.Fatalf("sampling ratio = %v, want 1.0", rc.SamplingRatio)
	}
	if rc.Headers == nil {
		t.Fatal("expected non-nil headers map")
	}
}

func TestNormalize_samplingAndLegacyInsecureMode(t *testing.T) {
	rc := normalize(Config{
		Enabled:      true,
		CollectorURL: "collector:4317",
		InsecureMode: "true",
		Sampling:     SamplingConfig{Ratio: 0.2},
		Resource: ResourceConfig{
			DeploymentEnvironment: "test",
			ServiceVersion:        "9.9.9",
		},
	})
	if !rc.Enabled {
		t.Fatal("expected enabled=true")
	}
	if !rc.Insecure {
		t.Fatal("expected insecure=true from insecure_mode")
	}
	if rc.Endpoint != "collector:4317" {
		t.Fatalf("endpoint = %q", rc.Endpoint)
	}
	if rc.SamplingRatio != 0.2 {
		t.Fatalf("sampling ratio = %v, want 0.2", rc.SamplingRatio)
	}
	if rc.DeploymentEnvironment != "test" {
		t.Fatalf("deployment env = %q", rc.DeploymentEnvironment)
	}
	if rc.ServiceVersion != "9.9.9" {
		t.Fatalf("service version = %q", rc.ServiceVersion)
	}
}

func TestNormalize_samplingClamped(t *testing.T) {
	rc := normalize(Config{
		Enabled:  true,
		Sampling: SamplingConfig{Ratio: 2.5},
	})
	if rc.SamplingRatio != 1.0 {
		t.Fatalf("ratio = %v, want clamped to 1.0", rc.SamplingRatio)
	}
}

func TestInit_disabledNoOp(t *testing.T) {
	shutdown, err := Init(t.Context(), Config{Enabled: false, ServiceName: "test"}, nil)
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	if IsEnabled() {
		t.Fatal("expected IsEnabled=false")
	}
	if err := shutdown(t.Context()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
}
