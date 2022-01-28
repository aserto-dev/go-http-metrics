package std

import (
	"bufio"
	"errors"
	"net"
	"net/http"
)

// ResponseWriterInterceptor is a simple wrapper to intercept set data on a
// ResponseWriter.
type ResponseWriterInterceptor struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
}

func NewResponseWriterInterceptor(w http.ResponseWriter) *ResponseWriterInterceptor {
	return &ResponseWriterInterceptor{
		statusCode:     http.StatusOK,
		ResponseWriter: w,
	}
}

func (w *ResponseWriterInterceptor) StatusCode() int {
	return w.statusCode
}

func (w *ResponseWriterInterceptor) BytesWritten() int64 {
	return w.bytesWritten
}

func (w *ResponseWriterInterceptor) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *ResponseWriterInterceptor) Write(p []byte) (int, error) {
	w.bytesWritten += int64(len(p))
	return w.ResponseWriter.Write(p)
}

func (w *ResponseWriterInterceptor) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("type assertion failed http.ResponseWriter not a http.Hijacker")
	}
	return h.Hijack()
}

func (w *ResponseWriterInterceptor) Flush() {
	f, ok := w.ResponseWriter.(http.Flusher)
	if !ok {
		return
	}

	f.Flush()
}

// Check interface implementations.
var (
	_ http.ResponseWriter = &ResponseWriterInterceptor{}
	_ http.Hijacker       = &ResponseWriterInterceptor{}
	_ http.Flusher        = &ResponseWriterInterceptor{}
)
