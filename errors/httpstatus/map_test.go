package httpstatus

import (
	"net/http"
	"testing"

	"github.com/gambitier/go-pkgs/errors/domainerr"
)

func TestCodeFromHTTPStatus_neverInternalFor4xx(t *testing.T) {
	for status := 400; status < 500; status++ {
		code := CodeFromHTTPStatus(status)
		if code == domainerr.CodeInternal {
			t.Fatalf("status %d mapped to INTERNAL", status)
		}
	}
}

func TestStatusFromCode_roundTrip(t *testing.T) {
	for status, wantCode := range codeByStatus {
		gotCode := CodeFromHTTPStatus(status)
		if gotCode != wantCode {
			t.Fatalf("status %d: expected code %s, got %s", status, wantCode, gotCode)
		}
		gotStatus := StatusFromCode(gotCode)
		if gotStatus != status {
			t.Fatalf("code %s: expected status %d, got %d", gotCode, status, gotStatus)
		}
	}
}

func TestMethodNotAllowedMapping(t *testing.T) {
	code := CodeFromHTTPStatus(http.StatusMethodNotAllowed)
	if code != domainerr.CodeMethodNotAllowed {
		t.Fatalf("expected METHOD_NOT_ALLOWED, got %s", code)
	}
	if StatusFromCode(code) != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", StatusFromCode(code))
	}
}

func TestToDomainError_methodNotAllowed(t *testing.T) {
	err := ToDomainError(http.StatusMethodNotAllowed, "Method Not Allowed", nil)
	if err.Code != domainerr.CodeMethodNotAllowed {
		t.Fatalf("expected METHOD_NOT_ALLOWED, got %s", err.Code)
	}
	if err.Message != "Method Not Allowed" {
		t.Fatalf("unexpected message: %q", err.Message)
	}
}

func TestToDomainError_emptyMessageUsesStatusText(t *testing.T) {
	err := ToDomainError(http.StatusNotFound, "", nil)
	if err.Message != http.StatusText(http.StatusNotFound) {
		t.Fatalf("expected %q, got %q", http.StatusText(http.StatusNotFound), err.Message)
	}
}

func TestCodeFromHTTPStatus_unknown5xxIsInternal(t *testing.T) {
	code := CodeFromHTTPStatus(599)
	if code != domainerr.CodeInternal {
		t.Fatalf("expected INTERNAL for unknown 5xx, got %s", code)
	}
}

func TestCodeFromHTTPStatus_unknown4xxIsInvalidArgument(t *testing.T) {
	code := CodeFromHTTPStatus(499)
	if code != domainerr.CodeInvalidArgument {
		t.Fatalf("expected INVALID_ARGUMENT for unknown 4xx, got %s", code)
	}
}
