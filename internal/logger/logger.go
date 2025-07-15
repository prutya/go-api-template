package logger

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"
)

var ErrUnknownLogLevel = errors.New("unknown log level")
var ErrUnknownLogFormat = errors.New("unknown log format")

type Logger struct {
	slog *slog.Logger
}

func New(levelStr string, format string) (*Logger, error) {
	level, err := parseLevel(levelStr)
	if err != nil {
		return nil, err
	}

	var baseHandler slog.Handler

	if format == "text" {
		baseHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
	} else if format == "json" {
		baseHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
	} else {
		return nil, ErrUnknownLogFormat
	}

	slogLogger := slog.New(baseHandler)
	slog.SetDefault(slogLogger)

	return &Logger{slog: slogLogger}, nil
}

func parseLevel(levelStr string) (slog.Level, error) {
	switch strings.ToLower(levelStr) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, ErrUnknownLogLevel
	}
}

func (l *Logger) DebugContext(ctx context.Context, msg string, args ...any) {
	debugContext(l, ctx, msg, args...)
}

func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.slog.InfoContext(ctx, msg, args...)
}

func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.slog.WarnContext(ctx, msg, args...)
}

func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.slog.ErrorContext(ctx, msg, args...)
}

func (l *Logger) FatalContext(ctx context.Context, msg string, args ...any) {
	l.slog.ErrorContext(ctx, msg, args...)
	os.Exit(1)
}

func (l *Logger) With(key string, value any) *Logger {
	return &Logger{
		slog: l.slog.With(key, value),
	}
}
