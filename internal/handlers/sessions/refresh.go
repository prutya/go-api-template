// TODO: Test

package sessions

import (
	"errors"
	"net/http"

	"prutya/go-api-template/internal/config"
	"prutya/go-api-template/internal/handlers/utils"
	"prutya/go-api-template/internal/services/authentication_service"
)

type RefreshResponse struct {
	AccessToken string `json:"accessToken"`
}

func NewRefreshHandler(
	config *config.Config,
	authenticationService authentication_service.AuthenticationService,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read the refresh token cookie
		refreshTokenCookie, err := r.Cookie(config.AuthenticationRefreshTokenCookieName)

		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				utils.RenderError(w, r, utils.ErrUnauthorized)
				return
			}

			utils.RenderError(w, r, err)
			return
		}

		newRefreshToken, newRefreshTokenExpiresAt, newAccessToken, err := authenticationService.Refresh(r.Context(), refreshTokenCookie.Value)

		if err != nil {
			if errors.Is(err, authentication_service.ErrInvalidRefreshTokenClaims) ||
				errors.Is(err, authentication_service.ErrRefreshTokenNotFound) ||
				errors.Is(err, authentication_service.ErrRefreshTokenInvalid) ||
				errors.Is(err, authentication_service.ErrRefreshTokenRevoked) ||
				errors.Is(err, authentication_service.ErrSessionNotFound) ||
				errors.Is(err, authentication_service.ErrSessionTerminated) {

				utils.RenderError(w, r, utils.ErrUnauthorized)
				return
			}

			utils.RenderError(w, r, err)
			return
		}

		// Set the refresh token in a cookie
		http.SetCookie(w, &http.Cookie{
			Name:     config.AuthenticationRefreshTokenCookieName,
			Domain:   config.AuthenticationRefreshTokenCookieDomain,
			Path:     config.AuthenticationRefreshTokenCookiePath,
			Value:    newRefreshToken,
			Expires:  newRefreshTokenExpiresAt,
			Secure:   config.AuthenticationRefreshTokenCookieSecure,
			HttpOnly: config.AuthenticationRefreshTokenCookieHttpOnly,
			SameSite: config.AuthenticationRefreshTokenCookieSameSite,
		})

		responseBody := RefreshResponse{
			AccessToken: newAccessToken,
		}

		utils.RenderJson(w, r, responseBody, http.StatusOK, nil)
	}
}
