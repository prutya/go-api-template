package logger

import (
	"errors"
	"log/slog"
	"os"
	"strings"
)

var ErrUnknownLogLevel = errors.New("unknown log level")
var ErrUnknownLogFormat = errors.New("unknown log format")

func New(levelStr string, format string) (*slog.Logger, error) {
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

	handler := &RedactingHandler{
		Handler: baseHandler,
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger, nil
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
