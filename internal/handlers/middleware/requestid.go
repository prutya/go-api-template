package middleware

import (
	"context"
	"net/http"
	"prutya/todo/internal/handlers/utils"
)

type RequestIdContextKeyType struct{}

type GenerateRequestIDFunc func(*http.Request) (string, error)

var RequestIdContextKey = RequestIdContextKeyType{}

func NewRequestID(generateRequestID GenerateRequestIDFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			requestId := r.Header.Get("x-request-id")

			if requestId == "" {
				randomString, err := generateRequestID(r)

				if err != nil {
					utils.RenderError(w, r, err)
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
	return r.WithContext(context.WithValue(r.Context(), RequestIdContextKey, requestId))
}

func GetRequestId(r *http.Request) (string, bool) {
	requestId, ok := r.Context().Value(RequestIdContextKey).(string)

	return requestId, ok
}
