package utils

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	loggerpkg "prutya/go-api-template/internal/logger"
)

type ResponseInfo struct {
	HttpStatus int
	ErrorCode  string
	InnerError error
}

type responseInfoContextKeyType struct{}

var responseInfoContextKey = responseInfoContextKeyType{}

func NewLoggerMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			requestLogger := logger

			// Add request_id to the logger output
			if requestId, requestIdOk := GetRequestId(r); requestIdOk {
				requestLogger = requestLogger.With("request_id", requestId)
			}

			responseInfo := new(ResponseInfo)

			// Store the logger in the request context to potentially be used by
			// handlers down the middleware stack
			r = SetRequestLogger(r, requestLogger)
			r = SetRequestResponseInfo(r, responseInfo)

			requestLogger.InfoContext(
				r.Context(),
				"Request started",
				"method", r.Method,
				"url", r.URL.String(),
			)

			// Measure the request duration
			start := time.Now()

			next.ServeHTTP(w, r)

			duration := time.Since(start)

			if responseInfo.ErrorCode != "" {
				requestLogger.InfoContext(
					r.Context(),
					"Request ended",
					"duration", duration,
					"status", responseInfo.HttpStatus,
					"error_code", responseInfo.ErrorCode,
					"error", responseInfo.InnerError,
				)
			} else {
				requestLogger.InfoContext(
					r.Context(),
					"Request ended",
					"duration", duration,
					"status", responseInfo.HttpStatus,
				)
			}
		}

		return http.HandlerFunc(fn)
	}
}

func SetRequestLogger(r *http.Request, logger *slog.Logger) *http.Request {
	ctxWithLogger := loggerpkg.NewContext(r.Context(), logger)

	return r.WithContext(ctxWithLogger)
}

func SetRequestResponseInfo(r *http.Request, responseInfo *ResponseInfo) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), responseInfoContextKey, responseInfo))
}

func GetRequestResponseInfo(r *http.Request) (*ResponseInfo, bool) {
	ri, ok := r.Context().Value(responseInfoContextKey).(*ResponseInfo)

	return ri, ok
}
