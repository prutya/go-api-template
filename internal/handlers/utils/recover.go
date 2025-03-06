package utils

import (
	"net/http"

	internal_logger "prutya/go-api-template/internal/logger"
)

func NewRecoverMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				rvr := recover()

				if rvr == nil {
					return
				}

				if rvr == http.ErrAbortHandler {
					// we don't recover http.ErrAbortHandler so the response
					// to the client is aborted, this should not be logged
					panic(rvr)
				}

				logger := internal_logger.MustFromContext(r.Context())
				logger.Error("Recovered from panic")

				RenderError(w, r, NewServerError(ErrCodeInternal, http.StatusInternalServerError))
			}()

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}
