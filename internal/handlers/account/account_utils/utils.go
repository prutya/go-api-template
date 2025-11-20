package account_utils

import (
	"net/http"
	"time"

	"prutya/go-api-template/internal/config"
)

func SetRefreshTokenCookie(config *config.Config, w http.ResponseWriter, value string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     config.AuthenticationRefreshTokenCookieName,
		Domain:   config.AuthenticationRefreshTokenCookieDomain,
		Path:     config.AuthenticationRefreshTokenCookiePath,
		Value:    value,
		Expires:  expiresAt,
		Secure:   config.AuthenticationRefreshTokenCookieSecure,
		HttpOnly: config.AuthenticationRefreshTokenCookieHttpOnly,
		SameSite: http.SameSiteStrictMode,
	})
}
