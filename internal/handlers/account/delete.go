package account

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"prutya/go-api-template/internal/config"
	"prutya/go-api-template/internal/handlers/account/account_utils"
	"prutya/go-api-template/internal/handlers/utils"
	"prutya/go-api-template/internal/services/authentication_service"
)

type DeleteAccountRequest struct {
	CurrentPassword string `json:"currentPassword" validate:"required,gte=1,lte=512"`
}

func NewDeleteAccountHandler(config *config.Config, authenticationService authentication_service.AuthenticationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentAccessTokenClaims := utils.GetAccessTokenClaimsFromContext(r.Context())

		reqBody := &DeleteAccountRequest{}

		if err := json.NewDecoder(r.Body).Decode(reqBody); err != nil {
			utils.RenderInvalidJsonError(w, r)
			return
		}

		if err := utils.Validate.Struct(reqBody); err != nil {
			utils.RenderError(w, r, err)
			return
		}

		err := authenticationService.DeleteAccount(
			r.Context(),
			currentAccessTokenClaims,
			reqBody.CurrentPassword,
		)
		if err != nil {
			if errors.Is(err, authentication_service.ErrInvalidCredentials) {
				utils.RenderError(w, r, utils.NewServerError(err.Error(), http.StatusUnprocessableEntity))
				return
			}

			utils.RenderError(w, r, err)
			return
		}

		// Remove the refresh token cookie
		account_utils.SetRefreshTokenCookie(config, w, "", time.Time{})

		utils.RenderNoContent(w, r, nil)
	}
}
