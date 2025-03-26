// TODO: Test

package users

import (
	"net/http"

	"prutya/go-api-template/internal/config"
	"prutya/go-api-template/internal/handlers/utils"
)

type ShowResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func NewShowCurrentHandler(config *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		currentUser := utils.GetUserFromContext(ctx)

		utils.RenderJson(w, r, &ShowResponse{
			ID:    currentUser.ID,
			Email: currentUser.Email,
		}, http.StatusOK, nil)
	}
}
