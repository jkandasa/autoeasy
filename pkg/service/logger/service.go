package service

import (
	loggerUtils "github.com/jkandasa/autoeasy/pkg/utils/logger"
	"go.uber.org/zap"
)

// Load logger
func Load(logLevel string, enableStacktrace bool) {
	logger := loggerUtils.GetLogger("record_all", logLevel, "console", false, 0, enableStacktrace)
	zap.ReplaceGlobals(logger)
}
