package main

import (
	"os"

	"github.com/pbergman/logger"
)

func GetLogger(config *Config) *logger.Logger {

	var handler = logger.NewWriterHandler(os.Stdout, logger.LogLevelDebug(), true)

	if false == config.Debug {
		handler = logger.NewThresholdHandler(handler, 10, logger.Error, true)
	}

	return logger.NewLogger("app", handler)
}
