package db

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/uptrace/bun"

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

	if event.Err != nil && !errors.Is(event.Err, sql.ErrNoRows) {
		logger.ErrorContext(ctx, "SQL query error", "query", query, "duration", queryDuration, "error", event.Err)
		return
	}

	logger.DebugContext(ctx, "SQL query", "query", query, "duration", queryDuration)
}
