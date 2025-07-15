package sessions

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"prutya/go-api-template/internal/config"
	"prutya/go-api-template/internal/handlers/account/account_utils"
	"prutya/go-api-template/internal/handlers/utils"
	"prutya/go-api-template/internal/helpers"
	"prutya/go-api-template/internal/services/authentication_service"
)

type TerminateResponse struct {
	HasTerminatedCurrentSession bool `json:"hasTerminatedCurrentSession"`
}

func NewSessionsTerminateHandler(
	config *config.Config,
	authenticationService authentication_service.AuthenticationService,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := chi.URLParam(r, "sessionID")

		if err := helpers.ValidateUUIDV7(sessionID); err != nil {
			utils.RenderError(w, r, utils.NewServerError(err.Error(), http.StatusUnprocessableEntity))
			return
		}

		hasTerminatedCurrentSession, err := authenticationService.TerminateUserSession(
			r.Context(),
			utils.GetAccessTokenClaimsFromContext(r.Context()),
			sessionID,
		)
		if err != nil {
			if errors.Is(err, authentication_service.ErrSessionNotFound) ||
				errors.Is(err, authentication_service.ErrSessionAlreadyTerminated) ||
				errors.Is(err, authentication_service.ErrSessionExpired) {
				utils.RenderError(w, r, utils.NewServerError(err.Error(), http.StatusUnprocessableEntity))
				return
			}

			utils.RenderError(w, r, err)
			return
		}

		// If this was the current session, also clear the refresh token cookie
		if hasTerminatedCurrentSession {
			account_utils.SetRefreshTokenCookie(config, w, "", time.Time{})
		}

		// Return the response
		utils.RenderJson(w, r, &TerminateResponse{
			HasTerminatedCurrentSession: hasTerminatedCurrentSession,
		}, http.StatusOK, nil)
	}
}
