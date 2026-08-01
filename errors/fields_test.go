package errors

import (
	"errors"
	"testing"
)

func TestCollectFieldsMergesWrappedDomainErrors(t *testing.T) {
	inner := Internal("storage read failed", errors.New("s3"), map[string]any{
		"bucket": "uploads",
		"key":    "path/to/file",
	})
	outer := Internal("failed to get image from storage", inner, map[string]any{
		"file_id": "f-123",
	})

	got := CollectFields(outer)
	if got["file_id"] != "f-123" {
		t.Fatalf("expected outer file_id, got %#v", got["file_id"])
	}
	if got["bucket"] != "uploads" {
		t.Fatalf("expected inner bucket, got %#v", got["bucket"])
	}
	if got["key"] != "path/to/file" {
		t.Fatalf("expected inner key, got %#v", got["key"])
	}
}

func TestCollectFieldsOuterWinsOnDuplicateKeys(t *testing.T) {
	inner := NotFound("missing", nil, map[string]any{"resource": "file", "id": "inner"})
	outer := NotFound("not found", inner, map[string]any{"resource": "customer", "id": "outer"})

	got := CollectFields(outer)
	if got["resource"] != "customer" {
		t.Fatalf("expected outer resource to win, got %#v", got["resource"])
	}
	if got["id"] != "outer" {
		t.Fatalf("expected outer id to win, got %#v", got["id"])
	}
}

func TestCollectFieldsIncludesWrapChain(t *testing.T) {
	base := NotFound("invoice not found", nil, map[string]any{"invoiceId": "inv_1"})
	wrapped := Wrap(Wrap(base, "failed to fetch invoice"), "handler failed")

	got := CollectFields(wrapped)
	if got["invoiceId"] != "inv_1" {
		t.Fatalf("expected fields from wrapped domain error, got %#v", got)
	}
}

func TestCollectFieldsFluentBuilderChain(t *testing.T) {
	inner := InternalError().
		Message("failed to download image").
		Fields(map[string]any{"storage_key": "k1"})
	outer := InternalError().
		Message("failed to get image from storage").
		Err(inner).
		Fields(map[string]any{"file_id": "f-1"})

	got := CollectFields(outer)
	if got["file_id"] != "f-1" {
		t.Fatalf("expected file_id, got %#v", got)
	}
	if got["storage_key"] != "k1" {
		t.Fatalf("expected storage_key from inner layer, got %#v", got)
	}
}

func TestCollectFieldsNilForNonDomainError(t *testing.T) {
	if got := CollectFields(errors.New("plain")); got != nil {
		t.Fatalf("expected nil fields, got %#v", got)
	}
}
