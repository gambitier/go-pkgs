package apiresponse

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	pkgerrors "github.com/gambitier/go-pkgs/errors"
)

func TestBuildProblem_notFound(t *testing.T) {
	err := pkgerrors.NotFound("item not found", nil, map[string]any{"id": "abc"})
	res := BuildProblem(err, BuildOptions{Instance: "/items/abc"})

	if res.Status != http.StatusNotFound {
		t.Fatalf("status: got %d", res.Status)
	}
	p := res.Problem
	if p.Type != DefaultProblemType {
		t.Fatalf("type: got %q", p.Type)
	}
	if p.Title != http.StatusText(http.StatusNotFound) {
		t.Fatalf("title: got %q", p.Title)
	}
	if p.Status != http.StatusNotFound {
		t.Fatalf("problem.status: got %d", p.Status)
	}
	if p.Detail != "item not found" {
		t.Fatalf("detail: got %q", p.Detail)
	}
	if p.Instance != "/items/abc" {
		t.Fatalf("instance: got %q", p.Instance)
	}
	if p.Code != string(pkgerrors.CodeNotFound) {
		t.Fatalf("code: got %q", p.Code)
	}
	if p.Fields["id"] != "abc" {
		t.Fatalf("fields: got %#v", p.Fields)
	}

	raw, marshalErr := json.Marshal(p)
	if marshalErr != nil {
		t.Fatal(marshalErr)
	}
	var decoded map[string]any
	if unmarshalErr := json.Unmarshal(raw, &decoded); unmarshalErr != nil {
		t.Fatal(unmarshalErr)
	}
	for _, key := range []string{"type", "title", "status", "detail", "instance", "code", "fields"} {
		if _, ok := decoded[key]; !ok {
			t.Fatalf("missing JSON key %q in %s", key, raw)
		}
	}
	if _, ok := decoded["requestId"]; ok {
		t.Fatalf("requestId must not appear in problem body: %s", raw)
	}
	if _, ok := decoded["correlationId"]; ok {
		t.Fatalf("correlationId must not appear in problem body: %s", raw)
	}
}

func TestBuildProblem_internalRedactsMessage(t *testing.T) {
	err := pkgerrors.Internal("db exploded", errors.New("secret"), nil)
	res := BuildProblem(err, BuildOptions{})
	if res.Status != http.StatusInternalServerError {
		t.Fatalf("status: got %d", res.Status)
	}
	if res.Problem.Detail != internalDetail {
		t.Fatalf("detail should be redacted, got %q", res.Problem.Detail)
	}
	if res.Problem.Code != string(pkgerrors.CodeInternal) {
		t.Fatalf("code: got %q", res.Problem.Code)
	}
}

func TestBuildProblem_unknownErrorIsInternal(t *testing.T) {
	res := BuildProblem(errors.New("boom"), BuildOptions{})
	if res.Status != http.StatusInternalServerError {
		t.Fatalf("status: got %d", res.Status)
	}
	if res.Problem.Detail != internalDetail {
		t.Fatalf("detail: got %q", res.Problem.Detail)
	}
}

func TestBuildProblem_wrapPreservesCode(t *testing.T) {
	base := pkgerrors.NotFound("customer not found", nil, map[string]any{"resource": "customer"})
	wrapped := pkgerrors.Wrap(base, "failed to load customer")
	res := BuildProblem(wrapped, BuildOptions{})
	if res.Problem.Code != string(pkgerrors.CodeNotFound) {
		t.Fatalf("code: got %q", res.Problem.Code)
	}
	if res.Problem.Detail != "customer not found" {
		t.Fatalf("detail: got %q", res.Problem.Detail)
	}
}

func TestContentTypeConstant(t *testing.T) {
	if ContentTypeProblemJSON != "application/problem+json" {
		t.Fatalf("unexpected content type %q", ContentTypeProblemJSON)
	}
}
