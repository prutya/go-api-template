package utils

import (
	"context"
	"net/http"
)

type RequestIdContextKeyType struct{}

type GenerateRequestIDFunc func(*http.Request) (string, error)

var HeaderXRequestID = http.CanonicalHeaderKey("X-Request-Id")
var RequestIDContextKey = RequestIdContextKeyType{}

func NewRequestIDMiddleware(generateRequestID GenerateRequestIDFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			requestId := r.Header.Get(HeaderXRequestID)

			if requestId == "" {
				randomString, err := generateRequestID(r)

				if err != nil {
					RenderError(w, r, err)
					return
				}

				requestId = randomString
			}

			r = SetRequestId(r, requestId)

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

func SetRequestId(r *http.Request, requestId string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), RequestIDContextKey, requestId))
}

func GetRequestId(r *http.Request) (string, bool) {
	requestId, ok := r.Context().Value(RequestIDContextKey).(string)

	return requestId, ok
}
