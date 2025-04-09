// TODO: Test

package sessions

import (
	"net/http"
	"time"

	"prutya/go-api-template/internal/config"
	"prutya/go-api-template/internal/handlers/utils"
	"prutya/go-api-template/internal/services/authentication_service"
)

func NewDeleteCurrentHandler(
	config *config.Config,
	authenticationService authentication_service.AuthenticationService,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		err := authenticationService.Logout(ctx, utils.GetAccessTokenClaimsFromContext(ctx))

		if err != nil {
			utils.RenderError(w, r, err)
			return
		}

		// Remove the refresh token cookie
		http.SetCookie(w, &http.Cookie{
			Name:     config.AuthenticationRefreshTokenCookieName,
			Domain:   config.AuthenticationRefreshTokenCookieDomain,
			Path:     config.AuthenticationRefreshTokenCookiePath,
			Value:    "",
			Expires:  time.Time{},
			Secure:   config.AuthenticationRefreshTokenCookieSecure,
			HttpOnly: config.AuthenticationRefreshTokenCookieHttpOnly,
			SameSite: config.AuthenticationRefreshTokenCookieSameSite,
		})

		utils.RenderNoContent(w, r, nil)
	}
}
