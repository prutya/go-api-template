// TODO: Test

package sessions

import (
	"net/http"

	"prutya/go-api-template/internal/config"
	"prutya/go-api-template/internal/handlers/utils"
	"prutya/go-api-template/internal/services/authentication_service"
)

func NewDeleteCurrentHandler(
	config *config.Config,
	authenticationService authentication_service.AuthenticationService,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		currentSession := utils.GetSessionFromContext(ctx)

		err := authenticationService.Logout(ctx, currentSession.ID)

		if err != nil {
			utils.RenderError(w, r, err)
			return
		}

		utils.RenderNoContent(w, r, nil)
	}
}
