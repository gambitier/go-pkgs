package logging

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
)

func TestLogHTTPErrorUsesWarnFor4xx(t *testing.T) {
	var out bytes.Buffer
	logger, err := New(Config{
		ServiceName: "example-service",
		Level:       "debug",
		Format:      "json",
		Output:      &out,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	LogHTTPError(logger, errors.New("invalid token"), HTTPErrorLog{
		Method:       "GET",
		Path:         "/items/api/v1/items",
		StatusCode:   401,
		ErrorCode:    "UNAUTHORIZED",
		ErrorMsg:     "token missing",
		ErrorContext: "token validation failed",
	})

	var payload map[string]any
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("failed to parse log output: %v", err)
	}
	if payload["level"] != "warning" {
		t.Fatalf("expected warning level, got %v", payload["level"])
	}
	if payload["msg"] != "invalid token" {
		t.Fatalf("unexpected message: %v", payload["msg"])
	}
}

func TestLogHTTPErrorUsesErrorFor5xx(t *testing.T) {
	var out bytes.Buffer
	logger, err := New(Config{
		ServiceName: "example-service",
		Level:       "debug",
		Format:      "json",
		Output:      &out,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	origErr := errors.New("db timeout")
	LogHTTPError(logger, origErr, HTTPErrorLog{
		Method:      "POST",
		Path:        "/items/api/v1/items",
		StatusCode:  500,
		ErrorCode:   "INTERNAL",
		ErrorMsg:    "internal server error",
		ErrorCause:  "db timeout",
		ErrorSource: "internal/app/service.go:42 Create",
		ErrorFrames: []string{"internal/app/service.go:42 Create"},
	})

	var payload map[string]any
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("failed to parse log output: %v", err)
	}
	if payload["level"] != "error" {
		t.Fatalf("expected error level, got %v", payload["level"])
	}
	if payload["error"] != "db timeout" {
		t.Fatalf("unexpected error field: %v", payload["error"])
	}
}

func TestBuildHTTPErrorLogUsesRootCause(t *testing.T) {
	inner := errors.New("connection refused")
	outer := fmt.Errorf("db query: %w", inner)

	payload := BuildHTTPErrorLog(outer, HTTPErrorLogInput{
		Method:     "GET",
		Path:       "/items/api/v1/items/1",
		StatusCode: 500,
		ErrorCode:  "INTERNAL",
		ErrorMsg:   "internal server error",
	})

	if payload.ErrorCause != "connection refused" {
		t.Fatalf("expected root cause, got %q", payload.ErrorCause)
	}
	if payload.ErrorMsg != "internal server error" {
		t.Fatalf("unexpected ErrorMsg: %q", payload.ErrorMsg)
	}
}

func TestResolveHTTPErrorLogMessagePrefersErrorOverClientMsg(t *testing.T) {
	got := resolveHTTPErrorLogMessage(errors.New("failed to get image from storage"), HTTPErrorLog{
		ErrorMsg: "internal server error",
	})
	if got != "failed to get image from storage" {
		t.Fatalf("expected domain error message, got %q", got)
	}
}
