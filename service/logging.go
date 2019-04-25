package service

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/uuid"
	"net/http"
	"os"
	"time"
)

// APILogger makes a global ApiLogger
var APILogger ApiLogger

type (
	richResponseWriter struct {
		http.ResponseWriter
		statusCode int
	}

	// ApiLogger holds the logger used by API and Service
	ApiLogger struct {
		Logger log.Logger
	}
)

func (w *richResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func newRichResponseWriter(w http.ResponseWriter) *richResponseWriter {
	return &richResponseWriter{w, http.StatusOK}
}

// InitJsonLogger function initiates a structured JSON logger, taking in the specified log level for what is displayed at runtime.
func (apiLogger *ApiLogger) InitJsonLogger(logLevel string) {
	apiLogger.Logger = log.NewJSONLogger(log.NewSyncWriter(os.Stdout))
	apiLogger.Logger = log.With(
		apiLogger.Logger,
		"service", "clamber-api",
		"node", uuid.New().String(),
	)
	switch logLevel {
	case "debug":
		apiLogger.Logger = level.NewFilter(apiLogger.Logger, level.AllowDebug())
	case "info":
		apiLogger.Logger = level.NewFilter(apiLogger.Logger, level.AllowInfo())
	case "error":
		apiLogger.Logger = level.NewFilter(apiLogger.Logger, level.AllowError())
	default:
		apiLogger.Logger = level.NewFilter(apiLogger.Logger, level.AllowInfo())
	}
}

// LogDebug function takes in a struct of keyvals and outputs a json log response to stdout at log level debug.
func (apiLogger *ApiLogger) LogDebug(keyvals ...interface{}) {
	_ = level.Debug(apiLogger.Logger).Log(keyvals...)
}

// LogInfo function takes in a struct of keyvals and outputs a json log response to stdout at log level info.
func (apiLogger *ApiLogger) LogInfo(keyvals ...interface{}) {
	_ = level.Info(apiLogger.Logger).Log(keyvals...)
}

// LogError function takes in a struct of keyvals and outputs a json log response to stdout at log level error.
func (apiLogger *ApiLogger) LogError(keyvals ...interface{}) {
	_ = level.Info(apiLogger.Logger).Log(keyvals...)
}

// HttpResponseLogger creates a custom logger which outputs HTTP response info as a json log to stdout.
func HttpResponseLogger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestUid := uuid.New()
		r.Header.Add("Clamber-Request-ID", requestUid.String())
		rw := newRichResponseWriter(w)
		handler.ServeHTTP(rw, r)
		APILogger.LogInfo(
			"uid", requestUid.String(),
			"uri", r.URL.Path+"?"+r.URL.RawQuery,
			"statusCode", rw.statusCode,
			"method", r.Method,
			"duration", fmt.Sprintf("%s", time.Since(start)),
		)
	})
}
