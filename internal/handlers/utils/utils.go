package utils

import (
	"encoding/json"
	"net/http"
	"strconv"
)

var EmptyHeaders = map[string]string{}

func RenderError(w http.ResponseWriter, r *http.Request, err error) {
	RenderJson(w, r, map[string]any{"error": err}, 500, EmptyHeaders)
}

func RenderJson(w http.ResponseWriter, r *http.Request, object any, httpStatusCode int, headers map[string]string) {
	json, marshalErr := json.Marshal(object)

	if marshalErr != nil {
		msg := []byte(ErrCodeInternal)

		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Length", strconv.Itoa(len(msg)))

		w.WriteHeader(http.StatusInternalServerError)

		w.Write(msg)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(json)))

	for k, v := range headers {
		w.Header().Set(k, v)
	}

	w.WriteHeader(httpStatusCode)

	w.Write(json)
}
