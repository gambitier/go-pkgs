package correlation

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

const HeaderName = "X-Correlation-ID"

type contextKey string

const correlationKey contextKey = "correlation_id"

func GetCorrelationID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	raw := ctx.Value(correlationKey)
	value, _ := raw.(string)
	return strings.TrimSpace(value)
}

func SetCorrelationID(ctx context.Context, correlationID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, correlationKey, strings.TrimSpace(correlationID))
}

func EnsureCorrelationID(ctx context.Context, incoming string) (context.Context, string) {
	value := strings.TrimSpace(incoming)
	if value == "" {
		value = uuid.NewString()
	}
	return SetCorrelationID(ctx, value), value
}
