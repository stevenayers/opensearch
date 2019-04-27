package api

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/uuid"
	"io"
	"net/http"
	"os"
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
		r.Header.Add("Clamber-Request-ID", requestUid.String())
		logger := InitJsonLogger(log.NewSyncWriter(os.Stdout), "info")
		rw := NewRichResponseWriter(w)
		handler.ServeHTTP(rw, r)
		_ = level.Info(logger).Log(
			"requestUid", requestUid.String(),
			"uri", r.URL.Path+"?"+r.URL.RawQuery,
			"StatusCode", rw.StatusCode,
			"method", r.Method,
			"duration", fmt.Sprintf("%s", time.Since(start)),
		)
	})
}

// InitJsonLogger function initiates a structured JSON logger, taking in the specified log level for what is displayed at runtime.
func InitJsonLogger(writer io.Writer, logLevel string) (logger log.Logger) {
	logger = log.NewJSONLogger(writer)
	logger = log.With(
		logger,
		"service", "clamber-api",
		"node", uuid.New().String(),
	)
	switch logLevel {
	case "debug":
		logger = level.NewFilter(logger, level.AllowDebug())
	case "info":
		logger = level.NewFilter(logger, level.AllowInfo())
	case "error":
		logger = level.NewFilter(logger, level.AllowError())
	default:
		logger = level.NewFilter(logger, level.AllowInfo())
	}
	return
}
