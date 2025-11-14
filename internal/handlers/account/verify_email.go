package account

import (
	"encoding/json"
	"errors"
	"net/http"

	"prutya/go-api-template/internal/config"
	"prutya/go-api-template/internal/handlers/account/account_utils"
	"prutya/go-api-template/internal/handlers/utils"
	"prutya/go-api-template/internal/services/authentication_service"
)

type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

type VerifyEmailResponse struct {
	AccessToken string `json:"accessToken"`
}

func NewVerifyEmailHandler(
	config *config.Config,
	authenticationService authentication_service.AuthenticationService,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqBody := &VerifyEmailRequest{}

		if err := json.NewDecoder(r.Body).Decode(reqBody); err != nil {
			utils.RenderInvalidJsonError(w, r)
			return
		}

		if err := utils.Validate.Struct(reqBody); err != nil {
			utils.RenderError(w, r, err)
			return
		}

		// Verify email
		loginResult, err := authenticationService.VerifyEmail(
			r.Context(),
			reqBody.Token,
			r.UserAgent(),
			r.RemoteAddr,
		)
		if err != nil {
			if errors.Is(err, authentication_service.ErrInvalidEmailVerificationTokenClaims) ||
				errors.Is(err, authentication_service.ErrEmailVerificationTokenNotFound) ||
				errors.Is(err, authentication_service.ErrEmailVerificationTokenInvalid) {
				utils.RenderError(w, r, utils.ErrUnauthorized)
				return
			}

			if errors.Is(err, authentication_service.ErrEmailAlreadyVerified) {
				utils.RenderError(w, r, utils.NewServerError(err.Error(), http.StatusUnprocessableEntity))
				return
			}

			utils.RenderError(w, r, err)
			return
		}

		// Set the refresh token cookie
		account_utils.SetRefreshTokenCookie(config, w, loginResult.RefreshToken, loginResult.RefreshTokenExpiresAt)

		// Render the response
		utils.RenderJson(w, r, &RefreshSessionResponse{
			AccessToken: loginResult.AccessToken,
		}, http.StatusOK, nil)
	}
}
