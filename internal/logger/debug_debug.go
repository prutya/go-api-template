//go:build debug

package logger

import "context"

const Debug = true

func debugContext(l *Logger, ctx context.Context, msg string, args ...any) {
	l.slog.DebugContext(ctx, msg, args...)
}
