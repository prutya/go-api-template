package account

import (
	"net/http"
	"time"

	"prutya/go-api-template/internal/config"
	"prutya/go-api-template/internal/handlers/account/account_utils"
	"prutya/go-api-template/internal/handlers/utils"
	"prutya/go-api-template/internal/services/authentication_service"
)

func NewLogoutHandler(
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
		account_utils.SetRefreshTokenCookie(config, w, "", time.Time{})

		utils.RenderNoContent(w, r, nil)
	}
}
