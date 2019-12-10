package logging

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/uuid"
	"io"
)

var Logger log.Logger

// InitJsonLogger function initiates a structured JSON logger, taking in the specified log level for what is displayed at runtime.
func InitJsonLogger(writer io.Writer, logLevel string, component string) {
	logger := log.NewJSONLogger(writer)
	logger = log.With(logger, "timestamp", log.DefaultTimestampUTC)
	logger = log.With(
		logger,
		"app", "clamber",
		"node", uuid.New().String(),
		"component", component,
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
	Logger = logger
}
