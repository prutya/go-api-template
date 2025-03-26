// TODO: Test

package sessions

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"prutya/go-api-template/internal/config"
	"prutya/go-api-template/internal/handlers/utils"
	"prutya/go-api-template/internal/services/authentication_service"
)

type CreateRequest struct {
	Email    string `json:"email" validate:"required,gte=3,lte=512"`
	Password string `json:"password" validate:"required,gte=1,lte=512"`
}

type Response struct {
	SessionToken string `json:"sessionToken"`
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

		// Prevent the potential attacker from measuring the response time
		loginStartTime := time.Now()

		sessionToken, err := authenticationService.Login(r.Context(), reqBody.Email, reqBody.Password)

		loginDuration := time.Since(loginStartTime)
		loginTimeLeft := config.TimingAttackDelay - loginDuration

		if loginTimeLeft > 0 {
			time.Sleep(loginTimeLeft)
		}

		if err != nil {
			if errors.Is(err, authentication_service.ErrInvalidCredentials) {
				utils.RenderError(w, r, utils.NewServerError(err.Error(), http.StatusUnprocessableEntity))
				return
			}

			utils.RenderError(w, r, err)
			return
		}

		responseBody := &Response{SessionToken: sessionToken}

		utils.RenderJson(w, r, responseBody, http.StatusOK, nil)
	}
}
