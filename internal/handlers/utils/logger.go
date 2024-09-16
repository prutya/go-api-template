package utils

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"

	internallogger "prutya/todo/internal/logger"
)

type ResponseInfo struct {
	HttpStatus int
	ErrorCode  string
	InnerError error
}

type ResponseInfoContextKey struct{}

func NewLoggerMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			requestId, ok := GetRequestId(r)

			requestLogger := logger

			if ok {
				requestLogger = requestLogger.With(zap.String("request_id", requestId))
			}

			responseInfo := new(ResponseInfo)

			r = SetRequestLogger(r, logger)
			r = SetRequestResponseInfo(r, responseInfo)

			requestLogger.Info(
				"Request started",
				zap.String("method", r.Method),
				zap.String("url", r.URL.String()),
			)

			start := time.Now()

			next.ServeHTTP(w, r)

			duration := time.Since(start)

			if responseInfo.ErrorCode != "" {
				requestLogger.Info(
					"Request ended",
					zap.Duration("duration", duration),
					zap.Int("status", responseInfo.HttpStatus),
					zap.String("error_code", responseInfo.ErrorCode),
					zap.Error(responseInfo.InnerError),
				)
			} else {
				requestLogger.Info(
					"Request ended",
					zap.Duration("duration", duration),
					zap.Int("status", responseInfo.HttpStatus),
				)
			}
		}

		return http.HandlerFunc(fn)
	}
}

func SetRequestLogger(r *http.Request, logger *zap.Logger) *http.Request {
	ctxWithLogger := internallogger.SetContextLogger(r.Context(), logger)

	return r.WithContext(ctxWithLogger)
}

func GetRequestLogger(r *http.Request) (*zap.Logger, bool) {
	logger, ok := internallogger.GetContextLogger(r.Context())

	return logger, ok
}

func SetRequestResponseInfo(r *http.Request, responseInfo *ResponseInfo) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), ResponseInfoContextKey{}, responseInfo))
}

func GetRequestResponseInfo(r *http.Request) (*ResponseInfo, bool) {
	ri, ok := r.Context().Value(ResponseInfoContextKey{}).(*ResponseInfo)

	return ri, ok
}
