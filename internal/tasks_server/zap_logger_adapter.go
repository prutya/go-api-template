package tasks_server

import (
	stringspkg "strings"

	"go.uber.org/zap"
)

type zapLoggerAdapter struct {
	logger *zap.Logger
}

func newZapLoggerAdapter(logger *zap.Logger) *zapLoggerAdapter {
	return &zapLoggerAdapter{logger: logger}
}

func (l *zapLoggerAdapter) Debug(args ...any) {
	l.logger.Debug(adaptLogEntry(args...))
}

func (l *zapLoggerAdapter) Info(args ...any) {
	l.logger.Info(adaptLogEntry(args...))
}

func (l *zapLoggerAdapter) Warn(args ...any) {
	l.logger.Warn(adaptLogEntry(args...))
}

func (l *zapLoggerAdapter) Error(args ...any) {
	l.logger.Error(adaptLogEntry(args...))
}

func (l *zapLoggerAdapter) Fatal(args ...any) {
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
