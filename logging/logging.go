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

func LogDebug(keyvals ...interface{}) {
	_ = level.Debug(apiLogger).Log(keyvals...)
}

func LogInfo(keyvals ...interface{}) {
	_ = level.Info(apiLogger).Log(keyvals...)
}

func LogError(keyvals ...interface{}) {
	_ = level.Info(apiLogger).Log(keyvals...)
}

func Logger(handler http.Handler) http.Handler {
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
