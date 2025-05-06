package logger

import (
	"context"
	"errors"
	"log/slog"
)

type loggerContextKeyType struct{}

var loggerContextKey = loggerContextKeyType{}

var ErrNoLoggerInContext = errors.New("no logger in context")

func NewContext(c context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(c, loggerContextKey, logger)
}

func MustFromContext(c context.Context) *slog.Logger {
	val, err := FromContext(c)

	if err != nil {
		panic(err)
	}

	return val
}

func FromContext(c context.Context) (*slog.Logger, error) {
	val, ok := c.Value(loggerContextKey).(*slog.Logger)

	if !ok {
		return nil, ErrNoLoggerInContext
	}

	return val, nil
}
