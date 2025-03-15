package tasks

import (
	stringspkg "strings"

	"go.uber.org/zap"
)

type LoggerAdapter struct {
	logger *zap.Logger
}

func NewLoggerAdapter(logger *zap.Logger) *LoggerAdapter {
	return &LoggerAdapter{logger: logger}
}

func (l *LoggerAdapter) Debug(args ...any) {
	l.logger.Debug(adaptLogEntry(args...))
}

func (l *LoggerAdapter) Info(args ...any) {
	l.logger.Info(adaptLogEntry(args...))
}

func (l *LoggerAdapter) Warn(args ...any) {
	l.logger.Warn(adaptLogEntry(args...))
}

func (l *LoggerAdapter) Error(args ...any) {
	l.logger.Error(adaptLogEntry(args...))
}

func (l *LoggerAdapter) Fatal(args ...any) {
	l.logger.Fatal(adaptLogEntry(args...))
}

func adaptLogEntry(args ...any) string {
	strings := make([]string, len(args))

	// Convert all arguments to strings, ingoring those that can't be converted
	for i, arg := range args {
		stringArg, ok := arg.(string)

		if !ok {
			continue
		}

		strings[i] = stringArg
	}

	return stringspkg.Join(strings, " ")
}
