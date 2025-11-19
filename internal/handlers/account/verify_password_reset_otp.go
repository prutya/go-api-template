package account

import (
	"encoding/json"
	"errors"
	"net/http"
	"prutya/go-api-template/internal/config"
	"prutya/go-api-template/internal/handlers/utils"
	"prutya/go-api-template/internal/logger"
	"prutya/go-api-template/internal/services/authentication_service"
)

type VerifyPasswordResetOTPRequest struct {
	Email string `json:"email" validate:"required,gte=3,lte=512,email"`
	OTP   string `json:"otp" validate:"required,len=6,numeric"`
}

type VerifyPasswordResetOTPResponse struct {
	PasswordResetToken string `json:"passwordResetToken"`
}

func NewVerifyPasswordResetOTPHandler(
	config *config.Config,
	authenticationService authentication_service.AuthenticationService,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqBody := &VerifyPasswordResetOTPRequest{}

		if err := json.NewDecoder(r.Body).Decode(reqBody); err != nil {
			utils.RenderInvalidJsonError(w, r)
			return
		}

		if err := utils.Validate.Struct(reqBody); err != nil {
			utils.RenderError(w, r, err)
			return
		}

		passwordResetToken, err := authenticationService.VerifyPasswordResetOTP(
			r.Context(),
			reqBody.Email,
			reqBody.OTP,
		)
		if err != nil {
			logger.MustWarnContext(r.Context(), "Password reset OTP verification failed", "error", err.Error())

			// Prevent user enumeration by handling errors and returning
			// 422 invalid_otp
			if errors.Is(err, authentication_service.ErrUserNotFound) ||
				errors.Is(err, authentication_service.ErrTooManyOTPAttempts) ||
				errors.Is(err, authentication_service.ErrPasswordResetNotRequested) ||
				errors.Is(err, authentication_service.ErrPasswordResetExpired) ||
				errors.Is(err, authentication_service.ErrInvalidOTP) {

				displayError := authentication_service.ErrInvalidOTP

				utils.RenderError(w, r, utils.NewServerError(displayError.Error(), http.StatusUnprocessableEntity))
				return
			}

			utils.RenderError(w, r, err)
			return
		}

		utils.RenderJson(w, r, &VerifyPasswordResetOTPResponse{
			PasswordResetToken: passwordResetToken,
		}, http.StatusOK, nil)
	}
}
