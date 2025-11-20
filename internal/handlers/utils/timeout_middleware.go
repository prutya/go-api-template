package utils

import (
	"context"
	"net/http"
	"time"
)

func NewTimeoutMiddleware(timeout time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)

			defer func() {
				cancel()

				if ctx.Err() == context.DeadlineExceeded {
					RenderError(w, r, ErrTimeout)
				}
			}()

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
