package logger

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggerContextKeyType struct{}

var loggerContextKey = LoggerContextKeyType{}

var ErrNoLoggerInContext = errors.New("no logger in context")

type LogsRedactKey struct{}

var logsRedactKey = LogsRedactKey{}

var ErrNoRedactedSecretsInContext = errors.New("no redacted secrets in context")

func New(level string, timeFormat string) (*zap.Logger, error) {
	loggerCfg := zap.NewProductionConfig()

	loggerLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		return nil, err
	}

	loggerCfg.Level.SetLevel(loggerLevel)

	var loggerTimeEncoder zapcore.TimeEncoder

	if err := loggerTimeEncoder.UnmarshalText([]byte(timeFormat)); err != nil {
		return nil, err
	}

	loggerCfg.EncoderConfig.EncodeTime = loggerTimeEncoder

	logger, err := loggerCfg.Build()

	if err != nil {
		return nil, err
	}

	return logger, nil
}

func NewContext(c context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(c, loggerContextKey, logger)
}

func MustFromContext(c context.Context) *zap.Logger {
	val, err := FromContext(c)

	if err != nil {
		panic(err)
	}

	return val
}

func FromContext(c context.Context) (*zap.Logger, error) {
	val, ok := c.Value(loggerContextKey).(*zap.Logger)

	if !ok {
		return nil, ErrNoLoggerInContext
	}

	return val, nil
}

func NewContextWithRedactedSecret(ctx context.Context, secret string) context.Context {
	existingSecrets, ok := ctx.Value(logsRedactKey).([]string)

	if ok {
		existingSecrets = append(existingSecrets, secret)
	} else {
		existingSecrets = []string{secret}
	}

	return context.WithValue(ctx, logsRedactKey, existingSecrets)
}

func GetContextRedactedSecrets(ctx context.Context) ([]string, bool) {
	redactStrings, ok := ctx.Value(logsRedactKey).([]string)

	return redactStrings, ok
}
