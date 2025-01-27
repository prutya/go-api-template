package timeout_check

import (
	"net/http"
	"prutya/go-api-template/internal/handlers/utils"
	"time"
)

// TODO: Test
func NewHandler() http.HandlerFunc {
	messageOk := []byte(`{"status":"ok"}`)

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		processTime := 61 * time.Second

		select {
		case <-ctx.Done():
			return

		case <-time.After(processTime):
			// The above channel simulates some hard work.
		}

		utils.RenderRawJson(w, r, messageOk, http.StatusOK, nil)
	}
}
