package service

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/uuid"
	"io"
)

// APILogger makes a global ApiLogger
var APILogger ApiLogger

type (

	// ApiLogger holds the logger used by API and Service
	ApiLogger struct {
		Logger log.Logger
	}
)

// InitJsonLogger function initiates a structured JSON logger, taking in the specified log level for what is displayed at runtime.
func (apiLogger *ApiLogger) InitJsonLogger(writer io.Writer, logLevel string) {
	apiLogger.Logger = log.NewJSONLogger(writer)
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
func (apiLogger *ApiLogger) LogDebug(keyvals ...interface{}) (err error) {
	err = level.Debug(apiLogger.Logger).Log(keyvals...)
	return
}

// LogInfo function takes in a struct of keyvals and outputs a json log response to stdout at log level info.
func (apiLogger *ApiLogger) LogInfo(keyvals ...interface{}) (err error) {
	err = level.Info(apiLogger.Logger).Log(keyvals...)
	return
}

// LogError function takes in a struct of keyvals and outputs a json log response to stdout at log level error.
func (apiLogger *ApiLogger) LogError(keyvals ...interface{}) (err error) {
	err = level.Error(apiLogger.Logger).Log(keyvals...)
	return
}
