package gorilla

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/slok/go-http-metrics/middleware"
	"github.com/slok/go-http-metrics/middleware/std"
)

func GorillaMuxMetricsMiddleware(m middleware.Middleware) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return GorillaMuxMetricsHandler(m, next)
	}
}

func GorillaMuxMetricsHandler(m middleware.Middleware, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wi := std.NewResponseWriterInterceptor(w)
		reporter := &gorillaReporter{
			CapturedResponse: wi,
			r:                r,
		}

		m.Measure("", reporter, func() {
			h.ServeHTTP(wi, r)
		})
	})
}

type gorillaReporter struct {
	std.CapturedResponse
	r *http.Request
}

func (s *gorillaReporter) Method() string { return s.r.Method }

func (s *gorillaReporter) Context() context.Context { return s.r.Context() }

// URLPath returns the path to be used as the metric label for the request.
// If the route contains any parameters (e.g. '/api/users/{id}') then the route template is used instead of the
// full URL in order to group all requests to the same route.
func (s *gorillaReporter) URLPath() string {
	if route, vars := mux.CurrentRoute(s.r), mux.Vars(s.r); route != nil && len(vars) > 0 {
		if pathTemlate, err := route.GetPathTemplate(); err == nil {
			return pathTemlate
		}
	}
	return s.r.URL.Path
}
