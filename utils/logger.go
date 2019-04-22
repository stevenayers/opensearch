package utils

import (
	"github.com/google/uuid"
	"log"
	"net/http"
	"time"
)

type richResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *richResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func newRichResponseWriter(w http.ResponseWriter) *richResponseWriter {
	return &richResponseWriter{w, http.StatusOK}
}

func Logger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestUid := uuid.New()
		r.Header.Add("Clamber-Request-ID", requestUid.String())
		rw := newRichResponseWriter(w)
		handler.ServeHTTP(rw, r)
		log.Printf(
			"%s: %s%s (status code: %d method: %s duration: %s)",
			requestUid.String(),
			r.URL.Path, "?"+r.URL.RawQuery,
			rw.statusCode,
			r.Method,
			time.Since(start),
		)

	})
}
