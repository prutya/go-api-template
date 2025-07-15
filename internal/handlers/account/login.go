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

type LoginRequest struct {
	Email    string `json:"email" validate:"required,gte=3,lte=512"`
	Password string `json:"password" validate:"required,gte=1,lte=512"`
}

type LoginResponse struct {
	AccessToken string `json:"accessToken"`
	CSRFToken   string `json:"csrfToken"`
}

func NewLoginHandler(
	config *config.Config,
	authenticationService authentication_service.AuthenticationService,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the request body
		reqBody := &LoginRequest{}
		if err := json.NewDecoder(r.Body).Decode(reqBody); err != nil {
			utils.RenderInvalidJsonError(w, r)
			return
		}

		// Validate the request body
		if err := utils.Validate.Struct(reqBody); err != nil {
			utils.RenderError(w, r, err)
			return
		}

		// Login
		loginResult, err := authenticationService.Login(
			r.Context(),
			reqBody.Email,
			reqBody.Password,
			r.UserAgent(),
			r.RemoteAddr,
		)
		if err != nil {
			if errors.Is(err, authentication_service.ErrInvalidCredentials) {
				utils.RenderError(w, r, utils.NewServerError(err.Error(), http.StatusUnprocessableEntity))
				return
			}

			utils.RenderError(w, r, err)
			return
		}

		// Set the refresh token cookie
		account_utils.SetRefreshTokenCookie(config, w, loginResult.RefreshToken, loginResult.RefreshTokenExpiresAt)

		// Render the response
		utils.RenderJson(w, r, &LoginResponse{
			AccessToken: loginResult.AccessToken,
			CSRFToken:   loginResult.CSRFToken,
		}, http.StatusOK, nil)
	}
}
