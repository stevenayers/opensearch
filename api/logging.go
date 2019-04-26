package api

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/stevenayers/clamber/service"
	"net/http"
	"time"
)

type (
	richResponseWriter struct {
		http.ResponseWriter
		statusCode int
	}
)

func (w *richResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func newRichResponseWriter(w http.ResponseWriter) *richResponseWriter {
	return &richResponseWriter{w, http.StatusOK}
}

// HttpResponseLogger creates a custom logger which outputs HTTP response info as a json log to stdout.
func HttpResponseLogger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestUid := uuid.New()
		r.Header.Add("Clamber-Request-ID", requestUid.String())
		rw := newRichResponseWriter(w)
		handler.ServeHTTP(rw, r)
		service.APILogger.LogInfo(
			"uid", requestUid.String(),
			"uri", r.URL.Path+"?"+r.URL.RawQuery,
			"statusCode", rw.statusCode,
			"method", r.Method,
			"duration", fmt.Sprintf("%s", time.Since(start)),
		)
	})
}
