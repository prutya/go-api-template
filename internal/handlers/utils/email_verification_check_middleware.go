package utils

import (
	"errors"
	"net/http"

	"prutya/go-api-template/internal/services/authentication_service"
)

func NewEmailVerificationCheckMiddleware(authenticationService authentication_service.AuthenticationService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			currentAccessTokenClaims := GetAccessTokenClaimsFromContext(r.Context())

			if err := authenticationService.CheckIfEmailIsVerified(r.Context(), currentAccessTokenClaims.UserID); err != nil {
				if errors.Is(err, authentication_service.ErrEmailNotVerified) {
					RenderError(w, r, NewServerError(err.Error(), http.StatusUnprocessableEntity))
					return
				}

				RenderError(w, r, err)
				return
			}

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}
