package users

import (
	"net/http"

	"prutya/go-api-template/internal/handlers/utils"
	"prutya/go-api-template/internal/services/user_service"
)

type CurrentResponse struct {
	ID              string  `json:"id"`
	Email           string  `json:"email"`
	EmailVerifiedAt *string `json:"emailVerifiedAt"`
	CreatedAt       string  `json:"createdAt"`
	UpdatedAt       string  `json:"updatedAt"`
}

func NewCurrentHandler(userService user_service.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user
		user, err := userService.FindByID(r.Context(), utils.GetAccessTokenClaimsFromContext(r.Context()).UserID)
		if err != nil {
			utils.RenderError(w, r, err)
			return
		}

		response := &CurrentResponse{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}

		if user.EmailVerifiedAt.Valid {
			emailVerifiedAtString := user.EmailVerifiedAt.Time.Format("2006-01-02T15:04:05Z")
			response.EmailVerifiedAt = &emailVerifiedAtString
		}

		utils.RenderJson(w, r, response, http.StatusOK, nil)
	}
}
