package logger

import (
	"context"
	"log/slog"
	"strings"
)

func NewContextWithRedactedSecret(ctx context.Context, secret string) context.Context {
	secretsInContext := getContextRedactedSecrets(ctx)

	secretsInContext[secret] = struct{}{}

	return context.WithValue(ctx, logsRedactKey, secretsInContext)
}

type RedactingHandler struct {
	slog.Handler
}

func (h *RedactingHandler) Handle(ctx context.Context, r slog.Record) error {
	secretsToRedact := getContextRedactedSecrets(ctx)

	if len(secretsToRedact) == 0 {
		return h.Handler.Handle(ctx, r)
	}

	// Copy record to modify it
	newAttrs := make([]slog.Attr, 0, r.NumAttrs())

	r.Attrs(func(attr slog.Attr) bool {
		if attr.Value.Kind() == slog.KindString {
			redactedValue := redactQuery(attr.Value.String(), secretsToRedact)
			attr = slog.String(attr.Key, redactedValue)
		}

		newAttrs = append(newAttrs, attr)

		return true
	})

	// Create a new record with redacted attributes
	newRecord := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)

	for _, attr := range newAttrs {
		newRecord.AddAttrs(attr)
	}

	return h.Handler.Handle(ctx, newRecord)
}

func (h *RedactingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &RedactingHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h *RedactingHandler) WithGroup(name string) slog.Handler {
	return &RedactingHandler{Handler: h.Handler.WithGroup(name)}
}

type logsRedactKeyType struct{}

var logsRedactKey = logsRedactKeyType{}

func getContextRedactedSecrets(ctx context.Context) secretsSet {
	redactStrings, ok := ctx.Value(logsRedactKey).(secretsSet)

	if !ok {
		return secretsSet{}
	}

	return redactStrings
}

type secretsSet map[string]struct{}

func redactQuery(query string, secrets secretsSet) string {
	for secret := range secrets {
		query = strings.ReplaceAll(query, secret, "REDACTED")
	}

	return query
}
