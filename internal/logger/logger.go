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

	switch format {
	case "text":
		baseHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
	case "json":
		baseHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
	default:
		return nil, ErrUnknownLogFormat
	}

	slogLogger := slog.New(baseHandler)
	slog.SetDefault(slogLogger)

	logger := &Logger{slog: slogLogger}

	if !Debug && level == slog.LevelDebug {
		logger.Warn("Debug logging was not enabled at compile time. The debug messages will not be printed.")
	}

	return logger, nil
}

func (l *Logger) Debug(msg string, args ...any) {
	debugContext(l, context.Background(), msg, args...)
}

func (l *Logger) Info(msg string, args ...any) {
	l.InfoContext(context.Background(), msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.WarnContext(context.Background(), msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.ErrorContext(context.Background(), msg, args...)
}

func (l *Logger) Fatal(msg string, args ...any) {
	l.ErrorContext(context.Background(), msg, args...)
	os.Exit(1)
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

func MustDebugContext(ctx context.Context, msg string, args ...any) {
	mustDebugContext(ctx, msg, args...)
}

func MustInfoContext(ctx context.Context, msg string, args ...any) {
	MustFromContext(ctx).InfoContext(ctx, msg, args...)
}

func MustWarnContext(ctx context.Context, msg string, args ...any) {
	MustFromContext(ctx).WarnContext(ctx, msg, args...)
}

func MustErrorContext(ctx context.Context, msg string, args ...any) {
	MustFromContext(ctx).ErrorContext(ctx, msg, args...)
}

func MustFatalContext(ctx context.Context, msg string, args ...any) {
	MustFromContext(ctx).ErrorContext(ctx, msg, args...)
	os.Exit(1)
}

func (l *Logger) With(key string, value any) *Logger {
	return &Logger{
		slog: l.slog.With(key, value),
	}
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
