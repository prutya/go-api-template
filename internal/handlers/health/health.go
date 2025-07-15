package health

import (
	"net/http"

	"prutya/go-api-template/internal/handlers/utils"
)

var MessageOK = []byte(`{"health":"ok"}`)

func NewHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		utils.RenderRawJson(w, r, MessageOK, http.StatusOK, nil)
	}
}
