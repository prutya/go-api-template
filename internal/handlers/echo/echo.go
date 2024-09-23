package echo

import (
	"encoding/json"
	"net/http"

	"prutya/go-api-template/internal/handlers/utils"
)

type Request struct {
	Message string `json:"message" validate:"required,gte=2,lte=16"`
}

type Response struct {
	Message string `json:"message"`
}

func NewHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the request
		reqBody := Request{}

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

		responseBody := &Response{Message: reqBody.Message}

		utils.RenderJson(w, r, responseBody, http.StatusOK, nil)
	}
}
