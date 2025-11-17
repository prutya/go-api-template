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
	Email string `json:"email" validate:"required,gte=3,lte=512,email"`
	OTP   string `json:"otp" validate:"required,len=6,numeric"`
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
			reqBody.Email,
			reqBody.OTP,
			r.UserAgent(),
			r.RemoteAddr,
		)
		if err != nil {
			// Prevent user enumeration by handling errors and returning
			// 422 invalid_otp
			if errors.Is(err, authentication_service.ErrUserNotFound) ||
				errors.Is(err, authentication_service.ErrTooManyOTPAttempts) ||
				errors.Is(err, authentication_service.ErrEmailVerificationExpired) ||
				errors.Is(err, authentication_service.ErrInvalidOTP) {

				displayError := authentication_service.ErrInvalidOTP

				utils.RenderError(w, r, utils.NewServerError(displayError.Error(), http.StatusUnprocessableEntity))
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
