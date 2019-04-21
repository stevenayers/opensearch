package logger

import (
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

func Logger(handler http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := newRichResponseWriter(w)
		handler.ServeHTTP(rw, r)
		log.Printf(
			"%s\t%d\t%s\t%s\t\t%s%s",
			r.Method,
			rw.statusCode,
			name,
			time.Since(start),
			r.URL.Path,
			"?"+r.URL.RawQuery, // This will add a question mark onto any request, regardless of query params, not good.
		)

	})
}
