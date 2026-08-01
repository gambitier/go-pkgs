package logging

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

// TraceFields returns trace_id and span_id when the context carries a valid span.
func TraceFields(ctx context.Context) Fields {
	if ctx == nil {
		return nil
	}
	span := trace.SpanFromContext(ctx)
	if span == nil || !span.SpanContext().IsValid() {
		return nil
	}
	sc := span.SpanContext()
	return Fields{
		"trace_id": sc.TraceID().String(),
		"span_id":  sc.SpanID().String(),
	}
}
