package account

import (
	"encoding/json"
	"errors"
	"net/http"

	"prutya/go-api-template/internal/handlers/utils"
	"prutya/go-api-template/internal/services/authentication_service"
)

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,gte=3,lte=512,email"`
	Password string `json:"password" validate:"required,gte=8,lte=512,containsUppercase,containsLowercase,containsDigit,containsSpecialCharacter"`
}

func NewRegisterHandler(authenticationService authentication_service.AuthenticationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the request body
		reqBody := &RegisterRequest{}
		if err := json.NewDecoder(r.Body).Decode(reqBody); err != nil {
			utils.RenderInvalidJsonError(w, r)
			return
		}

		// Validate the request body
		if err := utils.Validate.Struct(reqBody); err != nil {
			utils.RenderError(w, r, err)
			return
		}

		// Register
		if err := authenticationService.Register(r.Context(), reqBody.Email, reqBody.Password); err != nil {
			if errors.Is(err, authentication_service.ErrEmailDomainNotAllowed) {
				utils.RenderError(w, r, utils.NewServerError(err.Error(), http.StatusUnprocessableEntity))
				return
			}

			if errors.Is(err, authentication_service.ErrEmailIsTaken) {
				// We don't want to leak information about whether the email is already
				// registered, so when the frontend sees 204 No Content, it should show
				// a message like "If your email address is valid, you will receive a
				// verification code"
				utils.RenderNoContent(w, r, nil)
				return
			}

			utils.RenderError(w, r, err)
			return
		}

		utils.RenderNoContent(w, r, nil)
	}
}
