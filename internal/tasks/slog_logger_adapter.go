package tasks

import (
	"log/slog"
	"os"
	stringspkg "strings"
)

type slogLoggerAdapter struct {
	logger *slog.Logger
}

func NewSlogLoggerAdapter(logger *slog.Logger) *slogLoggerAdapter {
	return &slogLoggerAdapter{logger: logger}
}

func (l *slogLoggerAdapter) Debug(args ...any) {
	l.logger.Debug(adaptLogEntry(args...))
}

func (l *slogLoggerAdapter) Info(args ...any) {
	l.logger.Info(adaptLogEntry(args...))
}

func (l *slogLoggerAdapter) Warn(args ...any) {
	l.logger.Warn(adaptLogEntry(args...))
}

func (l *slogLoggerAdapter) Error(args ...any) {
	l.logger.Error(adaptLogEntry(args...))
}

func (l *slogLoggerAdapter) Fatal(args ...any) {
	l.logger.Error(adaptLogEntry(args...))
	os.Exit(1)
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
