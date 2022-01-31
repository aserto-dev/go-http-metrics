package urlpath

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/slok/go-http-metrics/middleware"
	"github.com/slok/go-http-metrics/middleware/std"
	"github.com/ucarion/urlpath"
)

// URLMatchMetricsMiddleware creates http middleware that reports metrics to prometheus using URL path matching.
//
// It is a variation on "github.com/slok/go-http-metrics/middleware/std" but instead of reporting the full URL path,
// this version uses a list of urlpath.Path objects and uses the matching path as the request's handler.
// This is done to reduce the cardinality of reported metrics and group metrics by route.
func URLMatchMetricsMiddleware(paths []urlpath.Path, m middleware.Middleware) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return URLMatchMetricsHandler(paths, m, next)
	}
}

func URLMatchMetricsHandler(paths []urlpath.Path, m middleware.Middleware, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wi := std.NewResponseWriterInterceptor(w)
		reporter := &matchingReporter{
			CapturedResponse: wi,
			r:                r,
			p:                paths,
		}

		m.Measure("", reporter, func() {
			h.ServeHTTP(wi, r)
		})
	})
}

type matchingReporter struct {
	std.CapturedResponse
	r *http.Request
	p []urlpath.Path
}

func (s *matchingReporter) Method() string { return s.r.Method }

func (s *matchingReporter) Context() context.Context { return s.r.Context() }

// Instead of always returning the full URL path from the incoming request, return the matching route as a string.
func (s *matchingReporter) URLPath() string {
	reqURLPath := s.r.URL.Path

	for _, matcher := range s.p {
		if _, matched := matcher.Match(reqURLPath); matched {
			return PathString(matcher)
		}
	}
	return reqURLPath
}

// PathString returns a string representation of a urlpath.Path. It should exactly match the string passed to
// urlpath.Path(...)
func PathString(p urlpath.Path) string {
	segments := []string{}

	for _, seg := range p.Segments {
		segment := seg.Const

		if seg.IsParam {
			segment = fmt.Sprintf(":%s", seg.Param)
		}

		segments = append(segments, segment)
	}

	if p.Trailing {
		segments = append(segments, "*")
	}

	return strings.Join(segments, "/")
}
