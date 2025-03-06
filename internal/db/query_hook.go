package db

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/uptrace/bun"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	internal_logger "prutya/go-api-template/internal/logger"
)

type QueryHook struct {
	bun.QueryHook
}

func (qh QueryHook) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
	return ctx
}

func (qh QueryHook) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	logger := internal_logger.MustFromContext(ctx)

	queryDuration := time.Since(event.StartTime)
	query := event.Query

	// Redact secrets provided in context
	if redactStrings, ok := internal_logger.GetContextRedactedSecrets(ctx); ok {
		for _, str := range redactStrings {
			escapedString := EscapeDbString(event.DB, str)
			query = strings.ReplaceAll(query, escapedString, "'[REDACTED]'")
		}
	}

	fields := []zapcore.Field{
		zap.String("operation", event.Operation()),
		zap.Duration("duration", queryDuration),
	}

	if event.Err != nil && !errors.Is(event.Err, sql.ErrNoRows) {
		fields = append(fields, zap.Error(event.Err))
		logger.Error(query, fields...)
		return
	}

	logger.Info(query, fields...)
}

func EscapeDbString(db bun.IDB, original string) string {
	escaped := []byte{}
	escaped = db.Dialect().AppendString(escaped, original)

	return string(escaped)
}
