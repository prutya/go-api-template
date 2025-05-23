// TODO: Test

package sessions

import (
	"encoding/json"
	"errors"
	"net/http"

	"prutya/go-api-template/internal/config"
	"prutya/go-api-template/internal/handlers/utils"
	"prutya/go-api-template/internal/services/authentication_service"
)

type CreateRequest struct {
	Email    string `json:"email" validate:"required,gte=3,lte=512"`
	Password string `json:"password" validate:"required,gte=1,lte=512"`
}

type Response struct {
	AccessToken string `json:"accessToken"`
}

func NewCreateHandler(
	config *config.Config,
	authenticationService authentication_service.AuthenticationService,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqBody := new(CreateRequest)

		decoder := json.NewDecoder(r.Body)

		if err := decoder.Decode(reqBody); err != nil {
			utils.RenderInvalidJsonError(w, r)
			return
		}

		if err := utils.Validate.Struct(reqBody); err != nil {
			utils.RenderError(w, r, err)
			return
		}

		refreshToken, refreshTokenExpiresAt, accessToken, err := authenticationService.Login(r.Context(), reqBody.Email, reqBody.Password)

		if err != nil {
			if errors.Is(err, authentication_service.ErrInvalidCredentials) {
				utils.RenderError(w, r, utils.NewServerError(err.Error(), http.StatusUnprocessableEntity))
				return
			}

			utils.RenderError(w, r, err)
			return
		}

		// Set the refresh token in a cookie
		http.SetCookie(w, &http.Cookie{
			Name:     config.AuthenticationRefreshTokenCookieName,
			Domain:   config.AuthenticationRefreshTokenCookieDomain,
			Path:     config.AuthenticationRefreshTokenCookiePath,
			Value:    refreshToken,
			Expires:  refreshTokenExpiresAt,
			Secure:   config.AuthenticationRefreshTokenCookieSecure,
			HttpOnly: config.AuthenticationRefreshTokenCookieHttpOnly,
			SameSite: config.AuthenticationRefreshTokenCookieSameSite,
		})

		responseBody := &Response{
			AccessToken: accessToken,
		}

		utils.RenderJson(w, r, responseBody, http.StatusOK, nil)
	}
}
