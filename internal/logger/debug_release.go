//go:build !debug

package logger

import "context"

const Debug = false

// No-op in release mode
func debugContext(l *Logger, ctx context.Context, msg string, args ...any) {
}

// No-op in release mode
func mustDebugContext(ctx context.Context, msg string, args ...any) {
}
