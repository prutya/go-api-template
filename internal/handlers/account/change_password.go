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

type ChangePasswordRequest struct {
	CurrentPassword        string `json:"currentPassword" validate:"required,gte=1,lte=512"`
	NewPassword            string `json:"newPassword" validate:"required,gte=8,lte=512,containsUppercase,containsLowercase,containsDigit,containsSpecialCharacter"`
	TerminateOtherSessions bool   `json:"terminateOtherSessions" validate:""`
}

func NewChangePasswordHandler(config *config.Config, authenticationService authentication_service.AuthenticationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentAccessTokenClaims := utils.GetAccessTokenClaimsFromContext(r.Context())

		reqBody := &ChangePasswordRequest{}

		if err := json.NewDecoder(r.Body).Decode(reqBody); err != nil {
			utils.RenderInvalidJsonError(w, r)
			return
		}

		if err := utils.Validate.Struct(reqBody); err != nil {
			utils.RenderError(w, r, err)
			return
		}

		if err := authenticationService.ChangePassword(
			r.Context(),
			currentAccessTokenClaims,
			reqBody.CurrentPassword,
			reqBody.NewPassword,
			reqBody.TerminateOtherSessions,
		); err != nil {
			logger.MustWarnContext(r.Context(), "Password change failed", "error", err.Error())

			if errors.Is(err, authentication_service.ErrInvalidCredentials) {
				utils.RenderError(w, r, utils.NewServerError(err.Error(), http.StatusUnprocessableEntity))
				return
			}

			utils.RenderError(w, r, err)
			return
		}

		utils.RenderNoContent(w, r, nil)
	}
}
