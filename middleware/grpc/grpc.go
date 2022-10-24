package grpc

import (
	"context"
	"net/http"

	"github.com/aserto-dev/go-http-metrics/middleware"
	"github.com/aserto-dev/go-http-metrics/middleware/std"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/metadata"
)

type key int

var pathPatternKey key

type GatewayPathPattern struct {
	PathPattern string
}

// CaptureGatewayRoute is a hook that plugs into a grpc-gateway ServeMux and stores the request's path-pattern
// in a context value.
// It must be attached to a ServeMux at creation time using runtime.WithMetadata.
func CaptureGatewayRoute(ctx context.Context, r *http.Request) metadata.MD {
	if pattern, ok := runtime.HTTPPathPattern(ctx); ok {
		if gwPathPattern := gatewayContextValue(r); gwPathPattern != nil {
			gwPathPattern.PathPattern = pattern
		}
	}
	return nil
}

// GatewayMuxMetricsMiddleware returns an HTTP middleware that reports metrics from grpc-gateway's runtime.ServeMux.
//
// Note: This middleware requires that CaptureGatewayRoute is attached to the runtime.ServeMux using
// runtime.WithMetadata.
func GatewayMuxMetricsMiddleware(m middleware.Middleware) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return GatewayMuxMetricsHandler(m, next)
	}
}

func GatewayMuxMetricsHandler(m middleware.Middleware, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(
			context.WithValue(r.Context(), pathPatternKey, &GatewayPathPattern{}),
		)
		wi := std.NewResponseWriterInterceptor(w)
		reporter := &gatewayMuxReporter{
			CapturedResponse: wi,
			r:                r,
		}

		m.Measure("", reporter, func() {
			h.ServeHTTP(wi, r)
		})
	})
}

type gatewayMuxReporter struct {
	std.CapturedResponse
	r *http.Request
}

func (s *gatewayMuxReporter) Method() string { return s.r.Method }

func (s *gatewayMuxReporter) Context() context.Context { return s.r.Context() }

func (s *gatewayMuxReporter) URLPath() string {
	if gwPathPattern := gatewayContextValue(s.r); gwPathPattern != nil {
		return gwPathPattern.PathPattern
	}

	return ""
}

func gatewayContextValue(r *http.Request) *GatewayPathPattern {
	gwPathPattern, ok := r.Context().Value(pathPatternKey).(*GatewayPathPattern)
	if !ok {
		return nil
	}

	return gwPathPattern
}
