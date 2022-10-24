package urlpath_test

import (
	"testing"

	metrics "github.com/aserto-dev/go-http-metrics/middleware/urlpath"
	"github.com/ucarion/urlpath"

	"github.com/stretchr/testify/assert"
)

func TestPathString(t *testing.T) {
	paths := []string{
		"/fixed/path",
		"/path/prefix/*",
		"/path/with/:param",
		"/path/:param/and/prefix/*",
		"*",
		"/",
	}

	for _, path := range paths {
		t.Run(path, verifyPath(path))
	}
}

func verifyPath(path string) func(t *testing.T) {
	return func(t *testing.T) {
		urlPath := urlpath.New(path)
		assert.Equal(t, path, metrics.PathString(urlPath))
	}
}
