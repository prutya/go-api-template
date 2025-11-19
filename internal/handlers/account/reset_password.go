package account

import (
	"encoding/json"
	"errors"
	"net/http"
	"prutya/go-api-template/internal/config"
	"prutya/go-api-template/internal/handlers/account/account_utils"
	"prutya/go-api-template/internal/handlers/utils"
	"prutya/go-api-template/internal/logger"
	"prutya/go-api-template/internal/services/authentication_service"
)

type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"newPassword" validate:"required,gte=8,lte=512,containsUppercase,containsLowercase,containsDigit,containsSpecialCharacter"`
}

func NewResetPasswordHandler(
	config *config.Config,
	authenticationService authentication_service.AuthenticationService,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqBody := &ResetPasswordRequest{}

		if err := json.NewDecoder(r.Body).Decode(reqBody); err != nil {
			utils.RenderInvalidJsonError(w, r)
			return
		}

		if err := utils.Validate.Struct(reqBody); err != nil {
			utils.RenderError(w, r, err)
			return
		}

		resetResult, err := authenticationService.ResetPassword(
			r.Context(),
			reqBody.Token,
			reqBody.NewPassword,
			r.UserAgent(),
			r.RemoteAddr,
		)
		if err != nil {
			logger.MustWarnContext(r.Context(), "Password reset failed", "error", err.Error())

			// Prevent user enumeration by handling errors and returning
			// 422 invalid_token
			if errors.Is(err, authentication_service.ErrInvalidPasswordResetToken) {
				displayError := authentication_service.ErrInvalidPasswordResetToken

				utils.RenderError(w, r, utils.NewServerError(displayError.Error(), http.StatusUnprocessableEntity))
				return
			}

			utils.RenderError(w, r, err)
			return
		}

		// Set the refresh token cookie
		account_utils.SetRefreshTokenCookie(config, w, resetResult.RefreshToken, resetResult.RefreshTokenExpiresAt)

		// Render the response
		utils.RenderJson(w, r, &RefreshSessionResponse{
			AccessToken: resetResult.AccessToken,
		}, http.StatusOK, nil)
	}
}
