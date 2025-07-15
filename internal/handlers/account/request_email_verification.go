package account

import (
	"encoding/json"
	"errors"
	"net/http"

	"prutya/go-api-template/internal/handlers/utils"
	"prutya/go-api-template/internal/services/authentication_service"
)

type RequestEmailVerificationRequest struct {
	Email string `json:"email" validate:"required,gte=3,lte=512,email"`
}

func NewRequestEmailVerificationHandler(
	authenticationService authentication_service.AuthenticationService,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the request body
		reqBody := &RequestEmailVerificationRequest{}
		if err := json.NewDecoder(r.Body).Decode(reqBody); err != nil {
			utils.RenderInvalidJsonError(w, r)
			return
		}

		// Validate the request body
		if err := utils.Validate.Struct(reqBody); err != nil {
			utils.RenderError(w, r, err)
			return
		}

		// Request new verification email
		if err := authenticationService.RequestNewVerificationEmail(r.Context(), reqBody.Email); err != nil {
			// We don't want to leak information about whether the email is already
			// registered, so we always return a 204 No Content response.
			if errors.Is(err, authentication_service.ErrUserNotFound) {
				utils.RenderNoContent(w, r, nil)
				return
			}

			utils.RenderError(w, r, err)
			return
		}

		utils.RenderNoContent(w, r, nil)
	}
}
