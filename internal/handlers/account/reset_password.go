package account

import (
	"encoding/json"
	"errors"
	"net/http"
	"prutya/go-api-template/internal/handlers/utils"
	"prutya/go-api-template/internal/services/authentication_service"
)

type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"newPassword" validate:"required,gte=8,lte=512,containsUppercase,containsLowercase,containsDigit,containsSpecialCharacter"`
}

func NewResetPasswordHandler(
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

		if err := authenticationService.ResetPassword(r.Context(), reqBody.Token, reqBody.NewPassword); err != nil {
			if errors.Is(err, authentication_service.ErrInvalidResetPasswordTokenClaims) ||
				errors.Is(err, authentication_service.ErrResetPasswordTokenNotFound) ||
				errors.Is(err, authentication_service.ErrResetPasswordTokenInvalid) {
				utils.RenderError(w, r, utils.ErrUnauthorized)
				return
			}

			utils.RenderError(w, r, err)
			return
		}

		utils.RenderNoContent(w, r, nil)
	}
}
