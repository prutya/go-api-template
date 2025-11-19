package account

import (
	"encoding/json"
	"errors"
	"net/http"
	"prutya/go-api-template/internal/handlers/utils"
	"prutya/go-api-template/internal/logger"
	"prutya/go-api-template/internal/services/authentication_service"
)

type RequestPasswordResetRequest struct {
	Email string `json:"email" validate:"required,gte=3,lte=512,email"`
}

func NewRequestPasswordResetHandler(
	authenticationService authentication_service.AuthenticationService,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the request body
		reqBody := &RequestPasswordResetRequest{}
		if err := json.NewDecoder(r.Body).Decode(reqBody); err != nil {
			utils.RenderInvalidJsonError(w, r)
			return
		}

		// Validate the request body
		if err := utils.Validate.Struct(reqBody); err != nil {
			utils.RenderError(w, r, err)
			return
		}

		// Request password reset
		if err := authenticationService.RequestPasswordReset(r.Context(), reqBody.Email); err != nil {
			logger.MustWarnContext(r.Context(), "Password reset request failed", "error", err.Error())

			// We don't want to leak information about whether the email is already
			// registered, so we always return a 204 No Content response.
			if errors.Is(err, authentication_service.ErrUserRecordLocked) ||
				errors.Is(err, authentication_service.ErrUserNotFound) ||
				errors.Is(err, authentication_service.ErrPasswordResetCooldown) {

				utils.RenderNoContent(w, r, nil)
				return
			}

			utils.RenderError(w, r, err)
			return
		}

		utils.RenderNoContent(w, r, nil)
	}
}
