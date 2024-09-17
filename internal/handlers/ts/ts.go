package ts

import (
	"net/http"
	"time"

	"prutya/go-api-template/internal/handlers/utils"
)

func NewHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentTime := time.Now()

		utils.RenderJson(w, r, map[string]int64{"ts": currentTime.Unix()}, http.StatusOK, nil)
	}
}
