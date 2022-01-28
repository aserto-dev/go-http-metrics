// Package std is a helper package to get a standard `http.Handler` compatible middleware.
package std

import (
	"context"
	"net/http"

	"github.com/slok/go-http-metrics/middleware"
)

// Handler returns an measuring standard http.Handler.
func Handler(handlerID string, m middleware.Middleware, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wi := NewResponseWriterInterceptor(w)
		reporter := &stdReporter{
			capturedResponse: wi,
			r:                r,
		}

		m.Measure(handlerID, reporter, func() {
			h.ServeHTTP(wi, r)
		})
	})
}

// HandlerProvider is a helper method that returns a handler provider. This kind of
// provider is a defacto standard in some frameworks (e.g: Gorilla, Chi...).
func HandlerProvider(handlerID string, m middleware.Middleware) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return Handler(handlerID, m, next)
	}
}

type capturedResponse interface {
	StatusCode() int
	BytesWritten() int64
}

type stdReporter struct {
	capturedResponse
	r *http.Request
}

func (s *stdReporter) Method() string { return s.r.Method }

func (s *stdReporter) Context() context.Context { return s.r.Context() }

func (s *stdReporter) URLPath() string { return s.r.URL.Path }
