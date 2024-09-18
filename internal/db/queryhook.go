package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/uptrace/bun"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	loggerpkg "prutya/go-api-template/internal/logger"
)

type QueryHook struct {
	bun.QueryHook
}

func (qh QueryHook) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
	return ctx
}

func (qh QueryHook) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	logger, ok := loggerpkg.GetContextLogger(ctx)
	if !ok {
		fmt.Println("WARN: Database logger is not configured")
		return
	}

	queryDuration := time.Since(event.StartTime)
	query := event.Query

	// Redact secrets provided in context
	if redactStrings, _ := loggerpkg.GetContextRedactedSecrets(ctx); ok {
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
