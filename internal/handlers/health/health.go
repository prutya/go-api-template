package health

import (
	"net/http"

	"prutya/go-api-template/internal/handlers/utils"
)

func NewHandler() http.HandlerFunc {
	messageOk := []byte(`{"health":"ok"}`)

	return func(w http.ResponseWriter, r *http.Request) {
		utils.RenderRawJson(w, r, messageOk, http.StatusOK, nil)
	}
}
