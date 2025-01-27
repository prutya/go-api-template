package panic_check

import (
	"net/http"
)

func NewHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		panic("Panic check")
	}
}
