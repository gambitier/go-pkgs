package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gambitier/go-pkgs/logging/correlation"
	"github.com/sirupsen/logrus"
)

func TestNewDoesNotMutateGlobalLogrusLogger(t *testing.T) {
	previous := logrus.StandardLogger().GetLevel()
	t.Cleanup(func() {
		logrus.StandardLogger().SetLevel(previous)
	})

	logrus.StandardLogger().SetLevel(logrus.PanicLevel)

	_, err := New(Config{Level: "debug", Format: "json"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := logrus.StandardLogger().GetLevel(); got != logrus.PanicLevel {
		t.Fatalf("expected standard logger level to remain panic, got %v", got)
	}
}

func TestWithFieldsAndMessageAreSeparated(t *testing.T) {
	var out bytes.Buffer
	logger, err := New(Config{
		ServiceName: "payments",
		Level:       "info",
		Format:      "json",
		Output:      &out,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	logger.Info("create checkout session", Fields{
		"customer_id": "cus_123",
		"plan_id":     "plan_basic",
	})

	var payload map[string]any
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("failed to parse log output: %v", err)
	}

	if payload["msg"] != "create checkout session" {
		t.Fatalf("unexpected msg field: %v", payload["msg"])
	}
	if payload["customer_id"] != "cus_123" {
		t.Fatalf("expected customer_id field in payload")
	}
	if payload["service"] != "payments" {
		t.Fatalf("expected service field in payload")
	}
}

func TestWithContextAddsCorrelationID(t *testing.T) {
	var out bytes.Buffer
	logger, err := New(Config{
		Level:  "info",
		Format: "json",
		Output: &out,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	ctx := correlation.SetCorrelationID(context.Background(), "cid-123")
	scoped := logger.WithContext(ctx)
	scoped.Info("request started", nil)

	if !strings.Contains(out.String(), "\"correlation_id\":\"cid-123\"") {
		t.Fatalf("expected correlation_id in output, got %s", out.String())
	}
}

func TestWithContextBindsContextToEntry(t *testing.T) {
	logger, err := New(Config{
		Level:  "info",
		Format: "json",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	ctx := context.WithValue(context.Background(), struct{}{}, "value")
	scoped := logger.WithContext(ctx)

	impl, ok := scoped.(*logrusLogger)
	if !ok {
		t.Fatalf("expected *logrusLogger, got %T", scoped)
	}
	if impl.entry.Context != ctx {
		t.Fatalf("expected entry context to be bound")
	}
}

func TestErrorLogIncludesErrorString(t *testing.T) {
	var out bytes.Buffer
	logger, err := New(Config{
		ServiceName: "example-service",
		Level:       "error",
		Format:      "json",
		Output:      &out,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	uploadErr := fmt.Errorf("storage set: %w", errors.New("AccessDenied: User is not allowed to access resource"))
	logger.Error("file upload failed", uploadErr, nil)

	var payload map[string]any
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("failed to parse log output: %v", err)
	}
	if payload["msg"] != "file upload failed" {
		t.Fatalf("unexpected msg: %v", payload["msg"])
	}
	errorValue, _ := payload["error"].(string)
	if !strings.Contains(errorValue, "AccessDenied") {
		t.Fatalf("expected AccessDenied in error field, got %q", errorValue)
	}
}

func TestNewWithUnsupportedSinkTypeReturnsError(t *testing.T) {
	_, err := New(Config{
		Level:  "info",
		Format: "json",
		Sinks: []SinkConfig{
			{Type: "unknown", Enabled: true},
		},
	})
	if err == nil {
		t.Fatalf("expected error for unsupported sink type")
	}
}

func TestNewWithFileSinkWritesToFile(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "service.log")

	logger, err := New(Config{
		ServiceName: "payments",
		Level:       "info",
		Format:      "json",
		Sinks: []SinkConfig{
			{
				Type:    "file",
				Enabled: true,
				File: FileSinkConfig{
					Path: logPath,
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	logger.Info("file sink message", nil)

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}
	if !strings.Contains(string(data), "file sink message") {
		t.Fatalf("expected message in file output, got %s", string(data))
	}
}
