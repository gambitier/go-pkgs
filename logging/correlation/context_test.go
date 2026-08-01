package correlation

import (
	"context"
	"testing"
)

func TestEnsureCorrelationIDUsesIncoming(t *testing.T) {
	ctx, id := EnsureCorrelationID(context.Background(), "incoming-id")
	if id != "incoming-id" {
		t.Fatalf("expected incoming correlation id, got %s", id)
	}
	if got := GetCorrelationID(ctx); got != "incoming-id" {
		t.Fatalf("expected context correlation id, got %s", got)
	}
}

func TestEnsureCorrelationIDGeneratesWhenMissing(t *testing.T) {
	ctx, id := EnsureCorrelationID(context.Background(), "")
	if id == "" {
		t.Fatalf("expected generated correlation id")
	}
	if got := GetCorrelationID(ctx); got == "" {
		t.Fatalf("expected generated id in context")
	}
}
