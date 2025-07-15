package account

import (
	"errors"
	"net/http"

	"prutya/go-api-template/internal/config"
	"prutya/go-api-template/internal/handlers/account/account_utils"
	"prutya/go-api-template/internal/handlers/utils"
	"prutya/go-api-template/internal/logger"
	"prutya/go-api-template/internal/services/authentication_service"
)

type RefreshSessionResponse struct {
	AccessToken string `json:"accessToken"`
	CSRFToken   string `json:"csrfToken"`
}

func NewRefreshSessionHandler(
	config *config.Config,
	authenticationService authentication_service.AuthenticationService,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logger.MustFromContext(r.Context())

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

		// Read the CSRF token from the request header
		clientCSRFToken := r.Header.Get(config.AuthenticationCSRFTokenHeaderName)
		if clientCSRFToken == "" {
			logger.WarnContext(r.Context(), "CSRF token is missing")

			utils.RenderError(w, r, utils.ErrUnauthorized)
			return
		}

		// Refresh
		refreshResult, err := authenticationService.Refresh(r.Context(), refreshTokenCookie.Value, clientCSRFToken)
		if err != nil {
			if errors.Is(err, authentication_service.ErrInvalidRefreshTokenClaims) ||
				errors.Is(err, authentication_service.ErrCSRFTokenMismatch) ||
				errors.Is(err, authentication_service.ErrRefreshTokenNotFound) ||
				errors.Is(err, authentication_service.ErrRefreshTokenInvalid) ||
				errors.Is(err, authentication_service.ErrRefreshTokenRevoked) ||
				errors.Is(err, authentication_service.ErrSessionNotFound) ||
				errors.Is(err, authentication_service.ErrSessionAlreadyTerminated) {

				utils.RenderError(w, r, utils.ErrUnauthorized)
				return
			}

			utils.RenderError(w, r, err)
			return
		}

		// Set the refresh token cookie
		account_utils.SetRefreshTokenCookie(config, w, refreshResult.RefreshToken, refreshResult.RefreshTokenExpiresAt)

		// Render the response
		utils.RenderJson(w, r, &RefreshSessionResponse{
			AccessToken: refreshResult.AccessToken,
			CSRFToken:   refreshResult.CSRFToken,
		}, http.StatusOK, nil)
	}
}
