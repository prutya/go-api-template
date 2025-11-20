package logger

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"
)

var ErrUnknownLogLevel = errors.New("unknown log level")
var ErrUnknownLogFormat = errors.New("unknown log format")

type Logger struct {
	slog              *slog.Logger
	showCallerInDebug bool
}

func New(levelStr string, format string, showCallerInDebug bool) (*Logger, error) {
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

	logger := &Logger{slog: slogLogger, showCallerInDebug: showCallerInDebug}

	if !Debug && level == slog.LevelDebug {
		logger.Warn("Debug logging was not enabled at compile time. Messages with level \"debug\" will not be printed.")
	}

	return logger, nil
}

func (l *Logger) Debug(msg string, args ...any) {
	if Debug {
		logContext(l, context.Background(), slog.LevelDebug, msg, args...)
	}
}

func (l *Logger) Info(msg string, args ...any) {
	logContext(l, context.Background(), slog.LevelInfo, msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	logContext(l, context.Background(), slog.LevelWarn, msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	logContext(l, context.Background(), slog.LevelError, msg, args...)
}

func (l *Logger) Fatal(msg string, args ...any) {
	logContext(l, context.Background(), slog.LevelError, msg, args...)
	os.Exit(1)
}

func MustDebugContext(ctx context.Context, msg string, args ...any) {
	if Debug {
		logContext(MustFromContext(ctx), ctx, slog.LevelDebug, msg, args...)
	}
}

func MustInfoContext(ctx context.Context, msg string, args ...any) {
	logContext(MustFromContext(ctx), ctx, slog.LevelInfo, msg, args...)
}

func MustWarnContext(ctx context.Context, msg string, args ...any) {
	logContext(MustFromContext(ctx), ctx, slog.LevelWarn, msg, args...)
}

func MustErrorContext(ctx context.Context, msg string, args ...any) {
	logContext(MustFromContext(ctx), ctx, slog.LevelError, msg, args...)
}

func MustFatalContext(ctx context.Context, msg string, args ...any) {
	logContext(MustFromContext(ctx), ctx, slog.LevelError, msg, args...)
	os.Exit(1)
}

func (l *Logger) DebugContext(ctx context.Context, msg string, args ...any) {
	if Debug {
		logContext(l, ctx, slog.LevelDebug, msg, args...)
	}
}

func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	logContext(l, ctx, slog.LevelInfo, msg, args...)
}

func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	logContext(l, ctx, slog.LevelWarn, msg, args...)
}

func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...any) {
	logContext(l, ctx, slog.LevelError, msg, args...)
}

func (l *Logger) FatalContext(ctx context.Context, msg string, args ...any) {
	logContext(l, ctx, slog.LevelError, msg, args...)
	os.Exit(1)
}

func (l *Logger) With(key string, value any) *Logger {
	return &Logger{
		slog:              l.slog.With(key, value),
		showCallerInDebug: l.showCallerInDebug,
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

func logContext(l *Logger, ctx context.Context, level slog.Level, msg string, args ...any) {
	if Debug && l.showCallerInDebug {
		l.slog.Log(ctx, level, msg, append(args, getCallerInfo()...)...)
	} else {
		l.slog.Log(ctx, level, msg, args...)
	}
}

func getCallerInfo() []any {
	_, file, line, ok := runtime.Caller(3)

	if ok {
		return []any{"caller", fmt.Sprintf("%s:%d", file, line)}
	}

	return []any{}
}
