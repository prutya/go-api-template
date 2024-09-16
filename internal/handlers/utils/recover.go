package utils

import (
	"net/http"
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

				logger, loggerOk := GetRequestLogger(r)
				if loggerOk {
					logger.Error("Recovered from panic")
				}

				RenderError(w, r, NewServerError(ErrCodeInternal, http.StatusInternalServerError))
			}()

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}
