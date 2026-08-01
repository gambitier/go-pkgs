package commonobservability

import (
	"fmt"
	"net/http"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"google.golang.org/grpc"
)

var healthCheckPaths = map[string]struct{}{
	"/health":  {},
	"/livez":   {},
	"/healthz": {},
	"/swagger": {},
}

// WrapHTTPHandler instruments a generic http.Handler with otelhttp.
func WrapHTTPHandler(serviceName string, handler http.Handler) http.Handler {
	if !IsEnabled() {
		return handler
	}
	return otelhttp.NewHandler(
		handler,
		serviceName,
		otelhttp.WithSpanNameFormatter(httpSpanNameFormatter),
		otelhttp.WithFilter(func(r *http.Request) bool {
			return !isHealthCheckPath(r.URL.Path)
		}),
	)
}

func httpSpanNameFormatter(_ string, r *http.Request) string {
	route := r.Pattern
	if route == "" {
		route = r.URL.Path
	}
	return fmt.Sprintf("%s %s", r.Method, route)
}

func NewHTTPClient(base *http.Client) *http.Client {
	if base == nil {
		base = http.DefaultClient
	}
	clone := *base
	if IsEnabled() {
		transport := base.Transport
		if transport == nil {
			transport = http.DefaultTransport
		}
		clone.Transport = otelhttp.NewTransport(transport)
	}
	return &clone
}

func isHealthCheckPath(path string) bool {
	path = strings.TrimSpace(path)
	if path == "" {
		return false
	}
	if _, ok := healthCheckPaths[path]; ok {
		return true
	}
	return strings.HasPrefix(path, "/swagger/")
}

func GRPCServerOptions() []grpc.ServerOption {
	if !IsEnabled() {
		return nil
	}
	return []grpc.ServerOption{
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	}
}

func GRPCDialOptions() []grpc.DialOption {
	if !IsEnabled() {
		return nil
	}
	return []grpc.DialOption{
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	}
}
