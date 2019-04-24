/*
Package logging provides custom logging to the clamber & api package. It's mostly a thin wrapper around go-kit's log.
It also implements a richResponseWriter which allows us to log the HTTP status code.
*/
package logging

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/uuid"
	"net/http"
	"os"
	"time"
)

var apiLogger log.Logger

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

// Initiate a structured JSON logger, taking in the specified log level for what is displayed at runtime.
func InitJsonLogger(logLevel string) {
	apiLogger = log.NewJSONLogger(log.NewSyncWriter(os.Stdout))
	apiLogger = log.With(
		apiLogger,
		"service", "clamber-api",
		"node", uuid.New().String(),
	)
	switch logLevel {
	case "debug":
		apiLogger = level.NewFilter(apiLogger, level.AllowDebug())
	case "info":
		apiLogger = level.NewFilter(apiLogger, level.AllowInfo())
	case "error":
		apiLogger = level.NewFilter(apiLogger, level.AllowError())
	default:
		apiLogger = level.NewFilter(apiLogger, level.AllowInfo())
	}
}

// Takes in a struct of keyvals and outputs a json log response to stdout at log level debug.
func LogDebug(keyvals ...interface{}) {
	_ = level.Debug(apiLogger).Log(keyvals...)
}

// Takes in a struct of keyvals and outputs a json log response to stdout at log level info.
func LogInfo(keyvals ...interface{}) {
	_ = level.Info(apiLogger).Log(keyvals...)
}

// Takes in a struct of keyvals and outputs a json log response to stdout at log level error.
func LogError(keyvals ...interface{}) {
	_ = level.Info(apiLogger).Log(keyvals...)
}

// Custom logger which outputs HTTP response info as a json log to stdout.
func HttpResponseLogger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestUid := uuid.New()
		r.Header.Add("Clamber-Request-ID", requestUid.String())
		rw := newRichResponseWriter(w)
		handler.ServeHTTP(rw, r)
		LogInfo(
			"uid", requestUid.String(),
			"uri", r.URL.Path+"?"+r.URL.RawQuery,
			"statusCode", rw.statusCode,
			"method", r.Method,
			"duration", fmt.Sprintf("%s", time.Since(start)),
		)
	})
}
