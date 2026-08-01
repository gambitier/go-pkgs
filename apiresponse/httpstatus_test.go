package apiresponse

import (
	"net/http"
	"testing"

	pkgerrors "github.com/gambitier/go-pkgs/errors"
)

func TestCodeFromHTTPStatus_neverInternalFor4xx(t *testing.T) {
	for status := 400; status < 500; status++ {
		code := CodeFromHTTPStatus(status)
		if code == pkgerrors.CodeInternal {
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

func TestToDomainError_methodNotAllowed(t *testing.T) {
	err := ToDomainError(http.StatusMethodNotAllowed, "Method Not Allowed", nil)
	if err.Code != pkgerrors.CodeMethodNotAllowed {
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
	if code != pkgerrors.CodeInternal {
		t.Fatalf("expected INTERNAL for unknown 5xx, got %s", code)
	}
}

func TestCodeFromHTTPStatus_unknown4xxIsInvalidArgument(t *testing.T) {
	code := CodeFromHTTPStatus(499)
	if code != pkgerrors.CodeInvalidArgument {
		t.Fatalf("expected INVALID_ARGUMENT for unknown 4xx, got %s", code)
	}
}
