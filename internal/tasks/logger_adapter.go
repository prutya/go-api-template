package tasks

import (
	"context"
	stringspkg "strings"

	loggerpkg "prutya/go-api-template/internal/logger"
)

type slogLoggerAdapter struct {
	logger *loggerpkg.Logger
}

func NewSlogLoggerAdapter(logger *loggerpkg.Logger) *slogLoggerAdapter {
	return &slogLoggerAdapter{logger: logger}
}

func (l *slogLoggerAdapter) Debug(args ...any) {
	l.logger.DebugContext(context.Background(), adaptLogEntry(args...))
}

func (l *slogLoggerAdapter) Info(args ...any) {
	l.logger.InfoContext(context.Background(), adaptLogEntry(args...))
}

func (l *slogLoggerAdapter) Warn(args ...any) {
	l.logger.WarnContext(context.Background(), adaptLogEntry(args...))
}

func (l *slogLoggerAdapter) Error(args ...any) {
	l.logger.ErrorContext(context.Background(), adaptLogEntry(args...))
}

func (l *slogLoggerAdapter) Fatal(args ...any) {
	l.logger.FatalContext(context.Background(), adaptLogEntry(args...))
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
