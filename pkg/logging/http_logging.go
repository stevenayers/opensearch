package logging

import (
	"fmt"
	"github.com/go-kit/kit/log/level"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type (
	// RichResponseWriter encapsulates status code and Response Writer
	RichResponseWriter struct {
		http.ResponseWriter
		StatusCode int
	}
)

// WriteHeader function Writers specified header to response
func (w *RichResponseWriter) WriteHeader(code int) {
	w.StatusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// NewRichResponseWriter function creates a new RichResponseWriter
func NewRichResponseWriter(w http.ResponseWriter) *RichResponseWriter {
	return &RichResponseWriter{w, http.StatusOK}
}

// HttpResponseLogger creates a custom logger which outputs HTTP response info as a json log to stdout.
func HttpResponseLogger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestUid := uuid.New()
		r.Header.Add("OpenSearch-Request-ID", requestUid.String())
		rw := NewRichResponseWriter(w)
		handler.ServeHTTP(rw, r)
		_ = level.Info(Logger).Log(
			"requestUid", requestUid.String(),
			"uri", r.URL.Path+"?"+r.URL.RawQuery,
			"StatusCode", rw.StatusCode,
			"method", r.Method,
			"duration", fmt.Sprintf("%s", time.Since(start)),
		)
	})
}
