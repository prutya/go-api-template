package echo

import (
	"encoding/json"
	"fmt"
	"net/http"

	"prutya/go-api-template/internal/handlers/utils"
)

type request struct {
	Name string `json:"name" validate:"required,gte=1,lte=255"`
}

type response struct {
	Message string `json:"message"`
}

func NewHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the request
		reqBody := request{}

		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&reqBody); err != nil {
			utils.RenderInvalidJsonError(w, r)
			return
		}

		// Validate the request
		if err := utils.Validate.Struct(reqBody); err != nil {
			utils.RenderError(w, r, err)
			return
		}

		responseBody := &response{Message: fmt.Sprintf("Hello, %s!", reqBody.Name)}

		utils.RenderJson(w, r, responseBody, http.StatusOK, nil)
	}
}
