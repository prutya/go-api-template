package logger

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggerContextKey struct{}
type LogsRedactKey struct{}

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

func SetContextLogger(c context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(c, LoggerContextKey{}, logger)
}

func GetContextLogger(c context.Context) (*zap.Logger, bool) {
	logger, ok := c.Value(LoggerContextKey{}).(*zap.Logger)

	return logger, ok
}

func ContextWithRedactedSecret(ctx context.Context, secret string) context.Context {
	existingSecrets, ok := ctx.Value(LogsRedactKey{}).([]string)

	if ok {
		existingSecrets = append(existingSecrets, secret)
	} else {
		existingSecrets = []string{secret}
	}

	return context.WithValue(ctx, LogsRedactKey{}, existingSecrets)
}

func GetContextRedactedSecrets(ctx context.Context) ([]string, bool) {
	redactStrings, ok := ctx.Value(LogsRedactKey{}).([]string)

	return redactStrings, ok
}
