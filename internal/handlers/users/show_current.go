// TODO: Test

package users

import (
	"net/http"

	"prutya/go-api-template/internal/config"
	"prutya/go-api-template/internal/handlers/utils"
	"prutya/go-api-template/internal/services/user_service"
)

type ShowResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func NewShowCurrentHandler(config *config.Config, userService user_service.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		currentAccessTokenClaims := utils.GetAccessTokenClaimsFromContext(ctx)

		currentUser, err := userService.GetUserById(ctx, currentAccessTokenClaims.UserID)
		if err != nil {
			utils.RenderError(w, r, err)
			return
		}

		utils.RenderJson(w, r, &ShowResponse{
			ID:    currentUser.ID,
			Email: currentUser.Email,
		}, http.StatusOK, nil)
	}
}
