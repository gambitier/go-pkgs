package httpresp

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/gambitier/go-pkgs/errors/domainerr"
	"github.com/gambitier/go-pkgs/errors/httpstatus"
)

func TestBuildErrorEnvelopeForDomainError(t *testing.T) {
	status, env, code, msg, fields := BuildErrorEnvelope(
		domainerr.InvalidArgument("invalid input", nil, map[string]any{"field": "email"}),
		"req-1",
	)

	if status != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", status)
	}
	if code != domainerr.CodeInvalidArgument {
		t.Fatalf("expected INVALID_ARGUMENT, got %s", code)
	}
	if msg != "invalid input" {
		t.Fatalf("unexpected message: %s", msg)
	}
	if fields["field"] != "email" {
		t.Fatalf("unexpected fields: %#v", fields)
	}
	if env.Error == nil || env.Error.RequestID != "req-1" || env.RequestID != "req-1" {
		t.Fatalf("request id not propagated: %#v", env)
	}
}

func TestBuildErrorEnvelopeForUnknownError(t *testing.T) {
	status, env, code, msg, _ := BuildErrorEnvelope(io.EOF, "req-2")
	if status != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", status)
	}
	if code != domainerr.CodeInternal || msg != InternalServerErrorMessage {
		t.Fatalf("unexpected fallback: code=%s msg=%s", code, msg)
	}
	if env.Error == nil || env.Error.ErrorCode != string(domainerr.CodeInternal) {
		t.Fatalf("unexpected envelope: %#v", env)
	}
}

func TestBuildErrorEnvelopeForDomainInternalError(t *testing.T) {
	err := domainerr.Internal("payment gateway error", errors.New("stripe timeout"), map[string]any{
		"operation": "CreatePortalSession",
	})
	wrapped := domainerr.Wrap(err, "failed to create portal session")

	status, env, code, msg, fields := BuildErrorEnvelope(wrapped, "req-internal")
	if status != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", status)
	}
	if code != domainerr.CodeInternal {
		t.Fatalf("expected INTERNAL, got %s", code)
	}
	if msg != InternalServerErrorMessage {
		t.Fatalf("expected generic internal message, got %q", msg)
	}
	if env.Error == nil || env.Error.ErrorCode != string(domainerr.CodeInternal) || env.Error.Message != InternalServerErrorMessage {
		t.Fatalf("unexpected envelope: %#v", env)
	}
	if fields["operation"] != "CreatePortalSession" {
		t.Fatalf("unexpected fields: %#v", fields)
	}
}

func TestBuildErrorEnvelopeForRawInternalErrorFallback(t *testing.T) {
	err := errors.New("db connection refused")
	status, env, code, msg, fields := BuildErrorEnvelope(err, "req-raw-internal")
	if status != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", status)
	}
	if code != domainerr.CodeInternal {
		t.Fatalf("expected INTERNAL, got %s", code)
	}
	if msg != InternalServerErrorMessage {
		t.Fatalf("expected generic internal message, got %q", msg)
	}
	if fields != nil {
		t.Fatalf("expected nil fields, got %#v", fields)
	}
	if env.Error == nil || env.Error.Message != InternalServerErrorMessage {
		t.Fatalf("unexpected envelope: %#v", env)
	}
}

func TestBuildErrorEnvelopePrefersDomainMessageOverWrapContext(t *testing.T) {
	base := domainerr.NotFound("customer not found", nil, map[string]any{"resource": "customer"})
	wrapped := domainerr.Wrap(base, "failed to load customer for portal")

	status, env, code, msg, fields := BuildErrorEnvelope(wrapped, "req-3")
	if status != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", status)
	}
	if code != domainerr.CodeNotFound {
		t.Fatalf("expected NOT_FOUND, got %s", code)
	}
	if msg != "customer not found" {
		t.Fatalf("expected domain-safe message, got %q", msg)
	}
	if env.Error == nil || env.Error.Message != "customer not found" {
		t.Fatalf("unexpected envelope message: %#v", env)
	}
	if fields["resource"] != "customer" {
		t.Fatalf("unexpected fields: %#v", fields)
	}
}

func TestBuildErrorEnvelopeForFmtWrappedDomainError(t *testing.T) {
	base := domainerr.NotFound("customer not found", nil, map[string]any{"resource": "customer"})
	wrapped := fmt.Errorf("service failed to load customer: %w", base)

	status, env, code, msg, fields := BuildErrorEnvelope(wrapped, "req-4")
	if status != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", status)
	}
	if code != domainerr.CodeNotFound {
		t.Fatalf("expected NOT_FOUND, got %s", code)
	}
	if msg != "customer not found" {
		t.Fatalf("expected domain-safe message, got %q", msg)
	}
	if env.Error == nil || env.Error.Message != "customer not found" {
		t.Fatalf("unexpected envelope message: %#v", env)
	}
	if fields["resource"] != "customer" {
		t.Fatalf("unexpected fields: %#v", fields)
	}
}

func TestBuildErrorEnvelopeForFmtThenDomainWrapChain(t *testing.T) {
	level1 := fmt.Errorf(
		"repository fetch failed: %w",
		errors.New("database timeout"),
	)
	level2 := domainerr.Wrap(level1, "service failed to load customer")

	status, env, code, msg, fields := BuildErrorEnvelope(level2, "req-7")
	if status != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", status)
	}
	if code != domainerr.CodeInternal {
		t.Fatalf("expected INTERNAL, got %s", code)
	}
	if msg != InternalServerErrorMessage {
		t.Fatalf("expected generic internal message, got %q", msg)
	}
	if env.Error == nil || env.Error.Message != InternalServerErrorMessage {
		t.Fatalf("unexpected envelope message: %#v", env)
	}
	if fields != nil {
		t.Fatalf("expected nil fields, got %#v", fields)
	}
}

func TestBuildErrorEnvelopeForMultiLevelNewDomainError(t *testing.T) {
	level1 := domainerr.NotFound("customer:123 not found", nil, map[string]any{"resource": "customer"})
	level2 := domainerr.NotFound("customer not found", level1, map[string]any{"resource": "customer"})

	status, env, code, msg, fields := BuildErrorEnvelope(level2, "req-5")
	if status != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", status)
	}
	if code != domainerr.CodeNotFound {
		t.Fatalf("expected NOT_FOUND, got %s", code)
	}
	if msg != "customer not found" {
		t.Fatalf("expected origin domain-safe message, got %q", msg)
	}
	if env.Error == nil || env.Error.Message != "customer not found" {
		t.Fatalf("unexpected envelope message: %#v", env)
	}
	if fields["resource"] != "customer" {
		t.Fatalf("unexpected fields: %#v", fields)
	}
}

func TestBuildErrorEnvelopeForInternalThenNotFound(t *testing.T) {
	level1 := domainerr.Internal("database timeout", errors.New("mongo timeout"), map[string]any{
		"operation": "GetByAccountID",
	})
	level2 := domainerr.NotFound("customer not found", level1, map[string]any{"resource": "customer"})

	status, env, code, msg, fields := BuildErrorEnvelope(level2, "req-6")
	if status != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", status)
	}
	if code != domainerr.CodeNotFound {
		t.Fatalf("expected NOT_FOUND, got %s", code)
	}
	if msg != "customer not found" {
		t.Fatalf("expected top-level domain-safe message, got %q", msg)
	}
	if env.Error == nil || env.Error.Message != "customer not found" {
		t.Fatalf("unexpected envelope message: %#v", env)
	}
	if fields["resource"] != "customer" {
		t.Fatalf("unexpected fields: %#v", fields)
	}
	if fields["operation"] != "GetByAccountID" {
		t.Fatalf("expected inner domain fields to be merged, got %#v", fields)
	}
}

func TestBuildErrorEnvelopeForMethodNotAllowedFiberError(t *testing.T) {
	err := httpstatus.ToDomainError(http.StatusMethodNotAllowed, "Method Not Allowed", nil)
	status, env, code, msg, _ := BuildErrorEnvelope(err, "req-405")
	if status != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", status)
	}
	if code != domainerr.CodeMethodNotAllowed {
		t.Fatalf("expected METHOD_NOT_ALLOWED, got %s", code)
	}
	if msg != "Method Not Allowed" {
		t.Fatalf("unexpected message: %q", msg)
	}
	if env.Error == nil || env.Error.ErrorCode != string(domainerr.CodeMethodNotAllowed) {
		t.Fatalf("unexpected envelope: %#v", env)
	}
}

func TestBuildSuccessEnvelopes(t *testing.T) {
	okEnv := BuildOKEnvelope(map[string]string{"status": "ok"}, "req-ok")
	if okEnv.RequestID != "req-ok" || okEnv.Data["status"] != "ok" {
		t.Fatalf("unexpected OK envelope: %#v", okEnv)
	}

	createdEnv := BuildCreatedEnvelope("created", "req-created")
	if createdEnv.RequestID != "req-created" || createdEnv.Data != "created" {
		t.Fatalf("unexpected Created envelope: %#v", createdEnv)
	}
}
